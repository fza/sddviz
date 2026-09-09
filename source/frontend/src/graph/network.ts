import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';

// `closesKind` and `supersedesKind` name the two links an entry declares outside
// its `refs` list. They share the namespace of the kinds a reference carries,
// which is what lets one drawn network hold all three.
export const closesKind = 'closes';
export const supersedesKind = 'supersedes';

export interface GraphLink {
	source: string;
	target: string;
	kind: string;
	desc: string;
}

// EntryNetwork is what the views draw: the links whose target the graph holds,
// indexed from both ends, and the identifiers named by a link the graph cannot
// resolve. A record reports what its file declares, so an entry naming one that
// no file carries reaches here and is drawn nowhere.
export interface EntryNetwork {
	links: GraphLink[];
	outgoing: Map<string, GraphLink[]>;
	incoming: Map<string, GraphLink[]>;
	degrees: Map<string, number>;
	unresolved: Map<string, string[]>;
}

function declaredLinks(entry: Entry): GraphLink[] {
	const links: GraphLink[] = (entry.refs ?? []).map((ref) => ({
		source: entry.id,
		target: ref.id,
		kind: ref.kind === '' ? 'related' : ref.kind,
		desc: ref.desc ?? '',
	}));

	for (const target of entry.closes ?? []) {
		links.push({ source: entry.id, target, kind: closesKind, desc: '' });
	}
	for (const target of entry.supersedes ?? []) {
		links.push({ source: entry.id, target, kind: supersedesKind, desc: '' });
	}
	return links;
}

function push(index: Map<string, GraphLink[]>, key: string, link: GraphLink) {
	const held = index.get(key);
	if (held === undefined) {
		index.set(key, [link]);
		return;
	}
	held.push(link);
}

export function buildNetwork(entries: Entry[]): EntryNetwork {
	const held = new Set(entries.map((entry) => entry.id));
	const network: EntryNetwork = {
		links: [],
		outgoing: new Map(),
		incoming: new Map(),
		degrees: new Map(),
		unresolved: new Map(),
	};

	for (const entry of entries) {
		network.degrees.set(entry.id, network.degrees.get(entry.id) ?? 0);
		for (const link of declaredLinks(entry)) {
			if (!held.has(link.target) || link.target === link.source) {
				const named = network.unresolved.get(entry.id) ?? [];
				named.push(link.target);
				network.unresolved.set(entry.id, named);
				continue;
			}
			network.links.push(link);
			push(network.outgoing, link.source, link);
			push(network.incoming, link.target, link);
			network.degrees.set(link.source, (network.degrees.get(link.source) ?? 0) + 1);
			network.degrees.set(link.target, (network.degrees.get(link.target) ?? 0) + 1);
		}
	}

	return network;
}

// neighbourhood reports how many links away each entry sits from the one given,
// following a link in either direction. The entry itself sits at zero, and an
// entry further than the depth asked for is absent.
export function neighbourhood(
	network: EntryNetwork,
	id: string,
	depth: number,
): Map<string, number> {
	const distances = new Map<string, number>([[id, 0]]);
	let frontier = [id];

	for (let step = 1; step <= depth; step += 1) {
		const next: string[] = [];
		for (const current of frontier) {
			const touching = [
				...(network.outgoing.get(current) ?? []),
				...(network.incoming.get(current) ?? []),
			];
			for (const link of touching) {
				const other = link.source === current ? link.target : link.source;
				if (distances.has(other)) {
					continue;
				}
				distances.set(other, step);
				next.push(other);
			}
		}
		frontier = next;
	}

	return distances;
}
