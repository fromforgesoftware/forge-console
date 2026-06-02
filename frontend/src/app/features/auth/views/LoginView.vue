<template>
	<Card class="w-full max-w-sm">
		<CardHeader class="items-center text-center gap-2">
			<div class="flex h-12 w-12 items-center justify-center rounded-full border border-border">
				<Hexagon class="h-6 w-6 text-primary" />
			</div>
			<CardTitle class="text-2xl font-bold tracking-tight">Sign in</CardTitle>
			<CardDescription>Sign in to the Forge control plane.</CardDescription>
		</CardHeader>

		<CardContent>
			<div role="alert" aria-live="assertive" class="sr-only">
				<template v-if="errorMessage">{{ errorMessage }}</template>
			</div>

			<form class="space-y-4" @submit.prevent="onSubmit">
				<div class="space-y-1.5">
					<Label for="login-email">Email</Label>
					<Input
						id="login-email"
						v-model="email"
						type="email"
						placeholder="you@example.com"
						autocomplete="email"
						:disabled="loading"
					/>
				</div>

				<div class="space-y-1.5">
					<Label for="login-password">Password</Label>
					<Input
						id="login-password"
						v-model="password"
						type="password"
						placeholder="••••••••"
						autocomplete="current-password"
						:disabled="loading"
					/>
				</div>

				<p v-if="errorMessage" class="text-xs text-destructive m-0" role="alert">
					{{ errorMessage }}
				</p>

				<Button type="submit" class="w-full" :disabled="!canSubmit" :loading="loading">
					Continue
				</Button>
			</form>

			<template v-if="providers.length > 0">
				<div class="my-6 flex items-center gap-3">
					<Divider class="flex-1" />
					<span class="text-xs text-muted-foreground">Or continue with</span>
					<Divider class="flex-1" />
				</div>
				<div class="flex flex-col gap-2">
					<Button
						v-for="p in providers"
						:key="p.id"
						variant="outline"
						class="w-full"
						@click="startOidc(p.id)"
					>
						{{ p.name }}
					</Button>
				</div>
			</template>
		</CardContent>
	</Card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Hexagon } from '@lucide/vue';
import {
	Card,
	CardHeader,
	CardTitle,
	CardDescription,
	CardContent,
	Button,
	Input,
	Label,
	Divider,
} from '@fromforgesoftware/vue-kit';
import { useAuthStore, type AuthProvider } from '@/app/core/auth/store';
import { useAppsStore } from '@/app/features/console/stores/apps';
import { useRealmStore } from '@/app/features/console/stores/realm';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const email = ref('');
const password = ref('');
const loading = ref(false);
const errorMessage = ref('');
const providers = ref<AuthProvider[]>([]);

const canSubmit = computed(
	() => email.value.trim() !== '' && password.value !== '' && !loading.value,
);

function redirectTarget(): string {
	const redirect = route.query.redirect;
	return typeof redirect === 'string' && redirect ? redirect : '/';
}

async function onSubmit() {
	if (!canSubmit.value) return;
	loading.value = true;
	errorMessage.value = '';
	try {
		const ok = await auth.login(email.value.trim(), password.value);
		if (!ok) {
			errorMessage.value = 'Invalid email or password.';
			return;
		}
		await Promise.all([useAppsStore().load(), useRealmStore().load()]);
		router.push(redirectTarget());
	} catch {
		errorMessage.value = 'Something went wrong. Please try again.';
	} finally {
		loading.value = false;
	}
}

function startOidc(id: string) {
	window.location.href = `/api/auth/oidc/${id}/start`;
}

onMounted(async () => {
	const err = route.query.error;
	if (err === 'not_a_user') {
		errorMessage.value = 'That account is not a Forge user.';
	} else if (typeof err === 'string' && err) {
		errorMessage.value = 'Sign-in failed. Please try again.';
	}
	providers.value = await auth.providers();
});
</script>
