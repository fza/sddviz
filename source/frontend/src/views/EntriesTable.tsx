import { useEffect, useRef } from 'react';
import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';
import { StatusBadge } from './StatusBadge';

export function EntriesTable({
	entries,
	selectedID,
	onSelect,
	onShowInGraph,
}: {
	entries: Entry[];
	selectedID: string | null;
	onSelect: (id: string) => void;
	onShowInGraph: (id: string) => void;
}) {
	const selectedRow = useRef<HTMLTableRowElement | null>(null);

	// A selection made in the graph names a row far outside the scrolled window,
	// so the table follows the selection rather than leaving a reader to find it.
	useEffect(() => {
		selectedRow.current?.scrollIntoView({ block: 'nearest' });
	}, [selectedID]);

	if (entries.length === 0) {
		return <p className="empty">This graph holds no entries.</p>;
	}

	return (
		<div className="entries-scroll">
			<table className="entries">
				<thead>
					<tr>
						<th>Entry</th>
						<th>Layer</th>
						<th>Kind</th>
						<th>Status</th>
						<th>Summary</th>
						<th aria-label="Draw" />
					</tr>
				</thead>
				<tbody>
					{entries.map((entry) => (
						<tr
							key={entry.id}
							ref={entry.id === selectedID ? selectedRow : null}
							className={`${entry.type === 'decision' ? 'decision' : 'signal'}${
								entry.id === selectedID ? ' selected' : ''
							}`}
							onClick={() => onSelect(entry.id)}
						>
							<td className="identifier">{entry.id}</td>
							<td>{entry.layer}</td>
							<td>{entry.kind}</td>
							<td>
								<StatusBadge kind={entry.status.kind} />
							</td>
							<td className="summary">{entry.summary}</td>
							<td className="draw">
								<button
									type="button"
									title="Draw this entry in the graph"
									onClick={(event) => {
										event.stopPropagation();
										onShowInGraph(entry.id);
									}}
								>
									Draw
								</button>
							</td>
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}
