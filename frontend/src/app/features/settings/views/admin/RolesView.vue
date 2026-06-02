<template>
	<div class="space-y-6">
		<div class="space-y-1">
			<h1 class="text-2xl font-semibold">Roles</h1>
			<p class="text-sm text-muted-foreground">Define roles and the permissions they grant.</p>
		</div>

		<Alert v-if="error" variant="destructive">
			<AlertTitle>Something went wrong</AlertTitle>
			<AlertDescription>{{ error }}</AlertDescription>
		</Alert>

		<DataTable
			:columns="columns"
			:data-source="{ data: roles, totalCount: roles.length }"
			:pagination="pagination"
			:get-row-id="(r) => r.slug"
			@update:pagination="pagination = $event"
		>
			<template #toolbar-end>
				<Button :disabled="!canWrite" @click="createOpen = true">
					<Plus class="mr-2 size-4" />
					New role
				</Button>
			</template>
		</DataTable>

		<RoleFormDrawer v-model:open="createOpen" :permissions="permissions" @saved="load" />
	</div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { Plus, Trash2 } from '@lucide/vue';
import {
	Alert,
	AlertTitle,
	AlertDescription,
	Badge,
	Button,
	DataTable,
	useDataTableState,
} from '@fromforgesoftware/vue-kit';
import type { ColumnDef } from '@fromforgesoftware/vue-kit';
import { adminService, type Permission, type Role } from '../../services/admin.service';
import { useAuthStore } from '@/app/core/auth/store';
import RoleFormDrawer from './forms/RoleFormDrawer.vue';

const { pagination } = useDataTableState({ key: 'admin-roles', defaultPageSize: 25 });

const auth = useAuthStore();
const canWrite = computed(() => auth.can('roles.write'));

const roles = ref<Role[]>([]);
const permissions = ref<Permission[]>([]);
const error = ref('');
const busy = ref(false);
const createOpen = ref(false);

function permissionBadges(role: Role) {
	if (role.permissions.includes('*.*') || role.permissions.includes('*')) {
		return [h(Badge, () => 'Full access')];
	}
	if (role.permissions.length === 0) {
		return [h('span', { class: 'text-muted-foreground text-sm' }, '—')];
	}
	const shown = role.permissions.slice(0, 3);
	const badges = shown.map((p) => h(Badge, { variant: 'secondary' }, () => p));
	if (role.permissions.length > shown.length) {
		badges.push(
			h(Badge, { variant: 'outline' }, () => `+${role.permissions.length - shown.length}`),
		);
	}
	return badges;
}

const columns: ColumnDef<Role, unknown>[] = [
	{ accessorKey: 'slug', header: 'Slug', cell: ({ row }) => h('span', row.original.slug) },
	{ accessorKey: 'name', header: 'Name', cell: ({ row }) => h('span', row.original.name) },
	{
		id: 'kind',
		header: 'Kind',
		cell: ({ row }) =>
			h(
				Badge,
				{ variant: row.original.kind === 'SYSTEM' ? 'default' : 'secondary' },
				() => row.original.kind,
			),
	},
	{
		id: 'permissions',
		header: 'Permissions',
		cell: ({ row }) => h('div', { class: 'flex flex-wrap gap-1' }, permissionBadges(row.original)),
	},
	{
		id: 'actions',
		header: '',
		cell: ({ row }) =>
			h('div', { class: 'flex justify-end' }, [
				h(
					Button,
					{
						variant: 'ghost',
						size: 'sm',
						disabled: busy.value || !canWrite.value || row.original.kind === 'SYSTEM',
						onClick: () => remove(row.original),
					},
					() => h(Trash2, { class: 'size-4 text-destructive' }),
				),
			]),
	},
];

async function load(): Promise<void> {
	error.value = '';
	try {
		[roles.value, permissions.value] = await Promise.all([
			adminService.listRoles(),
			adminService.listPermissions(),
		]);
	} catch {
		error.value = 'Failed to load roles.';
	}
}

async function remove(role: Role): Promise<void> {
	busy.value = true;
	try {
		await adminService.deleteRole(role.slug);
		await load();
	} catch {
		error.value = 'Failed to delete role.';
	} finally {
		busy.value = false;
	}
}

onMounted(load);
</script>
