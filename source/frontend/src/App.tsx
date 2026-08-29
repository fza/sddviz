import { useCallback, useEffect, useState } from 'react';
import { Service } from '../bindings/github.com/fza/sddviz/internal/graph';
import type { Entry, Repository } from '../bindings/github.com/fza/sddviz/internal/graph/models';
import { EntriesTable } from './views/EntriesTable';
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
	const [failure, setFailure] = useState<string | null>(null);
	const [opening, setOpening] = useState(false);

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
			setRepository(await Service.OpenRepository(path));
			setEntries((await Service.Entries()) ?? []);
		} catch (cause: unknown) {
			setFailure(`${cause}`);
		} finally {
			setOpening(false);
		}
	}, []);

	return (
		<div className="app">
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
							<span className="path">{repository.path}</span>
							<span className="count">{repository.entryCount} entries</span>
						</p>
					)}
				</footer>
			</nav>

			<main>
				<header>
					<h1>{viewTitles[view]}</h1>
				</header>

				{failure !== null && <p className="failure">{failure}</p>}

				{view === 'graph' && <GraphView entry={entries[0] ?? null} />}
				{view === 'entries' && <EntriesTable entries={entries} />}
			</main>
		</div>
	);
}
