<template>
	<Drawer v-model:open="openModel">
		<DrawerPanel size="md" class="flex flex-col">
			<DrawerHeader class="border-b border-border">
				<div class="flex items-center gap-3 pr-6">
					<div
						class="flex size-10 shrink-0 items-center justify-center rounded-xl"
						:class="iconWrapClass"
					>
						<component :is="kindMeta.icon" class="size-5" />
					</div>
					<div class="min-w-0 flex-1">
						<DrawerTitle class="truncate text-base leading-tight">{{
							node?.name ?? 'Node'
						}}</DrawerTitle>
						<p class="truncate text-xs text-muted-foreground">{{ subtitle }}</p>
					</div>
					<Badge v-if="node" :variant="statusVariant" class="shrink-0 capitalize">{{
						statusLabel
					}}</Badge>
				</div>
			</DrawerHeader>

			<div v-if="node" class="flex min-h-0 flex-1 flex-col gap-5 overflow-y-auto p-4">
				<!-- Replicas headline for workloads -->
				<div
					v-if="node.workload?.kind"
					class="flex items-center justify-between rounded-xl border border-border bg-muted/20 px-4 py-3"
				>
					<div>
						<p class="text-2xl font-semibold tabular-nums leading-none text-foreground">
							{{ node.replicas.ready
							}}<span class="text-muted-foreground">/{{ node.replicas.desired }}</span>
						</p>
						<p class="mt-1 text-xs text-muted-foreground">pods ready</p>
					</div>
					<div class="flex items-center gap-1.5">
						<span
							v-for="(dot, i) in replicaDots"
							:key="i"
							class="size-2.5 rounded-full"
							:class="dot ? toneBg : 'bg-border'"
						/>
					</div>
				</div>

				<!-- Details -->
				<dl class="overflow-hidden rounded-xl border border-border text-sm">
					<div
						v-for="(row, i) in detailRows"
						:key="row.label"
						class="flex items-start gap-3 px-4 py-2.5"
						:class="i % 2 ? 'bg-muted/20' : ''"
					>
						<dt class="w-24 shrink-0 text-muted-foreground">{{ row.label }}</dt>
						<dd
							class="min-w-0 flex-1 break-words text-right font-medium text-foreground"
							:class="row.mono ? 'font-mono text-xs' : ''"
						>
							{{ row.value }}
						</dd>
					</div>
				</dl>

				<!-- Workload actions -->
				<section
					v-if="hasWorkloadActions && (canManage || canDelete)"
					class="flex flex-col gap-2.5"
				>
					<h3 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
						Actions
					</h3>
					<div class="grid grid-cols-2 gap-2">
						<Button
							v-if="canManage"
							variant="outline"
							size="sm"
							class="justify-start"
							:disabled="busy"
							@click="run('restart', () => restartWorkload(target))"
						>
							<RotateCcw class="size-3.5" /> Restart
						</Button>
						<Button
							v-if="canManage && isDeployment && node.status !== 'paused'"
							variant="outline"
							size="sm"
							class="justify-start"
							:disabled="busy"
							@click="run('pause', () => pauseWorkload(target))"
						>
							<Pause class="size-3.5" /> Pause
						</Button>
						<Button
							v-if="canManage && isDeployment && node.status === 'paused'"
							variant="outline"
							size="sm"
							class="justify-start"
							:disabled="busy"
							@click="run('resume', () => resumeWorkload(target))"
						>
							<Play class="size-3.5" /> Resume
						</Button>
						<Button
							v-if="canDelete"
							variant="outline"
							size="sm"
							class="justify-start text-destructive hover:bg-destructive/10 hover:text-destructive"
							:disabled="busy"
							@click="
								confirmRun('delete', `Delete ${node.name}? This removes the workload.`, () =>
									deleteWorkload(target),
								)
							"
						>
							<Trash2 class="size-3.5" /> Delete
						</Button>
					</div>
					<div
						v-if="canManage"
						class="flex items-end gap-2 rounded-xl border border-border bg-muted/20 p-3"
					>
						<label class="flex flex-1 flex-col gap-1">
							<span class="text-xs text-muted-foreground">Scale replicas</span>
							<input
								v-model.number="replicas"
								type="number"
								min="0"
								class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm tabular-nums focus:border-primary focus:outline-none"
							/>
						</label>
						<Button
							variant="default"
							size="sm"
							:disabled="busy || replicas === node.replicas.desired"
							@click="run('scale', () => scaleWorkload(target, replicas))"
						>
							<Scaling class="size-3.5" /> Apply
						</Button>
					</div>
				</section>

				<!-- Node actions -->
				<section v-if="node.kind === 'worker' && canCluster" class="flex flex-col gap-2.5">
					<h3 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
						Node actions
					</h3>
					<div class="grid grid-cols-2 gap-2">
						<Button
							variant="outline"
							size="sm"
							class="justify-start"
							:disabled="busy"
							@click="run('cordon', () => nodeAction(node!.name, 'cordon'))"
						>
							<Ban class="size-3.5" /> Cordon
						</Button>
						<Button
							variant="outline"
							size="sm"
							class="justify-start"
							:disabled="busy"
							@click="run('uncordon', () => nodeAction(node!.name, 'uncordon'))"
						>
							<Play class="size-3.5" /> Uncordon
						</Button>
						<Button
							variant="outline"
							size="sm"
							class="col-span-2 justify-start text-destructive hover:bg-destructive/10 hover:text-destructive"
							:disabled="busy"
							@click="
								confirmRun('drain', `Drain ${node.name}? Pods will be evicted.`, () =>
									nodeAction(node!.name, 'drain'),
								)
							"
						>
							<TriangleAlert class="size-3.5" /> Drain node
						</Button>
					</div>
				</section>

				<!-- Logs -->
				<section v-if="logsAvailable" class="flex min-h-0 flex-1 flex-col gap-2">
					<div class="flex items-center justify-between">
						<h3 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
							Logs
						</h3>
						<Button variant="outline" size="sm" @click="toggleLogs">
							<component :is="streaming ? Square : ScrollText" class="size-3.5" />
							{{ streaming ? 'Stop' : 'Stream' }}
						</Button>
					</div>
					<pre
						v-if="streaming"
						ref="logBox"
						class="min-h-40 flex-1 overflow-auto whitespace-pre-wrap break-all rounded-lg bg-neutral-950 p-3 font-mono text-[11px] leading-relaxed text-neutral-200"
						>{{ logLines.join('\n') || '› waiting for output…' }}</pre
					>
				</section>
			</div>
		</DrawerPanel>
	</Drawer>
</template>

<script setup lang="ts">
import { computed, nextTick, onUnmounted, ref, watch } from 'vue';
import {
	Drawer,
	DrawerPanel,
	DrawerHeader,
	DrawerTitle,
	Button,
	Badge,
	useToast,
} from '@fromforgesoftware/vue-kit';
import {
	RotateCcw,
	Pause,
	Play,
	Trash2,
	Scaling,
	Ban,
	TriangleAlert,
	ScrollText,
	Square,
	Server,
	Database,
	Network,
	Cloud,
	Package,
	Cpu,
} from '@lucide/vue';
import type { Component } from 'vue';
import { useAuthStore } from '@/app/core/auth/store';
import type { TopologyNode, Topology } from '../domain/topology';
import { statusTone } from '../domain/topology';
import {
	restartWorkload,
	pauseWorkload,
	resumeWorkload,
	deleteWorkload,
	scaleWorkload,
	nodeAction,
} from '../data/topology';
import { useLogStream } from '../data/stream';

const props = defineProps<{ open: boolean; node: TopologyNode | null }>();
const emit = defineEmits<{ 'update:open': [value: boolean]; changed: [topology: Topology] }>();

const auth = useAuthStore();
const toast = useToast();
const busy = ref(false);
const replicas = ref(0);
const logBox = ref<HTMLElement | null>(null);

const openModel = computed({ get: () => props.open, set: (v) => emit('update:open', v) });

const { lines: logLines, start: startLogs, stop: stopLogs } = useLogStream();
const streaming = ref(false);
const logsAvailable = computed(
	() => !!props.node?.workload?.kind && (props.node?.pods?.length ?? 0) > 0,
);

function toggleLogs(): void {
	if (streaming.value) {
		stopLogs();
		streaming.value = false;
		return;
	}
	const pod = props.node?.pods?.[0];
	if (!pod || !props.node) return;
	startLogs({ namespace: props.node.namespace ?? 'default', pod });
	streaming.value = true;
}

watch(
	logLines,
	async () => {
		await nextTick();
		if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight;
	},
	{ deep: true },
);

function resetLogs(): void {
	stopLogs();
	streaming.value = false;
}

watch(
	() => props.node,
	(n) => {
		replicas.value = n?.replicas.desired ?? 0;
		resetLogs();
	},
);
watch(
	() => props.open,
	(open) => {
		if (!open) resetLogs();
	},
);
onUnmounted(stopLogs);

const KIND_ICONS: Record<TopologyNode['kind'], { icon: Component; label: string }> = {
	service: { icon: Server, label: 'Service' },
	database: { icon: Database, label: 'Database' },
	gateway: { icon: Network, label: 'Gateway' },
	external: { icon: Cloud, label: 'External' },
	lib: { icon: Package, label: 'Library' },
	worker: { icon: Cpu, label: 'Node' },
};
const kindMeta = computed(() => KIND_ICONS[props.node?.kind ?? 'service'] ?? KIND_ICONS.service);

const isDeployment = computed(() => props.node?.workload?.kind === 'Deployment');
const hasWorkloadActions = computed(() => !!props.node?.workload?.kind);
const canManage = computed(() => auth.can('platform.manage'));
const canDelete = computed(() => auth.can('platform:workload.delete'));
const canCluster = computed(() => auth.can('platform:cluster.manage'));

const target = computed(() => ({
	kind: props.node?.workload?.kind ?? '',
	namespace: props.node?.workload?.namespace ?? '',
	name: props.node?.workload?.name ?? '',
}));

const tone = computed(() => statusTone(props.node?.status));
const toneBg = computed(() => {
	switch (tone.value) {
		case 'success':
			return 'bg-success';
		case 'info':
			return 'bg-info';
		case 'destructive':
			return 'bg-destructive';
		default:
			return 'bg-border';
	}
});
const iconWrapClass = computed(() => {
	switch (tone.value) {
		case 'success':
			return 'bg-success/12 text-success';
		case 'info':
			return 'bg-info/12 text-info';
		case 'destructive':
			return 'bg-destructive/12 text-destructive';
		default:
			return 'bg-muted text-muted-foreground';
	}
});

const statusLabel = computed(() => (props.node?.status ?? '').replace('-', ' '));
const statusVariant = computed(() => {
	switch (props.node?.status) {
		case 'running':
			return 'success';
		case 'pending':
			return 'info';
		case 'degraded':
			return 'destructive';
		default:
			return 'secondary';
	}
});

const subtitle = computed(() => {
	const n = props.node;
	if (!n) return '';
	const parts = [kindMeta.value.label];
	if (n.namespace) parts.push(n.namespace);
	else if (n.placement === 'external') parts.push('external');
	return parts.join(' · ');
});

const replicaDots = computed(() => {
	const n = props.node;
	if (!n?.workload?.kind) return [];
	const total = Math.min(Math.max(n.replicas.desired, n.replicas.ready, 1), 8);
	return Array.from({ length: total }, (_, i) => i < n.replicas.ready);
});

const detailRows = computed(() => {
	const n = props.node;
	if (!n) return [] as { label: string; value: string; mono?: boolean }[];
	const rows: { label: string; value: string; mono?: boolean }[] = [];
	if (n.workload?.kind) rows.push({ label: 'Workload', value: n.workload.kind });
	if (n.project) rows.push({ label: 'Project', value: n.project });
	if (n.language) rows.push({ label: 'Language', value: n.language });
	if (n.engine) rows.push({ label: 'Engine', value: n.engine });
	if (n.placement) rows.push({ label: 'Placement', value: n.placement });
	if (n.host) rows.push({ label: 'Host', value: n.host, mono: true });
	if (n.image) rows.push({ label: 'Image', value: n.image, mono: true });
	if (n.meta?.region) rows.push({ label: 'Region', value: n.meta.region });
	if (n.pods?.length) rows.push({ label: 'Pods', value: String(n.pods.length) });
	return rows;
});

async function run(verb: string, fn: () => Promise<Topology>): Promise<void> {
	busy.value = true;
	try {
		const topo = await fn();
		emit('changed', topo);
		toast.success(`${verb} succeeded`);
	} catch (e) {
		toast.error(`${verb} failed: ${(e as Error).message}`);
	} finally {
		busy.value = false;
	}
}

function confirmRun(verb: string, message: string, fn: () => Promise<Topology>): void {
	if (window.confirm(message)) void run(verb, fn);
}
</script>
