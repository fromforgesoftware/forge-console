import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { ForgeConsolePlugin } from '@fromforgesoftware/forge-console-plugin';
import { resolvePlugins } from '@/app/features/console/application/registry';
import { useAppsStore } from './apps';

// usePluginsStore holds the ACTIVE console plugin set — the hybrid result of
// resolvePlugins over the live /apps list (runtime SystemJS plugin module where
// available, bundled plugin otherwise). The nav, dashboard, and lazy
// route-loader all read from here so the moment an app migrates to a plugin
// module (4d/4e) the whole console follows with no host edits.
//
// Resolution is LAZY per Grafana's model: nav renders from /apps metadata
// immediately; a remote plugin's code resolves on first navigation into its
// `/<slug>` section (see ensureSection). Bundled plugins resolve eagerly and
// synchronously since their code is already in the host bundle.
export const usePluginsStore = defineStore('console-plugins', () => {
	const active = ref<ForgeConsolePlugin[]>([]);
	const resolving = ref(false);
	// Slugs whose remote we've already attempted to resolve, so a flapping
	// remote isn't re-fetched on every navigation.
	const resolvedSlugs = ref<Set<string>>(new Set());

	const bySlug = computed(() => {
		const m = new Map<string, ForgeConsolePlugin>();
		for (const p of active.value) m.set(p.serviceId, p);
		return m;
	});

	// resolveAll runs the hybrid resolver over the full /apps list. Failure-
	// isolated inside resolvePlugins, so this never rejects in practice; the
	// try/catch is belt-and-suspenders so a thrown remote can't wedge the shell.
	async function resolveAll(): Promise<ForgeConsolePlugin[]> {
		resolving.value = true;
		try {
			const apps = useAppsStore().apps;
			active.value = await resolvePlugins(apps);
			resolvedSlugs.value = new Set(apps.map((a) => a.slug));
		} catch (err) {
			console.error('forge-console: plugin resolution failed', err);
		} finally {
			resolving.value = false;
		}
		return active.value;
	}

	// ensureSection lazily resolves the single app owning `slug` if it hasn't
	// been resolved yet — the per-route entry point the router guard calls so a
	// remote's code+routes load only on first navigation into its section.
	async function ensureSection(slug: string): Promise<ForgeConsolePlugin | undefined> {
		if (resolvedSlugs.value.has(slug)) return bySlug.value.get(slug);
		const app = useAppsStore().apps.find((a) => a.slug === slug);
		if (!app) return undefined;
		const [plugin] = await resolvePlugins([app]);
		resolvedSlugs.value.add(slug);
		if (plugin) {
			active.value = [...active.value.filter((p) => p.serviceId !== slug), plugin].sort(
				(a, b) => (a.order ?? Number.MAX_SAFE_INTEGER) - (b.order ?? Number.MAX_SAFE_INTEGER),
			);
		}
		return plugin;
	}

	return { active, resolving, bySlug, resolveAll, ensureSection };
});
