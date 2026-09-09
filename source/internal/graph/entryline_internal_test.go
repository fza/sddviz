package graph

import (
	"os"
	"testing"
)

func TestParseEntryLine(t *testing.T) {
	cases := []struct {
		name   string
		line   string
		id     string
		status Status
		ok     bool
	}{
		{
			name:   "an open signal",
			line:   "  20260829-200722-s-prc-9o1 process actor signal [confidence: high] (Felix) {status: open} <area-people> `Felix` is the canonical name",
			id:     "20260829-200722-s-prc-9o1",
			status: Status{Kind: "open"},
			ok:     true,
		},
		{
			name:   "an active decision",
			line:   "  20260829-200554-d-stg-0nd strategic aspiration decision [confidence: medium] (Felix) {status: active} <area-product> sddviz exists",
			id:     "20260829-200554-d-stg-0nd",
			status: Status{Kind: "active"},
			ok:     true,
		},
		{
			name:   "a superseded entry names its live head",
			line:   "  20260829-200314-s-prc-abi process fact signal [confidence: high] (Felix) {status: superseded-by 20260829-205109-s-prc-xv4} <area-scaffold> sddviz is",
			id:     "20260829-200314-s-prc-abi",
			status: Status{Kind: "superseded-by", By: "20260829-205109-s-prc-xv4"},
			ok:     true,
		},
		{
			name:   "a closed signal names its closer",
			line:   "  20260416-190732-s-prc-omw process gap signal [confidence: high] (Christopher, Claude) {status: closed-by 20260417-093934-s-prc-fxn} A closed signal",
			id:     "20260416-190732-s-prc-omw",
			status: Status{Kind: "closed-by", By: "20260417-093934-s-prc-fxn"},
			ok:     true,
		},
		{
			name:   "a settled directive",
			line:   "  20260416-151058-d-tac-set tactical directive decision [confidence: high] (Christopher) {status: settled} Born terminal",
			id:     "20260416-151058-d-tac-set",
			status: Status{Kind: "settled"},
			ok:     true,
		},
		{
			name:   "a guiding directive carries its intent before the participants",
			line:   "  20260416-151058-d-stg-gid strategic directive decision [confidence: medium] [intent: guiding] (Christopher, Claude) {status: active} A standing directive",
			id:     "20260416-151058-d-stg-gid",
			status: Status{Kind: "active"},
			ok:     true,
		},
		{
			name:   "an entry with no participants",
			line:   "  20260407-144751-s-ops-gn0 operational gap signal () {status: open} A signal with no participants recorded",
			id:     "20260407-144751-s-ops-gn0",
			status: Status{Kind: "open"},
			ok:     true,
		},
		{
			name:   "a done signal carries no status at all",
			line:   "  20260703-094500-s-prc-don process done signal [confidence: high] (Felix) The work landed",
			id:     "20260703-094500-s-prc-don",
			status: Status{},
			ok:     true,
		},
		{name: "a section heading is not an entry", line: "## Active", ok: false},
		{name: "an empty line is not an entry", line: "", ok: false},
		{name: "prose mentioning no id is not an entry", line: "  nothing here names an entry", ok: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			id, status, ok := parseEntryLine(testCase.line)
			if ok != testCase.ok {
				t.Fatalf("the line was recognised as %v, expected %v", ok, testCase.ok)
			}
			if !ok {
				return
			}
			if id != testCase.id {
				t.Errorf("the line yielded id %q, expected %q", id, testCase.id)
			}
			if status != testCase.status {
				t.Errorf("the line yielded status %+v, expected %+v", status, testCase.status)
			}
		})
	}
}

func TestParseEntryLinesOverRealOutput(t *testing.T) {
	rendered, err := os.ReadFile("testdata/rendered/as-list.txt")
	if err != nil {
		t.Fatalf("reading the captured view: %v", err)
	}

	statuses := parseEntryLines(string(rendered))
	if len(statuses) < 10 {
		t.Fatalf("the captured view yielded %d entries, expected every line it carries", len(statuses))
	}

	superseded, found := statuses["20260829-200314-s-prc-abi"]
	if !found {
		t.Fatal("the captured view names an entry the parser did not report")
	}
	if superseded.Kind != "superseded-by" || superseded.By == "" {
		t.Errorf("a superseded entry parsed as %+v, expected the status and the entry replacing it", superseded)
	}
}

func TestParseSearchScoresReadsARenderedSearch(t *testing.T) {
	rendered, err := os.ReadFile("testdata/rendered/search.txt")
	if err != nil {
		t.Fatalf("reading the rendered search: %v", err)
	}

	scores := parseSearchScores(string(rendered))

	if len(scores) == 0 {
		t.Fatal("the rendered search yielded no score at all")
	}
	if scores["20260829-205243-d-tac-7o5"] != 100 {
		t.Errorf("the strongest match scored %d, expected the hundred sdd normalises it to", scores["20260829-205243-d-tac-7o5"])
	}
	if scores["20260814-100000-s-prc-psr"] != 80 {
		t.Errorf("a weaker match scored %d", scores["20260814-100000-s-prc-psr"])
	}
	for id, score := range scores {
		if score < 0 || score > 100 {
			t.Errorf("%s scored %d, which is outside the range a percentage takes", id, score)
		}
	}
}

func TestParseSearchScoresKeepsTheStrongestCitationOfAnEntry(t *testing.T) {
	rendered := "  20260829-205243-d-tac-7o5 tactical directive decision (Felix) {status: settled} A summary\n" +
		"    ↳ 100%  ·  Summary  ·  the strongest part\n" +
		"    ↳  40%  ·  Body  ·  a weaker part\n"

	scores := parseSearchScores(rendered)

	if scores["20260829-205243-d-tac-7o5"] != 100 {
		t.Errorf("the entry scored %d, expected its strongest citation", scores["20260829-205243-d-tac-7o5"])
	}
}

func TestParseSearchScoresPassesOverWhatIsNotACitation(t *testing.T) {
	rendered := "  (no matches)\n    ↳ 100%  ·  Summary  ·  a citation naming no entry\n"

	if len(parseSearchScores(rendered)) != 0 {
		t.Error("a citation with no entry above it was counted anyway")
	}
}
