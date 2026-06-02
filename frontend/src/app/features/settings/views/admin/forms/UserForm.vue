<template>
	<div class="flex min-h-0 flex-1 flex-col">
		<div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
			<FormField label="Email" for="user-email">
				<Input id="user-email" v-model="form.email" type="email" placeholder="user@example.com" />
			</FormField>
			<FormField label="Display name" for="user-name">
				<Input id="user-name" v-model="form.displayName" placeholder="Jane Doe" />
			</FormField>
			<FormField
				label="Password"
				description="Leave blank to send an invite instead."
				for="user-password"
			>
				<Input
					id="user-password"
					v-model="form.password"
					type="password"
					autocomplete="new-password"
				/>
			</FormField>

			<p v-if="error" class="text-sm text-destructive">{{ error }}</p>
		</div>

		<DrawerFooter>
			<Button variant="outline" @click="emit('cancel')">Cancel</Button>
			<Button :disabled="busy || !form.email || !form.displayName" @click="submit">Create</Button>
		</DrawerFooter>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { Button, DrawerFooter, FormField, Input } from '@fromforgesoftware/vue-kit';
import { adminService } from '../../../services/admin.service';

const emit = defineEmits<{ saved: []; cancel: [] }>();

const form = ref({ email: '', displayName: '', password: '' });
const busy = ref(false);
const error = ref('');

async function submit(): Promise<void> {
	busy.value = true;
	error.value = '';
	try {
		await adminService.createUser({
			email: form.value.email,
			displayName: form.value.displayName,
			password: form.value.password || undefined,
		});
		emit('saved');
	} catch {
		error.value = 'Failed to create user.';
	} finally {
		busy.value = false;
	}
}
</script>
