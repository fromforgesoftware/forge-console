/// <reference types="vitest/config" />
import { fileURLToPath, URL } from 'node:url';
import { resolve } from 'node:path';
import { existsSync, readFileSync } from 'node:fs';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';
import federation from '@originjs/vite-plugin-federation';

// Module-Federation HOST. The console exposes nothing itself; it loads app
// plugin remotes at runtime (see src/.../console/application/runtime.ts).
// remotes is left empty here because remotes are discovered from /apps and
// registered dynamically — only `shared` matters at build time.
//
// `shared` pins every cross-boundary library as a SINGLETON whose
// requiredVersion is read straight from this host's package.json. That makes a
// remote reuse the host's already-loaded Vue/router/pinia/kits (the Grafana
// import-map equivalent) instead of bundling a second copy — two Vues or two
// pinias would silently break reactivity, the active router, and the store
// graph. Reading the version from package.json guarantees it can never drift
// from what we install.
const pkg = JSON.parse(readFileSync(resolve(__dirname, 'package.json'), 'utf8')) as {
	dependencies: Record<string, string>;
};

const singleton = (name: string) => ({
	[name]: { singleton: true, requiredVersion: pkg.dependencies[name] },
});

const federationShared = {
	...singleton('vue'),
	...singleton('vue-router'),
	...singleton('pinia'),
	...singleton('@fromforgesoftware/ts-kit'),
	// jsonapi-client is the only ts-kit subpath the host actually imports; share
	// it explicitly so a remote importing the same subpath dedupes to the host
	// copy rather than treating the deep path as a distinct module.
	'@fromforgesoftware/ts-kit/jsonapi-client': {
		singleton: true,
		requiredVersion: pkg.dependencies['@fromforgesoftware/ts-kit'],
	},
	...singleton('@fromforgesoftware/vue-kit'),
	...singleton('@fromforgesoftware/forge-console-plugin'),
};

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
	plugins: [
		vue(),
		tailwindcss(),
		federation({
			name: 'forge_host',
			// Remotes are registered at runtime from /apps, so none are listed
			// statically. The empty map still makes this build a federation host
			// and emits the `virtual:__federation__` dynamic-remote runtime the
			// loader uses (__federation_method_setRemote / getRemote).
			remotes: {},
			shared: federationShared,
		}),
	],
	// vite-plugin-federation requires a target that supports top-level await for
	// its shared-module runtime; the default esnext target satisfies this.
	build: { target: 'esnext' },
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
