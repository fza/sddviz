package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	sdd "github.com/networkteam/sdd/application"
	"gopkg.in/yaml.v3"
)

// DefaultGraphDir is where sdd keeps a graph when a repository's configuration
// names no other directory.
const DefaultGraphDir = ".sdd/graph"

// configPath is the file that makes a checkout a repository sddviz can open.
const configPath = ".sdd/config.yaml"

var errNotARepository = errors.New("no .sdd/config.yaml, so this is not a repository sddviz can open")

// Repository is what the front end learns about an opened repository. It
// carries counts rather than entries: the entries follow through their own
// call, once there is a view to draw them in.
type Repository struct {
	Name            string `json:"name"`
	Path            string `json:"path"`
	GraphDir        string `json:"graphDir"`
	RepoID          string `json:"repoId"`
	Revision        string `json:"revision"`
	EntryCount      int    `json:"entryCount"`
	UnreadableCount int    `json:"unreadableCount"`
}

type repositoryConfig struct {
	GraphDir string `yaml:"graph_dir"`
	RepoID   string `yaml:"repo_id"`
}

func readRepositoryConfig(repositoryPath string) (repositoryConfig, error) {
	raw, err := os.ReadFile(filepath.Join(repositoryPath, configPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repositoryConfig{}, fmt.Errorf("%s: %w", repositoryPath, errNotARepository)
		}
		return repositoryConfig{}, fmt.Errorf("reading %s of %s: %w", configPath, repositoryPath, err)
	}

	var config repositoryConfig
	err = yaml.Unmarshal(raw, &config)
	if err != nil {
		return repositoryConfig{}, fmt.Errorf("reading %s of %s: %w", configPath, repositoryPath, err)
	}
	if config.GraphDir == "" {
		config.GraphDir = DefaultGraphDir
	}
	return config, nil
}

// repositoryName is what a reader calls the open project: the name of the
// directory the graph lives in. A path is the front end's to display and never
// to take apart, so the split happens here.
func repositoryName(repositoryPath string) string {
	name := filepath.Base(filepath.Clean(repositoryPath))
	if name == "." || name == string(filepath.Separator) {
		return repositoryPath
	}
	return name
}

// projectID identifies the opened graph to sdd. A repository declaring a
// `repo_id` carries the identity every clone of it shares; one that declares none
// is local, and its own path is the only identity available.
func projectID(config repositoryConfig, repositoryPath string) sdd.ProjectID {
	if config.RepoID != "" {
		return sdd.ProjectID(config.RepoID)
	}
	return sdd.ProjectID(filepath.Clean(repositoryPath))
}
