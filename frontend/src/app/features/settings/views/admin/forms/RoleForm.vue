<template>
	<div class="flex min-h-0 flex-1 flex-col">
		<div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
			<FormField label="Slug" for="role-slug">
				<Input id="role-slug" v-model="form.slug" placeholder="operations" />
			</FormField>
			<FormField label="Name" for="role-name">
				<Input id="role-name" v-model="form.name" placeholder="Operations" />
			</FormField>

			<div class="space-y-3">
				<p class="text-sm font-medium">Permissions</p>
				<div v-for="group in groups" :key="group.resourceType" class="space-y-2">
					<p class="text-xs font-medium text-muted-foreground">{{ group.resourceType }}</p>
					<label
						v-for="perm in group.permissions"
						:key="perm.id"
						class="flex cursor-pointer items-center gap-2.5 text-sm"
					>
						<Checkbox
							:checked="form.permissions.includes(perm.id)"
							@update:checked="(v) => togglePermission(perm.id, v === true)"
						/>
						<span>{{ perm.verb }}</span>
						<span v-if="perm.description" class="text-xs text-muted-foreground">
							{{ perm.description }}
						</span>
					</label>
				</div>
				<p v-if="groups.length === 0" class="text-sm text-muted-foreground">
					No permissions available.
				</p>
			</div>

			<p v-if="error" class="text-sm text-destructive">{{ error }}</p>
		</div>

		<DrawerFooter>
			<Button variant="outline" @click="emit('cancel')">Cancel</Button>
			<Button :disabled="busy || !form.slug || !form.name" @click="submit">Create</Button>
		</DrawerFooter>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { Button, Checkbox, DrawerFooter, FormField, Input } from '@fromforgesoftware/vue-kit';
import { adminService, type Permission } from '../../../services/admin.service';

const props = defineProps<{ permissions: Permission[] }>();
const emit = defineEmits<{ saved: []; cancel: [] }>();

const form = ref({ slug: '', name: '', permissions: [] as string[] });
const busy = ref(false);
const error = ref('');

interface Group {
	resourceType: string;
	permissions: Permission[];
}

const groups = computed<Group[]>(() => {
	const byType = new Map<string, Permission[]>();
	for (const perm of props.permissions) {
		const list = byType.get(perm.resourceType) ?? [];
		list.push(perm);
		byType.set(perm.resourceType, list);
	}
	return [...byType.entries()].map(([resourceType, permissions]) => ({
		resourceType,
		permissions,
	}));
});

function togglePermission(id: string, on: boolean): void {
	if (on) {
		if (!form.value.permissions.includes(id)) form.value.permissions.push(id);
	} else {
		form.value.permissions = form.value.permissions.filter((p) => p !== id);
	}
}

async function submit(): Promise<void> {
	busy.value = true;
	error.value = '';
	try {
		await adminService.createRole({
			slug: form.value.slug,
			name: form.value.name,
			permissions: form.value.permissions,
		});
		emit('saved');
	} catch {
		error.value = 'Failed to create role.';
	} finally {
		busy.value = false;
	}
}
</script>
