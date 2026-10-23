import { describe, expect, it } from 'vitest';
import { entryName } from './names';

describe('entryName', () => {
	it('names an entry by the title its file carries', () => {
		expect(entryName({ title: 'Status on entry records', summary: 'A long claim.' })).toBe(
			'Status on entry records',
		);
	});

	it('falls back to the summary where the file carries no title', () => {
		expect(entryName({ title: '', summary: 'A long claim.' })).toBe('A long claim.');
	});
});
