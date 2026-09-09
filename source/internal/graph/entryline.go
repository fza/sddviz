package graph

import (
	"regexp"
	"strconv"
)

// Status is an entry's derived lifecycle state. sdd computes it from the whole
// graph and stores it nowhere, so it arrives on the line its view renders.
type Status struct {
	// Kind is one of sdd's status values: open, active, settled, closed-by,
	// superseded-by, cascade-closed, cascade-orphan. A done signal carries
	// none, and its Kind is empty.
	Kind string `json:"kind"`
	// By names the live entry that closed or superseded this one, resolved
	// forward through the chain. Empty for every other status.
	By string `json:"by,omitempty"`
}

// entryLine matches the line sdd renders per entry:
//
//	<id> <layer> <kind> <type> [confidence: x] [intent: y] (participants) {status: s} <topics> summary
//
// The shape is documented and held by sdd's own unit tests, which is what makes
// it firm enough to read. Measured against sdd 0.16.2.
var entryLine = regexp.MustCompile(
	`^\s*(?P<id>\d{8}-\d{6}-[a-z]-[a-z]{3}-[a-z0-9]+)\s` +
		`(?P<rest>.*)$`)

// statusField matches the status the line carries, with the live head that a
// closing or superseding status names.
var statusField = regexp.MustCompile(`\{status:\s*(?P<kind>[a-z-]+)(?:\s+(?P<by>[0-9a-z-]+))?\}`)

// citationLine matches the line a search renders beneath each entry it ranked:
//
//	↳ 100%  ·  Summary  ·  <snippet>
//
// The percentage is each citation's score against the strongest in the same
// result, so the first entry always renders 100% and the rest scale against it.
// It says how one match ranks beside the others in one search, and nothing
// about how good any of them is on its own. Measured against sdd 0.16.2.
var citationLine = regexp.MustCompile(`^\s*↳\s+(?P<score>\d{1,3})%\s`)

// parseSearchScores reads a rendered search into the relative score of every
// entry it ranked. An entry carries a citation for each part of it that
// matched, ordered strongest first, and the strongest is what stands for the
// entry.
func parseSearchScores(rendered string) map[string]int {
	scores := map[string]int{}
	current := ""
	for _, line := range splitLines(rendered) {
		id, _, isEntry := parseEntryLine(line)
		if isEntry {
			current = id
			continue
		}
		if current == "" {
			continue
		}
		match := citationLine.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if _, held := scores[current]; held {
			continue
		}
		score, err := strconv.Atoi(match[citationLine.SubexpIndex("score")])
		if err != nil {
			continue
		}
		scores[current] = score
	}
	return scores
}

// parseEntryLine reads one rendered line. A line that carries no entry ID is
// not one, which is how section headings and blank lines are passed over.
func parseEntryLine(line string) (string, Status, bool) {
	match := entryLine.FindStringSubmatch(line)
	if match == nil {
		return "", Status{}, false
	}
	id := match[entryLine.SubexpIndex("id")]
	rest := match[entryLine.SubexpIndex("rest")]

	status := statusField.FindStringSubmatch(rest)
	if status == nil {
		// A done signal renders no status at all, which is a fact about the
		// entry rather than a line the parser failed on.
		return id, Status{}, true
	}
	return id, Status{
		Kind: status[statusField.SubexpIndex("kind")],
		By:   status[statusField.SubexpIndex("by")],
	}, true
}

// parseEntryLines reads a rendered view into the status of every entry it names.
func parseEntryLines(rendered string) map[string]Status {
	statuses := map[string]Status{}
	for _, line := range splitLines(rendered) {
		id, status, ok := parseEntryLine(line)
		if !ok {
			continue
		}
		statuses[id] = status
	}
	return statuses
}

func splitLines(text string) []string {
	var lines []string
	start := 0
	for index := range len(text) {
		if text[index] == '\n' {
			lines = append(lines, text[start:index])
			start = index + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}
