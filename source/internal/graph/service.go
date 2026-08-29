// Package graph answers the front end's questions about a decision graph.
package graph

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	sdd "github.com/networkteam/sdd/application"
	local "github.com/networkteam/sdd/local"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const version = "0.1.0"

// viewer is the participant sdd attributes a read to. sddviz only ever reads,
// so the name never reaches a graph.
const viewer = "sddviz"

var errNoRepositoryOpen = errors.New("no repository is open")

// ErrNoWindowSystem is what the picker returns where no window system answers,
// which is every server build. The path is asked for another way there, and
// OpenRepository is the same call in both.
var ErrNoWindowSystem = errors.New("choosing a directory needs a window system")

// Service is the surface the front end calls.
type Service struct {
	mu          sync.RWMutex
	store       *Store
	application *sdd.Application
	project     sdd.ProjectID
	open        Repository
}

func NewService() *Service {
	return &Service{}
}

// BackendVersion reports the running back end, which is how the front end
// confirms the seam between the two halves works.
func (s *Service) BackendVersion() string {
	return version
}

// ChooseRepositoryDirectory asks the reader for a directory and returns it
// without opening anything. Desktop builds alone answer it.
func (s *Service) ChooseRepositoryDirectory() (string, error) {
	app := application.Get()
	if app == nil {
		return "", ErrNoWindowSystem
	}

	path, err := app.Dialog.OpenFile().
		SetTitle("Open a repository").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil {
		return "", fmt.Errorf("choosing a repository directory: %w", err)
	}
	return path, nil
}

// OpenRepository opens the graph of the repository at the given path and
// reports what it holds.
func (s *Service) OpenRepository(ctx context.Context, path string) (Repository, error) {
	if path == "" {
		return Repository{}, errNotARepository
	}

	config, err := readRepositoryConfig(path)
	if err != nil {
		return Repository{}, err
	}

	store, err := Open(projectID(config, path), filepath.Join(path, config.GraphDir))
	if err != nil {
		return Repository{}, err
	}

	snapshot, err := store.Current(ctx)
	if err != nil {
		return Repository{}, err
	}

	mounted, err := mountApplication(store, projectID(config, path))
	if err != nil {
		return Repository{}, err
	}

	opened := Repository{
		Path:            path,
		GraphDir:        config.GraphDir,
		RepoID:          config.RepoID,
		Revision:        snapshot.Revision(),
		EntryCount:      len(store.Documents()),
		UnreadableCount: len(store.Unreadable()),
	}

	s.mu.Lock()
	s.store = store
	s.application = mounted
	s.project = projectID(config, path)
	s.open = opened
	s.mu.Unlock()

	return opened, nil
}

// OpenedRepository reports the repository currently open, if any.
func (s *Service) OpenedRepository() Repository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.open
}

// mountApplication runs sdd over the store, which is how a read reaches the
// searching and the deriving sdd owns. Its session and staged-blob stores sit
// in scratch the viewer owns: sdd requires both, and a read writes to neither.
func mountApplication(store *Store, project sdd.ProjectID) (*sdd.Application, error) {
	scratch, err := os.MkdirTemp("", "sddviz-session-")
	if err != nil {
		return nil, fmt.Errorf("preparing the session scratch directory: %w", err)
	}

	sessions, err := local.NewFilesystemSessionStore(filepath.Join(scratch, "sessions"))
	if err != nil {
		return nil, fmt.Errorf("preparing the session store: %w", err)
	}
	blobs, err := local.NewFilesystemStagedBlobStore(filepath.Join(scratch, "blobs"))
	if err != nil {
		return nil, fmt.Errorf("preparing the staged blob store: %w", err)
	}

	runtime, err := sdd.NewProjectRuntime(sdd.ProjectRuntimeOptions{
		Project:     sdd.ProjectRef{ID: project, DisplayName: string(project)},
		Graph:       store,
		Sessions:    sessions,
		StagedBlobs: blobs,
		LLM: sdd.LLMExecutorFuncs{
			CapabilitiesFunc: func(context.Context) ([]string, error) { return nil, nil },
			ExecuteFunc: func(context.Context, sdd.LLMRequest) (sdd.LLMResult, error) {
				return sdd.LLMResult{}, errReadOnlyNeedsNoModel
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("mounting sdd over %s: %w", project, err)
	}

	mounted, err := sdd.NewApplication(readAccess{runtime: runtime})
	if err != nil {
		return nil, fmt.Errorf("mounting the sdd application: %w", err)
	}
	return mounted, nil
}

var errReadOnlyNeedsNoModel = errors.New("a read needs no model provider")

// readAccess grants the one caller read of the one open repository. sddviz
// serves the person at the machine, so there is no authorization to resolve.
type readAccess struct {
	runtime *sdd.ProjectRuntime
}

func (a readAccess) ResolvePrincipal(_ context.Context, identity sdd.RequestIdentity) (sdd.Principal, error) {
	return sdd.Principal{Subject: identity.Subject, Participant: identity.Subject}, nil
}

func (a readAccess) ListProjects(context.Context, sdd.Principal) (sdd.ProjectList, error) {
	return sdd.ProjectList{Projects: []sdd.ProjectSummary{{
		ProjectRef: a.runtime.Project(), CanRead: true, State: sdd.ProjectReady,
	}}}, nil
}

func (a readAccess) ResolveProject(_ context.Context, _ sdd.Principal, project sdd.ProjectID, _ sdd.Access) (*sdd.ProjectRuntime, error) {
	if project != a.runtime.Project().ID {
		return nil, &sdd.ApplicationError{Code: sdd.ErrorProjectUnavailable, Message: "unknown repository"}
	}
	return a.runtime, nil
}

func (a readAccess) ResolveDependency(context.Context, sdd.Principal, sdd.ProjectID, string) (*sdd.ProjectRuntime, error) {
	return nil, &sdd.ApplicationError{Code: sdd.ErrorProjectUnavailable, Message: "connected repositories are not read"}
}
