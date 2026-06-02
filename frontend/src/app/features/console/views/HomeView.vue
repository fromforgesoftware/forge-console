<template>
	<div class="flex flex-1 items-center justify-center text-sm text-muted-foreground">
		<template v-if="!firstPage">No apps enabled.</template>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { plugins } from '@/app/features/console/application/registry';
import { useAppsStore } from '@/app/features/console/stores/apps';

const router = useRouter();
const apps = useAppsStore();

// Land on the first enabled app's first page. Replaces the old dashboard —
// navigation lives in the sidebar + command palette, so the index is just a
// redirect (or an empty state when no app is installed).
const firstPage = computed(() => {
	for (const plugin of plugins) {
		if (!apps.slugs.includes(plugin.serviceId)) continue;
		const page = plugin.pages.find((p) => !p.path.includes('/'));
		if (page) return `${plugin.basePath}/${page.path}`;
	}
	return null;
});

function go() {
	if (firstPage.value) void router.replace(firstPage.value);
}

onMounted(go);
watch(firstPage, go);
</script>
