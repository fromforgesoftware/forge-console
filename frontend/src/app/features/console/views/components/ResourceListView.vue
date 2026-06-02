<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
	DataTable,
	useDataTableState,
	Alert,
	AlertDescription,
	Skeleton,
	Badge,
	Button,
	type ColumnDef,
} from '@fromforgesoftware/vue-kit';
import { Plus } from '@lucide/vue';
import { useAuthStore } from '@/app/core/auth/store';
import { fetchCollection, type JsonApiResource } from '@/app/core/http/jsonapi';

// Generic JSON:API list renderer: fetches {apiBase}/api/{type} and renders the
// resource attributes in a vue-kit DataTable (search / sort / paginate /
// column-toggle for free). Any service exposing a JSON:API collection gets a
// baseline grid with no per-type code.
const props = defineProps<{
	apiBase: string;
	type: string;
	title: string;
	columns?: string[];
}>();

const rows = ref<JsonApiResource[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

// Only offer the empty-state "Create new" CTA when this resource actually has
// a `<basePath>/<type>/new` page registered (e.g. realms/roles/bindings).
// Routes are named `${serviceId}:${page.path}`, so the create route is the
// current list route name suffixed with `/new`.
const createRouteName = computed(() => `${String(route.name)}/new`);
const canCreate = computed(() => router.hasRoute(createRouteName.value));

const { pagination, sorting, columnVisibility } = useDataTableState({
	key: `resource:${props.type}`,
});

function display(value: unknown): string {
	if (value === null || value === undefined) return '';
	if (typeof value === 'object') return JSON.stringify(value);
	return String(value);
}

// Enum-like columns render as colored badges so state reads at a glance.
const BADGE_KEYS = new Set(['status', 'kind', 'visibility', 'subjectType', 'type']);

type BadgeVariant =
	| 'default'
	| 'secondary'
	| 'outline'
	| 'soft-success'
	| 'soft-warning'
	| 'soft-destructive';

function badgeVariant(value: string): BadgeVariant {
	switch (value.toUpperCase()) {
		case 'ACTIVE':
		case 'ACCEPTED':
		case 'PUBLIC':
			return 'soft-success';
		case 'PENDING':
		case 'SUSPENDED':
			return 'soft-warning';
		case 'REVOKED':
			return 'soft-destructive';
		case 'ARCHIVED':
		case 'EXPIRED':
			return 'secondary';
		case 'SYSTEM':
		case 'PRIVATE':
		case 'ACTOR_SET':
			return 'secondary';
		case 'CUSTOM':
			return 'default';
		default:
			return 'outline';
	}
}

// Column keys: the declared list, else the union of the first row's attributes.
const attrKeys = computed<string[]>(() => {
	if (props.columns?.length) return props.columns;
	const first = rows.value[0];
	return first ? Object.keys(first.attributes) : [];
});

const tableColumns = computed<ColumnDef<JsonApiResource, unknown>[]>(() => [
	{
		id: 'id',
		header: 'id',
		accessorFn: (row) => row.id,
		cell: ({ row }) =>
			h('span', { class: 'font-mono text-xs text-muted-foreground' }, row.original.id),
	},
	...attrKeys.value.map<ColumnDef<JsonApiResource, unknown>>((key) => ({
		id: key,
		header: key,
		accessorFn: (row) => display(row.attributes[key]),
		cell: BADGE_KEYS.has(key)
			? ({ getValue }) => {
					const v = String(getValue() ?? '');
					return v ? h(Badge, { variant: badgeVariant(v), size: 'sm' }, () => v) : '';
				}
			: undefined,
	})),
]);

onMounted(async () => {
	try {
		rows.value = await fetchCollection(props.apiBase, props.type, auth.token);
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'request failed';
	} finally {
		loading.value = false;
	}
});
</script>

<template>
	<section class="space-y-4">
		<header v-if="title" class="flex items-end justify-between gap-4">
			<div class="space-y-0.5">
				<h1 class="text-xl font-semibold">{{ title }}</h1>
				<p v-if="!loading && !error" class="text-sm text-muted-foreground">
					{{ rows.length }} {{ rows.length === 1 ? 'item' : 'items' }}
				</p>
			</div>
			<Button v-if="canCreate" size="sm" @click="router.push({ name: createRouteName })">
				<Plus class="size-4" />
				New
			</Button>
		</header>

		<Alert v-if="error" variant="destructive">
			<AlertDescription>Failed to load: {{ error }}</AlertDescription>
		</Alert>

		<div v-else-if="loading" class="space-y-2">
			<Skeleton class="h-9 w-full" />
			<Skeleton v-for="n in 5" :key="n" class="h-12 w-full" />
		</div>

		<DataTable
			v-else
			:columns="tableColumns"
			:data-source="{ data: rows, totalCount: rows.length }"
			:pagination="pagination"
			:sorting="sorting"
			:column-visibility="columnVisibility"
			:get-row-id="(r: JsonApiResource) => r.id"
			:empty-title="`No ${type} found`"
			:empty-description="canCreate ? 'Get started by creating your first item.' : ''"
			:empty-action-label="canCreate ? 'Create new' : ''"
			@empty-action="router.push({ name: createRouteName })"
			@update:pagination="pagination = $event"
			@update:sorting="sorting = $event"
			@update:column-visibility="columnVisibility = $event"
		/>
	</section>
</template>
