// The widths a reader has not changed, in pixels, keyed by column. The summary
// carries none: it takes whatever the rest leave, so the prose a reader is
// actually reading gets every pixel the others do not need.
export const defaultWidths: Record<string, number> = {
	ordinal: 56,
	score: 78,
	captured: 178,
	short: 136,
	layer: 124,
	kind: 124,
	status: 128,
	draw: 68,
};

// A column narrower than this hides its own heading, so a pull stops here.
export const minimumWidth = 44;

// summaryFloor is the room the summary keeps at the table's narrowest, below
// which the table scrolls sideways rather than squeezing the prose to nothing.
const summaryFloor = 260;

export type Widths = Record<string, number>;

export function widthOf(widths: Widths, column: string): number {
	return widths[column] ?? defaultWidths[column] ?? minimumWidth;
}

// resizeTo reports the widths that follow a pull, clamping the column so it
// cannot fall below what its heading needs.
export function resizeTo(widths: Widths, column: string, width: number): Widths {
	return { ...widths, [column]: Math.max(minimumWidth, Math.round(width)) };
}

// tableFloor is the width below which the table scrolls sideways: every column
// carrying a width, at the width it carries, plus the room the summary keeps.
export function tableFloor(widths: Widths, columns: string[]): number {
	const fixed = columns
		.filter((column) => column !== 'summary')
		.reduce((total, column) => total + widthOf(widths, column), 0);
	return fixed + summaryFloor;
}


// The detail panel's width, in pixels, and the bounds a pull holds it inside.
// Narrower than the floor hides its own headings; wider than the ceiling leaves
// the view it sits beside too little room to read.
export const defaultDetailWidth = 440;
const detailFloor = 320;
const detailCeiling = 900;

export function resizeDetail(width: number): number {
	return Math.min(detailCeiling, Math.max(detailFloor, Math.round(width)));
}
