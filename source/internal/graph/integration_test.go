package graph_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sdd "github.com/networkteam/sdd/pkg/application"
	"github.com/networkteam/sdd/pkg/llm"
	local "github.com/networkteam/sdd/pkg/local"
)

// access is the smallest resolver sdd's application needs: one project, one
// caller, read access.
type access struct {
	runtime *sdd.ProjectRuntime
}

func (a access) ResolvePrincipal(_ context.Context, identity sdd.RequestIdentity) (sdd.Principal, error) {
	return sdd.Principal{Subject: identity.Subject}, nil
}

func (a access) ResolveParticipant(_ context.Context, principal sdd.Principal, _ sdd.ProjectID) (string, error) {
	return principal.Subject, nil
}

func (a access) ListProjects(context.Context, sdd.Principal) (sdd.ProjectList, error) {
	return sdd.ProjectList{Projects: []sdd.ProjectSummary{{
		ProjectRef: a.runtime.Project(), CanRead: true, State: sdd.ProjectReady,
	}}}, nil
}

func (a access) ResolveProject(_ context.Context, _ sdd.Principal, project sdd.ProjectID, _ sdd.Access) (*sdd.ProjectRuntime, error) {
	if project != a.runtime.Project().ID {
		return nil, &sdd.ApplicationError{Code: sdd.ErrorProjectUnavailable, Message: "unknown project"}
	}
	return a.runtime, nil
}

func (a access) ResolveDependency(context.Context, sdd.Principal, sdd.ProjectID, string) (*sdd.ProjectRuntime, error) {
	return nil, &sdd.ApplicationError{Code: sdd.ErrorProjectUnavailable, Message: "no dependencies"}
}

func (a access) AuthorizeSession(context.Context, sdd.SessionAccessRequest) error {
	return &sdd.ApplicationError{Code: sdd.ErrorWriteDenied, Message: "no sessions"}
}

// newApplication mounts sdd over the store under test. Reads need no model
// provider, so the executor here refuses rather than pretending to answer.
func newApplication(t *testing.T) *sdd.Application {
	t.Helper()

	store := openFixture(t)
	sessions, err := local.NewFilesystemSessionStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("session store: %v", err)
	}
	blobs, err := local.NewFilesystemStagedBlobStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("staged blob store: %v", err)
	}

	runtime, err := sdd.NewProjectRuntime(sdd.ProjectRuntimeOptions{
		Project: sdd.ProjectRef{ID: "fixture", DisplayName: "Fixture"},
		Graph:   store,
		LLM: llm.RunnerFunc(func(context.Context, llm.Request) (llm.Result, error) {
			return llm.Result{}, fmt.Errorf("a read needs no model provider")
		}),
	})
	if err != nil {
		t.Fatalf("project runtime: %v", err)
	}

	application, err := sdd.NewApplication(sdd.ApplicationOptions{
		Access:      access{runtime: runtime},
		Sessions:    sessions,
		StagedBlobs: blobs,
	})
	if err != nil {
		t.Fatalf("application: %v", err)
	}
	return application
}

func TestSddRendersTheStoresEntries(t *testing.T) {
	application := newApplication(t)
	identity := sdd.RequestIdentity{Subject: "Felix"}

	result, err := application.View(t.Context(), identity, "fixture", sdd.ViewRequest{
		Layout: "source(graph):as-list",
	})
	if err != nil {
		t.Fatalf("View: %v", err)
	}
	if !strings.Contains(result.Sections, "s-prc-abi") {
		t.Error("the rendered view names none of the fixture's entries")
	}
	if !strings.Contains(result.Sections, "{status:") {
		t.Error("the rendered view carries no status, which is what the entry line exists to report")
	}
}

func TestSddSearchesTheStoresEntries(t *testing.T) {
	application := newApplication(t)
	identity := sdd.RequestIdentity{Subject: "Felix"}

	result, err := application.Search(t.Context(), identity, "fixture", sdd.SearchRequest{
		Terms:    []string{"sddviz"},
		Limit:    10,
		SyncMode: sdd.SearchSyncLocal,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(result.EntryIDs) == 0 {
		t.Fatal("the search found nothing, though the fixture's entries name sddviz")
	}
}
