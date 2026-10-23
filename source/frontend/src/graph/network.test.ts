import { describe, expect, it } from 'vitest';
import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';
import { buildNetwork, neighbourhood } from './network';

function entry(id: string, fields: Partial<Entry> = {}): Entry {
	return {
		id,
		type: 'decision',
		layer: 'tactical',
		kind: 'directive',
		confidence: 'high',
		participants: [],
		topics: [],
		title: '',
		summary: `the summary of ${id}`,
		body: '',
		refs: [],
		closes: [],
		supersedes: [],
		status: { kind: 'settled' },
		...fields,
	};
}

const first = '20260829-221000-d-cpt-d8y';
const second = '20260829-221103-d-tac-tz1';
const absent = '20260829-000000-d-tac-gone';

describe('buildNetwork', () => {
	it('draws a link for every reference whose target the graph holds', () => {
		const network = buildNetwork([
			entry(first, { refs: [{ id: second, kind: 'refines', desc: 'narrows it' }] }),
			entry(second),
		]);

		expect(network.links).toEqual([
			{ source: first, target: second, kind: 'refines', desc: 'narrows it' },
		]);
		expect(network.outgoing.get(first)).toHaveLength(1);
		expect(network.incoming.get(second)).toHaveLength(1);
		expect(network.unresolved.size).toBe(0);
	});

	it('draws closing and superseding as links of their own', () => {
		const network = buildNetwork([
			entry(first, { closes: [second], supersedes: [second] }),
			entry(second),
		]);

		expect(network.links.map((link) => link.kind)).toEqual(['closes', 'supersedes']);
	});

	it('leaves a reference naming an entry the graph does not hold undrawn', () => {
		const network = buildNetwork([
			entry(first, { refs: [{ id: absent, kind: 'related', desc: '' }] }),
		]);

		expect(network.links).toHaveLength(0);
		expect(network.unresolved.get(first)).toEqual([absent]);
	});

	it('leaves an entry naming itself undrawn', () => {
		const network = buildNetwork([entry(first, { closes: [first] })]);

		expect(network.links).toHaveLength(0);
		expect(network.unresolved.get(first)).toEqual([first]);
	});

	it('counts the links at both ends of an entry', () => {
		const third = '20260829-223328-d-tac-3qx';
		const network = buildNetwork([
			entry(first, { refs: [{ id: second, kind: 'refines', desc: '' }] }),
			entry(third, { refs: [{ id: second, kind: 'grounded-in', desc: '' }] }),
			entry(second),
		]);

		expect(network.degrees.get(second)).toBe(2);
		expect(network.degrees.get(first)).toBe(1);
	});

	it('reports an entry that declares no link at all', () => {
		const network = buildNetwork([entry(first, { refs: null, closes: null, supersedes: null })]);

		expect(network.links).toHaveLength(0);
		expect(network.degrees.get(first)).toBe(0);
	});
});

describe('neighbourhood', () => {
	const a = '20260829-221000-d-cpt-d8y';
	const b = '20260829-221103-d-tac-tz1';
	const c = '20260829-223328-d-tac-3qx';
	const d = '20260829-235605-d-prc-zbm';

	const network = buildNetwork([
		entry(a, { refs: [{ id: b, kind: 'refines', desc: '' }] }),
		entry(b, { refs: [{ id: c, kind: 'related', desc: '' }] }),
		entry(c, { refs: [{ id: d, kind: 'related', desc: '' }] }),
		entry(d),
	]);

	it('places the entry itself at no distance', () => {
		expect(neighbourhood(network, a, 2).get(a)).toBe(0);
	});

	it('follows a link in either direction', () => {
		expect(neighbourhood(network, b, 1).get(a)).toBe(1);
		expect(neighbourhood(network, b, 1).get(c)).toBe(1);
	});

	it('stops at the depth it is asked for', () => {
		const reached = neighbourhood(network, a, 2);

		expect(reached.get(c)).toBe(2);
		expect(reached.has(d)).toBe(false);
	});

	it('keeps the shortest distance to an entry reachable two ways', () => {
		const linked = buildNetwork([
			entry(a, { refs: [{ id: b, kind: 'related', desc: '' }, { id: c, kind: 'related', desc: '' }] }),
			entry(b, { refs: [{ id: c, kind: 'related', desc: '' }] }),
			entry(c),
		]);

		expect(neighbourhood(linked, a, 2).get(c)).toBe(1);
	});
});
