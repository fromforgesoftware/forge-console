<template>
	<div class="space-y-1">
		<h1 class="text-2xl font-semibold">Profile</h1>
	</div>

	<!-- Info banner -->
	<Alert variant="info" class="my-6">
		<Info />
		<AlertDescription>
			Your profile information is shared across the control plane.
		</AlertDescription>
	</Alert>

	<!-- Avatar -->
	<div class="mb-8 flex flex-col items-center space-y-2">
		<Avatar
			:name="displayName || email"
			size="xl"
			class="!size-20 shrink-0 rounded-full [&>span]:bg-primary [&>span]:text-primary-foreground [&>span]:text-2xl"
		/>
	</div>

	<!-- Display name -->
	<div class="space-y-2">
		<Label for="profile-display-name">Display name</Label>
		<Input id="profile-display-name" v-model="displayName" placeholder="Your name" />
	</div>

	<!-- Email (managed by an administrator) -->
	<div class="mt-4 space-y-2">
		<Label for="profile-email">Email</Label>
		<Input id="profile-email" :model-value="email" disabled />
	</div>

	<p v-if="error" class="mt-3 text-xs text-destructive">{{ error }}</p>

	<div class="mt-6 flex justify-end">
		<Button :disabled="saving || !canSave" @click="save">
			{{ saving ? 'Saving…' : 'Save changes' }}
		</Button>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import {
	Alert,
	AlertDescription,
	Avatar,
	Button,
	Input,
	Label,
	toast,
} from '@fromforgesoftware/vue-kit';
import { Info } from '@lucide/vue';
import { useAuthStore } from '@/app/core/auth/store';
import { accountService } from '../../services/account.service';

const auth = useAuthStore();

const displayName = ref(auth.user?.displayName ?? '');
const email = ref(auth.user?.email ?? '');
const saving = ref(false);
const error = ref('');

const canSave = computed(
	() => displayName.value.trim() !== '' && displayName.value.trim() !== auth.user?.displayName,
);

async function save(): Promise<void> {
	saving.value = true;
	error.value = '';
	try {
		const updated = await accountService.updateProfile(displayName.value.trim());
		auth.user = updated;
		displayName.value = updated.displayName;
		toast.success('Profile updated');
	} catch {
		error.value = 'Something went wrong while saving your profile.';
	} finally {
		saving.value = false;
	}
}
</script>
