import { describe, expect, it } from 'vitest';
import type { PlatformReport } from './platform';
import { reservesWindowControls } from './platform';

function report(fields: Partial<PlatformReport> = {}): PlatformReport {
	return { mac: true, desktop: true, mode: '', ...fields };
}

describe('reservesWindowControls', () => {
	it('reserves the corner in a native macOS window', () => {
		expect(reservesWindowControls(report())).toBe(true);
	});

	it('reserves nothing for a page a browser loaded from the server build', () => {
		expect(reservesWindowControls(report({ desktop: false, mode: 'server' }))).toBe(false);
	});

	it('reserves nothing where the runtime reports the server mode alone', () => {
		expect(reservesWindowControls(report({ mode: 'server' }))).toBe(false);
	});

	it('reserves nothing away from macOS, whose window alone hides its title bar', () => {
		expect(reservesWindowControls(report({ mac: false }))).toBe(false);
	});

	it('reserves nothing where the runtime answered nothing at all', () => {
		expect(reservesWindowControls({ mac: false, desktop: false, mode: '' })).toBe(false);
	});
});
