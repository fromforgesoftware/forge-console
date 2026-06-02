import { ShieldCheck } from '@lucide/vue';
import type { ForgeConsolePlugin, ForgeConsolePage } from '@/app/features/console/domain/plugin';
import { apiBaseFor } from '@/app/core/http/services';
import ResourceListView from '@/app/features/console/views/components/ResourceListView.vue';
import ResourceCreateForm from '@/app/features/console/views/components/ResourceCreateForm.vue';
import ActionForm from '@/app/features/console/views/components/ActionForm.vue';
import RoleBuilder from '@/app/features/console/views/components/RoleBuilder.vue';
import BindingForm from '@/app/features/console/views/components/BindingForm.vue';
import AegisOverview from '@/app/features/console/views/components/AegisOverview.vue';

// The Aegis console plugin: an overview dashboard plus the admin surface over
// the JSON:API admin API — list pages via the generic renderer, plus the
// custom-role builder and binding editor.
function list(apiBase: string, type: string, name: string, columns: string[]): ForgeConsolePage {
	return {
		path: type,
		name,
		component: ResourceListView,
		props: { apiBase, type, title: name, columns },
	};
}

export function aegisPlugin(): ForgeConsolePlugin {
	const apiBase = apiBaseFor('aegis');
	return {
		serviceId: 'aegis',
		title: 'Aegis',
		basePath: '/aegis',
		apiBase,
		icon: ShieldCheck,
		order: 1,
		pages: [
			{ path: 'overview', name: 'Overview', component: AegisOverview, props: { apiBase } },

			list(apiBase, 'realms', 'Realms', ['name', 'displayName']),
			list(apiBase, 'organizations', 'Organizations', ['name', 'slug', 'status']),
			list(apiBase, 'invitations', 'Invitations', ['email', 'status', 'roleId']),
			list(apiBase, 'session-states', 'Sessions', ['accountId', 'currentShard', 'region']),
			list(apiBase, 'service-accounts', 'Service accounts', ['name', 'clientId', 'lastUsedAt']),
			list(apiBase, 'roles', 'Roles', ['name', 'resourceType', 'kind']),
			list(apiBase, 'permissions', 'Permissions', ['resourceType', 'verb']),
			list(apiBase, 'resources', 'Resources', ['type', 'parentId', 'visibility']),
			list(apiBase, 'bindings', 'Bindings', ['resourceId', 'roleId', 'subjectType', 'subjectId']),
			list(apiBase, 'clients', 'OIDC clients', ['name', 'type']),
			list(apiBase, 'audit-events', 'Audit log', ['action', 'actorId', 'resourceType']),
			{
				path: 'realms/new',
				name: 'New realm',
				component: ResourceCreateForm,
				props: {
					apiBase,
					type: 'realms',
					title: 'New realm',
					fields: [
						{ name: 'name', label: 'Name', required: true },
						{ name: 'displayName', label: 'Display name' },
					],
				},
			},
			{
				path: 'organizations/new',
				name: 'New organization',
				component: ResourceCreateForm,
				props: {
					apiBase,
					type: 'organizations',
					title: 'New organization',
					fields: [
						{ name: 'realmId', label: 'Realm ID', required: true },
						{ name: 'name', label: 'Name', required: true },
						{ name: 'slug', label: 'Slug', required: true },
					],
				},
			},
			{ path: 'roles/new', name: 'New role', component: RoleBuilder, props: { apiBase } },
			{ path: 'bindings/new', name: 'New binding', component: BindingForm, props: { apiBase } },
			{
				path: 'accounts/ban',
				name: 'Ban account',
				component: ActionForm,
				props: {
					apiBase,
					path: '/api/accounts/ban',
					type: 'accountBans',
					title: 'Ban account',
					submitLabel: 'Ban',
					fields: [
						{ name: 'accountId', label: 'Account ID', required: true },
						{ name: 'reason', label: 'Reason' },
						{ name: 'until', label: 'Until (RFC3339, blank = permanent)' },
					],
				},
			},
			{
				path: 'accounts/unban',
				name: 'Unban account',
				component: ActionForm,
				props: {
					apiBase,
					path: '/api/accounts/unban',
					type: 'accountBans',
					title: 'Unban account',
					submitLabel: 'Unban',
					fields: [{ name: 'accountId', label: 'Account ID', required: true }],
				},
			},
			{
				path: 'accounts/merge',
				name: 'Merge accounts',
				component: ActionForm,
				props: {
					apiBase,
					path: '/api/accounts/merge',
					type: 'accountMerges',
					title: 'Merge accounts',
					submitLabel: 'Merge',
					fields: [
						{ name: 'sourceId', label: 'Source account ID', required: true },
						{ name: 'targetId', label: 'Target (survivor) account ID', required: true },
					],
				},
			},
			{
				path: 'risk-policy',
				name: 'Risk policy',
				component: ActionForm,
				props: {
					apiBase,
					path: '/api/realm-risk-policies',
					type: 'realmRiskPolicies',
					title: 'Realm risk policy',
					submitLabel: 'Save',
					fields: [
						{ name: 'realmId', label: 'Realm ID', required: true },
						{ name: 'newIpWeight', label: 'New IP weight', type: 'number' },
						{ name: 'newDeviceWeight', label: 'New device weight', type: 'number' },
						{ name: 'failureWeight', label: 'Failure weight', type: 'number' },
						{ name: 'stepUpThreshold', label: 'Step-up threshold', type: 'number' },
						{ name: 'denyThreshold', label: 'Deny threshold', type: 'number' },
					],
				},
			},
			{
				path: 'mfa-policy',
				name: 'MFA policy',
				component: ActionForm,
				props: {
					apiBase,
					path: '/api/realm-acr-policies',
					type: 'realmAcrPolicies',
					title: 'Realm MFA policy',
					submitLabel: 'Save',
					fields: [
						{ name: 'realmId', label: 'Realm ID', required: true },
						{ name: 'mfaRequired', label: 'Require MFA', type: 'checkbox' },
						{ name: 'requiredAcr', label: 'Required ACR (e.g. aal2)' },
					],
				},
			},
			{
				path: 'service-accounts/new',
				name: 'New service account',
				component: ResourceCreateForm,
				props: {
					apiBase,
					type: 'service-accounts',
					title: 'New service account',
					fields: [
						{ name: 'realmId', label: 'Realm ID', required: true },
						{ name: 'name', label: 'Name', required: true },
					],
				},
			},
		],
	};
}
