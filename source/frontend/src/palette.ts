// The colours the two views share, so a decision drawn in the graph and the same
// decision listed in the table read as one thing.

export const typeColours: Record<string, string> = {
	decision: '#6ec4e8',
	signal: '#e8b246',
};

export const linkColours: Record<string, string> = {
	'grounded-in': '#7d8fd6',
	'builds-on': '#6fc9a8',
	refines: '#c78fe0',
	addresses: '#e2a05f',
	surfaces: '#dd7f8e',
	'surfaced-by': '#dd7f8e',
	'depends-on': '#5fa8d3',
	'required-by': '#5fa8d3',
	related: '#6a6a74',
	closes: '#7ee2a0',
	supersedes: '#e07b7b',
};

// The statuses that mean a reader is looking at something the graph has retired.
export const retiredStatuses = new Set([
	'closed-by',
	'superseded-by',
	'cascade-closed',
	'cascade-orphan',
]);

export const statusLabels: Record<string, string> = {
	open: 'Open',
	active: 'Active',
	settled: 'Settled',
	'closed-by': 'Closed',
	'superseded-by': 'Superseded',
	'cascade-closed': 'Cascade closed',
	'cascade-orphan': 'Cascade orphan',
};
