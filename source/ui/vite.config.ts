import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
	plugins: [react()],
	// Tauri owns the terminal output while the app runs.
	clearScreen: false,
	server: {
		port: 1420,
		strictPort: true,
		watch: {
			ignored: ['**/../backend/**'],
		},
	},
});
