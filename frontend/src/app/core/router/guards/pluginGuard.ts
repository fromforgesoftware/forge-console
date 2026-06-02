import type { NavigationGuard, Router } from 'vue-router';
import { useAppsStore } from '@/app/features/console/stores/apps';
import { usePluginsStore } from '@/app/features/console/stores/plugins';
import { registerPluginRoutes } from '@/app/features/console/application/registry';

// makePluginGuard builds the Grafana-style LAZY route loader, bound to the live
// router (passed in to avoid a circular import with the router module). The nav
// renders every enabled app from /apps metadata immediately, but a remote
// plugin's code and routes are only fetched on first navigation into its
// `/<slug>` section.
//
// Bundled plugins already have their routes in CONSOLE_ROUTES at startup, so
// they match synchronously and this guard is a no-op for them (no regression).
// A remote-backed app has no static route, so the first hit to `/<slug>/...`
// would 404 — here we resolve that one app's plugin, register its routes under
// `app-shell`, then redirect to the same path so the freshly added route
// matches. Failure-isolated: a remote that errors simply leaves the section
// unregistered and falls through to the not-found route, never crashing.
export function makePluginGuard(router: Router): NavigationGuard {
	return async (to) => {
		const slug = sectionSlug(to.path);
		if (!slug) return true;

		// Already a matched, named plugin route -> nothing to lazily load.
		if (typeof to.name === 'string' && to.name.startsWith(`${slug}:`)) return true;

		const app = useAppsStore().apps.find((a) => a.slug === slug);
		// Not a known app (or apps not loaded yet) -> let normal matching decide.
		if (!app) return true;

		const plugin = await usePluginsStore().ensureSection(slug);
		if (!plugin) return true;

		const added = registerPluginRoutes(router, plugin);
		// Re-run navigation so the newly added route can match this same URL.
		if (added) return to.fullPath;
		return true;
	};
}

// sectionSlug extracts the leading `/<slug>` segment of a path, or undefined
// for the dashboard / non-section routes.
function sectionSlug(path: string): string | undefined {
	const seg = path.replace(/^\//, '').split('/')[0];
	return seg || undefined;
}
