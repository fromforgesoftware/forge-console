import type { RouteRecordRaw } from 'vue-router';
import type { ForgeConsolePlugin } from '../domain/plugin';
import { aegisPlugin } from '../plugins/aegis';
import { gleipnirPlugin } from '../plugins/gleipnir';
import { talosPlugin } from '../plugins/talos';
import { gjallarhornPlugin } from '../plugins/gjallarhorn';

// Compile-time plugin registry: every first-party console plugin is listed
// here (no runtime federation in v1). Adding a service = dropping a manifest
// file + one line here; the manifest carries its own icon/order, so the
// sidebar, dashboard, and command palette need no per-service edits.
export const plugins: ForgeConsolePlugin[] = [
	aegisPlugin(),
	talosPlugin(),
	gjallarhornPlugin(),
	gleipnirPlugin(),
].sort((a, b) => (a.order ?? Number.MAX_SAFE_INTEGER) - (b.order ?? Number.MAX_SAFE_INTEGER));

// enabledPlugins filters the registry to the apps the backend reports as
// installed/accessible (useAppsStore().slugs), preserving sort order.
export function enabledPlugins(appSlugs: string[]): ForgeConsolePlugin[] {
	return plugins.filter((p) => appSlugs.includes(p.serviceId));
}

// pluginRoutes flattens each plugin's pages into authenticated routes mounted
// under the plugin basePath.
export function pluginRoutes(list: ForgeConsolePlugin[] = plugins): RouteRecordRaw[] {
	return list.flatMap((plugin) =>
		plugin.pages.map((page) => ({
			path: `${plugin.basePath}/${page.path}`,
			name: `${plugin.serviceId}:${page.path}`,
			component: page.component,
			props: page.props,
			meta: { requiresAuth: true, plugin: plugin.serviceId },
		})),
	);
}
