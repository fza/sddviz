import { describe, expect, it } from 'vitest';
import { bodyBlocks } from './body';

describe('bodyBlocks', () => {
	it('rejoins the wrapped lines of running prose into one paragraph', () => {
		const blocks = bodyBlocks('A rule the file states\nacross two wrapped lines.');

		expect(blocks).toEqual([{ text: 'A rule the file states across two wrapped lines.', structured: false }]);
	});

	it('separates paragraphs on a blank line', () => {
		const blocks = bodyBlocks('First claim.\n\nSecond claim.');

		expect(blocks.map((block) => block.text)).toEqual(['First claim.', 'Second claim.']);
	});

	it('keeps the lines of a list exactly as written', () => {
		const blocks = bodyBlocks('- one thing\n- another thing');

		expect(blocks).toEqual([{ text: '- one thing\n- another thing', structured: true }]);
	});

	it('keeps a heading on its own line', () => {
		const blocks = bodyBlocks('## What holds\n\nRunning prose after it.');

		expect(blocks[0]).toEqual({ text: '## What holds', structured: true });
		expect(blocks[1]).toEqual({ text: 'Running prose after it.', structured: false });
	});

	it('keeps a fenced block as written', () => {
		const blocks = bodyBlocks('```bash\ndevbox run test\n```');

		expect(blocks[0].structured).toBe(true);
		expect(blocks[0].text).toBe('```bash\ndevbox run test\n```');
	});

	it('reports nothing for an empty body', () => {
		expect(bodyBlocks('')).toEqual([]);
		expect(bodyBlocks('\n\n  \n')).toEqual([]);
	});
});
