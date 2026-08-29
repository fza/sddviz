import { useEffect, useState } from 'react';
import { invoke } from '@tauri-apps/api/core';

export function App() {
	const [backendVersion, setBackendVersion] = useState<string | null>(null);
	const [backendError, setBackendError] = useState<string | null>(null);

	useEffect(() => {
		invoke<string>('backend_version')
			.then(setBackendVersion)
			.catch((cause: unknown) => setBackendError(`${cause}`));
	}, []);

	return (
		<main className="app">
			<h1>sddviz</h1>
			<p className="tagline">A read-only viewer for SDD decision graphs.</p>
			<p className="status">
				{backendError !== null
					? `Backend unreachable: ${backendError}`
					: backendVersion !== null
						? `Backend ${backendVersion}`
						: 'Connecting to the backend'}
			</p>
		</main>
	);
}
