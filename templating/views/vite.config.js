import { resolve } from "path";
import { defineConfig } from "vite";
export default defineConfig({
	build: {
		root: resolve(__dirname, ""),
		lib: {
			entry: "./transform/main.js",
			name: "PC-UI",
			fileName: "scripts",
			formats: ["es"],
		},
		outDir: resolve(__dirname, "assets/"),
	},
});
