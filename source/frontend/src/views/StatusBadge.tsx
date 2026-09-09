import { statusLabels } from '../palette';

export function StatusBadge({ kind }: { kind: string }) {
	if (kind === '') {
		return <span className="status status-none">None</span>;
	}
	return <span className={`status status-${kind}`}>{statusLabels[kind] ?? kind}</span>;
}
