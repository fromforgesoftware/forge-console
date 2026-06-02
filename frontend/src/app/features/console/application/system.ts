// SystemJS host setup — the Grafana-style replacement for Module Federation.
//
// App plugins are built as a SINGLE SystemJS `module.js` that EXTERNALISES the
// shared singletons (vue, vue-router, pinia, the forge kits, this contract) —
// see @fromforgesoftware/forge-console-plugin/build (consolePluginModule). They
// do NOT bundle those libs; they `import` the bare specifiers at runtime. For
// that to resolve to the HOST's already-loaded instances (one Vue reactivity
// system, one pinia root, one router, one set of kit clients) the host must
// register its own copies into SystemJS BEFORE any plugin module is imported.
//
// We do this with the standard SystemJS seam used by the package's interop
// proof: an import map mapping each bare specifier to a host-owned id, plus
// `System.set(id, namespace)` seeding that id with the host module's live
// namespace. A plugin's `import 'vue'` then resolves through the map to our id
// and gets the SAME object — real sharing, not a second copy.
//
// `systemjs/dist/system.js` is a side-effecting script that installs the global
// `System` loader (and `System.addImportMap` / `System.set` / `System.import`).
// Importing it for effect gives us one loader instance the whole app shares.
import 'systemjs/dist/system.js';

// The host's OWN module namespaces for every shared specifier. These are the
// exact modules the host bundle already imports, so registering them hands
// plugins the host's live instances. `import * as` captures the full namespace
// (named + default) so a plugin reading either shape resolves correctly.
import * as vue from 'vue';
import * as vueRouter from 'vue-router';
import * as pinia from 'pinia';
import * as tsKit from '@fromforgesoftware/ts-kit';
import * as tsKitJsonApiClient from '@fromforgesoftware/ts-kit/jsonapi-client';
import * as vueKit from '@fromforgesoftware/vue-kit';
import * as consolePlugin from '@fromforgesoftware/forge-console-plugin';
import * as consolePluginUi from '@fromforgesoftware/forge-console-plugin/ui';

// The configured loader. The browser build installs it on the global; in the
// jsdom test env the same script runs and installs it there too. We read it
// from globalThis so there is exactly ONE loader for bootstrap and for every
// runtime `System.import()`.
const System = (globalThis as unknown as { System: SystemLoader }).System;

interface SystemLoader {
	import(id: string): Promise<unknown>;
	set(id: string, namespace: object): void;
	addImportMap(map: { imports?: Record<string, string>; scopes?: Record<string, Record<string, string>> }): void;
}

// SHARED_SINGLETONS pairs each bare specifier a plugin may import with the host
// module namespace that must back it. The keys mirror
// CONSOLE_SHARED_SINGLETONS (+ the subpaths the host actually uses) from
// @fromforgesoftware/forge-console-plugin/build, so the host import map is the
// exact dual of what the plugin build externalises.
const SHARED_SINGLETONS: Record<string, object> = {
	vue,
	'vue-router': vueRouter,
	pinia,
	'@fromforgesoftware/ts-kit': tsKit,
	'@fromforgesoftware/ts-kit/jsonapi-client': tsKitJsonApiClient,
	'@fromforgesoftware/vue-kit': vueKit,
	'@fromforgesoftware/forge-console-plugin': consolePlugin,
	'@fromforgesoftware/forge-console-plugin/ui': consolePluginUi,
};

// SystemJS rejects bare specifiers as registry ids (they must be URL-like), so
// each shared specifier is mapped to a `forge-singleton:<spec>` id that we then
// `System.set`. This is exactly the import-map + set pattern the package's
// interop test proves (it maps `vue` -> `forge:vue` and sets that id).
const SINGLETON_ID = (spec: string) => `forge-singleton:${spec}`;

let registered = false;

// registerSharedSingletons wires the host's singleton instances into SystemJS.
// Idempotent — bootstrap calls it once before the first plugin loads; a second
// call (e.g. in tests) is a no-op. After this, `System.import(moduleUri)` on any
// plugin resolves its externalised imports to these host instances.
export function registerSharedSingletons(): void {
	if (registered) return;
	const imports: Record<string, string> = {};
	for (const [spec, namespace] of Object.entries(SHARED_SINGLETONS)) {
		const id = SINGLETON_ID(spec);
		imports[spec] = id;
		System.set(id, namespace);
	}
	System.addImportMap({ imports });
	registered = true;
}

// systemImport is the loader seam the runtime registry uses: resolve a plugin
// module.js URL to its loaded SystemJS namespace via the host's configured
// loader (the one whose import map registers the shared singletons above). This
// is the SystemImporter the package's loader.ts expects: (uri) => System.import(uri).
export function systemImport(uri: string): Promise<unknown> {
	return System.import(uri);
}
