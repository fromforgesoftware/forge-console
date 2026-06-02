<template>
	<div class="space-y-8">
		<section class="space-y-1">
			<h1 class="text-2xl font-semibold">Security</h1>
			<p class="text-sm text-muted-foreground">Update the password for your user account.</p>
		</section>

		<section class="space-y-3">
			<h2 class="text-xs font-medium uppercase tracking-wide text-muted-foreground">Password</h2>
			<div class="divide-y divide-border rounded-xl border border-border bg-card">
				<!-- Current password -->
				<SettingsRow
					title="Current password"
					description="Confirm the password you use to sign in today."
				>
					<div class="relative w-64">
						<Input
							v-model="current"
							:type="showCurrent ? 'text' : 'password'"
							autocomplete="current-password"
							placeholder="Current password"
							class="pr-10"
						/>
						<button
							type="button"
							class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
							@click="showCurrent = !showCurrent"
						>
							<Eye v-if="!showCurrent" class="size-4" />
							<EyeOff v-else class="size-4" />
						</button>
					</div>
				</SettingsRow>

				<!-- New password -->
				<SettingsRow
					title="New password"
					description="Use a strong password you don't use elsewhere."
				>
					<div class="relative w-64">
						<Input
							v-model="next"
							:type="showNext ? 'text' : 'password'"
							autocomplete="new-password"
							placeholder="New password"
							class="pr-10"
						/>
						<button
							type="button"
							class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
							@click="showNext = !showNext"
						>
							<Eye v-if="!showNext" class="size-4" />
							<EyeOff v-else class="size-4" />
						</button>
					</div>
				</SettingsRow>

				<!-- Confirm password -->
				<SettingsRow
					title="Confirm new password"
					description="Re-enter the new password to confirm."
				>
					<div class="relative w-64">
						<Input
							v-model="confirm"
							:type="showConfirm ? 'text' : 'password'"
							autocomplete="new-password"
							placeholder="Confirm password"
							class="pr-10"
							:variant="passwordsMatch ? 'default' : 'error'"
						/>
						<button
							type="button"
							class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
							@click="showConfirm = !showConfirm"
						>
							<Eye v-if="!showConfirm" class="size-4" />
							<EyeOff v-else class="size-4" />
						</button>
					</div>
				</SettingsRow>

				<!-- Strength indicator -->
				<div v-if="next.length > 0" class="space-y-3 px-6 py-4">
					<div class="flex items-center justify-between">
						<span class="text-xs font-medium text-muted-foreground">Password strength</span>
						<span class="text-xs font-medium" :class="strengthColor">{{ strengthLabel }}</span>
					</div>
					<Progress :model-value="strengthPercent" class="h-1.5" />
					<ul class="grid grid-cols-2 gap-1.5">
						<li
							v-for="req in requirements"
							:key="req.label"
							class="flex items-center gap-1.5 text-xs"
							:class="req.met ? 'text-success' : 'text-muted-foreground'"
						>
							<Check v-if="req.met" class="size-3" />
							<X v-else class="size-3" />
							{{ req.label }}
						</li>
					</ul>
				</div>

				<!-- Mismatch error -->
				<div v-if="!passwordsMatch" class="px-6 py-3">
					<p class="text-xs text-destructive">Passwords do not match.</p>
				</div>
			</div>
		</section>

		<div class="flex justify-end">
			<Button :disabled="!canSubmit" @click="save">
				{{ saving ? 'Saving…' : 'Change password' }}
			</Button>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { Button, Input, Progress, toast } from '@fromforgesoftware/vue-kit';
import { Eye, EyeOff, Check, X } from '@lucide/vue';
import { accountService } from '../../services/account.service';
import SettingsRow from '../SettingsRow.vue';

const current = ref('');
const next = ref('');
const confirm = ref('');
const showCurrent = ref(false);
const showNext = ref(false);
const showConfirm = ref(false);
const saving = ref(false);

const requirements = computed(() => [
	{ label: 'At least 8 characters', met: next.value.length >= 8 },
	{ label: 'A lowercase letter', met: /[a-z]/.test(next.value) },
	{ label: 'An uppercase letter', met: /[A-Z]/.test(next.value) },
	{ label: 'A number', met: /[0-9]/.test(next.value) },
	{ label: 'A special character', met: /[^A-Za-z0-9]/.test(next.value) },
]);

const strengthScore = computed(() => requirements.value.filter((r) => r.met).length);

const strengthLabel = computed(() => {
	if (next.value.length === 0) return '';
	if (strengthScore.value <= 1) return 'Weak';
	if (strengthScore.value <= 3) return 'Fair';
	if (strengthScore.value === 4) return 'Good';
	return 'Strong';
});

const strengthPercent = computed(() =>
	next.value.length === 0 ? 0 : (strengthScore.value / 5) * 100,
);

const strengthColor = computed(() => {
	if (strengthScore.value <= 1) return 'text-destructive';
	if (strengthScore.value <= 3) return 'text-warning';
	return 'text-success';
});

const passwordsMatch = computed(() => {
	if (confirm.value.length === 0) return true;
	return next.value === confirm.value;
});

// The backend enforces a minimum of 8 characters; require that here, plus a
// confirmation match. The full strength meter is advisory.
const canSubmit = computed(
	() =>
		!!current.value &&
		next.value.length >= 8 &&
		confirm.value.length > 0 &&
		passwordsMatch.value &&
		!saving.value,
);

async function save(): Promise<void> {
	saving.value = true;
	try {
		await accountService.changePassword(current.value, next.value);
		toast.success('Password updated');
		current.value = '';
		next.value = '';
		confirm.value = '';
	} catch {
		toast.error('Could not change password. Check your current password and try again.');
	} finally {
		saving.value = false;
	}
}
</script>
