import { describe, expect, it } from 'vitest';
import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';
import { includingDone, withoutDone } from './kinds';

function entry(id: string, kind: string): Entry {
	return {
		id,
		type: 'signal',
		layer: 'tactical',
		kind,
		confidence: 'high',
		participants: [],
		topics: [],
		title: '',
		summary: '',
		body: '',
		refs: [],
		closes: [],
		supersedes: [],
		status: { kind: '' },
	};
}

const entries = [entry('one', 'done'), entry('two', 'gap'), entry('three', 'done')];

describe('withoutDone', () => {
	it('drops what records that work happened', () => {
		expect(withoutDone(entries).map((held) => held.id)).toEqual(['two']);
	});

	it('leaves a set carrying none of them alone', () => {
		const open = [entry('two', 'gap')];

		expect(withoutDone(open)).toEqual(open);
	});
});

describe('includingDone', () => {
	it('returns every entry while they are included', () => {
		expect(includingDone(entries, true)).toBe(entries);
	});

	it('drops them once they are not', () => {
		expect(includingDone(entries, false).map((held) => held.id)).toEqual(['two']);
	});
});
