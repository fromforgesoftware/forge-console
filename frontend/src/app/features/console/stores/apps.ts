import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { many } from '@/app/core/http/api';

// AppInfo is a managed app the backend reports as enabled. Drives which
// console plugins appear — "Forge + Hallmark only" simply omits the rest.
export interface AppInfo {
	slug: string;
	name: string;
	kind: string;
}

export const useAppsStore = defineStore('apps', () => {
	const apps = ref<AppInfo[]>([]);
	const slugs = computed(() => apps.value.map((p) => p.slug));

	async function load(): Promise<void> {
		try {
			apps.value = (await many('/apps')).map((r) => ({
				slug: r.attributes.slug as string,
				name: r.attributes.name as string,
				kind: r.attributes.kind as string,
			}));
		} catch {
			apps.value = [];
		}
	}

	return { apps, slugs, load };
});
