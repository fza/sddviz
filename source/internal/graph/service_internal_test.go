package graph

import (
	"context"
	"errors"
	"testing"

	sdd "github.com/networkteam/sdd/pkg/application"
	"github.com/networkteam/sdd/pkg/llm"
)

const fixtureProject sdd.ProjectID = "fixture"

func fixtureRuntime(t *testing.T) *sdd.ProjectRuntime {
	t.Helper()
	store, err := Open(fixtureProject, "testdata/graph")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	runtime, err := sdd.NewProjectRuntime(sdd.ProjectRuntimeOptions{
		Project: sdd.ProjectRef{ID: fixtureProject},
		Graph:   store,
		LLM: llm.RunnerFunc(func(context.Context, llm.Request) (llm.Result, error) {
			return llm.Result{}, errReadOnlyNeedsNoModel
		}),
	})
	if err != nil {
		t.Fatalf("NewProjectRuntime: %v", err)
	}
	return runtime
}

func TestReadAccessRefusesEveryWrite(t *testing.T) {
	access := readAccess{runtime: fixtureRuntime(t)}
	principal := sdd.Principal{Subject: viewer}

	_, err := access.ResolveProject(t.Context(), principal, fixtureProject, sdd.AccessRead)
	if err != nil {
		t.Fatalf("a read of the open repository was refused: %v", err)
	}

	cases := []struct {
		name  string
		write func() error
	}{
		{
			name: "write access to the open repository",
			write: func() error {
				_, err := access.ResolveProject(t.Context(), principal, fixtureProject, sdd.AccessWrite)
				return err
			},
		},
		{
			name: "continuing a session",
			write: func() error {
				return access.AuthorizeSession(t.Context(), sdd.SessionAccessRequest{Actor: principal, Owner: principal, Project: fixtureProject})
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var refusal *sdd.ApplicationError
			err := testCase.write()
			if !errors.As(err, &refusal) || refusal.Code != sdd.ErrorWriteDenied {
				t.Errorf("the resolver answered %v, expected a refused write", err)
			}
		})
	}
}

func TestTheStoreImplementsNoWritePort(t *testing.T) {
	var store any = &Store{}

	if _, writes := store.(sdd.PublicationStore); writes {
		t.Error("the store can publish an entry, so a write has a path into a reader's repository")
	}
	if _, acquires := store.(sdd.TargetAcquirer); acquires {
		t.Error("the store hands out mutation targets, so a write has a path into a reader's repository")
	}
	if _, reads := store.(sdd.GraphStore); !reads {
		t.Error("the store no longer satisfies the read port sdd mounts")
	}
}
