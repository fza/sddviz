import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';

// `entryName` is what names an entry wherever the viewer lists or draws one. sdd
// writes a short title into each entry from graph schema 2 on, and a graph
// written before that carries none, so the summary stands in for it there.
export function entryName(entry: Pick<Entry, 'title' | 'summary'>): string {
	return entry.title !== '' ? entry.title : entry.summary;
}
