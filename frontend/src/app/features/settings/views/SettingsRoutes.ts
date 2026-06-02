import type { RouteRecordRaw } from 'vue-router';
import SettingsLayout from './SettingsLayout.vue';
import { useAuthStore } from '@/app/core/auth/store';

// Admin pages are guarded client-side for UX (the nav already hides them and
// the backend 403s users without the permission). Users lacking the required
// read permission are bounced to Profile.
const requirePerm = (action: string) => (): boolean | { path: string } => {
	if (useAuthStore().can(action)) return true;
	return { path: '/settings/account/profile' };
};

export const SETTINGS_ROUTES: RouteRecordRaw[] = [
	{
		path: '/settings',
		component: SettingsLayout,
		meta: { requiresAuth: true },
		children: [
			{ path: '', redirect: '/settings/account/profile' },
			{
				path: 'account/profile',
				name: 'settings-account-profile',
				component: () => import('./account/ProfileView.vue'),
			},
			{
				path: 'account/preferences',
				name: 'settings-account-preferences',
				component: () => import('./account/PreferencesView.vue'),
			},
			{
				path: 'account/security',
				name: 'settings-account-security',
				component: () => import('./account/SecurityView.vue'),
			},
			{
				path: 'admin/users',
				name: 'settings-admin-users',
				meta: { maxWidth: 'max-w-5xl' },
				beforeEnter: requirePerm('users.read'),
				component: () => import('./admin/UsersView.vue'),
			},
			{
				path: 'admin/roles',
				name: 'settings-admin-roles',
				meta: { maxWidth: 'max-w-5xl' },
				beforeEnter: requirePerm('roles.read'),
				component: () => import('./admin/RolesView.vue'),
			},
			{
				path: 'admin/apps',
				name: 'settings-admin-apps',
				meta: { maxWidth: 'max-w-5xl' },
				beforeEnter: requirePerm('apps.read'),
				component: () => import('./admin/AppsView.vue'),
			},
			{
				path: 'admin/service-accounts',
				name: 'settings-admin-service-accounts',
				meta: { maxWidth: 'max-w-5xl' },
				beforeEnter: requirePerm('service_accounts.read'),
				component: () => import('./admin/ServiceAccountsView.vue'),
			},
		],
	},
];
