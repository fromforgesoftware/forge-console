<template>
	<div class="space-y-6">
		<section class="space-y-1">
			<h1 class="text-2xl font-semibold">Apps</h1>
			<p class="text-sm text-muted-foreground">
				Configure the apps available in the control plane.
			</p>
		</section>

		<Alert v-if="error" variant="destructive">
			<AlertTitle>Something went wrong</AlertTitle>
			<AlertDescription>{{ error }}</AlertDescription>
		</Alert>

		<DataTable
			:columns="columns"
			:data-source="{ data: apps, totalCount: apps.length }"
			:pagination="pagination"
			:get-row-id="(r) => r.slug"
			@update:pagination="pagination = $event"
		/>

		<AppFormDrawer v-model:open="editOpen" :app="editing" @saved="load" />
	</div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { Pencil } from '@lucide/vue';
import {
	Alert,
	AlertTitle,
	AlertDescription,
	Button,
	DataTable,
	useDataTableState,
	Switch,
} from '@fromforgesoftware/vue-kit';
import type { ColumnDef } from '@fromforgesoftware/vue-kit';
import { adminService, type App } from '../../services/admin.service';
import { useAuthStore } from '@/app/core/auth/store';
import AppFormDrawer from './forms/AppFormDrawer.vue';

const { pagination } = useDataTableState({ key: 'admin-apps', defaultPageSize: 25 });

const auth = useAuthStore();
const canWrite = computed(() => auth.can('apps.write'));

const apps = ref<App[]>([]);
const error = ref('');
const busy = ref(false);

const editOpen = ref(false);
const editing = ref<App | null>(null);

const columns: ColumnDef<App, unknown>[] = [
	{ accessorKey: 'slug', header: 'Slug', cell: ({ row }) => h('span', row.original.slug) },
	{ accessorKey: 'name', header: 'Name', cell: ({ row }) => h('span', row.original.name) },
	{ accessorKey: 'kind', header: 'Kind', cell: ({ row }) => h('span', row.original.kind) },
	{
		id: 'enabled',
		header: 'Enabled',
		cell: ({ row }) =>
			h(Switch, {
				checked: row.original.enabled,
				disabled: busy.value || !canWrite.value,
				'onUpdate:checked': (v: boolean) => toggleEnabled(row.original, v),
			}),
	},
	{
		accessorKey: 'adminBaseURL',
		header: 'Admin base URL',
		cell: ({ row }) =>
			h('span', { class: 'text-muted-foreground truncate' }, row.original.adminBaseURL),
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
						disabled: !canWrite.value,
						onClick: () => openEdit(row.original),
					},
					() => h(Pencil, { class: 'size-4' }),
				),
			]),
	},
];

async function load(): Promise<void> {
	error.value = '';
	try {
		apps.value = await adminService.listApps();
	} catch {
		error.value = 'Failed to load apps.';
	}
}

function openEdit(app: App): void {
	editing.value = { ...app };
	editOpen.value = true;
}

async function toggleEnabled(app: App, enabled: boolean): Promise<void> {
	busy.value = true;
	try {
		const { slug, ...rest } = { ...app, enabled };
		await adminService.updateApp(slug, rest);
		await load();
	} catch {
		error.value = 'Failed to update app.';
	} finally {
		busy.value = false;
	}
}

onMounted(load);
</script>
