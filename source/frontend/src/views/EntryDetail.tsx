import type { Entry } from '../../bindings/github.com/fza/sddviz/internal/graph/models';
import type { PointerEvent as ReactPointerEvent } from 'react';
import { useCallback } from 'react';
import Markdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { resizeDetail } from '../graph/columns';
import type { EntryNetwork, GraphLink } from '../graph/network';
import { linkColours } from '../palette';
import { StatusBadge } from './StatusBadge';

function LinkList({
	title,
	links,
	end,
	summaries,
	onSelect,
}: {
	title: string;
	links: GraphLink[];
	end: 'source' | 'target';
	summaries: Map<string, string>;
	onSelect: (id: string) => void;
}) {
	if (links.length === 0) {
		return null;
	}

	return (
		<section className="links">
			<h3>
				{title} <span className="count">{links.length}</span>
			</h3>
			<ul>
				{links.map((link, index) => {
					const other = link[end];
					return (
						<li key={`${link.kind}-${other}-${index}`}>
							<button type="button" onClick={() => onSelect(other)}>
								<span className="kind" style={{ color: linkColours[link.kind] ?? linkColours.related }}>
									{link.kind}
								</span>
								<span className="identifier">{other}</span>
								<span className="line">{link.desc === '' ? summaries.get(other) : link.desc}</span>
							</button>
						</li>
					);
				})}
			</ul>
		</section>
	);
}

export function EntryDetail({
	entry,
	network,
	summaries,
	onSelect,
	onClose,
	width,
	onWidth,
}: {
	entry: Entry;
	network: EntryNetwork;
	summaries: Map<string, string>;
	onSelect: (id: string) => void;
	onClose: () => void;
	width: number;
	onWidth: (width: number) => void;
}) {
	const unresolved = network.unresolved.get(entry.id) ?? [];

	// The panel sits at the right edge, so pulling its grip leftwards widens it.
	const startResize = useCallback(
		(event: ReactPointerEvent<HTMLDivElement>) => {
			event.preventDefault();
			const originX = event.clientX;
			const from = width;

			// The window carries the listeners rather than the grip, so a pull
			// that runs past the edge of the panel keeps being followed.
			const move = (moved: PointerEvent) => onWidth(resizeDetail(from + originX - moved.clientX));
			const stop = () => {
				window.removeEventListener('pointermove', move);
				window.removeEventListener('pointerup', stop);
				window.removeEventListener('pointercancel', stop);
			};
			window.addEventListener('pointermove', move);
			window.addEventListener('pointerup', stop);
			window.addEventListener('pointercancel', stop);
		},
		[width, onWidth],
	);

	return (
		<aside className="detail">
			<div
				className="grip"
				role="separator"
				aria-label="Resize the detail panel"
				onPointerDown={startResize}
			/>
			<header>
				<span className="identifier">{entry.id}</span>
				<button type="button" className="close" onClick={onClose} aria-label="Close the detail panel">
					×
				</button>
			</header>

			<div className="facets">
				<span className={`facet type-${entry.type}`}>{entry.type}</span>
				<span className="facet">{entry.layer}</span>
				<span className="facet">{entry.kind}</span>
				{entry.intent !== undefined && entry.intent !== '' && (
					<span className="facet">{entry.intent}</span>
				)}
				<StatusBadge kind={entry.status.kind} />
			</div>

			{entry.status.by !== undefined && entry.status.by !== '' && (
				<p className="succeeded">
					Live head{' '}
					<button type="button" className="identifier" onClick={() => onSelect(entry.status.by ?? '')}>
						{entry.status.by}
					</button>
				</p>
			)}

			<p className="summary">{entry.summary}</p>

			<div className="topics">
				{(entry.topics ?? []).map((topic) => (
					<span key={topic} className={topic.startsWith('area-') ? 'topic area' : 'topic'}>
						{topic}
					</span>
				))}
			</div>

			<p className="attribution">
				{(entry.participants ?? []).join(', ')}
				{entry.confidence !== '' && ` · ${entry.confidence} confidence`}
			</p>

			<LinkList
				title="Points at"
				links={network.outgoing.get(entry.id) ?? []}
				end="target"
				summaries={summaries}
				onSelect={onSelect}
			/>
			<LinkList
				title="Pointed at by"
				links={network.incoming.get(entry.id) ?? []}
				end="source"
				summaries={summaries}
				onSelect={onSelect}
			/>

			{unresolved.length > 0 && (
				<section className="links unresolved">
					<h3>Names an entry this graph does not hold</h3>
					<ul>
						{unresolved.map((id) => (
							<li key={id} className="identifier">
								{id}
							</li>
						))}
					</ul>
				</section>
			)}

			<section className="body">
				<h3>Body</h3>
				{/* The renderer builds elements rather than HTML, so markup written
				    inside an entry is text a reader sees and never markup the page
				    runs. A link opens where it points and nowhere else. */}
				<Markdown
					remarkPlugins={[remarkGfm]}
					components={{
						a: ({ href, children }) => (
							<a href={href} target="_blank" rel="noreferrer noopener">
								{children}
							</a>
						),
					}}
				>
					{entry.body}
				</Markdown>
			</section>
		</aside>
	);
}
