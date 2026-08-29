import { useEffect, useRef } from 'react';
import ForceGraph3D from '3d-force-graph';
import type { ForceGraph3DInstance, NodeObject } from '3d-force-graph';
import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';

const signalColour = '#e8b246';
const decisionColour = '#6ec4e8';

// EntryNode is what the graph draws per entry. The library types a node as an
// open object, so the fields the accessors read are declared here.
type EntryNode = NodeObject & {
	id: string;
	type: string;
	summary: string;
};

// One entry alone, which is what shows the canvas, the camera and the picking
// all work before a whole graph is drawn into it.
export function GraphView({ entry }: { entry: Entry | null }) {
	const container = useRef<HTMLDivElement | null>(null);

	useEffect(() => {
		const element = container.current;
		if (element === null || entry === null) {
			return;
		}

		// The library declares its default export as a non-generic constructor,
		// so the node type reaches the accessors through the instance type it
		// exports rather than through a type argument.
		const graph = new ForceGraph3D(element) as unknown as ForceGraph3DInstance<EntryNode>;
		graph
			.backgroundColor('#12141a')
			.nodeLabel((node: EntryNode) => node.summary)
			.nodeColor((node: EntryNode) => (node.type === 'decision' ? decisionColour : signalColour))
			.nodeVal(8)
			.graphData({
				nodes: [{ id: entry.id, type: entry.type, summary: entry.summary }],
				links: [],
			});

		const resize = () => {
			graph.width(element.clientWidth).height(element.clientHeight);
		};
		resize();
		window.addEventListener('resize', resize);

		return () => {
			window.removeEventListener('resize', resize);
			graph._destructor();
			element.replaceChildren();
		};
	}, [entry]);

	if (entry === null) {
		return <p className="empty">Open a repository to draw its graph.</p>;
	}

	return <div className="graph" ref={container} />;
}
