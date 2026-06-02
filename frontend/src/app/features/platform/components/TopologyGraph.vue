<template>
	<div class="relative h-full w-full overflow-hidden rounded-xl border border-border bg-background">
		<VueFlow
			:nodes="layoutNodes"
			:edges="layoutEdges"
			:nodes-draggable="false"
			:nodes-connectable="false"
			:elements-selectable="true"
			:pan-on-drag="true"
			:zoom-on-scroll="true"
			:min-zoom="0.3"
			:max-zoom="1.6"
			fit-view-on-init
			class="topology-flow"
			@node-click="onNodeClick"
		>
			<Background pattern-color="#cbd5e1" :gap="16" :size="1" />
			<Controls position="bottom-right" :show-interactive="false" />
			<MiniMap pannable zoomable :node-color="miniMapColor" />
			<template #node-topology="nodeProps">
				<TopologyNode :data="nodeProps.data" />
			</template>
		</VueFlow>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { VueFlow, Position } from '@vue-flow/core';
import { Background } from '@vue-flow/background';
import { Controls } from '@vue-flow/controls';
import { MiniMap } from '@vue-flow/minimap';
import dagre from '@dagrejs/dagre';
import type { Edge, Node } from '@vue-flow/core';
import TopologyNode from './TopologyNode.vue';
import type { TopologyNode as TNode, TopologyEdge } from '../domain/topology';
import { statusTone } from '../domain/topology';

import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';

const props = defineProps<{ nodes: TNode[]; edges: TopologyEdge[] }>();
const emit = defineEmits<{ (e: 'select', node: TNode): void }>();

const NODE_WIDTH = 280;
const NODE_HEIGHT = 120;

// Positions are cached by node id and only recomputed when the SET of nodes
// changes. Live status updates (the common case) reuse the stored positions, so
// nodes never jump around between realtime snapshots.
const layoutNodes = ref<Node[]>([]);
const layoutEdges = ref<Edge[]>([]);
const positions = new Map<string, { x: number; y: number }>();
let lastKey = '';

function relayout(nodes: TNode[], edges: TopologyEdge[]): void {
	const g = new dagre.graphlib.Graph();
	g.setDefaultEdgeLabel(() => ({}));
	g.setGraph({ rankdir: 'TB', nodesep: 80, ranksep: 110 });
	for (const n of nodes) g.setNode(n.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
	for (const e of edges) {
		if (g.hasNode(e.source) && g.hasNode(e.target)) g.setEdge(e.source, e.target);
	}
	dagre.layout(g);
	positions.clear();
	for (const n of nodes) {
		const p = g.node(n.id);
		positions.set(n.id, { x: p.x - NODE_WIDTH / 2, y: p.y - NODE_HEIGHT / 2 });
	}
}

watch(
	[() => props.nodes, () => props.edges],
	([nodes, edges]) => {
		const key = nodes
			.map((n) => n.id)
			.sort()
			.join('|');
		if (key !== lastKey) {
			relayout(nodes, edges);
			lastKey = key;
		}
		layoutNodes.value = nodes.map((n) => ({
			id: n.id,
			type: 'topology',
			position: positions.get(n.id) ?? { x: 0, y: 0 },
			data: n,
			sourcePosition: Position.Bottom,
			targetPosition: Position.Top,
			draggable: false,
		}));
		const ids = new Set(nodes.map((n) => n.id));
		layoutEdges.value = edges
			.filter((e) => ids.has(e.source) && ids.has(e.target))
			.map((e) => ({
				id: e.id,
				source: e.source,
				target: e.target,
				label: e.label,
				animated: e.status === 'pending',
				style: { stroke: edgeStroke(e), strokeWidth: 1.5, opacity: 0.55 },
			}));
	},
	{ immediate: true, deep: true },
);

function edgeStroke(e: TopologyEdge): string {
	switch (statusTone(e.status)) {
		case 'success':
			return 'var(--color-success)';
		case 'info':
			return 'var(--color-info)';
		case 'destructive':
			return 'var(--color-destructive)';
		default:
			return 'var(--color-border)';
	}
}

function miniMapColor(node: Node): string {
	const tone = statusTone((node.data as TNode)?.status);
	return `var(--color-${tone === 'muted' ? 'border' : tone})`;
}

function onNodeClick({ node }: { node: Node }) {
	emit('select', node.data as TNode);
}
</script>

<style>
.topology-flow .vue-flow__panel.vue-flow__attribution {
	display: none;
}
</style>
