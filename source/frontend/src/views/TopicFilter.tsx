import { useEffect, useRef, useState } from 'react';
import { isArea } from '../graph/topics';

export function TopicFilter({
	topics,
	chosen,
	onChange,
}: {
	topics: string[];
	chosen: string[];
	onChange: (chosen: string[]) => void;
}) {
	const [open, setOpen] = useState(false);
	const holder = useRef<HTMLDivElement | null>(null);

	// Selecting anywhere else closes the list, which is what a reader expects of
	// something that covers the view beneath it.
	useEffect(() => {
		if (!open) {
			return;
		}
		const close = (event: MouseEvent) => {
			if (holder.current !== null && !holder.current.contains(event.target as Node)) {
				setOpen(false);
			}
		};
		document.addEventListener('mousedown', close);
		return () => document.removeEventListener('mousedown', close);
	}, [open]);

	const toggle = (topic: string) => {
		onChange(chosen.includes(topic) ? chosen.filter((held) => held !== topic) : [...chosen, topic]);
	};

	return (
		<div className="topicfilter" ref={holder}>
			<button type="button" className={chosen.length > 0 ? 'chosen' : ''} onClick={() => setOpen(!open)}>
				{/* Three bars narrowing to a point, the mark a filter carries. */}
				<svg viewBox="0 0 16 16" width="13" height="13" aria-hidden="true" focusable="false">
					<rect x="1.5" y="3" width="13" height="1.9" rx="0.95" fill="currentColor" />
					<rect x="3.8" y="7.05" width="8.4" height="1.9" rx="0.95" fill="currentColor" />
					<rect x="6.1" y="11.1" width="3.8" height="1.9" rx="0.95" fill="currentColor" />
				</svg>
				{chosen.length === 0 ? 'Topics' : `Topics · ${chosen.length}`}
			</button>

			{open && (
				<div className="list">
					{chosen.length > 0 && (
						<button type="button" className="reset" onClick={() => onChange([])}>
							Clear all
						</button>
					)}
					{topics.length === 0 && <p className="empty">This graph carries no topics.</p>}
					{topics.map((topic) => (
						<label key={topic} className={isArea(topic) ? 'area' : ''}>
							<input
								type="checkbox"
								checked={chosen.includes(topic)}
								onChange={() => toggle(topic)}
							/>
							{topic}
						</label>
					))}
				</div>
			)}
		</div>
	);
}
