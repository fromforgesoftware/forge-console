<template>
	<SidebarGroup v-if="visible">
		<SidebarGroupContent>
			<SidebarMenu>
				<SidebarMenuItem>
					<SidebarMenuButton
						tooltip="Platform"
						:is-active="active"
						@click="router.push('/platform/topology')"
					>
						<Network />
						<span>Platform</span>
					</SidebarMenuButton>
				</SidebarMenuItem>
			</SidebarMenu>
		</SidebarGroupContent>
	</SidebarGroup>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Network } from '@lucide/vue';
import {
	SidebarGroup,
	SidebarGroupContent,
	SidebarMenu,
	SidebarMenuItem,
	SidebarMenuButton,
} from '@fromforgesoftware/vue-kit';
import { useAuthStore } from '@/app/core/auth/store';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const visible = computed(() => auth.can('platform.read'));
const active = computed(() => route.path.startsWith('/platform'));
</script>
