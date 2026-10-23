package graph

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	sdd "github.com/networkteam/sdd/pkg/application"
	"gopkg.in/yaml.v3"
)

// DefaultGraphDir is where sdd keeps a graph when a repository's configuration
// names no other directory.
const DefaultGraphDir = ".sdd/graph"

// configPath is the file that makes a checkout a repository sddviz can open.
const configPath = ".sdd/config.yaml"

// localConfigPath carries what one machine adds to a repository's shared
// configuration, an embedding endpoint and its key above all. It overlays the
// shared file field by field and is absent from most checkouts.
const localConfigPath = ".sdd/config.local.yaml"

// userConfigFile is sdd's machine-wide configuration, relative to the
// configuration directory. sdd reads it beneath both repository files, and it
// is where a machine keeps the embedding settings every repository shares.
// Measured against sdd v0.18.0-alpha.2.fza.24.
const userConfigFile = "sdd/config.yaml"

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
	GraphDir  string          `yaml:"graph_dir"`
	RepoID    string          `yaml:"repo_id"`
	Embedding embeddingConfig `yaml:"embedding"`
}

// UserConfigPath is where sdd keeps its machine-wide configuration:
// `$XDG_CONFIG_HOME/sdd/config.yaml`, or `~/.config/sdd/config.yaml` where the
// variable is unset. It is empty where no home directory resolves, and a
// repository then reads its own files alone.
func UserConfigPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, userConfigFile)
}

// readRepositoryConfig resolves a repository's configuration in the order sdd
// reads it: the machine-wide file first, then the shared file, then the local
// one, each overlaying the one beneath it field by field. The shared file
// alone never sets an embedding endpoint or key.
func readRepositoryConfig(repositoryPath, userConfigPath string) (repositoryConfig, error) {
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

	user, err := readOptionalConfig(userConfigPath)
	if err != nil {
		return repositoryConfig{}, err
	}
	local, err := readOptionalConfig(filepath.Join(repositoryPath, localConfigPath))
	if err != nil {
		return repositoryConfig{}, err
	}
	// The shared file arrives with whatever a repository's committers wrote, so
	// it may not choose where a request goes or which key it carries. A key
	// from either machine-owned file would otherwise travel to an endpoint a
	// stranger named the moment a reader opened their repository.
	shared := config.Embedding
	shared.Endpoint = ""
	shared.APIKeys = nil
	config.Embedding = mergeEmbeddingConfig(mergeEmbeddingConfig(user.Embedding, shared), local.Embedding)
	if local.GraphDir != "" {
		config.GraphDir = local.GraphDir
	}
	if local.RepoID != "" {
		config.RepoID = local.RepoID
	}
	return config, nil
}

// readOptionalConfig reads a configuration file that may be absent, which is
// no failure: a machine without one, or a checkout without a local file, reads
// what the other layers say.
func readOptionalConfig(path string) (repositoryConfig, error) {
	if path == "" {
		return repositoryConfig{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repositoryConfig{}, nil
		}
		return repositoryConfig{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var config repositoryConfig
	err = yaml.Unmarshal(raw, &config)
	if err != nil {
		return repositoryConfig{}, fmt.Errorf("reading %s: %w", path, err)
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
