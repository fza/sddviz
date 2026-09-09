package graph

import (
	"context"
	"fmt"
	"strings"

	sdd "github.com/networkteam/sdd/application"
)

// searchLimit caps what one search reports. sdd answers with ten where a
// request names no limit, which fills a table too thinly to read a graph by.
const searchLimit = 200

// vectorMode is the word sdd's Info uses for the mode that reads an embedding
// index. A mount carrying no embeddings provider and no index reports `text`
// alone, and every search then runs over terms.
const vectorMode = "vector"

// textMode is the word sdd's Info uses for matching terms literally.
const textMode = "text"

// SearchOutcome is what one search reports: the entries sdd matched, in the
// order it ranked them, and the mode that produced them.
type SearchOutcome struct {
	Query   string  `json:"query"`
	Mode    string  `json:"mode"`
	Entries []Entry `json:"entries"`
}

// Search reports the entries sdd matches for what a reader typed. A phrase
// reaches sdd's own searching, never a comparison written here.
func (s *Service) Search(ctx context.Context, phrase string) (SearchOutcome, error) {
	s.mu.RLock()
	store, application, project := s.store, s.application, s.project
	s.mu.RUnlock()

	if store == nil {
		return SearchOutcome{}, errNoRepositoryOpen
	}

	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return SearchOutcome{Query: phrase, Mode: "", Entries: []Entry{}}, nil
	}

	available, err := availableSearchMode(ctx, application, project)
	if err != nil {
		return SearchOutcome{}, err
	}

	words := strings.Fields(phrase)
	mode := chooseSearchMode(available, words)

	// One citation per entry is the strongest part of it that matched, and the
	// score of that citation is what stands for the entry. Asking for none
	// renders no citation line at all, and with it no score.
	request := sdd.SearchRequest{Limit: searchLimit, MaxCitations: 1, IncludeSuperseded: true}
	if mode == vectorMode {
		request.Phrase = phrase
	} else {
		request.Terms = words
	}

	result, err := application.Search(ctx, sdd.RequestIdentity{Subject: viewer}, project, request)
	if err != nil {
		return SearchOutcome{}, fmt.Errorf("searching for %q: %w", phrase, err)
	}

	statuses, err := s.deriveStatuses(ctx, application, project)
	if err != nil {
		return SearchOutcome{}, err
	}

	scores := parseSearchScores(result.Results)
	matched := entriesByID(store.Documents())
	entries := make([]Entry, 0, len(result.EntryIDs))
	for _, id := range result.EntryIDs {
		document, found := matched[id]
		if !found {
			// sdd searches the snapshot it was handed, so an identifier it
			// reports that no document carries names a dependency graph rather
			// than this one.
			continue
		}
		entry := entryFromFrontmatter(id, document)
		entry.Status = statuses[id]
		entry.Score = scores[id]
		entries = append(entries, entry)
	}

	return SearchOutcome{Query: phrase, Mode: mode, Entries: entries}, nil
}

// chooseSearchMode decides how one phrase is answered. A single word is what a
// reader types to find a known token, and matching it literally beats the
// nearest neighbours of its meaning. Several words describe a thing rather than
// name it, which is what the vector space answers well.
func chooseSearchMode(available string, words []string) string {
	if available == vectorMode && len(words) > 1 {
		return vectorMode
	}
	return textMode
}

func availableSearchMode(ctx context.Context, application *sdd.Application, project sdd.ProjectID) (string, error) {
	info, err := application.Info(ctx, sdd.RequestIdentity{Subject: viewer}, project, sdd.InfoRequest{})
	if err != nil {
		return "", fmt.Errorf("reading which search modes answer: %w", err)
	}
	for _, mode := range strings.Split(info.Search, ",") {
		if strings.TrimSpace(mode) == vectorMode {
			return vectorMode, nil
		}
	}
	return textMode, nil
}

func entriesByID(documents []sdd.EntryDocument) map[string]sdd.EntryDocument {
	byID := make(map[string]sdd.EntryDocument, len(documents))
	for _, document := range documents {
		id, err := entryID(document.LogicalPath)
		if err != nil {
			continue
		}
		byID[id] = document
	}
	return byID
}
