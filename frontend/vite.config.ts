/// <reference types="vitest/config" />
import { fileURLToPath, URL } from 'node:url';
import { resolve } from 'node:path';
import { existsSync } from 'node:fs';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';

const tsKit = (sub: string) => resolve(__dirname, `../../ts-kit/src/${sub}`);
const consolePlugin = (sub: string) =>
	resolve(__dirname, `../../forge-console-plugin/src/${sub}`);

// Local dev links the sibling ts-kit/vue-kit/forge-console-plugin source
// checkouts so changes there are picked up live. CI (and any consumer without
// those checkouts) instead resolves the published @fromforgesoftware/* packages
// pinned in package.json — the same versions production ships. Auto-detect via
// the presence of the sibling source so both paths work with no extra config.
const useKitSource =
	process.env.FORGE_USE_PUBLISHED_KIT !== '1' &&
	existsSync(resolve(__dirname, '../../ts-kit/src/index.ts'));

const kitAliases = useKitSource
	? {
			'@fromforgesoftware/ts-kit/jsonapi-client': tsKit('jsonapi-client/index.ts'),
			'@fromforgesoftware/ts-kit/jsonapi': tsKit('jsonapi/index.ts'),
			'@fromforgesoftware/ts-kit/resource-state': tsKit('resource-state/index.ts'),
			'@fromforgesoftware/ts-kit/http': tsKit('http/index.ts'),
			'@fromforgesoftware/ts-kit/errors': tsKit('errors/index.ts'),
			'@fromforgesoftware/ts-kit/date': tsKit('date/index.ts'),
			'@fromforgesoftware/ts-kit': tsKit('index.ts'),
			'@fromforgesoftware/vue-kit': resolve(__dirname, '../../vue-kit/src/index.ts'),
		}
	: {};

// The forge-console-plugin contract + generic renderers follow the same
// sibling-source-or-published rule, gated on its own source presence so the
// host still builds when only the kits are checked out.
const useConsolePluginSource =
	process.env.FORGE_USE_PUBLISHED_KIT !== '1' &&
	existsSync(resolve(__dirname, '../../forge-console-plugin/src/index.ts'));

const consolePluginAliases = useConsolePluginSource
	? {
			'@fromforgesoftware/forge-console-plugin/ui': consolePlugin('ui/index.ts'),
			'@fromforgesoftware/forge-console-plugin': consolePlugin('index.ts'),
		}
	: {};

export default defineConfig({
	plugins: [vue(), tailwindcss()],
	resolve: {
		alias: {
			'@': fileURLToPath(new URL('./src', import.meta.url)),
			...kitAliases,
				...consolePluginAliases,
		},
	},
	test: {
		name: 'unit',
		include: ['src/**/*.test.ts'],
		environment: 'jsdom',
		globals: true,
		server: {
			// When resolving against the published packages (CI / no sibling
			// source), inline the kits so Vite's resolver transforms them rather
			// than handing them to Node's stricter native-ESM loader.
			deps: {
				inline: [/@fromforgesoftware\/(ts|vue)-kit/, /@fromforgesoftware\/forge-console-plugin/],
			},
		},
	},
});
