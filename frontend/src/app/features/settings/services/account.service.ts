import { one, send, type JsonApiResource } from '@/app/core/http/api';
import type { User } from '@/app/core/auth/store';

// Account self-service over the shared JSON:API client: profile/me are the
// `users` resource, preferences are the `user-settings` resource.
export interface UserSettings {
	theme: string;
}

function toUser(r: JsonApiResource): User {
	const a = r.attributes;
	return {
		id: r.id,
		email: String(a.email ?? ''),
		displayName: String(a.displayName ?? ''),
		avatar: a.avatar as string | undefined,
		isAdmin: Boolean(a.isAdmin),
		roles: (a.roles as string[]) ?? [],
		permissions: (a.permissions as string[]) ?? [],
		settings: a.settings as { theme: string } | undefined,
	};
}

export const accountService = {
	getMe: async (): Promise<User> => toUser(await one('GET', '/users/me')),

	updateProfile: async (displayName: string): Promise<User> =>
		toUser(await one('PATCH', '/users/me', 'users', { displayName })),

	changePassword: (currentPassword: string, newPassword: string): Promise<void> =>
		send('PUT', '/users/me/password', 'users', { currentPassword, newPassword }),

	getSettings: async (): Promise<UserSettings> => {
		const r = await one('GET', '/users/me/settings');
		return { theme: String(r.attributes.theme ?? 'system') };
	},

	updateSettings: async (theme: string): Promise<UserSettings> => {
		const r = await one('PUT', '/users/me/settings', 'user-settings', { theme });
		return { theme: String(r.attributes.theme ?? theme) };
	},
};
