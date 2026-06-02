import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { apiBaseFor } from '@/app/core/http/services';
import { fetchCollection } from '@/app/core/http/jsonapi';
import { useAppsStore } from './apps';

// Realm is an Aegis realm (the platform's "tenant"). The sidebar switcher lets
// a user pick which one they're administering.
export interface Realm {
	id: string;
	name: string;
	displayName: string;
}

const CURRENT_KEY = 'forge.realm';

export const useRealmStore = defineStore('realm', () => {
	const realms = ref<Realm[]>([]);
	const currentId = ref<string | null>(localStorage.getItem(CURRENT_KEY));

	const current = computed(
		() => realms.value.find((r) => r.id === currentId.value) ?? realms.value[0] ?? null,
	);

	// load fetches realms through the gateway — only when the Aegis app is
	// installed (otherwise there are no realms to switch between).
	async function load(): Promise<void> {
		if (!useAppsStore().slugs.includes('aegis')) {
			realms.value = [];
			return;
		}
		try {
			const rows = await fetchCollection(apiBaseFor('aegis'), 'realms', null);
			realms.value = rows.map((d) => ({
				id: d.id,
				name: String(d.attributes.name ?? ''),
				displayName: String(d.attributes.displayName || d.attributes.name || ''),
			}));
			if (!currentId.value && realms.value[0]) currentId.value = realms.value[0].id;
		} catch {
			realms.value = [];
		}
	}

	function select(id: string): void {
		currentId.value = id;
		localStorage.setItem(CURRENT_KEY, id);
	}

	return { realms, current, currentId, load, select };
});
