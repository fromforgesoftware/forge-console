<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue';
import {
	Alert,
	AlertDescription,
	Button,
	Checkbox,
	FormField,
	Input,
} from '@fromforgesoftware/vue-kit';
import { useAuthStore } from '@/app/core/auth/store';
import { fetchCollection, createResource } from '@/app/core/http/jsonapi';
import { roleAttributes } from '@/app/features/console/application/forms';

// Custom-role builder: pick permissions of a resource_type and POST /api/roles.
// The backend accepts a synthetic `permissions` attribute that seeds the
// role's permission set atomically.
const props = defineProps<{ apiBase: string }>();

const auth = useAuthStore();
const form = reactive({ realmId: '', name: '', resourceType: '', kind: 'CUSTOM' });
const allPermissions = ref<{ id: string; resourceType: string }[]>([]);
const selected = ref<Set<string>>(new Set());
const error = ref<string | null>(null);
const createdId = ref<string | null>(null);
const submitting = ref(false);

onMounted(async () => {
	try {
		const rows = await fetchCollection(props.apiBase, 'permissions', auth.token);
		allPermissions.value = rows.map((r) => ({
			id: r.id,
			resourceType: String(r.attributes.resourceType ?? ''),
		}));
	} catch {
		// permission catalog optional for rendering; surfaced on submit if empty
	}
});

function toggle(id: string, checked: boolean) {
	const next = new Set(selected.value);
	if (checked) next.add(id);
	else next.delete(id);
	selected.value = next;
}

async function submit() {
	error.value = null;
	if (!form.realmId || !form.name || !form.resourceType) {
		error.value = 'realm, name and resource type are required';
		return;
	}
	submitting.value = true;
	try {
		const created = await createResource(
			props.apiBase,
			'roles',
			roleAttributes(form, [...selected.value]),
			auth.token,
		);
		createdId.value = created.id;
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'create failed';
	} finally {
		submitting.value = false;
	}
}
</script>

<template>
	<section class="mx-auto w-full max-w-2xl space-y-4">
		<h1 class="text-xl font-semibold">New custom role</h1>

		<Alert v-if="createdId" variant="success">
			<AlertDescription>
				Created role <span class="font-mono">{{ createdId }}</span
				>.
			</AlertDescription>
		</Alert>

		<form class="space-y-4" @submit.prevent="submit">
			<FormField label="Realm ID" for="role-realm">
				<Input id="role-realm" v-model="form.realmId" />
			</FormField>
			<FormField label="Name" for="role-name">
				<Input id="role-name" v-model="form.name" />
			</FormField>
			<FormField label="Resource type" for="role-resource-type">
				<Input id="role-resource-type" v-model="form.resourceType" />
			</FormField>

			<FormField label="Permissions">
				<div class="w-full space-y-1">
					<p v-if="allPermissions.length === 0" class="text-sm text-muted-foreground">
						No permissions in catalog.
					</p>
					<label
						v-for="p in allPermissions"
						:key="p.id"
						class="flex items-center gap-2 py-0.5 text-sm"
					>
						<Checkbox
							:checked="selected.has(p.id)"
							@update:checked="(v) => toggle(p.id, v === true)"
						/>
						<span class="font-mono">{{ p.id }}</span>
					</label>
				</div>
			</FormField>

			<Alert v-if="error" variant="destructive">
				<AlertDescription>{{ error }}</AlertDescription>
			</Alert>

			<div class="flex justify-end">
				<Button type="submit" :disabled="submitting">
					{{ submitting ? 'Saving…' : 'Create role' }}
				</Button>
			</div>
		</form>
	</section>
</template>
