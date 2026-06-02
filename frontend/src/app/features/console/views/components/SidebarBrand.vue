<template>
	<div class="flex items-center gap-2 px-0.5">
		<SidebarCollapsibleShow>
			<Avatar
				name="Forge"
				size="lg"
				class="shrink-0 rounded-xl [&>span]:bg-sidebar-primary [&>span]:text-sidebar-primary-foreground"
			/>
		</SidebarCollapsibleShow>
		<SidebarCollapsibleHide class="flex items-center gap-2.5 min-w-0 flex-1">
			<Avatar
				name="Forge"
				size="lg"
				class="shrink-0 rounded-xl [&>span]:bg-sidebar-primary [&>span]:text-sidebar-primary-foreground"
			/>
			<div class="flex flex-col min-w-0 flex-1">
				<span class="text-sm font-semibold leading-none truncate">Forge</span>
				<span class="text-xs text-muted-foreground block truncate mt-0.5">Control plane</span>
			</div>
			<SidebarTrigger
				class="hidden lg:flex size-7 shrink-0 rounded-lg border border-sidebar-border/60 !bg-background text-foreground shadow-none"
			>
				<PanelLeftClose class="size-4" />
			</SidebarTrigger>
		</SidebarCollapsibleHide>
	</div>
</template>

<script setup lang="ts">
import { PanelLeftClose } from '@lucide/vue';
import {
	Avatar,
	SidebarCollapsibleHide,
	SidebarCollapsibleShow,
	SidebarTrigger,
} from '@fromforgesoftware/vue-kit';
</script>
