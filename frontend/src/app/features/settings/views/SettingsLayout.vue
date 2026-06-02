<template>
	<SidebarProvider :collapsible="isDesktop ? 'none' : 'offcanvas'" :open="true">
		<Sidebar>
			<SidebarContent class="pt-2">
				<div class="px-2 pb-2">
					<Button
						variant="ghost"
						size="sm"
						class="group w-full justify-start gap-2 text-muted-foreground hover:text-foreground"
						@click="router.push('/')"
					>
						<ArrowLeft class="size-4 transition-transform group-hover:-translate-x-0.5" />
						<span>Go back</span>
					</Button>
				</div>

				<SidebarGroup v-for="section in visibleSections" :key="section.label">
					<SidebarGroupLabel>{{ section.label }}</SidebarGroupLabel>
					<SidebarGroupContent>
						<SidebarMenu>
							<SidebarMenuItem v-for="item in section.items" :key="item.name">
								<SidebarMenuButton
									:is-active="isItemActive(item.name)"
									@click="router.push(item.path)"
								>
									<component :is="item.icon" />
									<span>{{ item.label }}</span>
								</SidebarMenuButton>
							</SidebarMenuItem>
						</SidebarMenu>
					</SidebarGroupContent>
				</SidebarGroup>
			</SidebarContent>
		</Sidebar>

		<div class="flex flex-1 flex-col min-h-0 min-w-0 lg:p-2 lg:pl-0">
			<main
				class="flex flex-1 flex-col min-w-0 bg-background overflow-hidden lg:rounded-xl lg:border lg:border-border"
			>
				<header class="flex h-12 shrink-0 items-center gap-2 border-b border-border px-4 lg:px-6">
					<SidebarTrigger class="lg:hidden text-foreground shrink-0">
						<Menu class="size-5" />
					</SidebarTrigger>
					<Breadcrumb>
						<BreadcrumbList>
							<BreadcrumbItem>
								<BreadcrumbLink class="cursor-pointer" @click="router.push(sectionDefaultPath)">
									Settings
								</BreadcrumbLink>
							</BreadcrumbItem>
							<BreadcrumbSeparator />
							<BreadcrumbItem>
								<BreadcrumbPage>{{ currentPage }}</BreadcrumbPage>
							</BreadcrumbItem>
						</BreadcrumbList>
					</Breadcrumb>
				</header>

				<div class="flex flex-1 min-h-0 min-w-0 overflow-hidden">
					<div class="flex-1 overflow-y-auto min-h-0 min-w-0">
						<div class="flex flex-col items-center py-10 px-6">
							<div :class="`w-full ${pageMaxWidth}`">
								<RouterView />
							</div>
						</div>
					</div>
				</div>
			</main>
		</div>
	</SidebarProvider>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import {
	SidebarProvider,
	Sidebar,
	SidebarContent,
	SidebarGroup,
	SidebarGroupLabel,
	SidebarGroupContent,
	SidebarMenu,
	SidebarMenuItem,
	SidebarMenuButton,
	SidebarTrigger,
	Breadcrumb,
	BreadcrumbList,
	BreadcrumbItem,
	BreadcrumbLink,
	BreadcrumbPage,
	BreadcrumbSeparator,
	Button,
} from '@fromforgesoftware/vue-kit';
import {
	ArrowLeft,
	Menu,
	User,
	SlidersHorizontal,
	Shield,
	Users,
	ShieldCheck,
	Boxes,
	KeyRound,
} from '@lucide/vue';
import type { Component } from 'vue';
import { useAuthStore } from '@/app/core/auth/store';

interface NavItem {
	label: string;
	name: string;
	path: string;
	icon: Component;
	visible?: () => boolean;
}

interface NavSection {
	label: string;
	items: NavItem[];
	visible?: () => boolean;
}

const router = useRouter();
const route = useRoute();
const auth = useAuthStore();

const media = window.matchMedia('(min-width: 1024px)');
const isDesktop = ref(media.matches);
const onMediaChange = (e: MediaQueryListEvent) => (isDesktop.value = e.matches);
media.addEventListener('change', onMediaChange);
onUnmounted(() => media.removeEventListener('change', onMediaChange));

const sections = computed<NavSection[]>(() => [
	{
		label: 'Account',
		items: [
			{
				label: 'Profile',
				name: 'settings-account-profile',
				path: '/settings/account/profile',
				icon: User,
			},
			{
				label: 'Preferences',
				name: 'settings-account-preferences',
				path: '/settings/account/preferences',
				icon: SlidersHorizontal,
			},
			{
				label: 'Security',
				name: 'settings-account-security',
				path: '/settings/account/security',
				icon: Shield,
			},
		],
	},
	{
		label: 'Administration',
		visible: () =>
			auth.can('users.read') ||
			auth.can('roles.read') ||
			auth.can('apps.read') ||
			auth.can('service_accounts.read'),
		items: [
			{
				label: 'Users',
				name: 'settings-admin-users',
				path: '/settings/admin/users',
				icon: Users,
				visible: () => auth.can('users.read'),
			},
			{
				label: 'Roles',
				name: 'settings-admin-roles',
				path: '/settings/admin/roles',
				icon: ShieldCheck,
				visible: () => auth.can('roles.read'),
			},
			{
				label: 'Apps',
				name: 'settings-admin-apps',
				path: '/settings/admin/apps',
				icon: Boxes,
				visible: () => auth.can('apps.read'),
			},
			{
				label: 'Service accounts',
				name: 'settings-admin-service-accounts',
				path: '/settings/admin/service-accounts',
				icon: KeyRound,
				visible: () => auth.can('service_accounts.read'),
			},
		],
	},
]);

const visibleSections = computed(() =>
	sections.value
		.filter((s) => !s.visible || s.visible())
		.map((s) => ({ ...s, items: s.items.filter((i) => !i.visible || i.visible()) })),
);

const routeName = computed(() => String(route.name ?? ''));

const owner = computed(() => {
	for (const section of visibleSections.value) {
		for (const item of section.items) {
			if (item.name === routeName.value) return { section, item };
		}
	}
	return null;
});

function isItemActive(itemName: string): boolean {
	return owner.value?.item.name === itemName;
}

const currentSection = computed(() => owner.value?.section ?? visibleSections.value[0]);
const currentPage = computed(() => owner.value?.item.label ?? '');

const pageMaxWidth = computed(() => (route.meta.maxWidth as string) ?? 'max-w-2xl');

const sectionDefaultPath = computed(
	() => currentSection.value?.items[0]?.path ?? '/settings/account/profile',
);
</script>
