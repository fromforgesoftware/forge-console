<template>
	<div class="space-y-6">
		<div class="space-y-1">
			<h1 class="text-2xl font-semibold">Users</h1>
			<p class="text-sm text-muted-foreground">Manage control plane users and their roles.</p>
		</div>

		<Alert v-if="error" variant="destructive">
			<AlertTitle>Something went wrong</AlertTitle>
			<AlertDescription>{{ error }}</AlertDescription>
		</Alert>

		<DataTable
			:columns="columns"
			:data-source="{ data: users, totalCount: users.length }"
			:pagination="pagination"
			:get-row-id="(r) => r.id"
			@update:pagination="pagination = $event"
		>
			<template #toolbar-end>
				<Button :disabled="!canWrite" @click="createOpen = true">
					<Plus class="mr-2 size-4" />
					New user
				</Button>
			</template>
		</DataTable>

		<UserFormDrawer v-model:open="createOpen" @saved="load" />
		<UserRolesDrawer
			v-model:open="rolesOpen"
			:user="rolesTarget"
			:roles="allRoles"
			:can-write="canWrite"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { Plus, Shield } from '@lucide/vue';
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
import { adminService, type AdminUser, type Role } from '../../services/admin.service';
import { useAuthStore } from '@/app/core/auth/store';
import UserFormDrawer from './forms/UserFormDrawer.vue';
import UserRolesDrawer from './forms/UserRolesDrawer.vue';

const { pagination } = useDataTableState({ key: 'admin-users', defaultPageSize: 25 });

const auth = useAuthStore();
const canWrite = computed(() => auth.can('users.write'));

const users = ref<AdminUser[]>([]);
const allRoles = ref<Role[]>([]);
const error = ref('');
const busy = ref(false);

const createOpen = ref(false);
const rolesOpen = ref(false);
const rolesTarget = ref<AdminUser | null>(null);

const columns: ColumnDef<AdminUser, unknown>[] = [
	{ accessorKey: 'email', header: 'Email', cell: ({ row }) => h('span', row.original.email) },
	{
		accessorKey: 'displayName',
		header: 'Name',
		cell: ({ row }) => h('span', row.original.displayName),
	},
	{
		accessorKey: 'status',
		header: 'Status',
		cell: ({ row }) =>
			h(
				Badge,
				{ variant: row.original.status === 'ENABLED' ? 'default' : 'secondary' },
				() => row.original.status,
			),
	},
	{
		id: 'actions',
		header: '',
		cell: ({ row }) =>
			h('div', { class: 'flex justify-end gap-2' }, [
				h(Button, { variant: 'ghost', size: 'sm', onClick: () => openRoles(row.original) }, () =>
					h(Shield, { class: 'size-4' }),
				),
				h(
					Button,
					{
						variant: 'outline',
						size: 'sm',
						disabled: busy.value || !canWrite.value,
						onClick: () => toggleStatus(row.original),
					},
					() => (row.original.status === 'ENABLED' ? 'Disable' : 'Enable'),
				),
			]),
	},
];

async function load(): Promise<void> {
	error.value = '';
	try {
		[users.value, allRoles.value] = await Promise.all([
			adminService.listUsers(),
			adminService.listRoles(),
		]);
	} catch {
		error.value = 'Failed to load users.';
	}
}

async function toggleStatus(op: AdminUser): Promise<void> {
	busy.value = true;
	try {
		await adminService.setUserStatus(op.id, op.status === 'ENABLED' ? 'DISABLED' : 'ENABLED');
		await load();
	} catch {
		error.value = 'Failed to update status.';
	} finally {
		busy.value = false;
	}
}

function openRoles(op: AdminUser): void {
	rolesTarget.value = op;
	rolesOpen.value = true;
}

onMounted(load);
</script>
