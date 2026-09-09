import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';

export function SearchBar({
	phrase,
	searching,
	mode,
	matchCount,
	onSearch,
	onClear,
	includeDone,
	onIncludeDone,
	filter,
}: {
	phrase: string;
	searching: boolean;
	mode: string;
	matchCount: number | null;
	onSearch: (phrase: string) => void;
	onClear: () => void;
	includeDone: boolean;
	onIncludeDone: (include: boolean) => void;
	filter?: ReactNode;
}) {
	const [typed, setTyped] = useState(phrase);

	// A search made elsewhere, or cleared elsewhere, shows in the field.
	useEffect(() => setTyped(phrase), [phrase]);

	return (
		<div className="searchbar">
			<div className="field">
				<input
					type="search"
					value={typed}
					placeholder="Search this graph"
					aria-label="Search this graph"
					onChange={(event) => setTyped(event.target.value)}
					onKeyDown={(event) => {
						if (event.key === 'Enter' && typed.trim() !== '') {
							onSearch(typed);
						}
					}}
				/>
				{phrase !== '' && (
					<button type="button" className="clear" aria-label="Show every entry again" onClick={onClear}>
						×
					</button>
				)}
			</div>

			{searching && <span className="spinner" role="status" aria-label="Searching" />}

			{!searching && matchCount !== null && (
				<span className="matches">
					{matchCount} matching, {mode === 'vector' ? 'by meaning' : 'by words'}
				</span>
			)}

			<div className="filler" />

			<label className="donefilter">
				<input
					type="checkbox"
					checked={includeDone}
					onChange={(event) => onIncludeDone(event.target.checked)}
				/>
				Done
			</label>

			{filter}
		</div>
	);
}
