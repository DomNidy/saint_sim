import { defineConfig } from "@hey-api/openapi-ts";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const configDir = dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	input: resolve(configDir, "../api/openapi.yaml"),
	output: resolve(configDir, "src/lib/saint-api/generated"),
	plugins: [
		{
			name: "@hey-api/typescript",
			enums: "javascript",
			comments: true,
		},
		"@hey-api/client-fetch",
		"@hey-api/sdk",
		{
			name: "zod",
			definitions: true,
			comments: true,
		},
	],
});
