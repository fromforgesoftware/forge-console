<template>
	<DropdownMenu>
		<div class="flex items-center gap-2.5 px-0.5">
			<SidebarCollapsibleShow>
				<DropdownMenuTrigger
					class="rounded-xl cursor-pointer focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
					:aria-label="`Open menu for ${displayName}`"
				>
					<Avatar
						:name="displayName"
						size="lg"
						class="shrink-0 rounded-xl [&>span]:bg-primary [&>span]:text-primary-foreground"
					/>
				</DropdownMenuTrigger>
			</SidebarCollapsibleShow>
			<SidebarCollapsibleHide class="flex items-center gap-2.5 min-w-0 flex-1">
				<Avatar
					:name="displayName"
					size="lg"
					class="shrink-0 rounded-xl [&>span]:bg-primary [&>span]:text-primary-foreground"
				/>
				<div class="min-w-0 flex-1">
					<DropdownMenuTrigger as-child>
						<button
							type="button"
							class="inline-flex items-center gap-1 w-fit max-w-full cursor-pointer rounded-sm focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
							:aria-label="`Open menu for ${displayName}`"
						>
							<span class="text-sm font-semibold leading-none truncate">{{ displayName }}</span>
							<ChevronDown class="size-3 shrink-0 text-foreground" :stroke-width="3" />
						</button>
					</DropdownMenuTrigger>
					<span class="text-xs text-muted-foreground block truncate">{{ email }}</span>
				</div>
			</SidebarCollapsibleHide>
		</div>

		<DropdownMenuContent class="w-64" side="top" align="start" :side-offset="4">
			<div class="flex items-center gap-2.5 px-3 py-2.5">
				<Avatar
					:name="displayName"
					size="default"
					class="shrink-0 rounded-lg [&>span]:bg-primary [&>span]:text-primary-foreground"
				/>
				<div class="min-w-0 flex-1">
					<p class="text-sm font-semibold truncate">
						{{ displayName }}
						<Tooltip
							v-if="roles.length"
							:content="roles.join(', ')"
							side="right"
							:delay-duration="0"
						>
							<Shield
								class="inline size-3.5 text-muted-foreground cursor-default align-baseline ml-0.5"
							/>
						</Tooltip>
					</p>
					<p class="text-xs text-muted-foreground truncate">{{ email }}</p>
				</div>
			</div>
			<DropdownMenuSeparator />
			<DropdownMenuItem @select="router.push('/settings/account/profile')">
				<Settings class="mr-2 size-4" />
				<span>Settings</span>
			</DropdownMenuItem>
			<DropdownMenuSeparator />
			<DropdownMenuItem @select="onLogout">
				<LogOut class="mr-2 size-4" />
				<span>Log out</span>
			</DropdownMenuItem>
		</DropdownMenuContent>
	</DropdownMenu>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { ChevronDown, LogOut, Settings, Shield } from '@lucide/vue';
import {
	Avatar,
	DropdownMenu,
	DropdownMenuTrigger,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	SidebarCollapsibleShow,
	SidebarCollapsibleHide,
	Tooltip,
} from '@fromforgesoftware/vue-kit';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/app/core/auth/store';

const router = useRouter();
const auth = useAuthStore();

const displayName = computed(() => auth.user?.displayName || 'User');
const email = computed(() => auth.user?.email ?? '');
const roles = computed(() => auth.user?.roles ?? []);

async function onLogout() {
	await auth.logout();
	router.push({ name: 'login' });
}
</script>
