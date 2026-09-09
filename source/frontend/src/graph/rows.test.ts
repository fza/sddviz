import { describe, expect, it } from 'vitest';
import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';
import { nextSort, sortEntries, splitIdentifier } from './rows';

function entry(id: string, fields: Partial<Entry> = {}): Entry {
	return {
		id,
		type: 'decision',
		layer: 'tactical',
		kind: 'directive',
		confidence: 'high',
		participants: [],
		topics: [],
		summary: '',
		body: '',
		refs: [],
		closes: [],
		supersedes: [],
		status: { kind: 'settled' },
		...fields,
	};
}

describe('splitIdentifier', () => {
	it('separates the moment an entry was captured from what it is', () => {
		expect(splitIdentifier('20260907-161758-d-tac-eu1')).toEqual({
			captured: '2026-09-07 16:17',
			short: 'd-tac-eu1',
		});
	});

	it('leaves an identifier it does not recognise whole', () => {
		expect(splitIdentifier('register')).toEqual({ captured: '', short: 'register' });
	});
});

describe('sortEntries', () => {
	const entries = [
		entry('20260907-161758-d-tac-eu1', { layer: 'tactical' }),
		entry('20260901-100000-s-ops-aaa', { layer: 'operational' }),
		entry('20260903-120000-d-cpt-bbb', { layer: 'conceptual' }),
	];

	it('keeps the given order when nothing is sorted', () => {
		expect(sortEntries(entries, null)).toBe(entries);
	});

	it('orders by the moment of capture', () => {
		const sorted = sortEntries(entries, { column: 'captured', ascending: true });

		expect(sorted.map((held) => held.id)).toEqual([
			'20260901-100000-s-ops-aaa',
			'20260903-120000-d-cpt-bbb',
			'20260907-161758-d-tac-eu1',
		]);
	});

	it('turns the order around', () => {
		const sorted = sortEntries(entries, { column: 'captured', ascending: false });

		expect(sorted[0].id).toBe('20260907-161758-d-tac-eu1');
	});

	it('orders by a field the entries share', () => {
		const sorted = sortEntries(entries, { column: 'layer', ascending: true });

		expect(sorted.map((held) => held.layer)).toEqual(['conceptual', 'operational', 'tactical']);
	});

	it('leaves the entries it was given untouched', () => {
		const given = [...entries];
		sortEntries(entries, { column: 'layer', ascending: true });

		expect(entries).toEqual(given);
	});

	it('leaves the newest first where a column ties', () => {
		const tied = [
			entry('20260901-100000-d-tac-aaa', { kind: 'directive' }),
			entry('20260907-100000-d-tac-bbb', { kind: 'directive' }),
		];
		const sorted = sortEntries(tied, { column: 'kind', ascending: true });

		expect(sorted[0].id).toBe('20260907-100000-d-tac-bbb');
	});
});

describe('sortEntries by score', () => {
	const entries = [
		entry('20260901-100000-d-tac-aaa', { score: 9 }),
		entry('20260902-100000-d-tac-bbb', { score: 100 }),
		entry('20260903-100000-d-tac-ccc', { score: 40 }),
	];

	it('orders a score as a number rather than as text', () => {
		const sorted = sortEntries(entries, { column: 'score', ascending: true });

		expect(sorted.map((held) => held.score)).toEqual([9, 40, 100]);
	});

	it('turns the order around', () => {
		const sorted = sortEntries(entries, { column: 'score', ascending: false });

		expect(sorted.map((held) => held.score)).toEqual([100, 40, 9]);
	});

	it('treats an entry carrying no score as the weakest', () => {
		const mixed = [entry('20260901-100000-d-tac-aaa'), entry('20260902-100000-d-tac-bbb', { score: 5 })];
		const sorted = sortEntries(mixed, { column: 'score', ascending: true });

		expect(sorted[0].score).toBeUndefined();
	});
});

describe('nextSort', () => {
	it('opens a newly chosen column ascending', () => {
		expect(nextSort(null, 'layer')).toEqual({ column: 'layer', ascending: true });
		expect(nextSort({ column: 'kind', ascending: false }, 'layer')).toEqual({
			column: 'layer',
			ascending: true,
		});
	});

	it('turns the sorted column around, then clears it', () => {
		expect(nextSort({ column: 'layer', ascending: true }, 'layer')).toEqual({
			column: 'layer',
			ascending: false,
		});
		expect(nextSort({ column: 'layer', ascending: false }, 'layer')).toBeNull();
	});
});
