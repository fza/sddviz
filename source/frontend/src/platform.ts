import { System } from '@wailsio/runtime';

// A native macOS window paints the system's own close, minimise and zoom
// controls over the top-left corner of the page, because the window hides its
// title bar and lets the page draw to the top edge. A page served to a browser
// carries nothing there. Only the first has a corner to keep clear.
export interface PlatformReport {
	mac: boolean;
	desktop: boolean;
	mode: string;
}

export function reservesWindowControls(report: PlatformReport): boolean {
	return report.mac && report.desktop && report.mode !== 'server';
}

export async function readPlatform(): Promise<PlatformReport> {
	try {
		const environment = await System.Environment();
		const info: Record<string, unknown> = environment.PlatformInfo;
		const mode = info.mode;
		return {
			mac: System.IsMac(),
			desktop: System.IsDesktop(),
			mode: typeof mode === 'string' ? mode : '',
		};
	} catch {
		// A page that cannot reach the runtime has no window controls over it.
		return { mac: false, desktop: false, mode: '' };
	}
}
