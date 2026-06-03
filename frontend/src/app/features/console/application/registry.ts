import type { RouteRecordRaw, Router } from 'vue-router';
import type { ConsolePluginModule, ForgeConsolePlugin } from '@fromforgesoftware/forge-console-plugin';
import { gleipnirPlugin } from '../plugins/gleipnir';
import { talosPlugin } from '../plugins/talos';
import { gjallarhornPlugin } from '../plugins/gjallarhorn';
import type { AppInfo } from '@/app/features/console/stores/apps';
import { apiBaseFor } from '@/app/core/http/services';
import { importModule, hasRemote } from './runtime';

// Compile-time plugin registry: every first-party console plugin bundled into
// the host. In the hybrid model (4c) these are the FALLBACK — an app uses its
// bundled plugin until it ships a runtime SystemJS plugin module (moduleUri on
// its /apps entry), at which point the runtime plugin takes over. Adding a
// bundled service = dropping a manifest file + one line here; the manifest
// carries its own icon/order, so sidebar, dashboard, and palette need no
// per-service edits.
const ORDER_LAST = Number.MAX_SAFE_INTEGER;
const byOrder = (a: ForgeConsolePlugin, b: ForgeConsolePlugin) =>
	(a.order ?? ORDER_LAST) - (b.order ?? ORDER_LAST);

export const plugins: ForgeConsolePlugin[] = [
	talosPlugin(),
	gjallarhornPlugin(),
	gleipnirPlugin(),
].sort(byOrder);

// bundledFor looks up the compile-time plugin for a slug, if one exists.
function bundledFor(slug: string): ForgeConsolePlugin | undefined {
	return plugins.find((p) => p.serviceId === slug);
}

// enabledPlugins filters the bundled registry to the apps the backend reports
// as installed/accessible (useAppsStore().slugs), preserving sort order. Used
// by the synchronous nav fallback and tests.
export function enabledPlugins(appSlugs: string[]): ForgeConsolePlugin[] {
	return plugins.filter((p) => appSlugs.includes(p.serviceId));
}

// resolvePlugin unwraps a loaded SystemJS module namespace to a
// ForgeConsolePlugin, accepting a direct plugin object or a zero-arg factory,
// via `default` or `plugin` (mirrors the package loader's contract).
function resolvePlugin(mod: ConsolePluginModule, apiBase: string): ForgeConsolePlugin | undefined {
	const exported = mod.default ?? mod.plugin;
	// A remote module.js is built before its apiBase is known, so a factory
	// receives it here and threads it into its page props (the plugin's top-level
	// apiBase is set again by the caller as a backstop).
	if (typeof exported === 'function')
		return (exported as (ctx: { apiBase: string }) => ForgeConsolePlugin)({ apiBase });
	if (exported && typeof exported === 'object') return exported;
	return undefined;
}

// resolvePlugins computes the ACTIVE plugin set from the live /apps list,
// hybrid + failure-isolated:
//   - app has a moduleUri AND its module loads      -> use the runtime plugin
//   - otherwise (no module, or it errored)          -> fall back to bundled
//   - neither a runtime module nor a bundled plugin -> skip the app
// This lets apps migrate to SystemJS plugin modules one at a time while
// un-migrated ones keep working unchanged. A 404/throwing module NEVER crashes
// the console — it just degrades that one app to its bundled plugin (or drops
// it if none exists).
export async function resolvePlugins(apps: AppInfo[]): Promise<ForgeConsolePlugin[]> {
	const resolved: ForgeConsolePlugin[] = [];
	for (const app of apps) {
		const bundled = bundledFor(app.slug);
		if (hasRemote(app)) {
			try {
				const mod = await importModule(app.moduleUri);
				const plugin = resolvePlugin(mod, apiBaseFor(app.slug));
				if (!plugin) {
					throw new Error(`plugin module "${app.slug}" exposes no ForgeConsolePlugin`);
				}
				resolved.push({
					...plugin,
					serviceId: plugin.serviceId || app.slug,
					apiBase: plugin.apiBase || apiBaseFor(app.slug),
					type: plugin.type ?? 'app',
				});
				continue;
			} catch (err) {
				console.error(
					`forge-console: plugin module "${app.slug}" failed to load from ${app.moduleUri}; ` +
						`${bundled ? 'falling back to bundled plugin' : 'no bundled fallback, skipping'}`,
					err,
				);
			}
		}
		if (bundled) resolved.push(bundled);
	}
	return resolved.sort(byOrder);
}

// routesForPlugin flattens one plugin's pages into authenticated routes mounted
// under its basePath. `relative` drops the leading slash for routes added as
// children of the `/` app-shell (router.addRoute / CONSOLE_ROUTES); absolute
// paths are used by pluginRoutes' standalone callers/tests.
function routesForPlugin(plugin: ForgeConsolePlugin, relative = false): RouteRecordRaw[] {
	return plugin.pages.map((page) => {
		const path = `${plugin.basePath}/${page.path}`;
		return {
			path: relative ? path.replace(/^\//, '') : path,
			name: `${plugin.serviceId}:${page.path}`,
			component: page.component,
			props: page.props,
			meta: { requiresAuth: true, plugin: plugin.serviceId },
		};
	});
}

// pluginRoutes flattens each plugin's pages into authenticated routes mounted
// under the plugin basePath (absolute paths).
export function pluginRoutes(list: ForgeConsolePlugin[] = plugins): RouteRecordRaw[] {
	return list.flatMap((plugin) => routesForPlugin(plugin));
}

// registerPluginRoutes adds a (typically runtime-resolved) plugin's pages to
// the live router as children of the `app-shell` route, the Grafana-style lazy
// step run on first navigation into the plugin's section. Idempotent: a page
// whose route name already exists (e.g. a bundled plugin already mounted at
// startup) is skipped, so re-resolving never double-registers. Returns true if
// any new route was added (the caller re-resolves navigation to hit it).
export function registerPluginRoutes(router: Router, plugin: ForgeConsolePlugin): boolean {
	let added = false;
	for (const route of routesForPlugin(plugin, true)) {
		if (route.name && router.hasRoute(route.name)) continue;
		router.addRoute('app-shell', route);
		added = true;
	}
	return added;
}
