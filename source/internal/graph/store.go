// Package graph reads an SDD decision graph and answers the front end's
// questions about it.
package graph

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"

	sdd "github.com/networkteam/sdd/application"
	"gopkg.in/yaml.v3"
)

// ErrReadOnly is what every write method of the store returns. sddviz never
// writes to a repository it reads, and refusing here means no call path exists
// that could, whatever an embedded caller attempts.
var ErrReadOnly = errors.New("sddviz reads a graph and never writes to it")

var (
	errNotADirectory           = errors.New("not a directory")
	errAttachmentNamesAPath    = errors.New("attachment names a path rather than a file")
	errNoFrontmatter           = errors.New("the document carries no YAML frontmatter")
	errUnterminatedFrontmatter = errors.New("the document's YAML frontmatter is unterminated")
)

// entryPath matches the layout an entry file takes inside the graph directory,
// `2026/08/29-205109-s-prc-xv4.md`. sdd validates the same shape while loading,
// and rejects a document whose path it cannot turn into an entry ID, so the
// walk below has to recognise one before offering it. Measured against sdd
// 0.16.2.
var entryPath = regexp.MustCompile(`^\d{4}/\d{2}/\d{2}-\d{6}-[a-z]-[a-z]{3}-[a-z0-9]+\.md$`)

// DocumentIssue names one document the walk could not decode: its path inside
// the graph directory and why it failed. sdd 0.16.2 carries no slot for these
// in its snapshot, so the store holds them and the reader is told which file
// failed rather than left with a shorter list.
type DocumentIssue struct {
	LogicalPath string
	Message     string
}

// Store reads a graph directory and satisfies sdd's GraphStore port. It keeps
// the documents it read, which the snapshot handed to sdd does not expose, so
// one walk answers both the tool and the views.
type Store struct {
	project sdd.ProjectID
	dir     string

	mu         sync.RWMutex
	data       sdd.SnapshotData
	unreadable []DocumentIssue
	snapshot   *sdd.Snapshot
}

func Open(project sdd.ProjectID, graphDir string) (*Store, error) {
	info, err := os.Stat(graphDir)
	if err != nil {
		return nil, fmt.Errorf("opening graph directory %s: %w", graphDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("graph directory %s: %w", graphDir, errNotADirectory)
	}
	return &Store{project: project, dir: graphDir}, nil
}

func (s *Store) Current(ctx context.Context) (*sdd.Snapshot, error) {
	data, unreadable, err := s.read()
	if err != nil {
		return nil, err
	}
	snapshot, err := sdd.BuildSnapshot(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("building the snapshot of %s: %w", s.dir, err)
	}

	s.mu.Lock()
	s.data = data
	s.unreadable = unreadable
	s.snapshot = snapshot
	s.mu.Unlock()

	return snapshot, nil
}

// Documents returns the entry documents behind the last snapshot, which is what
// the views draw. sdd's Snapshot carries the same documents and exposes none of
// them.
func (s *Store) Documents() []sdd.EntryDocument {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Entries
}

// Unreadable names the documents the walk could not decode. One malformed file
// leaves the rest of a graph readable, and the reader is told which one failed
// rather than left with a shorter list.
func (s *Store) Unreadable() []DocumentIssue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.unreadable
}

func (s *Store) Apply(context.Context, string, sdd.MutationBatch, sdd.StagedBlobReader) (sdd.ApplyResult, error) {
	return sdd.ApplyResult{State: sdd.MutationNotApplied}, ErrReadOnly
}

func (s *Store) Reconcile(context.Context, string, string) (sdd.ApplyResult, error) {
	return sdd.ApplyResult{State: sdd.MutationNotApplied}, ErrReadOnly
}

func (s *Store) ReadAttachmentPage(_ context.Context, entryID, filename string, offset int64, maxBytes int) (sdd.AttachmentPage, error) {
	relDir, err := sdd.AttachmentDirRelPath(entryID)
	if err != nil {
		return sdd.AttachmentPage{}, fmt.Errorf("resolving the attachment directory of %s: %w", entryID, err)
	}
	if strings.Contains(filename, "/") || filename == "." || filename == ".." {
		return sdd.AttachmentPage{}, fmt.Errorf("attachment %q: %w", filename, errAttachmentNamesAPath)
	}

	content, err := os.ReadFile(path.Join(s.dir, relDir, filename))
	if err != nil {
		return sdd.AttachmentPage{}, fmt.Errorf("reading attachment %s of %s: %w", filename, entryID, err)
	}
	if offset > int64(len(content)) {
		offset = int64(len(content))
	}
	page := content[offset:]
	if maxBytes > 0 && len(page) > maxBytes {
		page = page[:maxBytes]
	}
	next := offset + int64(len(page))
	return sdd.AttachmentPage{
		Filename:   filename,
		Content:    page,
		Offset:     offset,
		NextOffset: next,
		TotalSize:  int64(len(content)),
		More:       next < int64(len(content)),
	}, nil
}

// read walks the graph directory once and produces the canonical documents sdd
// builds a snapshot from. It mirrors what sdd's own filesystem loader collects:
// entry documents, work-in-progress markers, and the documents neither could be
// decoded.
func (s *Store) read() (sdd.SnapshotData, []DocumentIssue, error) {
	fsys := os.DirFS(s.dir)
	data := sdd.SnapshotData{Project: s.project}
	var unreadable []DocumentIssue
	digest := sha256.New()

	err := fs.WalkDir(fsys, ".", func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if filename != "." && strings.HasPrefix(entry.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}

		raw, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return err
		}
		digest.Write([]byte(filename))
		digest.Write(raw)

		if strings.HasPrefix(filename, "wip/") {
			data.WIP = append(data.WIP, sdd.WIPDocument{LogicalPath: filename, Content: string(raw)})
			return nil
		}
		if !entryPath.MatchString(filename) {
			return nil
		}

		document, err := parseEntryDocument(filename, raw)
		if err != nil {
			// One document that will not decode leaves the rest of a graph
			// readable, so the walk records it and carries on rather than
			// failing the whole read.
			unreadable = append(unreadable, DocumentIssue{LogicalPath: filename, Message: err.Error()})
			return nil //nolint:nilerr // the failure is reported, not swallowed
		}
		data.Entries = append(data.Entries, document)
		return nil
	})
	if err != nil {
		return sdd.SnapshotData{}, nil, fmt.Errorf("walking the graph documents in %s: %w", s.dir, err)
	}

	err = s.collectAttachments(fsys, data.Entries)
	if err != nil {
		return sdd.SnapshotData{}, nil, err
	}

	// The revision identifies the content rather than a commit, because a
	// reader's working tree carries entries no commit holds yet.
	data.Revision = hex.EncodeToString(digest.Sum(nil))
	return data, unreadable, nil
}

func (s *Store) collectAttachments(fsys fs.FS, entries []sdd.EntryDocument) error {
	for index := range entries {
		document := &entries[index]
		attachmentDir := strings.TrimSuffix(document.LogicalPath, ".md")

		attachments, err := fs.ReadDir(fsys, attachmentDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return fmt.Errorf("reading the attachments of %s: %w", document.LogicalPath, err)
		}
		for _, attachment := range attachments {
			if attachment.IsDir() {
				continue
			}
			document.Attachments = append(document.Attachments, attachmentDir+"/"+attachment.Name())
		}
		sort.Strings(document.Attachments)
	}
	return nil
}

// parseEntryDocument splits an entry into its frontmatter and its body. The
// frontmatter stays a map rather than a struct: sdd owns what the fields mean,
// and it validates them while building the snapshot.
func parseEntryDocument(logicalPath string, raw []byte) (sdd.EntryDocument, error) {
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") {
		return sdd.EntryDocument{}, errNoFrontmatter
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return sdd.EntryDocument{}, errUnterminatedFrontmatter
	}
	end += 4

	var frontmatter map[string]any
	err := yaml.Unmarshal([]byte(text[4:end]), &frontmatter)
	if err != nil {
		return sdd.EntryDocument{}, err
	}

	body := strings.TrimPrefix(text[end+4:], "\n")
	return sdd.EntryDocument{
		LogicalPath: logicalPath,
		Frontmatter: frontmatter,
		Body:        strings.TrimSpace(body),
	}, nil
}
