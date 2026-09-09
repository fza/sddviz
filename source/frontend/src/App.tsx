import type { MouseEvent as ReactMouseEvent } from 'react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { Window } from '@wailsio/runtime';
import { Service } from '../bindings/github.com/fza/sddviz/internal/graph';
import type {
	Entry,
	Repository,
	SearchOutcome,
} from '../bindings/github.com/fza/sddviz/internal/graph/models';
import type { Widths } from './graph/columns';
import { includingDone } from './graph/kinds';
import { defaultDetailWidth } from './graph/columns';
import { buildNetwork } from './graph/network';
import type { Sort, SortColumn } from './graph/rows';
import { nextSort, sortEntries } from './graph/rows';
import { availableTopics, withTopics } from './graph/topics';
import { EntriesTable } from './views/EntriesTable';
import { EntryDetail } from './views/EntryDetail';
import { GraphView } from './views/GraphView';
import { readPlatform, reservesWindowControls } from './platform';
import { SearchBar } from './views/SearchBar';
import { TopicFilter } from './views/TopicFilter';

type View = 'entries' | 'graph';

// Entries opens a repository, because a list of what a graph holds is what a
// reader can read straight away.
const views: { key: View; title: string }[] = [
	{ key: 'entries', title: 'Entries' },
	{ key: 'graph', title: 'Graph' },
];

export function App() {
	const [view, setView] = useState<View>('entries');
	const [repository, setRepository] = useState<Repository | null>(null);
	const [entries, setEntries] = useState<Entry[]>([]);
	const [selectedID, setSelectedID] = useState<string | null>(null);
	const [failure, setFailure] = useState<string | null>(null);
	const [opening, setOpening] = useState(false);
	const [reservesControls, setReservesControls] = useState(false);
	// A search outlives a switch between the views, so it lives here rather than
	// inside the table that draws its result.
	const [outcome, setOutcome] = useState<SearchOutcome | null>(null);
	const [searching, setSearching] = useState(false);
	const [chosenTopics, setChosenTopics] = useState<string[]>([]);
	// A done signal records that work happened, and a reader following what is
	// still open reads the list without them. They are there until asked away.
	const [includeDone, setIncludeDone] = useState(true);
	const [sort, setSort] = useState<Sort | null>(null);
	// Column widths outlive a switch between the views, as a search does.
	const [widths, setWidths] = useState<Widths>({});
	const [detailWidth, setDetailWidth] = useState(defaultDetailWidth);

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

	// Double-clicking a window's title area zooms it, which is what the system
	// does everywhere else. A page served to a browser has no window to zoom,
	// and the call reporting that is the whole answer.
	const zoomWindow = useCallback((event: ReactMouseEvent<HTMLElement>) => {
		if ((event.target as HTMLElement).closest('button') !== null) {
			return;
		}
		Window.ToggleMaximise().catch(() => undefined);
	}, []);

	const forget = useCallback(() => {
		setRepository(null);
		setEntries([]);
		setSelectedID(null);
		setOutcome(null);
		setChosenTopics([]);
		setSort(null);
		setView('entries');
	}, []);

	const openRepository = useCallback(async () => {
		setFailure(null);
		setOpening(true);
		try {
			const path = await Service.ChooseRepositoryDirectory();
			if (path === '') {
				return;
			}
			forget();
			setRepository(await Service.OpenRepository(path));
			setEntries((await Service.Entries()) ?? []);
		} catch (cause: unknown) {
			setFailure(`${cause}`);
		} finally {
			setOpening(false);
		}
	}, [forget]);

	const eject = useCallback(async () => {
		setFailure(null);
		forget();
		try {
			await Service.CloseRepository();
		} catch (cause: unknown) {
			setFailure(`${cause}`);
		}
	}, [forget]);

	const search = useCallback(async (phrase: string) => {
		setFailure(null);
		setSearching(true);
		// A topic chosen against the whole graph rarely fits what a search
		// returns, and one that fits nothing hides every match. A sort left in
		// place would replace the ranking before a reader ever saw it. A search
		// starts from every entry it found, in the order it ranked them.
		setChosenTopics([]);
		setSort(null);
		try {
			setOutcome(await Service.Search(phrase));
		} catch (cause: unknown) {
			setFailure(`${cause}`);
		} finally {
			setSearching(false);
		}
	}, []);

	const network = useMemo(() => buildNetwork(entries), [entries]);
	const byID = useMemo(() => new Map(entries.map((entry) => [entry.id, entry])), [entries]);
	const summaries = useMemo(
		() => new Map(entries.map((entry) => [entry.id, entry.summary])),
		[entries],
	);

	const answered = outcome === null ? entries : (outcome.entries ?? []);
	const matched = useMemo(() => includingDone(answered, includeDone), [answered, includeDone]);
	const filtered = useMemo(() => withTopics(matched, chosenTopics), [matched, chosenTopics]);
	// With nothing sorted the table keeps the order it was given: newest first
	// over every entry, and sdd's ranking over a search.
	const listed = useMemo(() => sortEntries(filtered, sort), [filtered, sort]);
	// The topics on offer follow what is listed, so choosing one never leads to
	// an empty table and a search offers only the topics its matches carry.
	const topics = useMemo(() => availableTopics(matched, chosenTopics), [matched, chosenTopics]);

	const selected = selectedID === null ? undefined : byID.get(selectedID);
	const loaded = repository !== null;

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
		<div
			className={shell.join(' ')}
			style={
				selected === undefined
					? undefined
					: { gridTemplateColumns: `minmax(0, 1fr) ${detailWidth}px` }
			}
		>
			<header className="topbar" onDoubleClick={zoomWindow}>
				<span className="repository">{repository?.name ?? ''}</span>
				{repository !== null && (
					<span className="revision">{repository.revision.slice(0, 12)}</span>
				)}
				<div className="filler" />
				{loaded && (
					<button
						type="button"
						className="eject"
						onClick={eject}
						title="Eject this repository"
						aria-label="Eject this repository"
					>
						{/* The system's own eject mark: a triangle over a bar. */}
						<svg viewBox="0 0 16 16" width="15" height="15" aria-hidden="true" focusable="false">
							<path d="M8 2.6 14.2 9.6H1.8z" fill="currentColor" />
							<rect x="1.8" y="11.4" width="12.4" height="2" rx="0.7" fill="currentColor" />
						</svg>
					</button>
				)}
			</header>

			{/* Nothing is shown where nothing can be chosen: the strip arrives with
			    its views the moment a graph opens. */}
			{loaded && (
				<nav className="tabs">
					<div className="segmented" role="tablist">
						{views.map((candidate) => (
							<button
								key={candidate.key}
								type="button"
								role="tab"
								aria-selected={candidate.key === view}
								className={candidate.key === view ? 'selected' : ''}
								onClick={() => setView(candidate.key)}
							>
								{candidate.title}
							</button>
						))}
					</div>
				</nav>
			)}

			{/* The graph reads as the window's own surface, so it takes the whole
			    area rather than sitting inside its padding. */}
			<main className={loaded && view === 'graph' ? 'bleed' : ''}>
				{failure !== null && <p className="failure">{failure}</p>}

				{!loaded && (
					<div className="welcome">
						<button type="button" className="load" onClick={openRepository} disabled={opening}>
							{opening ? 'Opening' : 'Load repository'}
						</button>
						<p>Open a checkout carrying an .sdd directory to read its graph.</p>
					</div>
				)}

				{loaded && (
					<SearchBar
						phrase={outcome?.query ?? ''}
						searching={searching}
						mode={outcome?.mode ?? ''}
						matchCount={outcome === null ? null : listed.length}
						onSearch={search}
						onClear={() => setOutcome(null)}
						includeDone={includeDone}
						onIncludeDone={setIncludeDone}
						filter={
							<TopicFilter topics={topics} chosen={chosenTopics} onChange={setChosenTopics} />
						}
					/>
				)}

				{loaded && view === 'graph' && (
					<GraphView entries={listed} selectedID={selectedID} onSelect={setSelectedID} />
				)}
				{loaded && view === 'entries' && (
					<EntriesTable
						entries={listed}
						selectedID={selectedID}
						onSelect={setSelectedID}
						onShowInGraph={showInGraph}
						showScore={outcome !== null}
						widths={widths}
						onWidths={setWidths}
						sort={sort}
						onSort={(column: SortColumn) => setSort((held) => nextSort(held, column))}
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
					width={detailWidth}
					onWidth={setDetailWidth}
				/>
			)}
		</div>
	);
}
