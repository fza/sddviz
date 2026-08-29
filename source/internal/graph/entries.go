package graph

import (
	"context"
	"fmt"
	"sort"

	sdd "github.com/networkteam/sdd/application"
)

// statusLayout renders every entry in the graph on one line each, which is
// where derived status comes from. `source(graph)` is the unfiltered
// population, so closed and superseded entries appear beside live ones.
const statusLayout = "source(graph):as-list"

// Entry is one entry as the views draw it: what its frontmatter holds, plus the
// status sdd derives.
type Entry struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Layer        string   `json:"layer"`
	Kind         string   `json:"kind"`
	Confidence   string   `json:"confidence"`
	Intent       string   `json:"intent,omitempty"`
	Participants []string `json:"participants"`
	Topics       []string `json:"topics"`
	Summary      string   `json:"summary"`
	Body         string   `json:"body"`
	Status       Status   `json:"status"`
}

// Entries reports every entry of the open repository, newest first. The order
// is the one an entry ID already carries, since it opens with the moment the
// entry was written.
func (s *Service) Entries(ctx context.Context) ([]Entry, error) {
	s.mu.RLock()
	store, application, project := s.store, s.application, s.project
	s.mu.RUnlock()

	if store == nil {
		return nil, errNoRepositoryOpen
	}

	statuses, err := s.deriveStatuses(ctx, application, project)
	if err != nil {
		return nil, err
	}

	documents := store.Documents()
	entries := make([]Entry, 0, len(documents))
	for _, document := range documents {
		id, err := entryID(document.LogicalPath)
		if err != nil {
			continue
		}
		entry := entryFromFrontmatter(id, document)
		entry.Status = statuses[id]
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(first, second int) bool {
		return entries[first].ID > entries[second].ID
	})
	return entries, nil
}

func (s *Service) deriveStatuses(ctx context.Context, application *sdd.Application, project sdd.ProjectID) (map[string]Status, error) {
	result, err := application.View(ctx, sdd.RequestIdentity{Subject: viewer}, project, sdd.ViewRequest{
		Layout: statusLayout,
	})
	if err != nil {
		return nil, fmt.Errorf("reading the status of every entry: %w", err)
	}
	return parseEntryLines(result.Sections), nil
}

func entryFromFrontmatter(id string, document sdd.EntryDocument) Entry {
	return Entry{
		ID:           id,
		Type:         stringField(document.Frontmatter, "type"),
		Layer:        stringField(document.Frontmatter, "layer"),
		Kind:         stringField(document.Frontmatter, "kind"),
		Confidence:   stringField(document.Frontmatter, "confidence"),
		Intent:       stringField(document.Frontmatter, "intent"),
		Participants: stringsField(document.Frontmatter, "participants"),
		Topics:       stringsField(document.Frontmatter, "topics"),
		Summary:      stringField(document.Frontmatter, "summary"),
		Body:         document.Body,
	}
}

func stringField(frontmatter map[string]any, name string) string {
	value, found := frontmatter[name].(string)
	if !found {
		return ""
	}
	return value
}

func stringsField(frontmatter map[string]any, name string) []string {
	values, found := frontmatter[name].([]any)
	if !found {
		return []string{}
	}
	strings := make([]string, 0, len(values))
	for _, value := range values {
		text, isText := value.(string)
		if !isText {
			continue
		}
		strings = append(strings, text)
	}
	return strings
}
