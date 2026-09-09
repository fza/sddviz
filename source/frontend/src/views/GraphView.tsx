import { useEffect, useMemo, useRef } from 'react';
import ForceGraph3D from '3d-force-graph';
import type { ForceGraph3DInstance, NodeObject } from '3d-force-graph';
import type { Object3D } from 'three';
import { CanvasTexture, Sprite, SpriteMaterial } from 'three';
import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';
import { buildNetwork, closesKind, neighbourhood, supersedesKind } from '../graph/network';
import { linkColours, retiredStatuses, typeColours } from '../palette';

// EntryNode is what the drawing carries per entry. The library types a node as
// an open object, so the fields the accessors read are declared here.
type EntryNode = NodeObject & {
	id: string;
	type: string;
	summary: string;
	statusKind: string;
	weight: number;
};

type EntryLink = {
	source: string;
	target: string;
	kind: string;
	desc: string;
};

const selectedColour = '#ffffff';
const retiredColour = '#5c5c66';

// A corona marks how far an entry sits from the selected one: the selection
// itself, then what it links to, then what those link to. Each ring is smaller
// than the one before, so depth reads off size as well as colour.
const coronaTiers = [
	{ colour: '#ef4444', size: 34 },
	{ colour: '#a855f7', size: 22 },
	{ colour: '#a855f7', size: 14 },
];

const coronaDepth = coronaTiers.length - 1;

// The simulation settles rather than running for as long as a reader watches. A
// graph of a thousand entries keeps a frame budget busy indefinitely otherwise,
// and the reading a viewer is for starts once the drawing holds still.
const settleMilliseconds = 12000;

// How far the camera sits from an entry it flies to, in the units the layout
// works in.
const focusDistance = 480;

// Repulsion reaching across the whole drawing pushes an entry group that links
// to nothing else out past every other, which leaves a fit to the whole graph
// showing a speck. Capping its reach keeps those groups beside the body they
// belong to.
const repulsionReach = 260;

function coronaTexture(colour: string): CanvasTexture {
	const size = 128;
	const canvas = document.createElement('canvas');
	canvas.width = size;
	canvas.height = size;

	const context = canvas.getContext('2d');
	if (context !== null) {
		const gradient = context.createRadialGradient(size / 2, size / 2, 0, size / 2, size / 2, size / 2);
		gradient.addColorStop(0, `${colour}00`);
		gradient.addColorStop(0.55, `${colour}00`);
		gradient.addColorStop(0.72, `${colour}cc`);
		gradient.addColorStop(1, `${colour}00`);
		context.fillStyle = gradient;
		context.fillRect(0, 0, size, size);
	}

	return new CanvasTexture(canvas);
}

function escapeHTML(text: string): string {
	return text
		.replaceAll('&', '&amp;')
		.replaceAll('<', '&lt;')
		.replaceAll('>', '&gt;')
		.replaceAll('"', '&quot;');
}

// focusOn flies the camera to one entry. It answers whether it found the entry
// to fly to, so a caller can fall back to a fit over everything.
function focusOn(
	drawn: ForceGraph3DInstance<EntryNode, EntryLink>,
	id: string | null,
	milliseconds: number,
): boolean {
	if (id === null) {
		return false;
	}
	const node = drawn.graphData().nodes.find((candidate) => candidate.id === id);
	if (node?.x === undefined || node.y === undefined || node.z === undefined) {
		return false;
	}
	const reach = Math.hypot(node.x, node.y, node.z);
	const ratio = reach === 0 ? 1 : 1 + focusDistance / reach;
	drawn.cameraPosition(
		{ x: node.x * ratio, y: node.y * ratio, z: node.z * ratio },
		{ x: node.x, y: node.y, z: node.z },
		milliseconds,
	);
	return true;
}

function nodeColour(node: EntryNode, selectedID: string | null): string {
	if (node.id === selectedID) {
		return selectedColour;
	}
	if (retiredStatuses.has(node.statusKind)) {
		return retiredColour;
	}
	return typeColours[node.type] ?? typeColours.signal;
}

export function GraphView({
	entries,
	selectedID,
	onSelect,
}: {
	entries: Entry[];
	selectedID: string | null;
	onSelect: (id: string | null) => void;
}) {
	const container = useRef<HTMLDivElement | null>(null);
	const graph = useRef<ForceGraph3DInstance<EntryNode, EntryLink> | null>(null);
	const select = useRef(onSelect);
	select.current = onSelect;

	const network = useMemo(() => buildNetwork(entries), [entries]);
	const distances = useMemo(
		() => (selectedID === null ? new Map<string, number>() : neighbourhood(network, selectedID, coronaDepth)),
		[network, selectedID],
	);
	const paint = useRef({ selectedID, distances });
	paint.current = { selectedID, distances };

	const textures = useMemo(() => coronaTiers.map((tier) => coronaTexture(tier.colour)), []);

	useEffect(() => {
		const element = container.current;
		if (element === null || entries.length === 0) {
			return;
		}

		// The library declares its default export as a non-generic constructor,
		// so the node and link types reach the accessors through the instance
		// type it exports rather than through type arguments.
		const drawn = new ForceGraph3D(element) as unknown as ForceGraph3DInstance<EntryNode, EntryLink>;
		graph.current = drawn;

		const nodes: EntryNode[] = entries.map((entry) => ({
			id: entry.id,
			type: entry.type,
			summary: entry.summary,
			statusKind: entry.status.kind,
			weight: 1 + Math.sqrt(network.degrees.get(entry.id) ?? 0) * 2,
		}));
		const links: EntryLink[] = network.links.map((link) => ({ ...link }));

		// A node without a corona draws as the default sphere alone, which the
		// library asks for by an accessor answering nothing.
		const corona = (node: EntryNode): Object3D | null => {
			const distance = paint.current.distances.get(node.id);
			if (distance === undefined) {
				return null;
			}
			const tier = coronaTiers[distance];
			const sprite = new Sprite(
				new SpriteMaterial({ map: textures[distance], transparent: true, depthWrite: false }),
			);
			sprite.scale.set(tier.size, tier.size, 1);
			return sprite;
		};

		drawn
			.backgroundColor('#12141a')
			.showNavInfo(false)
			.nodeLabel((node: EntryNode) => `<b>${escapeHTML(node.id)}</b><br>${escapeHTML(node.summary)}`)
			.nodeColor((node: EntryNode) => nodeColour(node, paint.current.selectedID))
			.nodeVal((node: EntryNode) => node.weight)
			.nodeResolution(8)
			.nodeOpacity(0.9)
			.nodeThreeObjectExtend(true)
			.nodeThreeObject(corona as (node: EntryNode) => Object3D)
			.linkColor((link: EntryLink) => linkColours[link.kind] ?? linkColours.related)
			.linkOpacity(0.35)
			.linkWidth((link: EntryLink) =>
				link.kind === closesKind || link.kind === supersedesKind ? 1.2 : 0.6,
			)
			.linkDirectionalArrowLength(2.5)
			.linkDirectionalArrowRelPos(1)
			.linkLabel((link: EntryLink) =>
				escapeHTML(link.desc === '' ? link.kind : `${link.kind}: ${link.desc}`),
			)
			.cooldownTime(settleMilliseconds)
			// A settled layout frames what the reader is looking at: the selected
			// entry where there is one, the whole graph otherwise.
			.onEngineStop(() => {
				if (!focusOn(drawn, paint.current.selectedID, 600)) {
					drawn.zoomToFit(600, 60);
				}
			})
			.onNodeClick((node: EntryNode) => select.current(node.id))
			.onBackgroundClick(() => select.current(null))
			.graphData({ nodes, links });

		drawn.d3Force('charge')?.distanceMax(repulsionReach);

		// The drawing follows its own container rather than the window, because
		// opening and closing the detail panel resizes it while the window
		// itself never changes.
		const resize = new ResizeObserver(() => {
			drawn.width(element.clientWidth).height(element.clientHeight);
		});
		drawn.width(element.clientWidth).height(element.clientHeight);
		resize.observe(element);

		return () => {
			resize.disconnect();
			graph.current = null;
			drawn._destructor();
			element.replaceChildren();
		};
	}, [entries, network, textures]);

	// Selection repaints the drawing rather than rebuilding it, because a
	// rebuild throws away a settled layout and starts the simulation again. The
	// camera then flies to the entry, which a fit over a thousand entries draws
	// too small to find by eye.
	useEffect(() => {
		const drawn = graph.current;
		if (drawn === null) {
			return;
		}
		drawn.refresh();
		focusOn(drawn, selectedID, 900);
	}, [selectedID, distances]);

	useEffect(() => () => textures.forEach((texture) => texture.dispose()), [textures]);

	if (entries.length === 0) {
		return <p className="empty">Open a repository to draw its graph.</p>;
	}

	return (
		<div className="graph-view">
			<div className="graph" ref={container} />
			<dl className="legend">
				<div>
					<dt style={{ background: typeColours.decision }} />
					<dd>Decision</dd>
				</div>
				<div>
					<dt style={{ background: typeColours.signal }} />
					<dd>Signal</dd>
				</div>
				<div>
					<dt style={{ background: retiredColour }} />
					<dd>Retired</dd>
				</div>
				<div>
					<dt className="ring" style={{ borderColor: coronaTiers[0].colour }} />
					<dd>Selected</dd>
				</div>
				<div>
					<dt className="ring" style={{ borderColor: coronaTiers[1].colour }} />
					<dd>One link away</dd>
				</div>
				<div>
					<dt className="ring small" style={{ borderColor: coronaTiers[2].colour }} />
					<dd>Two links away</dd>
				</div>
				{Object.entries(linkColours).map(([kind, colour]) => (
					<div key={kind}>
						<dt className="line" style={{ background: colour }} />
						<dd>{kind}</dd>
					</div>
				))}
			</dl>
			<p className="graph-count">
				{entries.length} entries, {network.links.length} links
			</p>
		</div>
	);
}
