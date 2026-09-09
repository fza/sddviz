import { describe, expect, it } from 'vitest';
import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';
import { availableTopics, isArea, topicsOf, withTopics } from './topics';

function entry(id: string, topics: string[]): Entry {
	return {
		id,
		type: 'decision',
		layer: 'tactical',
		kind: 'directive',
		confidence: 'high',
		participants: [],
		topics,
		summary: '',
		body: '',
		refs: [],
		closes: [],
		supersedes: [],
		status: { kind: 'settled' },
	};
}

describe('topicsOf', () => {
	it('reports each topic once, areas ahead of facets', () => {
		const topics = topicsOf([
			entry('one', ['stack', 'area-scaffold']),
			entry('two', ['area-people', 'stack']),
		]);

		expect(topics).toEqual(['area-people', 'area-scaffold', 'stack']);
	});

	it('reports nothing for entries carrying no topic', () => {
		expect(topicsOf([entry('one', [])])).toEqual([]);
	});
});

describe('withTopics', () => {
	const entries = [
		entry('one', ['area-scaffold', 'stack']),
		entry('two', ['area-scaffold', 'testing']),
		entry('three', ['area-people']),
	];

	it('returns every entry when nothing is chosen', () => {
		expect(withTopics(entries, [])).toHaveLength(3);
	});

	it('keeps the entries carrying the chosen topic', () => {
		expect(withTopics(entries, ['area-scaffold']).map((e) => e.id)).toEqual(['one', 'two']);
	});

	it('narrows to the overlap when several are chosen', () => {
		expect(withTopics(entries, ['area-scaffold', 'testing']).map((e) => e.id)).toEqual(['two']);
	});

	it('returns nothing when no entry carries them all', () => {
		expect(withTopics(entries, ['area-people', 'stack'])).toHaveLength(0);
	});
});

describe('availableTopics', () => {
	const entries = [
		entry('one', ['area-scaffold', 'stack']),
		entry('two', ['area-scaffold', 'testing']),
		entry('three', ['area-people', 'hiring']),
	];

	it('offers every topic when nothing is chosen', () => {
		expect(availableTopics(entries, [])).toEqual([
			'area-people',
			'area-scaffold',
			'hiring',
			'stack',
			'testing',
		]);
	});

	it('drops the topics that would leave nothing to read', () => {
		const offered = availableTopics(entries, ['area-scaffold']);

		expect(offered).toEqual(['area-scaffold', 'stack', 'testing']);
		expect(offered).not.toContain('area-people');
		expect(offered).not.toContain('hiring');
	});

	it('keeps a chosen topic offered so a reader can undo it', () => {
		expect(availableTopics(entries, ['area-scaffold', 'stack'])).toContain('area-scaffold');
		expect(availableTopics(entries, ['area-scaffold', 'stack'])).toContain('stack');
	});

	it('narrows further as topics are added', () => {
		expect(availableTopics(entries, ['area-scaffold', 'stack'])).toEqual(['area-scaffold', 'stack']);
	});

	it('keeps the areas ahead of the facets', () => {
		expect(availableTopics(entries, [])[0]).toBe('area-people');
	});

	it('offers nothing for a set carrying no entries', () => {
		expect(availableTopics([], [])).toEqual([]);
	});
});

describe('isArea', () => {
	it('tells an area from a facet by its prefix', () => {
		expect(isArea('area-scaffold')).toBe(true);
		expect(isArea('stack')).toBe(false);
	});
});
