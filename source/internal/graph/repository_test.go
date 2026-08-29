package graph_test

import (
	"strings"
	"testing"

	"github.com/fza/sddviz/internal/graph"
)

func TestOpenRepositoryReportsWhatTheGraphHolds(t *testing.T) {
	service := graph.NewService("")

	opened, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}
	if opened.EntryCount != 2 {
		t.Errorf("the repository reports %d entries, expected the ones its graph carries", opened.EntryCount)
	}
	if opened.UnreadableCount != 0 {
		t.Errorf("the repository reports %d undecodable documents, expected none", opened.UnreadableCount)
	}
	if opened.RepoID != "example.com/fixture/repository" {
		t.Errorf("the repository reports id %q, expected the one its configuration declares", opened.RepoID)
	}
	if opened.GraphDir != ".sdd/graph" {
		t.Errorf("the repository reports graph directory %q", opened.GraphDir)
	}
	if opened.Revision == "" {
		t.Error("the repository reports no revision")
	}

	if service.OpenedRepository() != opened {
		t.Error("the service does not report the repository it just opened")
	}
}

func TestOpenRepositoryDefaultsTheGraphDirectory(t *testing.T) {
	service := graph.NewService("")

	opened, err := service.OpenRepository(t.Context(), "testdata/defaulted")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}
	if opened.GraphDir != graph.DefaultGraphDir {
		t.Errorf("a configuration naming no graph directory produced %q, expected the default", opened.GraphDir)
	}
	if opened.EntryCount != 1 {
		t.Errorf("the repository reports %d entries", opened.EntryCount)
	}
}

func TestOpenRepositoryRefusesWhatIsNotARepository(t *testing.T) {
	service := graph.NewService("")

	cases := []struct {
		name string
		path string
	}{
		{name: "a directory carrying no configuration", path: "testdata/notarepo"},
		{name: "a path that does not exist", path: "testdata/absent"},
		{name: "no path at all", path: ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := service.OpenRepository(t.Context(), testCase.path)
			if err == nil {
				t.Fatal("OpenRepository accepted something that is not a repository")
			}
			if !strings.Contains(err.Error(), "repository") {
				t.Errorf("the refusal reads %q, which does not tell a reader what was wrong", err)
			}
		})
	}
}

func TestEntriesCarryFrontmatterAndDerivedStatus(t *testing.T) {
	service := graph.NewService("")

	_, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}

	entries, err := service.Entries(t.Context())
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("the repository reported %d entries, expected the ones its graph carries", len(entries))
	}

	for _, entry := range entries {
		if entry.ID == "" {
			t.Error("an entry carries no id")
		}
		if entry.Type == "" || entry.Layer == "" {
			t.Errorf("%s carries no type or layer, so its frontmatter was not read", entry.ID)
		}
		if entry.Summary == "" {
			t.Errorf("%s carries no summary", entry.ID)
		}
		if entry.Status.Kind == "" {
			t.Errorf("%s carries no derived status, which is what the view call exists to supply", entry.ID)
		}
	}

	if entries[0].ID < entries[1].ID {
		t.Error("the entries are not ordered newest first")
	}
}

func TestEntriesRefusesWithoutAnOpenRepository(t *testing.T) {
	service := graph.NewService("")

	_, err := service.Entries(t.Context())
	if err == nil {
		t.Fatal("Entries answered though no repository is open")
	}
}

func TestStartupOpensThePathTheCommandLineNamed(t *testing.T) {
	service := graph.NewService("testdata/repository")

	state := service.Startup(t.Context())
	if !state.Requested {
		t.Fatal("a path was given, so the startup state should report one was requested")
	}
	if state.Failure != "" {
		t.Fatalf("opening the named repository failed: %s", state.Failure)
	}
	if state.Repository.EntryCount != 2 {
		t.Errorf("the opened repository reports %d entries", state.Repository.EntryCount)
	}

	entries, err := service.Entries(t.Context())
	if err != nil {
		t.Fatalf("Entries after startup: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("the service reported %d entries after startup", len(entries))
	}
}

func TestStartupReportsAPathThatWillNotOpen(t *testing.T) {
	service := graph.NewService("testdata/notarepo")

	state := service.Startup(t.Context())
	if !state.Requested {
		t.Fatal("a path was given, so the startup state should report one was requested")
	}
	if state.Failure == "" {
		t.Fatal("a path that is not a repository opened without complaint")
	}
	if state.Repository.Path != "" {
		t.Error("a failed startup reports a repository anyway")
	}
}

func TestStartupWithNoPathRequestsNothing(t *testing.T) {
	service := graph.NewService("")

	state := service.Startup(t.Context())
	if state.Requested {
		t.Error("no path was given, yet the startup state reports one was requested")
	}
	if state.Failure != "" {
		t.Errorf("no path was given, yet startup failed: %s", state.Failure)
	}
}
