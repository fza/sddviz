import { useCallback, useEffect, useMemo, useState } from 'react';
import { Service } from '../bindings/github.com/fza/sddviz/internal/graph';
import type { Entry, Repository } from '../bindings/github.com/fza/sddviz/internal/graph/models';
import { buildNetwork } from './graph/network';
import { readPlatform, reservesWindowControls } from './platform';
import { EntriesTable } from './views/EntriesTable';
import { EntryDetail } from './views/EntryDetail';
import { GraphView } from './views/GraphView';

type View = 'graph' | 'entries';

const viewTitles: Record<View, string> = {
	graph: 'Graph',
	entries: 'Entries',
};

export function App() {
	const [view, setView] = useState<View>('graph');
	const [repository, setRepository] = useState<Repository | null>(null);
	const [entries, setEntries] = useState<Entry[]>([]);
	const [selectedID, setSelectedID] = useState<string | null>(null);
	const [failure, setFailure] = useState<string | null>(null);
	const [opening, setOpening] = useState(false);
	const [reservesControls, setReservesControls] = useState(false);

	useEffect(() => {
		readPlatform().then((report) => setReservesControls(reservesWindowControls(report)));
	}, []);

	// A path on the command line opens a repository before the reader asks,
	// which is the only way in where no directory picker exists.
	useEffect(() => {
		Service.Startup()
			.then(async (state) => {
				if (!state.requested) {
					return;
				}
				if (state.failure !== '') {
					setFailure(state.failure);
					return;
				}
				setRepository(state.repository);
				setEntries((await Service.Entries()) ?? []);
			})
			.catch((cause: unknown) => setFailure(`${cause}`));
	}, []);

	const openRepository = useCallback(async () => {
		setFailure(null);
		setOpening(true);
		try {
			const path = await Service.ChooseRepositoryDirectory();
			if (path === '') {
				return;
			}
			setSelectedID(null);
			setRepository(await Service.OpenRepository(path));
			setEntries((await Service.Entries()) ?? []);
		} catch (cause: unknown) {
			setFailure(`${cause}`);
		} finally {
			setOpening(false);
		}
	}, []);

	const network = useMemo(() => buildNetwork(entries), [entries]);
	const byID = useMemo(() => new Map(entries.map((entry) => [entry.id, entry])), [entries]);
	const summaries = useMemo(
		() => new Map(entries.map((entry) => [entry.id, entry.summary])),
		[entries],
	);

	const selected = selectedID === null ? undefined : byID.get(selectedID);

	const showInGraph = useCallback((id: string) => {
		setSelectedID(id);
		setView('graph');
	}, []);

	const shell = ['app'];
	if (selected !== undefined) {
		shell.push('detailed');
	}
	if (reservesControls) {
		shell.push('reserves-controls');
	}

	return (
		<div className={shell.join(' ')}>
			<header className="topbar">
				<span className="repository">{repository?.name ?? 'No repository open'}</span>
				{repository !== null && <span className="revision">{repository.revision.slice(0, 12)}</span>}
			</header>

			<nav className="sidebar">
				<p className="section">Repository</p>
				<ul>
					{(Object.keys(viewTitles) as View[]).map((candidate) => (
						<li key={candidate}>
							<button
								type="button"
								className={candidate === view ? 'selected' : ''}
								onClick={() => setView(candidate)}
							>
								{viewTitles[candidate]}
							</button>
						</li>
					))}
				</ul>

				<footer>
					<button type="button" onClick={openRepository} disabled={opening}>
						{opening ? 'Opening' : 'Open a repository'}
					</button>
					{repository !== null && (
						<p className="opened">
							<span className="path" title={repository.path}>
								{repository.path}
							</span>
							<span className="count">
								{repository.entryCount} entries, {network.links.length} links
							</span>
						</p>
					)}
				</footer>
			</nav>

			<main>
				<header>
					<h1>{viewTitles[view]}</h1>
				</header>

				{failure !== null && <p className="failure">{failure}</p>}

				{view === 'graph' && (
					<GraphView entries={entries} selectedID={selectedID} onSelect={setSelectedID} />
				)}
				{view === 'entries' && (
					<EntriesTable
						entries={entries}
						selectedID={selectedID}
						onSelect={setSelectedID}
						onShowInGraph={showInGraph}
					/>
				)}
			</main>

			{selected !== undefined && (
				<EntryDetail
					entry={selected}
					network={network}
					summaries={summaries}
					onSelect={setSelectedID}
					onClose={() => setSelectedID(null)}
				/>
			)}
		</div>
	);
}
