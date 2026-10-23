import { describe, expect, it } from 'vitest';
import {
	defaultDetailWidth,
	defaultWidths,
	minimumWidth,
	resizeDetail,
	resizeTo,
	tableFloor,
	widthOf,
} from './columns';

describe('widthOf', () => {
	it('reports what a reader chose', () => {
		expect(widthOf({ layer: 200 }, 'layer')).toBe(200);
	});

	it('falls back to the width the column carries by default', () => {
		expect(widthOf({}, 'layer')).toBe(defaultWidths.layer);
	});

	it('falls back to the floor for a column it does not know', () => {
		expect(widthOf({}, 'nothing')).toBe(minimumWidth);
	});
});

describe('resizeTo', () => {
	it('takes the width a pull arrived at', () => {
		expect(resizeTo({}, 'layer', 180)).toEqual({ layer: 180 });
	});

	it('rounds to whole pixels', () => {
		expect(resizeTo({}, 'layer', 180.6)).toEqual({ layer: 181 });
	});

	it('stops at the floor rather than letting a column vanish', () => {
		expect(resizeTo({}, 'layer', 4)).toEqual({ layer: minimumWidth });
		expect(resizeTo({}, 'layer', -50)).toEqual({ layer: minimumWidth });
	});

	it('leaves every other column alone', () => {
		expect(resizeTo({ kind: 90 }, 'layer', 180)).toEqual({ kind: 90, layer: 180 });
	});
});

describe('tableFloor', () => {
	it('adds every column carrying a width to the room the name keeps', () => {
		const floor = tableFloor({}, ['ordinal', 'layer', 'name']);

		expect(floor).toBe(defaultWidths.ordinal + defaultWidths.layer + 260);
	});

	it('grows as a column is widened', () => {
		const columns = ['ordinal', 'layer', 'name'];

		expect(tableFloor({ layer: 300 }, columns)).toBeGreaterThan(tableFloor({}, columns));
	});

	it('counts the score column only where it is shown', () => {
		const without = tableFloor({}, ['ordinal', 'name']);
		const shown = tableFloor({}, ['ordinal', 'score', 'name']);

		expect(shown - without).toBe(defaultWidths.score);
	});

	it('leaves the name out of the sum, since it takes what is left', () => {
		expect(tableFloor({ name: 9999 }, ['ordinal', 'name'])).toBe(
			defaultWidths.ordinal + 260,
		);
	});
});


describe('resizeDetail', () => {
	it('takes the width a pull arrived at', () => {
		expect(resizeDetail(600)).toBe(600);
	});

	it('rounds to whole pixels', () => {
		expect(resizeDetail(600.4)).toBe(600);
	});

	it('stops before the panel is too narrow to read', () => {
		expect(resizeDetail(40)).toBe(320);
	});

	it('stops before the panel crowds out the view beside it', () => {
		expect(resizeDetail(4000)).toBe(900);
	});

	it('holds the width it opens at', () => {
		expect(resizeDetail(defaultDetailWidth)).toBe(defaultDetailWidth);
	});
});
