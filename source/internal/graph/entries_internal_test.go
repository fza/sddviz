package graph

import "testing"

func TestRefsFieldSkipsWhatItCannotRead(t *testing.T) {
	cases := []struct {
		name        string
		frontmatter map[string]any
		expected    []Ref
	}{
		{
			name:        "no refs key at all",
			frontmatter: map[string]any{},
			expected:    []Ref{},
		},
		{
			name:        "a refs key that is not a list",
			frontmatter: map[string]any{"refs": "20260829-223328-d-tac-3qx"},
			expected:    []Ref{},
		},
		{
			name: "an element that is not a mapping, beside one that is",
			frontmatter: map[string]any{"refs": []any{
				"20260829-223328-d-tac-3qx",
				map[string]any{"id": "20260829-221103-d-tac-tz1", "kind": "grounded-in"},
			}},
			expected: []Ref{{ID: "20260829-221103-d-tac-tz1", Kind: "grounded-in"}},
		},
		{
			name: "a mapping naming no entry, beside one that does",
			frontmatter: map[string]any{"refs": []any{
				map[string]any{"kind": "related", "desc": "names nothing"},
				map[string]any{"id": "20260829-221000-d-cpt-d8y", "kind": "refines", "desc": "narrows it"},
			}},
			expected: []Ref{{ID: "20260829-221000-d-cpt-d8y", Kind: "refines", Desc: "narrows it"}},
		},
		{
			name: "a reference whose kind and description are absent",
			frontmatter: map[string]any{"refs": []any{
				map[string]any{"id": "20260829-221000-d-cpt-d8y"},
			}},
			expected: []Ref{{ID: "20260829-221000-d-cpt-d8y"}},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			refs := refsField(testCase.frontmatter, "refs")
			if refs == nil {
				t.Fatal("the reader reports a missing field rather than an empty list")
			}
			if len(refs) != len(testCase.expected) {
				t.Fatalf("the reader kept %d references, expected the readable ones", len(refs))
			}
			for index, want := range testCase.expected {
				if refs[index] != want {
					t.Errorf("reference %d reads %+v, expected the readable one in that position", index, refs[index])
				}
			}
		})
	}
}
