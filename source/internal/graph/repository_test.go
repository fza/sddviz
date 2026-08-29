package graph_test

import (
	"strings"
	"testing"

	"github.com/fza/sddviz/internal/graph"
)

func TestOpenRepositoryReportsWhatTheGraphHolds(t *testing.T) {
	service := graph.NewService()

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
	service := graph.NewService()

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
	service := graph.NewService()

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
