package graph_test

import (
	"errors"
	"testing"

	sdd "github.com/networkteam/sdd/application"

	"github.com/fza/sddviz/internal/graph"
)

const fixtureGraph = "testdata/graph"

func openFixture(t *testing.T) *graph.Store {
	t.Helper()
	store, err := graph.Open("fixture", fixtureGraph)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return store
}

func TestCurrentReadsTheFixtureGraph(t *testing.T) {
	store := openFixture(t)

	snapshot, err := store.Current(t.Context())
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if snapshot.Project() != "fixture" {
		t.Errorf("the snapshot names project %q, expected the one the store was opened with", snapshot.Project())
	}
	if snapshot.Revision() == "" {
		t.Error("the snapshot carries no revision, so a caller cannot tell one read from another")
	}

	documents := store.Documents()
	if len(documents) != 3 {
		t.Fatalf("the store kept %d documents, expected the readable entries of the fixture", len(documents))
	}
	for _, document := range documents {
		if document.Frontmatter["type"] == nil {
			t.Errorf("%s carries no type in its frontmatter, so it was not decoded", document.LogicalPath)
		}
		if document.Body == "" {
			t.Errorf("%s carries an empty body", document.LogicalPath)
		}
	}
}

func TestRevisionFollowsContent(t *testing.T) {
	store := openFixture(t)

	first, err := store.Current(t.Context())
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	second, err := store.Current(t.Context())
	if err != nil {
		t.Fatalf("Current, second read: %v", err)
	}
	if first.Revision() != second.Revision() {
		t.Error("two reads of an unchanged graph report different revisions")
	}
}

func TestAnUndecodableDocumentIsReportedAndTheRestLoad(t *testing.T) {
	store := openFixture(t)

	_, err := store.Current(t.Context())
	if err != nil {
		t.Fatalf("Current: %v", err)
	}

	unreadable := store.Unreadable()
	if len(unreadable) != 1 {
		t.Fatalf("the store reported %d undecodable documents, expected the one the fixture carries", len(unreadable))
	}
	if unreadable[0].LogicalPath != "2026/08/29-000000-s-tac-bad.md" {
		t.Errorf("the report names %q rather than the document that failed", unreadable[0].LogicalPath)
	}
	if unreadable[0].Message == "" {
		t.Error("the report carries no reason, so a reader cannot act on it")
	}
}

func TestEveryWriteRefuses(t *testing.T) {
	store := openFixture(t)

	cases := []struct {
		name  string
		write func() (sdd.ApplyResult, error)
	}{
		{
			name: "Apply",
			write: func() (sdd.ApplyResult, error) {
				return store.Apply(t.Context(), "any-revision", sdd.MutationBatch{ID: "batch"}, nil)
			},
		},
		{
			name: "Reconcile",
			write: func() (sdd.ApplyResult, error) {
				return store.Reconcile(t.Context(), "batch", "digest")
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := testCase.write()
			if !errors.Is(err, graph.ErrReadOnly) {
				t.Errorf("the write returned %v, expected a refusal", err)
			}
			if result.State != sdd.MutationNotApplied {
				t.Errorf("the write reported state %q, expected nothing to have been applied", result.State)
			}
		})
	}
}

func TestOpenRefusesSomethingThatIsNotADirectory(t *testing.T) {
	_, err := graph.Open("fixture", "testdata/graph/2026/08/29-200314-s-prc-abi.md")
	if err == nil {
		t.Fatal("Open accepted a file where a graph directory belongs")
	}
}
