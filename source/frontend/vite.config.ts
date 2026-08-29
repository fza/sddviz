import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// The Go module embeds the built assets from its own tree, which is what
// `go:embed` requires.
export default defineConfig({
	plugins: [react()],
	build: {
		outDir: 'dist',
		emptyOutDir: true,
	},
	clearScreen: false,
	server: {
		host: '127.0.0.1',
		port: Number(process.env.WAILS_VITE_PORT) || 9245,
		strictPort: true,
	},
});
