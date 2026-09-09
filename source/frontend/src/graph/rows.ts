import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';

// An entry identifier opens with the moment the graph received it and closes
// with what the entry is: `20260907-161758-d-tac-eu1` was written on the 7th of
// September 2026 and is the tactical decision `d-tac-eu1`.
const identifier = /^(\d{4})(\d{2})(\d{2})-(\d{2})(\d{2})(\d{2})-(.+)$/;

export interface SplitIdentifier {
	captured: string;
	short: string;
}

export function splitIdentifier(id: string): SplitIdentifier {
	const match = identifier.exec(id);
	if (match === null) {
		return { captured: '', short: id };
	}
	const [, year, month, day, hour, minute, , short] = match;
	return { captured: `${year}-${month}-${day} ${hour}:${minute}`, short };
}

// A summary is prose, and ordering a graph by the first letter of its prose
// tells a reader nothing, so it carries no sort.
export type SortColumn = 'captured' | 'short' | 'layer' | 'kind' | 'status' | 'score';

export interface Sort {
	column: SortColumn;
	ascending: boolean;
}

function sortValue(entry: Entry, column: SortColumn): string {
	switch (column) {
		case 'captured':
			return entry.id;
		case 'short':
			return splitIdentifier(entry.id).short;
		case 'layer':
			return entry.layer;
		case 'kind':
			return entry.kind;
		case 'status':
			return entry.status.kind;
		case 'score':
			return '';
	}
}

// A score is a number, and comparing it as text puts 9 above 10.
function compare(first: Entry, second: Entry, column: SortColumn): number {
	if (column === 'score') {
		return (first.score ?? 0) - (second.score ?? 0);
	}
	return sortValue(first, column).localeCompare(sortValue(second, column));
}

// sortEntries orders a copy, so the order sdd reported survives beside it.
export function sortEntries(entries: Entry[], sort: Sort | null): Entry[] {
	if (sort === null) {
		return entries;
	}
	const direction = sort.ascending ? 1 : -1;
	return [...entries].sort((first, second) => {
		const compared = compare(first, second, sort.column);
		if (compared !== 0) {
			return compared * direction;
		}
		// A column many entries share leaves the newest first, which is the
		// order the table carries with nothing sorted.
		return first.id < second.id ? 1 : -1;
	});
}

// nextSort turns a chosen column into the sort that follows it: a new column
// opens ascending, and the column already sorted turns around, then clears.
export function nextSort(current: Sort | null, column: SortColumn): Sort | null {
	if (current === null || current.column !== column) {
		return { column, ascending: true };
	}
	if (current.ascending) {
		return { column, ascending: false };
	}
	return null;
}
