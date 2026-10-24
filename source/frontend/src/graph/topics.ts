import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';

// An area names the body of work an entry belongs to and a facet names a
// concern it touches. Both share one namespace, and the prefix tells them
// apart, so the areas sort ahead of the facets.
const areaPrefix = 'area-';

export function isArea(topic: string): boolean {
	return topic.startsWith(areaPrefix);
}

// `topicsOf` reports every topic the entries carry, areas first and each group in
// alphabetical order.
export function topicsOf(entries: Entry[]): string[] {
	const held = new Set<string>();
	for (const entry of entries) {
		for (const topic of entry.topics ?? []) {
			held.add(topic);
		}
	}

	return [...held].sort((first, second) => {
		if (isArea(first) !== isArea(second)) {
			return isArea(first) ? -1 : 1;
		}
		return first.localeCompare(second);
	});
}

// `availableTopics` reports the topics worth offering: those carried by the
// entries that survive the current choice, which is every topic that would
// still leave something to read, plus the chosen ones so a reader can undo
// them. A topic leading to nothing is absent rather than disabled.
export function availableTopics(entries: Entry[], chosen: string[]): string[] {
	const reachable = new Set(topicsOf(withTopics(entries, chosen)));
	for (const topic of chosen) {
		reachable.add(topic);
	}
	return topicsOf(entries).filter((topic) => reachable.has(topic));
}

// `withTopics` narrows entries to those carrying every topic chosen, because
// choosing an area and a facet together asks for the overlap rather than the
// union.
export function withTopics(entries: Entry[], chosen: string[]): Entry[] {
	if (chosen.length === 0) {
		return entries;
	}
	return entries.filter((entry) => {
		const topics = new Set(entry.topics ?? []);
		return chosen.every((topic) => topics.has(topic));
	});
}
