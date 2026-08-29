import { useEffect, useState } from 'react';
import { Service } from './bindings/github.com/fza/sddviz/internal/graph';

export function App() {
	const [backendVersion, setBackendVersion] = useState<string | null>(null);
	const [backendError, setBackendError] = useState<string | null>(null);

	useEffect(() => {
		Service.BackendVersion()
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
