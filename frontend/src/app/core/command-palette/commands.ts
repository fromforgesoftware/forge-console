import {
	LogOut,
	User,
	SlidersHorizontal,
	Shield,
	Users,
	ShieldCheck,
	Boxes,
	KeyRound,
} from '@lucide/vue';
import { enabledPlugins } from '@/app/features/console/application/registry';
import { useAuthStore } from '@/app/core/auth/store';
import { type useAppsStore } from '@/app/features/console/stores/apps';
import type { Command } from './types';

// buildCommands derives navigation commands from the registered plugin pages
// (one per visible top-level page) plus a sign-out action. Each plugin's icon
// comes from its manifest. Only apps the user actually has are surfaced.
export function buildCommands(apps: ReturnType<typeof useAppsStore>): Command[] {
	const commands: Command[] = [];

	for (const plugin of enabledPlugins(apps.slugs)) {
		for (const page of plugin.pages) {
			if (page.path.includes('/')) continue;
			const path = `${plugin.basePath}/${page.path}`;
			commands.push({
				id: `route.${plugin.serviceId}.${page.path}`,
				title: page.name,
				group: 'navigation',
				icon: plugin.icon,
				parent: plugin.title,
				keywords: [plugin.serviceId, plugin.title],
				handler: ({ router }) => {
					void router.push(path);
				},
			});
		}
	}

	const settingsNav: Command[] = [
		{
			id: 'route.settings.profile',
			title: 'Profile',
			group: 'navigation',
			icon: User,
			parent: 'Settings',
			keywords: ['account', 'settings'],
			handler: ({ router }) => void router.push('/settings/account/profile'),
		},
		{
			id: 'route.settings.preferences',
			title: 'Preferences',
			group: 'navigation',
			icon: SlidersHorizontal,
			parent: 'Settings',
			keywords: ['theme', 'appearance', 'settings'],
			handler: ({ router }) => void router.push('/settings/account/preferences'),
		},
		{
			id: 'route.settings.security',
			title: 'Security',
			group: 'navigation',
			icon: Shield,
			parent: 'Settings',
			keywords: ['password', 'settings'],
			handler: ({ router }) => void router.push('/settings/account/security'),
		},
	];
	commands.push(...settingsNav);

	const auth = useAuthStore();
	if (auth.can('users.read')) {
		commands.push({
			id: 'route.admin.users',
			title: 'Users',
			group: 'navigation',
			icon: Users,
			parent: 'Administration',
			keywords: ['admin', 'users'],
			handler: ({ router }) => void router.push('/settings/admin/users'),
		});
	}
	if (auth.can('roles.read')) {
		commands.push({
			id: 'route.admin.roles',
			title: 'Roles',
			group: 'navigation',
			icon: ShieldCheck,
			parent: 'Administration',
			keywords: ['admin', 'roles', 'permissions'],
			handler: ({ router }) => void router.push('/settings/admin/roles'),
		});
	}
	if (auth.can('apps.read')) {
		commands.push({
			id: 'route.admin.apps',
			title: 'Apps',
			group: 'navigation',
			icon: Boxes,
			parent: 'Administration',
			keywords: ['admin', 'apps'],
			handler: ({ router }) => void router.push('/settings/admin/apps'),
		});
	}
	if (auth.can('service_accounts.read')) {
		commands.push({
			id: 'route.admin.service-accounts',
			title: 'Service accounts',
			group: 'navigation',
			icon: KeyRound,
			parent: 'Administration',
			keywords: ['admin', 'service', 'accounts', 'machine', 'token', 'credentials'],
			handler: ({ router }) => void router.push('/settings/admin/service-accounts'),
		});
	}

	commands.push({
		id: 'action.sign-out',
		title: 'Sign out',
		group: 'action',
		icon: LogOut,
		keywords: ['logout', 'log out', 'exit'],
		handler: async ({ router }) => {
			await useAuthStore().logout();
			void router.push({ name: 'login' });
		},
	});

	return commands;
}
