<template>
	<div class="flex min-h-0 flex-1 flex-col">
		<div class="min-h-0 flex-1 space-y-2 overflow-y-auto p-4">
			<label
				v-for="role in roles"
				:key="role.slug"
				class="flex cursor-pointer items-center gap-2.5 text-sm"
			>
				<Checkbox
					:checked="selected.includes(role.slug)"
					@update:checked="(v) => toggle(role.slug, v === true)"
				/>
				<span>{{ role.name }}</span>
			</label>
			<p v-if="roles.length === 0" class="text-sm text-muted-foreground">No roles defined.</p>
			<p v-if="error" class="text-sm text-destructive">{{ error }}</p>
		</div>

		<DrawerFooter>
			<Button variant="outline" @click="emit('cancel')">Cancel</Button>
			<Button :disabled="busy || !canWrite" @click="submit">Save</Button>
		</DrawerFooter>
	</div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { Button, Checkbox, DrawerFooter } from '@fromforgesoftware/vue-kit';
import { adminService, type ServiceAccount, type Role } from '../../../services/admin.service';

const props = defineProps<{ account: ServiceAccount; roles: Role[]; canWrite: boolean }>();
const emit = defineEmits<{ saved: []; cancel: [] }>();

const selected = ref<string[]>([]);
const busy = ref(false);
const error = ref('');

function toggle(slug: string, on: boolean): void {
	if (on) {
		if (!selected.value.includes(slug)) selected.value.push(slug);
	} else {
		selected.value = selected.value.filter((s) => s !== slug);
	}
}

async function submit(): Promise<void> {
	busy.value = true;
	error.value = '';
	try {
		await adminService.setServiceAccountRoles(props.account.id, selected.value);
		emit('saved');
	} catch {
		error.value = 'Failed to save roles.';
	} finally {
		busy.value = false;
	}
}

onMounted(async () => {
	try {
		const assigned = await adminService.getServiceAccountRoles(props.account.id);
		selected.value = assigned.map((r) => r.slug);
	} catch {
		error.value = 'Failed to load assigned roles.';
	}
});
</script>
