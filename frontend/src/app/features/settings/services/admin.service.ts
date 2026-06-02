import { one, many, send, type JsonApiResource } from '@/app/core/http/api';

// The admin API speaks JSON:API via the shared client: writes wrap attributes
// in `{ data: { type, attributes } }`, reads return `{ data: <resource|[]> }`.
// These mappers turn each `{ id, type, attributes }` into the flat view types
// the admin UI works with. Paths are relative to the client basePath (`/api`).

export type UserStatus = 'ENABLED' | 'DISABLED';

export interface AdminUser {
	id: string;
	email: string;
	displayName: string;
	status: UserStatus;
}

const toUser = (r: JsonApiResource): AdminUser => ({
	id: r.id,
	email: String(r.attributes.email ?? ''),
	displayName: String(r.attributes.displayName ?? ''),
	status: (r.attributes.status as UserStatus) ?? 'ENABLED',
});

export interface NewUser {
	email: string;
	displayName: string;
	password?: string;
}

export type RoleKind = 'SYSTEM' | 'CUSTOM';

export interface Role {
	slug: string;
	name: string;
	kind: RoleKind;
	permissions: string[];
}

const toRole = (r: JsonApiResource): Role => ({
	slug: String(r.attributes.slug ?? r.id),
	name: String(r.attributes.name ?? ''),
	kind: (r.attributes.kind as RoleKind) ?? 'CUSTOM',
	permissions: (r.attributes.permissions as string[]) ?? [],
});

export interface NewRole {
	slug: string;
	name: string;
	permissions: string[];
}

export interface Permission {
	id: string;
	resourceType: string;
	verb: string;
	description: string;
}

const toPermission = (r: JsonApiResource): Permission => ({
	id: r.id,
	resourceType: String(r.attributes.resourceType ?? ''),
	verb: String(r.attributes.verb ?? ''),
	description: String(r.attributes.description ?? ''),
});

export interface App {
	slug: string;
	name: string;
	kind: string;
	adminBaseURL: string;
	enabled: boolean;
}

const toApp = (r: JsonApiResource): App => ({
	slug: String(r.attributes.slug ?? r.id),
	name: String(r.attributes.name ?? ''),
	kind: String(r.attributes.kind ?? ''),
	adminBaseURL: String(r.attributes.adminBaseURL ?? ''),
	enabled: Boolean(r.attributes.enabled),
});

export interface ServiceAccount {
	id: string;
	name: string;
	clientId: string;
	status: UserStatus;
	lastUsedAt?: string;
}

const toServiceAccount = (r: JsonApiResource): ServiceAccount => ({
	id: r.id,
	name: String(r.attributes.name ?? ''),
	clientId: String(r.attributes.clientId ?? ''),
	status: (r.attributes.status as UserStatus) ?? 'ENABLED',
	lastUsedAt: r.attributes.lastUsedAt as string | undefined,
});

export interface NewServiceAccountResult {
	id: string;
	name: string;
	clientId: string;
	clientSecret: string;
}

export const adminService = {
	listUsers: async (): Promise<AdminUser[]> => (await many('/admin/users')).map(toUser),
	createUser: async (user: NewUser): Promise<AdminUser> =>
		toUser(await one('POST', '/admin/users', 'users', { ...user })),
	setUserStatus: async (id: string, status: UserStatus): Promise<AdminUser> =>
		toUser(await one('PATCH', `/admin/users/${id}`, 'users', { status })),
	getUserRoles: async (id: string): Promise<Role[]> =>
		(await many(`/admin/users/${id}/roles`)).map(toRole),
	setUserRoles: (id: string, roles: string[]) =>
		one('PUT', `/admin/users/${id}/roles`, 'users', { roles }),

	listRoles: async (): Promise<Role[]> => (await many('/admin/roles')).map(toRole),
	createRole: async (role: NewRole): Promise<Role> =>
		toRole(await one('POST', '/admin/roles', 'roles', { ...role })),
	deleteRole: (slug: string) => send('DELETE', `/admin/roles/${slug}`),

	listPermissions: async (): Promise<Permission[]> =>
		(await many('/admin/permissions')).map(toPermission),

	listApps: async (): Promise<App[]> => (await many('/admin/apps')).map(toApp),
	createApp: async (slug: string, app: Omit<App, 'slug'>): Promise<App> =>
		toApp(await one('POST', `/admin/apps/${slug}`, 'apps', { ...app })),
	updateApp: async (slug: string, app: Omit<App, 'slug'>): Promise<App> =>
		toApp(await one('PUT', `/admin/apps/${slug}`, 'apps', { ...app })),

	listServiceAccounts: async (): Promise<ServiceAccount[]> =>
		(await many('/admin/service-accounts')).map(toServiceAccount),
	createServiceAccount: async (name: string): Promise<NewServiceAccountResult> => {
		const r = await one('POST', '/admin/service-accounts', 'service-accounts', { name });
		return {
			id: r.id,
			name: String(r.attributes.name ?? ''),
			clientId: String(r.attributes.clientId ?? ''),
			clientSecret: String(r.attributes.clientSecret ?? ''),
		};
	},
	deleteServiceAccount: (id: string) => send('DELETE', `/admin/service-accounts/${id}`),
	getServiceAccountRoles: async (id: string): Promise<Role[]> =>
		(await many(`/admin/service-accounts/${id}/roles`)).map(toRole),
	setServiceAccountRoles: (id: string, roles: string[]) =>
		one('PUT', `/admin/service-accounts/${id}/roles`, 'service-accounts', { roles }),
};
