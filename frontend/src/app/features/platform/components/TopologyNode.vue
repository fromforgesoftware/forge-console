<template>
	<div class="relative" :data-slot="`topology-node-${data.status}`">
		<!-- Kind tab attached to the card's top-left -->
		<span
			class="pointer-events-auto absolute bottom-full left-0 inline-flex items-center gap-1.5 rounded-t-md border border-b-0 border-border bg-card px-2.5 py-1 text-xs font-medium leading-tight text-muted-foreground"
		>
			<component :is="kind.icon" class="size-3" />
			{{ kind.label }}
		</span>

		<!-- Status pill floating above-right -->
		<span
			v-if="pill"
			class="pointer-events-auto absolute bottom-full right-0 mb-1.5 inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[11px] font-medium leading-tight"
			:class="pillClass"
		>
			<component :is="pill.icon" class="size-2.5" />
			{{ pill.label }}
		</span>

		<div
			class="w-[280px] rounded-xl rounded-tl-none border bg-card shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition-colors"
			:class="borderClass"
		>
			<div class="flex items-center gap-2 px-3 pt-3 pb-2">
				<div
					class="flex size-6 shrink-0 items-center justify-center rounded-md"
					:class="iconWrapClass"
				>
					<component :is="kind.icon" class="size-3.5" />
				</div>
				<p class="flex-1 truncate text-sm font-semibold leading-tight text-foreground">
					{{ data.name }}
				</p>
				<span
					v-if="badge"
					class="inline-flex shrink-0 items-center rounded-md border border-border bg-muted/40 px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground"
				>
					{{ badge }}
				</span>
			</div>
			<div class="border-t border-border/70 px-3 py-2">
				<p class="truncate text-xs leading-snug text-muted-foreground">{{ subtitle }}</p>
			</div>
		</div>

		<Handle
			type="target"
			:position="Position.Top"
			class="!size-2.5 !border !border-border !bg-card"
		/>
		<Handle
			type="source"
			:position="Position.Bottom"
			class="!size-2.5 !border !border-border !bg-card"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Handle, Position } from '@vue-flow/core';
import {
	Server,
	Database,
	Network,
	Cloud,
	Package,
	Cpu,
	Check,
	LoaderCircle,
	TriangleAlert,
	Pause,
	CircleDashed,
} from '@lucide/vue';
import type { Component } from 'vue';
import type { TopologyNode } from '../domain/topology';
import { statusTone } from '../domain/topology';

const props = defineProps<{ data: TopologyNode }>();

const KINDS: Record<TopologyNode['kind'], { icon: Component; label: string }> = {
	service: { icon: Server, label: 'Service' },
	database: { icon: Database, label: 'Database' },
	gateway: { icon: Network, label: 'Gateway' },
	external: { icon: Cloud, label: 'External' },
	lib: { icon: Package, label: 'Library' },
	worker: { icon: Cpu, label: 'Node' },
};

const kind = computed(() => KINDS[props.data.kind] ?? KINDS.service);

const pill = computed<{ label: string; icon: Component } | null>(() => {
	switch (props.data.status) {
		case 'running':
			return { label: 'Running', icon: Check };
		case 'pending':
			return { label: 'Pending', icon: LoaderCircle };
		case 'degraded':
			return { label: 'Degraded', icon: TriangleAlert };
		case 'paused':
			return { label: 'Paused', icon: Pause };
		case 'not-deployed':
			return { label: 'Not deployed', icon: CircleDashed };
		default:
			return null;
	}
});

const tone = computed(() => statusTone(props.data.status));

const pillClass = computed(() => {
	switch (tone.value) {
		case 'success':
			return 'bg-success/12 text-success border-success/30';
		case 'info':
			return 'bg-info/12 text-info border-info/30 [&_svg]:animate-spin';
		case 'destructive':
			return 'bg-destructive/12 text-destructive border-destructive/30';
		default:
			return 'bg-muted/40 text-muted-foreground border-border';
	}
});

const borderClass = computed(() => {
	if (props.data.status === 'not-deployed') return 'border-border border-dashed';
	switch (tone.value) {
		case 'success':
			return 'border-success/45';
		case 'info':
			return 'border-info/50';
		case 'destructive':
			return 'border-destructive/45';
		default:
			return 'border-border';
	}
});

const iconWrapClass = computed(() => {
	switch (tone.value) {
		case 'success':
			return 'bg-success/10 text-success';
		case 'info':
			return 'bg-info/10 text-info';
		case 'destructive':
			return 'bg-destructive/10 text-destructive';
		default:
			return 'bg-muted text-muted-foreground';
	}
});

const badge = computed(() => props.data.engine || props.data.namespace || '');

const subtitle = computed(() => {
	const d = props.data;
	if (d.kind === 'worker') return d.meta?.region ? `region ${d.meta.region}` : 'worker node';
	if (d.kind === 'external') return d.host || 'external dependency';
	if (d.kind === 'lib') return d.language ? `${d.language} library` : 'library';
	if (d.workload?.kind || d.replicas.desired)
		return `${d.replicas.ready}/${d.replicas.desired} ready · ${d.image || d.workload?.kind || ''}`;
	return d.image || d.placement || '';
});
</script>
