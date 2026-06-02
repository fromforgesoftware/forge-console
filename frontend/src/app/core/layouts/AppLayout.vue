<template>
	<SidebarProvider collapsible="icon">
		<Sidebar>
			<SidebarHeader>
				<SidebarBrand />
			</SidebarHeader>

			<SidebarContent>
				<!-- Command palette trigger -->
				<SidebarGroup class="pb-0">
					<SidebarGroupContent>
						<SidebarCollapsibleHide class="px-1 pb-1">
							<Button
								variant="outline"
								class="w-full justify-between font-normal text-muted-foreground"
								@click="commandPaletteOpen = true"
							>
								<Search class="size-4 shrink-0" />
								<span class="flex-1 truncate text-left">Search here...</span>
								<KbdGroup size="sm" class="pointer-events-none hidden sm:flex">
									<Kbd>⌘</Kbd>
									<Kbd>K</Kbd>
								</KbdGroup>
							</Button>
						</SidebarCollapsibleHide>
						<SidebarCollapsibleShow>
							<SidebarMenu>
								<SidebarMenuItem>
									<SidebarMenuButton tooltip="Search" @click="commandPaletteOpen = true">
										<Search />
									</SidebarMenuButton>
								</SidebarMenuItem>
							</SidebarMenu>
						</SidebarCollapsibleShow>
					</SidebarGroupContent>
				</SidebarGroup>

				<NavPlatform />
				<NavMain />
			</SidebarContent>

			<SidebarFooter>
				<NavUser />
			</SidebarFooter>
		</Sidebar>

		<div class="flex flex-1 flex-col min-h-0 min-w-0 lg:p-2 lg:pl-0">
			<main
				class="flex flex-1 flex-col min-w-0 bg-background overflow-hidden lg:rounded-xl lg:border lg:border-border"
			>
				<header class="flex h-12 shrink-0 items-center gap-2 border-b border-border px-3">
					<SidebarTrigger class="lg:hidden text-foreground">
						<Menu class="size-5" />
					</SidebarTrigger>
					<SidebarCollapsibleShow class="hidden lg:contents">
						<SidebarTrigger class="text-foreground">
							<PanelLeftOpen class="size-4" />
						</SidebarTrigger>
					</SidebarCollapsibleShow>
					<nav
						class="flex flex-1 items-center gap-1.5 min-w-0 overflow-hidden text-sm text-muted-foreground"
					>
						<span class="font-medium text-foreground shrink-0">Forge</span>
						<template v-if="pageTitle">
							<span class="text-muted-foreground/40 shrink-0">/</span>
							<span class="truncate font-medium text-foreground">{{ pageTitle }}</span>
						</template>
					</nav>
				</header>

				<div
					data-slot="app-layout-scroll"
					class="relative flex flex-col flex-1 min-h-0 min-w-0 overflow-auto p-2 sm:p-3 lg:p-4"
				>
					<RouterView :key="route.fullPath" />
				</div>
			</main>
		</div>
	</SidebarProvider>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { RouterView, useRoute } from 'vue-router';
import { Menu, PanelLeftOpen, Search } from '@lucide/vue';
import {
	Sidebar,
	SidebarProvider,
	SidebarHeader,
	SidebarContent,
	SidebarFooter,
	SidebarGroup,
	SidebarGroupContent,
	SidebarMenu,
	SidebarMenuItem,
	SidebarMenuButton,
	SidebarCollapsibleShow,
	SidebarCollapsibleHide,
	SidebarTrigger,
	Button,
	Kbd,
	KbdGroup,
} from '@fromforgesoftware/vue-kit';
import SidebarBrand from '@/app/features/console/views/components/SidebarBrand.vue';
import NavPlatform from '@/app/features/platform/views/NavPlatform.vue';
import NavMain from '@/app/features/console/views/components/NavMain.vue';
import NavUser from '@/app/features/console/views/components/NavUser.vue';
import { useCommandPalette } from '@/app/core/command-palette';

const route = useRoute();
const { commandPaletteOpen } = useCommandPalette();

// Route names are `${serviceId}:${pageName}`; surface the page name as the
// breadcrumb tail. The bare dashboard route has no colon.
const pageTitle = computed(() => {
	const name = String(route.name ?? '');
	if (name.includes(':')) return name.split(':')[1];
	return '';
});
</script>
