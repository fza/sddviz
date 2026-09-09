import {
	useCallback,
	useEffect,
	useLayoutEffect,
	useRef,
	useState,
} from "react";
import type { Entry } from "../../bindings/github.com/fza/sddviz/internal/graph/models";
import type { Widths } from "../graph/columns";
import { resizeTo, tableFloor, widthOf } from "../graph/columns";
import type { Sort, SortColumn } from "../graph/rows";
import { splitIdentifier } from "../graph/rows";
import { StatusBadge } from "./StatusBadge";

// rowHeight is what every row measures, in pixels. Windowing needs a height it
// can multiply rather than measure, and `.entries tbody tr` holds rows to it.
const rowHeight = 34;

// overscan is how many rows above and below the visible window are drawn, so a
// scroll of a few rows finds them already there.
const overscan = 12;

type Column = {
	key: SortColumn | "ordinal" | "score" | "summary" | "filler" | "draw";
	title: string;
	// A column carrying a width holds it. The filler carries none and takes
	// whatever the rest leave, which is what keeps the button column pinned to
	// the right at its own size.
	sized: boolean;
	// A grip on the column's edge is what a reader pulls to change its width.
	resizable: boolean;
	sortable: boolean;
};

const baseColumns: Column[] = [
	{ key: "ordinal", title: "#", sized: true, resizable: true, sortable: false },
	{
		key: "captured",
		title: "Captured",
		sized: true,
		resizable: true,
		sortable: true,
	},
	{
		key: "short",
		title: "Entry",
		sized: true,
		resizable: true,
		sortable: true,
	},
	{
		key: "layer",
		title: "Layer",
		sized: true,
		resizable: true,
		sortable: true,
	},
	{ key: "kind", title: "Kind", sized: true, resizable: true, sortable: true },
	{
		key: "status",
		title: "Status",
		sized: true,
		resizable: true,
		sortable: true,
	},
	{
		key: "summary",
		title: "Summary",
		sized: true,
		resizable: true,
		sortable: false,
	},
	{ key: "filler", title: "", sized: false, resizable: false, sortable: false },
	{ key: "draw", title: "", sized: true, resizable: false, sortable: false },
];

const scoreColumn: Column = {
	key: "score",
	title: "Score",
	sized: true,
	resizable: true,
	sortable: true,
};

function columnsFor(showScore: boolean): Column[] {
	if (!showScore) {
		return baseColumns;
	}
	// The score joins after the counter and leaves every other width alone, so
	// the columns a reader already knows do not move.
	return [baseColumns[0], scoreColumn, ...baseColumns.slice(1)];
}

function marker(sort: Sort | null, column: Column): string {
	if (sort === null || sort.column !== column.key) {
		return "";
	}
	return sort.ascending ? " ▲" : " ▼";
}

export function EntriesTable({
	entries,
	selectedID,
	onSelect,
	onShowInGraph,
	sort,
	onSort,
	showScore = false,
	widths,
	onWidths,
}: {
	entries: Entry[];
	selectedID: string | null;
	onSelect: (id: string) => void;
	onShowInGraph: (id: string) => void;
	// A search arrives in the order sdd ranked it, and sorting replaces that
	// order. Clearing the sort brings the ranking back.
	sort?: Sort | null;
	onSort?: (column: SortColumn) => void;
	// The score ranks one match against the strongest of the same search, so it
	// belongs to a search and to nothing else.
	showScore?: boolean;
	widths: Widths;
	onWidths: (widths: Widths) => void;
}) {
	const scroller = useRef<HTMLDivElement | null>(null);
	const widthsRef = useRef(widths);
	widthsRef.current = widths;
	const onWidthsRef = useRef(onWidths);
	onWidthsRef.current = onWidths;
	const [window, setWindow] = useState({ top: 0, height: 600 });
	const sortable = onSort !== undefined;
	const columns = columnsFor(showScore);

	// A pull tracks the pointer it captured, so the width follows the pointer
	// even where it leaves the heading it started on.
	const startResize = useCallback(
		(
			event: React.PointerEvent<HTMLDivElement>,
			column: string,
			from: number,
		) => {
			event.preventDefault();
			event.stopPropagation();
			const handle = event.currentTarget;
			const originX = event.clientX;
			handle.setPointerCapture(event.pointerId);

			const move = (moved: PointerEvent) => {
				onWidthsRef.current(
					resizeTo(widthsRef.current, column, from + moved.clientX - originX),
				);
			};
			const stop = () => {
				handle.releasePointerCapture(event.pointerId);
				handle.removeEventListener("pointermove", move);
				handle.removeEventListener("pointerup", stop);
				handle.removeEventListener("pointercancel", stop);
			};
			handle.addEventListener("pointermove", move);
			handle.addEventListener("pointerup", stop);
			handle.addEventListener("pointercancel", stop);
		},
		[],
	);

	const measure = useCallback(() => {
		const element = scroller.current;
		if (element !== null) {
			setWindow({ top: element.scrollTop, height: element.clientHeight });
		}
	}, []);

	useLayoutEffect(() => {
		measure();
		const element = scroller.current;
		if (element === null) {
			return;
		}
		const resize = new ResizeObserver(measure);
		resize.observe(element);
		return () => resize.disconnect();
	}, [measure]);

	// A selection made in the graph names a row the window may not hold, so the
	// table scrolls to its place rather than searching the drawn rows for it.
	useEffect(() => {
		const element = scroller.current;
		if (element === null || selectedID === null) {
			return;
		}
		const index = entries.findIndex((entry) => entry.id === selectedID);
		if (index < 0) {
			return;
		}
		const top = index * rowHeight;
		if (
			top < element.scrollTop ||
			top + rowHeight > element.scrollTop + element.clientHeight
		) {
			element.scrollTop = Math.max(0, top - element.clientHeight / 2);
		}
	}, [selectedID, entries]);

	if (entries.length === 0) {
		return <p className="empty">No entry matches what is asked for.</p>;
	}

	const first = Math.max(0, Math.floor(window.top / rowHeight) - overscan);
	const last = Math.min(
		entries.length,
		Math.ceil((window.top + window.height) / rowHeight) + overscan,
	);
	const drawn = entries.slice(first, last);
	const above = first * rowHeight;
	const below = (entries.length - last) * rowHeight;

	return (
		<div className="entries-scroll" ref={scroller} onScroll={measure}>
			<table
				className="entries"
				style={{
					minWidth: `${tableFloor(
						widths,
						columns.map((column) => column.key),
					)}px`,
				}}
			>
				<colgroup>
					{columns.map((column) => (
						<col
							key={column.key}
							style={
								column.sized
									? { width: `${widthOf(widths, column.key)}px` }
									: undefined
							}
						/>
					))}
				</colgroup>
				<thead>
					<tr>
						{columns.map((column) => (
							<th
								key={column.key}
								// The heading takes a prefixed class so a column named after
								// something the page already styles, `status` above all, cannot
								// pick that styling up and fall out of the table's layout.
								className={`col-${column.key}${sortable && column.sortable ? " sortable" : ""}`}
							>
								{sortable && column.sortable ? (
									<button
										type="button"
										onClick={() => onSort(column.key as SortColumn)}
									>
										{column.title}
										{marker(sort ?? null, column)}
									</button>
								) : (
									column.title
								)}
								{column.resizable && (
									<div
										className="grip"
										role="separator"
										aria-label={`Resize the ${column.title || "last"} column`}
										onPointerDown={(event) =>
											startResize(
												event,
												column.key,
												widthOf(widths, column.key),
											)
										}
									/>
								)}
							</th>
						))}
					</tr>
				</thead>
				<tbody>
					{above > 0 && (
						<tr className="spacer" style={{ height: `${above}px` }} />
					)}
					{drawn.map((entry, offset) => {
						const split = splitIdentifier(entry.id);
						return (
							<tr
								key={entry.id}
								className={`${entry.type === "decision" ? "decision" : "signal"}${
									entry.id === selectedID ? " selected" : ""
								}`}
								onClick={() => onSelect(entry.id)}
							>
								<td className="ordinal">{first + offset + 1}</td>
								{showScore && <td className="score">{entry.score ?? 0}%</td>}
								<td className="captured">{split.captured}</td>
								<td className="identifier">{split.short}</td>
								<td>{entry.layer}</td>
								<td>{entry.kind}</td>
								<td>
									<StatusBadge kind={entry.status.kind} />
								</td>
								<td className="summary" title={entry.summary}>
									{entry.summary}
								</td>
								<td className="filler" />
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
						);
					})}
					{below > 0 && (
						<tr className="spacer" style={{ height: `${below}px` }} />
					)}
				</tbody>
			</table>
		</div>
	);
}
