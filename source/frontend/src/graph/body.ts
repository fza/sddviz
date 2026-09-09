// An entry body is Markdown, hard wrapped by the tool that wrote it. Those wraps
// were chosen for a file at a fixed width and fight every panel of a different
// one, so a block of running prose is rejoined into one paragraph while a block
// carrying structure keeps its lines exactly as written.

const structuralLine = /^\s*(?:[-*+>|]|#{1,6}\s|\d+[.)]\s|```|\t| {4})/;

export interface BodyBlock {
	text: string;
	structured: boolean;
}

export function bodyBlocks(body: string): BodyBlock[] {
	const blocks: BodyBlock[] = [];

	for (const block of body.split(/\n\s*\n/)) {
		const lines = block.split('\n').filter((line) => line.trim() !== '');
		if (lines.length === 0) {
			continue;
		}
		const structured = lines.some((line) => structuralLine.test(line));
		blocks.push({
			text: structured ? lines.join('\n') : lines.map((line) => line.trim()).join(' '),
			structured,
		});
	}

	return blocks;
}
