# Rules - the vocabulary this repository fixes

**This document decides which words this repository uses for its own parts, and in which form.**
Binding on every document, every commit message, every code comment, every sentence written to a
reader, and every developer-facing value.

Companion to the authoring rules in [`AGENTS.md`](../AGENTS.md#authoring). What splits them is what
checks each one. A rule here is a word a text checker can find, and a rule there is a judgement no
checker can make.

---

## 1. How these rules work

### The words themselves are Vale rules, never prose

**A check carries every word this document fixes.** A table and a check say the same thing until
someone edits one, and the one nobody runs is the one that rots. What each word must be is in the
rule's own `message`, which is what a writer actually reads when it fires.

| Kind | Rule | Where it lives |
|---|---|---|
| **Banned** | Retired in every sense for this repository's own parts. No justification reintroduces one | `.vale/styles/Taxonomy/Entry.yml`, at **error** |
| **Explicit** | Use without restriction. The check asks whether a reader could take the word for one of its named senses and be wrong | `.vale/styles/Taxonomy/{Node,Topic}.yml`, at **suggestion** |

**The doubt test, for every explicit word:** whether a reader could take this for one of its named
senses and be wrong. Yes, then qualify. Already clear from the sentence, then leave it. A suggestion
is the check asking that question, not an instruction to rewrite.

**The set starts small and earns each entry.** A word joins it when a real confusion appears, never
in anticipation.

### What these rules bind

| | Bound |
|---|---|
| **Prose** - documents, commit messages, code and doc comments, anything a reader sees | **yes, always** |
| **Developer-facing values** - Tauri command names, configuration keys, error text | **yes**, through the breaking-change gate |
| **Internal identifiers** - Rust and TypeScript names, JSON keys, file paths | **no** |

So a sentence naming an entry by a retired synonym violates the first row, while
"the entry is superseded" does not, and a three.js `node` variable in the back end stands under the
third:

<vale Taxonomy.Entry = NO>
"the card is superseded"
</vale Taxonomy.Entry>

- **Never rename an existing identifier to satisfy a prose rule.** That is a refactor needing its own
  justification.
- **A new identifier takes the aligned name where it costs nothing.**
- **An identifier ambiguous at its use site becomes a signal.** Never silently renamed, never ignored.

### Backticks are not an exemption, and are not a loophole either

**Vale does not read a code span**, so a backticked word is never reported. That is correct where the
backticks name an identifier, a flag or a wire value, which the table above leaves unbound. It is not
licence to backtick a word to silence a finding. The rule binds the prose, and quoting a bare word to
get past a check is the violation the check exists to catch.

### Mentioning a banned word

A word this document **mentions** rather than uses has to survive its own ban, and no text check
tells the two apart. Wrap the line, name **one** rule, and switch it back on immediately:

```markdown
<vale Taxonomy.Entry = NO>
- an early draft called each entry a card
</vale Taxonomy.Entry>
```

**Never a whole-file exemption.** Wrap only a quoted defect, a qualified form the check cannot see,
or a quoted claim in someone else's voice.

### Standing obligations

- **Never coin a competing term.** Needing a name this document lacks means changing this document,
  and adding the check alongside it.
- **In doubt: stop and ask.**
- **Correct on sight.** Fix a violation met while doing something else.
- **This outranks brevity, house style, and matching surrounding text.**

### The graph's words are SDD's, deliberately

**Where fdbox's equivalent rule keeps a tool's vocabulary out of product prose, this one does the
opposite for one bounded set.** sddviz displays an SDD graph, so the words for what it displays are
SDD's own, spelled as SDD spells them. Correcting one toward a word this project prefers would put a
name on screen that the tool a reader also uses never says.

That licence stops at the graph. Everything sddviz builds of its own takes a name decided here.

---

## 2. The graph's words

**Every term in this section is SDD's.** Its definition lives in that project, and this document
fixes only the spelling and the ban on competitors.

### entry

The unit of the graph. An `entry` is a `signal` or a `decision`, never both, and the pair are its two
types.

**Banned: `card`, `vertex`.** Reported by `Taxonomy/Entry.yml` at error. Neither is SDD's word, and
both invite a reader to think the viewer holds something other than what the tool wrote.

### node

**Explicit.** A `node` is the object drawn in the graph view. An `entry` is what the graph holds.
Reported by `Taxonomy/Node.yml` at suggestion, which asks which of the two a sentence means. A
sentence about geometry, a scene graph or a layout keeps the word.

### topic, area

A `topic` is a label on an entry, hierarchical and joined with `/`. An `area` is a topic whose first
component opens with `area-`, and an entry carries exactly one.

**Explicit: `tag`, `label`.** Reported by `Taxonomy/Topic.yml` at suggestion. Both are correct for a
git tag or an on-screen label, and neither names a topic.

### ref

A `ref` is a typed pointer from one entry to another, carrying one of SDD's nine kinds. The word for
the relationship is `ref`, never a synonym.

`supersedes` and `closes` are not refs. They are their own fields, and they are what status derives
from.

### status

`status` is derived, never stored. Its values are SDD's and are written exactly as the tool writes
them: `open`, `active`, `settled`, `closed-by`, `superseded-by`, `cascade-closed`, `cascade-orphan`,
and the empty status a done signal carries.

**Never invent a shorthand.** "retired", "dead" and "resolved" each blur two of those values.

### layer, kind, confidence, participant

SDD's words for an entry's other fixed fields, spelled as the tool spells them. The layers are
`strategic`, `conceptual`, `tactical`, `operational`, `process`.

---

## 3. This repository's own words

### the front end, the back end

`source/ui` is **the front end** and `source/backend` is **the back end**. Both are written as two
words in prose, and the directory names are identifiers rather than prose.

### repository

A `repository` is what a reader opens: a checkout carrying an `.sdd/` directory. The word for it is
`repository`, never "project", which SDD uses for something else, and never "workspace".
