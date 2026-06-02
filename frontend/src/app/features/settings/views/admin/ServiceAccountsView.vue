<template>
	<div class="space-y-6">
		<div class="space-y-1">
			<h1 class="text-2xl font-semibold">Service accounts</h1>
			<p class="text-sm text-muted-foreground">
				Machine identities that authenticate via client-credentials.
			</p>
		</div>

		<Alert v-if="error" variant="destructive">
			<AlertTitle>Something went wrong</AlertTitle>
			<AlertDescription>{{ error }}</AlertDescription>
		</Alert>

		<DataTable
			:columns="columns"
			:data-source="{ data: accounts, totalCount: accounts.length }"
			:pagination="pagination"
			:get-row-id="(r) => r.id"
			@update:pagination="pagination = $event"
		>
			<template #toolbar-end>
				<Button :disabled="!canWrite" @click="createOpen = true">
					<Plus class="mr-2 size-4" />
					New service account
				</Button>
			</template>
		</DataTable>

		<ServiceAccountFormDrawer v-model:open="createOpen" @saved="load" />
		<ServiceAccountRolesDrawer
			v-model:open="rolesOpen"
			:account="rolesTarget"
			:roles="allRoles"
			:can-write="canWrite"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { Plus, Shield, Trash2 } from '@lucide/vue';
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
import { adminService, type ServiceAccount, type Role } from '../../services/admin.service';
import { useAuthStore } from '@/app/core/auth/store';
import ServiceAccountFormDrawer from './forms/ServiceAccountFormDrawer.vue';
import ServiceAccountRolesDrawer from './forms/ServiceAccountRolesDrawer.vue';

const { pagination } = useDataTableState({ key: 'admin-service-accounts', defaultPageSize: 25 });

const auth = useAuthStore();
const canWrite = computed(() => auth.can('service_accounts.write'));

const accounts = ref<ServiceAccount[]>([]);
const allRoles = ref<Role[]>([]);
const error = ref('');
const busy = ref(false);

const createOpen = ref(false);
const rolesOpen = ref(false);
const rolesTarget = ref<ServiceAccount | null>(null);

const columns: ColumnDef<ServiceAccount, unknown>[] = [
	{ accessorKey: 'name', header: 'Name', cell: ({ row }) => h('span', row.original.name) },
	{
		accessorKey: 'clientId',
		header: 'Client ID',
		cell: ({ row }) => h('span', { class: 'font-mono text-xs' }, row.original.clientId),
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
		accessorKey: 'lastUsedAt',
		header: 'Last used',
		cell: ({ row }) =>
			h(
				'span',
				{ class: 'text-sm text-muted-foreground' },
				row.original.lastUsedAt ? new Date(row.original.lastUsedAt).toLocaleString() : 'Never',
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
						variant: 'ghost',
						size: 'sm',
						disabled: busy.value || !canWrite.value,
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
		[accounts.value, allRoles.value] = await Promise.all([
			adminService.listServiceAccounts(),
			adminService.listRoles(),
		]);
	} catch {
		error.value = 'Failed to load service accounts.';
	}
}

async function remove(sa: ServiceAccount): Promise<void> {
	busy.value = true;
	try {
		await adminService.deleteServiceAccount(sa.id);
		await load();
	} catch {
		error.value = 'Failed to delete service account.';
	} finally {
		busy.value = false;
	}
}

function openRoles(sa: ServiceAccount): void {
	rolesTarget.value = sa;
	rolesOpen.value = true;
}

onMounted(load);
</script>
