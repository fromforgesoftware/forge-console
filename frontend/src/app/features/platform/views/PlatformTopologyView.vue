<template>
	<div class="flex min-h-0 flex-1 flex-col gap-3">
		<header class="flex items-center justify-between">
			<div>
				<h1 class="text-lg font-semibold text-foreground">Platform topology</h1>
				<p class="text-sm text-muted-foreground">
					<template v-if="topology?.cluster.available">
						{{ topology.cluster.version || 'kubernetes' }} ·
						{{ topology.cluster.nodeCount }} node(s)
					</template>
					<template v-else>Live cluster view</template>
				</p>
			</div>
			<div class="flex items-center gap-2">
				<Legend />
				<Button variant="outline" size="sm" :disabled="loading" @click="load">
					<RefreshCw class="size-3.5" :class="loading ? 'animate-spin' : ''" /> Refresh
				</Button>
			</div>
		</header>

		<div class="relative min-h-0 flex-1">
			<div v-if="loading && !topology" class="flex h-full items-center justify-center">
				<Spinner />
			</div>
			<EmptyState v-else-if="error" title="Couldn't load topology" :description="error" />
			<EmptyState
				v-else-if="topology && !topology.cluster.available"
				title="Cluster unavailable"
				description="Forge could not reach a Kubernetes cluster. The topology appears once it is deployed in-cluster."
			/>
			<EmptyState
				v-else-if="topology && topology.nodes.length === 0"
				title="No workloads found"
				description="Nothing is deployed in the cluster yet."
			/>
			<TopologyGraph
				v-else-if="topology"
				:nodes="topology.nodes"
				:edges="topology.edges"
				@select="onSelect"
			/>
		</div>

		<NodeDetailDrawer v-model:open="drawerOpen" :node="selected" @changed="onChanged" />
	</div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { Button, Spinner, EmptyState } from '@fromforgesoftware/vue-kit';
import { RefreshCw } from '@lucide/vue';
import TopologyGraph from '../components/TopologyGraph.vue';
import NodeDetailDrawer from '../components/NodeDetailDrawer.vue';
import Legend from '../components/TopologyLegend.vue';
import { fetchTopology } from '../data/topology';
import { usePlatformStream } from '../data/stream';
import type { Topology, TopologyNode } from '../domain/topology';

const topology = ref<Topology | null>(null);
const loading = ref(false);
const error = ref('');
const selected = ref<TopologyNode | null>(null);
const drawerOpen = ref(false);

async function load(): Promise<void> {
	loading.value = true;
	error.value = '';
	try {
		topology.value = await fetchTopology();
	} catch (e) {
		error.value = (e as Error).message;
	} finally {
		loading.value = false;
	}
}

function onSelect(node: TopologyNode): void {
	selected.value = node;
	drawerOpen.value = true;
}

function onChanged(updated: Topology): void {
	topology.value = updated;
}

// Live layer: informer-driven snapshots keep the graph current without polling.
usePlatformStream((t) => {
	topology.value = t;
	loading.value = false;
	error.value = '';
});

onMounted(load);
</script>
