import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';

// A done signal records that a piece of work happened. It carries no status of
// its own, and a reader following what is still open often wants it out of the
// way.
export const doneKind = 'done';

export function withoutDone(entries: Entry[]): Entry[] {
	return entries.filter((entry) => entry.kind !== doneKind);
}

export function includingDone(entries: Entry[], include: boolean): Entry[] {
	return include ? entries : withoutDone(entries);
}
