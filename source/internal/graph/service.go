// Package graph answers the front end's questions about a decision graph.
package graph

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const version = "0.1.0"

// ErrNoWindowSystem is what the picker returns where no window system answers,
// which is every server build. The path is asked for another way there, and
// OpenRepository is the same call in both.
var ErrNoWindowSystem = errors.New("choosing a directory needs a window system")

// Service is the surface the front end calls.
type Service struct {
	mu    sync.RWMutex
	store *Store
	open  Repository
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
