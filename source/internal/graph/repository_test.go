package graph_test

import (
	"strings"
	"testing"

	"github.com/fza/sddviz/internal/graph"
)

const fixtureEntryCount = 9

func TestOpenRepositoryReportsWhatTheGraphHolds(t *testing.T) {
	service := graph.NewService("", "")

	opened, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}
	if opened.EntryCount != fixtureEntryCount {
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
	service := graph.NewService("", "")

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
	service := graph.NewService("", "")

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
	service := graph.NewService("", "")

	_, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}

	entries, err := service.Entries(t.Context())
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	if len(entries) != fixtureEntryCount {
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
		if entry.Title == "" {
			t.Errorf("%s carries no title, though its file declares one", entry.ID)
		}
		if entry.Status.Kind == "" {
			t.Errorf("%s carries no derived status, which is what the view call exists to supply", entry.ID)
		}
	}

	if entries[0].ID < entries[1].ID {
		t.Error("the entries are not ordered newest first")
	}
}

func TestEntriesCarryTheLinksTheirFilesDeclare(t *testing.T) {
	service := graph.NewService("", "")

	_, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}

	entries, err := service.Entries(t.Context())
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}

	byID := map[string]graph.Entry{}
	for _, entry := range entries {
		byID[entry.ID] = entry
	}

	t.Run("a declared reference keeps its target, its kind and its description", func(t *testing.T) {
		refs := byID["20260829-223328-d-tac-3qx"].Refs
		expected := []graph.Ref{
			{
				ID:   "20260829-221103-d-tac-tz1",
				Kind: "grounded-in",
				Desc: "the view call this record's status arrives on",
			},
			{
				ID:   "20260829-221000-d-cpt-d8y",
				Kind: "refines",
				Desc: "narrows the embedding's reading surface to the record shape the front end receives",
			},
		}
		if len(refs) != len(expected) {
			t.Fatalf("the entry carries %d references, expected the ones its file declares", len(refs))
		}
		for index, want := range expected {
			if refs[index] != want {
				t.Errorf("reference %d reads %+v, expected the one the file declares in that position", index, refs[index])
			}
		}
	})

	t.Run("a closing entry carries what it closes", func(t *testing.T) {
		closes := byID["20260829-235605-d-prc-zbm"].Closes
		if len(closes) != 1 || closes[0] != "20260829-221424-s-prc-8nm" {
			t.Errorf("the entry closes %v, expected the signal its file names", closes)
		}
	})

	t.Run("a superseding entry carries what it retires", func(t *testing.T) {
		supersedes := byID["20260829-205243-d-tac-7o5"].Supersedes
		if len(supersedes) != 1 || supersedes[0] != "20260829-202759-d-tac-8mk" {
			t.Errorf("the entry supersedes %v, expected the decision its file names", supersedes)
		}
	})

	t.Run("an entry declaring no link carries empty lists", func(t *testing.T) {
		entry := byID["20260829-200722-s-prc-9o1"]
		if entry.Refs == nil || entry.Closes == nil || entry.Supersedes == nil {
			t.Error("an entry declaring no link reports a missing field rather than an empty list")
		}
		if len(entry.Refs) != 0 || len(entry.Closes) != 0 || len(entry.Supersedes) != 0 {
			t.Errorf("an entry declaring no link reports links anyway: %+v", entry)
		}
	})
}

func TestEntriesRefusesWithoutAnOpenRepository(t *testing.T) {
	service := graph.NewService("", "")

	_, err := service.Entries(t.Context())
	if err == nil {
		t.Fatal("Entries answered though no repository is open")
	}
}

func TestStartupOpensThePathTheCommandLineNamed(t *testing.T) {
	service := graph.NewService("testdata/repository", "")

	state := service.Startup(t.Context())
	if !state.Requested {
		t.Fatal("a path was given, so the startup state should report one was requested")
	}
	if state.Failure != "" {
		t.Fatalf("opening the named repository failed: %s", state.Failure)
	}
	if state.Repository.EntryCount != fixtureEntryCount {
		t.Errorf("the opened repository reports %d entries", state.Repository.EntryCount)
	}

	entries, err := service.Entries(t.Context())
	if err != nil {
		t.Fatalf("Entries after startup: %v", err)
	}
	if len(entries) != fixtureEntryCount {
		t.Errorf("the service reported %d entries after startup", len(entries))
	}
}

func TestStartupReportsAPathThatWillNotOpen(t *testing.T) {
	service := graph.NewService("testdata/notarepo", "")

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
	service := graph.NewService("", "")

	state := service.Startup(t.Context())
	if state.Requested {
		t.Error("no path was given, yet the startup state reports one was requested")
	}
	if state.Failure != "" {
		t.Errorf("no path was given, yet startup failed: %s", state.Failure)
	}
}

func TestOpenRepositoryNamesTheDirectoryTheGraphLivesIn(t *testing.T) {
	service := graph.NewService("", "")

	cases := []struct {
		name     string
		path     string
		expected string
	}{
		{name: "a plain path", path: "testdata/repository", expected: "repository"},
		{name: "a path with a trailing separator", path: "testdata/repository/", expected: "repository"},
		{name: "a path naming the directory another way", path: "testdata/./repository", expected: "repository"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			opened, err := service.OpenRepository(t.Context(), testCase.path)
			if err != nil {
				t.Fatalf("OpenRepository: %v", err)
			}
			if opened.Name != testCase.expected {
				t.Errorf("the repository reports the name %q, expected the directory its graph lives in", opened.Name)
			}
		})
	}
}

func TestSearchReportsWhatSddMatched(t *testing.T) {
	service := graph.NewService("", "")

	_, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}

	outcome, err := service.Search(t.Context(), "frontmatter")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(outcome.Entries) == 0 {
		t.Fatal("the search found nothing, though the fixture's entries name frontmatter")
	}
	if outcome.Query != "frontmatter" {
		t.Errorf("the outcome reports the phrase %q", outcome.Query)
	}
	for _, entry := range outcome.Entries {
		if entry.ID == "" || entry.Summary == "" {
			t.Errorf("a matched entry arrived without its fields: %+v", entry)
		}
	}
}

func TestSearchReportsTheModeItRan(t *testing.T) {
	service := graph.NewService("", "")

	_, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}

	outcome, err := service.Search(t.Context(), "frontmatter")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// The mount carries no embeddings provider and no search index, so sdd
	// offers terms alone.
	if outcome.Mode != "text" {
		t.Errorf("the search ran in mode %q, expected the one a mount without embeddings offers", outcome.Mode)
	}
}

func TestSearchOnAnEmptyPhraseMatchesNothing(t *testing.T) {
	service := graph.NewService("", "")

	_, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}

	outcome, err := service.Search(t.Context(), "   ")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(outcome.Entries) != 0 {
		t.Errorf("an empty phrase matched %d entries", len(outcome.Entries))
	}
}

func TestSearchRefusesWithoutAnOpenRepository(t *testing.T) {
	service := graph.NewService("", "")

	_, err := service.Search(t.Context(), "frontmatter")
	if err == nil {
		t.Fatal("Search answered though no repository is open")
	}
}

func TestSearchScoresRankMatchesAgainstTheStrongest(t *testing.T) {
	service := graph.NewService("", "")

	_, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}

	outcome, err := service.Search(t.Context(), "frontmatter")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(outcome.Entries) == 0 {
		t.Fatal("the search found nothing to score")
	}

	if outcome.Entries[0].Score != 100 {
		t.Errorf("the strongest match scored %d, expected the hundred sdd normalises it to", outcome.Entries[0].Score)
	}
	for _, entry := range outcome.Entries {
		if entry.Score <= 0 || entry.Score > 100 {
			t.Errorf("%s scored %d, which is outside the range a percentage takes", entry.ID, entry.Score)
		}
	}
}

func TestCloseRepositoryLetsGoOfWhatWasOpen(t *testing.T) {
	service := graph.NewService("", "")

	opened, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}
	if opened.EntryCount == 0 {
		t.Fatal("the repository opened with no entries to let go of")
	}

	service.CloseRepository()

	if service.OpenedRepository() != (graph.Repository{}) {
		t.Error("the service still reports a repository after letting go of it")
	}
	_, err = service.Entries(t.Context())
	if err == nil {
		t.Error("Entries answered after the repository was closed")
	}
	_, err = service.Search(t.Context(), "frontmatter")
	if err == nil {
		t.Error("Search answered after the repository was closed")
	}
}

func TestCloseRepositoryOnNothingOpenIsNoFailure(t *testing.T) {
	service := graph.NewService("", "")

	service.CloseRepository()

	if service.OpenedRepository() != (graph.Repository{}) {
		t.Error("closing nothing left a repository behind")
	}
}

func TestOpeningAgainAfterClosingWorks(t *testing.T) {
	service := graph.NewService("", "")

	_, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository: %v", err)
	}
	service.CloseRepository()

	reopened, err := service.OpenRepository(t.Context(), "testdata/repository")
	if err != nil {
		t.Fatalf("OpenRepository after closing: %v", err)
	}
	if reopened.EntryCount != fixtureEntryCount {
		t.Errorf("the reopened repository reports %d entries", reopened.EntryCount)
	}
	entries, err := service.Entries(t.Context())
	if err != nil {
		t.Fatalf("Entries after reopening: %v", err)
	}
	if len(entries) != fixtureEntryCount {
		t.Errorf("the reopened repository listed %d entries", len(entries))
	}
}
