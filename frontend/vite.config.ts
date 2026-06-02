/// <reference types="vitest/config" />
import { fileURLToPath, URL } from 'node:url';
import { resolve } from 'node:path';
import { existsSync } from 'node:fs';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';

// SystemJS plugin HOST (Grafana-style). The console exposes nothing itself; it
// loads each app's runtime plugin module.js at runtime via SystemJS (see
// src/.../console/application/system.ts + runtime.ts). There is NO Module
// Federation and no build-time `shared` map: singleton sharing happens at
// RUNTIME — the host registers its own vue/router/pinia/kit/contract instances
// in a SystemJS import map at bootstrap, and a plugin's externalised imports
// resolve to those host instances. So this build is an ordinary Vite app build.

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
	// The runtime SystemJS host (system.ts) registers the host's own vue-kit /
	// console-plugin namespaces as shared singletons, so any code path that loads
	// it (the app, and tests that import the registry transitively) pulls the
	// FULL sibling-source barrels — including vue-kit chart components' `.svg?url`
	// assets, which live outside this project's root. Allow the sibling kit
	// checkouts (workspace root) so Vite's fs guard transforms those assets rather
	// than denying them. No effect when resolving the published packages.
	server: {
		fs: {
			allow: [resolve(__dirname, '..', '..')],
		},
	},
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
