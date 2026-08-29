import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';

const statusLabels: Record<string, string> = {
	open: 'Open',
	active: 'Active',
	settled: 'Settled',
	'closed-by': 'Closed',
	'superseded-by': 'Superseded',
	'cascade-closed': 'Cascade closed',
	'cascade-orphan': 'Cascade orphan',
};

function StatusBadge({ kind }: { kind: string }) {
	if (kind === '') {
		return <span className="status status-none">None</span>;
	}
	return <span className={`status status-${kind}`}>{statusLabels[kind] ?? kind}</span>;
}

export function EntriesTable({ entries }: { entries: Entry[] }) {
	if (entries.length === 0) {
		return <p className="empty">This graph holds no entries.</p>;
	}

	return (
		<table className="entries">
			<thead>
				<tr>
					<th>Entry</th>
					<th>Layer</th>
					<th>Kind</th>
					<th>Status</th>
					<th>Summary</th>
				</tr>
			</thead>
			<tbody>
				{entries.map((entry) => (
					<tr key={entry.id} className={entry.type === 'decision' ? 'decision' : 'signal'}>
						<td className="identifier">{entry.id}</td>
						<td>{entry.layer}</td>
						<td>{entry.kind}</td>
						<td>
							<StatusBadge kind={entry.status.kind} />
						</td>
						<td className="summary">{entry.summary}</td>
					</tr>
				))}
			</tbody>
		</table>
	);
}
