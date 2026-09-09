// Package graph answers the front end's questions about a decision graph.
package graph

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

// StartupState is what the front end learns when it starts: the repository a
// path on the command line opened, or why that path did not open. Both are
// empty where no path was given.
type StartupState struct {
	Repository Repository `json:"repository"`
	Failure    string     `json:"failure"`
	Requested  bool       `json:"requested"`
}

// Service is the surface the front end calls.
type Service struct {
	startupPath string

	mu          sync.RWMutex
	store       *Store
	application *sdd.Application
	project     sdd.ProjectID
	open        Repository
	scratch     string
	startup     StartupState
	started     bool
}

// NewService takes the repository path the command line named, which is empty
// where it named none.
func NewService(startupPath string) *Service {
	return &Service{startupPath: startupPath}
}

// Startup opens the repository the command line named and reports what
// happened. Calling it again returns the same answer rather than reopening.
func (s *Service) Startup(ctx context.Context) StartupState {
	s.mu.RLock()
	started, state := s.started, s.startup
	s.mu.RUnlock()
	if started {
		return state
	}

	state = StartupState{Requested: s.startupPath != ""}
	if state.Requested {
		opened, err := s.OpenRepository(ctx, s.startupPath)
		if err != nil {
			state.Failure = err.Error()
		} else {
			state.Repository = opened
		}
	}

	s.mu.Lock()
	s.startup = state
	s.started = true
	s.mu.Unlock()

	return state
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

	mounted, scratch, err := mountApplication(store, projectID(config, path), config, path)
	if err != nil {
		return Repository{}, err
	}

	opened := Repository{
		Name:            repositoryName(path),
		Path:            path,
		GraphDir:        config.GraphDir,
		RepoID:          config.RepoID,
		Revision:        snapshot.Revision(),
		EntryCount:      len(store.Documents()),
		UnreadableCount: len(store.Unreadable()),
	}

	s.mu.Lock()
	previous := s.scratch
	s.store = store
	s.application = mounted
	s.project = projectID(config, path)
	s.open = opened
	s.scratch = scratch
	s.mu.Unlock()

	discardScratch(previous)

	return opened, nil
}

// CloseRepository lets go of the repository the viewer holds, so a reader can
// leave one open graph without opening another. It reports nothing: closing
// what is already closed is what a reader asked for either way.
func (s *Service) CloseRepository() {
	s.mu.Lock()
	scratch := s.scratch
	s.store = nil
	s.application = nil
	s.project = ""
	s.open = Repository{}
	s.scratch = ""
	s.mu.Unlock()

	discardScratch(scratch)
}

// discardScratch removes the session and staged-blob directories one mount
// owned. They sit in the machine's temporary directory rather than in any
// repository, so removing them reaches nothing a reader owns.
func discardScratch(scratch string) {
	if scratch == "" {
		return
	}
	err := os.RemoveAll(scratch)
	if err != nil {
		slog.Warn("removing the session scratch directory", "path", scratch, "error", err)
	}
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
//
// A repository configuring an embedding provider the viewer speaks also gets
// sdd's vector search, over the stored index sdd's own tool fills. That index
// sits in the machine's cache directory rather than in the repository, so
// reading and reconciling it leaves the repository untouched.
func mountApplication(store *Store, project sdd.ProjectID, config repositoryConfig, repositoryPath string) (*sdd.Application, string, error) {
	scratch, err := os.MkdirTemp("", "sddviz-session-")
	if err != nil {
		return nil, "", fmt.Errorf("preparing the session scratch directory: %w", err)
	}

	sessions, err := local.NewFilesystemSessionStore(filepath.Join(scratch, "sessions"))
	if err != nil {
		return nil, "", fmt.Errorf("preparing the session store: %w", err)
	}
	blobs, err := local.NewFilesystemStagedBlobStore(filepath.Join(scratch, "blobs"))
	if err != nil {
		return nil, "", fmt.Errorf("preparing the staged blob store: %w", err)
	}

	options := sdd.ProjectRuntimeOptions{
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
	}

	embeddings, index, err := vectorSearchOptions(config, repositoryPath, project)
	if err != nil {
		return nil, "", err
	}
	// sdd refuses a runtime carrying one without the other, and a repository
	// configuring no provider searches over terms alone.
	if embeddings != nil {
		options.Embeddings = embeddings
		options.SearchIndex = index
	}

	runtime, err := sdd.NewProjectRuntime(options)
	if err != nil {
		return nil, "", fmt.Errorf("mounting sdd over %s: %w", project, err)
	}

	mounted, err := sdd.NewApplication(readAccess{runtime: runtime})
	if err != nil {
		return nil, "", fmt.Errorf("mounting the sdd application: %w", err)
	}
	return mounted, scratch, nil
}

// vectorSearchOptions reports what sdd needs for vector search, or nothing at
// all where the repository configures no provider the viewer speaks. A
// configuration the viewer cannot use is not a failure: the search runs over
// terms instead.
func vectorSearchOptions(config repositoryConfig, repositoryPath string, project sdd.ProjectID) (sdd.EmbeddingExecutor, sdd.SearchIndexStore, error) {
	executor, err := newEmbedder(config.Embedding)
	if err != nil {
		if errors.Is(err, errEmbeddingUnconfigured) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	cacheRoot, err := indexCacheRoot()
	if err != nil {
		return nil, nil, err
	}

	absolute, err := filepath.Abs(repositoryPath)
	if err != nil {
		return nil, nil, fmt.Errorf("resolving the path of %s: %w", repositoryPath, err)
	}

	index := local.NewPersistentSearchIndexStore(project, cacheRoot, indexRepoKey(config.RepoID, absolute))
	return executor, index, nil
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
