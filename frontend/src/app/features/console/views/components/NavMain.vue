<template>
	<SidebarGroup>
		<SidebarGroupContent>
			<SidebarMenu>
				<!-- One collapsible group per enabled app -->
				<SidebarMenuItem v-for="app in enabledApps" :key="app.serviceId">
					<SidebarMenuCollapsible
						:label="app.title"
						:items="subItemsFor(app)"
						:open="expanded[app.serviceId] ?? isAppActive(app)"
						:active-item="activeItem"
						@update:open="(v: boolean) => (expanded[app.serviceId] = v)"
						@select="(key: string) => router.push(`/${key}`)"
					>
						<template #icon><component :is="app.icon" /></template>
					</SidebarMenuCollapsible>
				</SidebarMenuItem>
			</SidebarMenu>
		</SidebarGroupContent>
	</SidebarGroup>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
	SidebarGroup,
	SidebarGroupContent,
	SidebarMenu,
	SidebarMenuItem,
	SidebarMenuCollapsible,
} from '@fromforgesoftware/vue-kit';
import { enabledPlugins } from '@/app/features/console/application/registry';
import type { ForgeConsolePlugin } from '@/app/features/console/domain/plugin';
import { useAppsStore } from '@/app/features/console/stores/apps';

const route = useRoute();
const router = useRouter();
const apps = useAppsStore();

const expanded = reactive<Record<string, boolean>>({});

const enabledApps = computed(() => enabledPlugins(apps.slugs));

// Top-level pages only — sub-paths like `realms/new` stay out of the nav.
function subItemsFor(app: ForgeConsolePlugin) {
	return app.pages
		.filter((page) => !page.path.includes('/'))
		.map((page) => ({
			key: `${app.basePath.replace(/^\//, '')}/${page.path}`,
			label: page.name,
		}));
}

function isAppActive(app: ForgeConsolePlugin): boolean {
	return route.path.startsWith(app.basePath);
}

// Active sub-item key matches the path without its leading slash.
const activeItem = computed(() => route.path.replace(/^\//, ''));
</script>
