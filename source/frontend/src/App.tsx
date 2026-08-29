import { useCallback, useEffect, useState } from 'react';
import { Service } from '../bindings/github.com/fza/sddviz/internal/graph';
import type { Repository } from '../bindings/github.com/fza/sddviz/internal/graph/models';

export function App() {
	const [backendVersion, setBackendVersion] = useState<string | null>(null);
	const [repository, setRepository] = useState<Repository | null>(null);
	const [failure, setFailure] = useState<string | null>(null);
	const [opening, setOpening] = useState(false);

	useEffect(() => {
		Service.BackendVersion()
			.then(setBackendVersion)
			.catch((cause: unknown) => setFailure(`${cause}`));
	}, []);

	const openRepository = useCallback(async () => {
		setFailure(null);
		setOpening(true);
		try {
			const path = await Service.ChooseRepositoryDirectory();
			if (path === '') {
				return;
			}
			setRepository(await Service.OpenRepository(path));
		} catch (cause: unknown) {
			setFailure(`${cause}`);
		} finally {
			setOpening(false);
		}
	}, []);

	return (
		<main className="app">
			<header>
				<h1>sddviz</h1>
				<p className="tagline">A read-only viewer for SDD decision graphs.</p>
			</header>

			<button type="button" onClick={openRepository} disabled={opening}>
				{opening ? 'Opening' : 'Open a repository'}
			</button>

			{failure !== null && <p className="failure">{failure}</p>}

			{repository !== null && (
				<dl className="repository">
					<dt>Path</dt>
					<dd>{repository.path}</dd>
					<dt>Graph</dt>
					<dd>{repository.graphDir}</dd>
					{repository.repoId !== '' && (
						<>
							<dt>Identity</dt>
							<dd>{repository.repoId}</dd>
						</>
					)}
					<dt>Entries</dt>
					<dd>{repository.entryCount}</dd>
					{repository.unreadableCount > 0 && (
						<>
							<dt>Undecodable</dt>
							<dd>{repository.unreadableCount}</dd>
						</>
					)}
				</dl>
			)}

			<p className="status">
				{backendVersion !== null ? `Backend ${backendVersion}` : 'Connecting to the backend'}
			</p>
		</main>
	);
}
