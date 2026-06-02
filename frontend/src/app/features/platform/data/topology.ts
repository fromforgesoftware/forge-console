import { api } from '@/app/core/http/api';
import type { Topology, TopologyNode, TopologyEdge, ClusterInfo } from '../domain/topology';

// The platform endpoints are forge-native (not gateway-proxied); they go
// through the shared JSON:API client (cookie auth, media-type headers, error
// parsing) under the `/platform` path.
interface JsonApiResource {
	type: string;
	id: string;
	attributes: Record<string, unknown>;
}

async function call(
	path: string,
	init?: { method?: string; body?: unknown },
): Promise<JsonApiResource | null> {
	const res = await api.request({
		method: init?.method ?? 'GET',
		path: `/platform${path}`,
		body: init?.body,
	});
	if (res.status === 204) return null;
	const body = res.body as { data?: JsonApiResource | JsonApiResource[] };
	const data = Array.isArray(body.data) ? body.data[0] : body.data;
	return data ?? null;
}

function toTopology(res: JsonApiResource | null): Topology {
	const attrs = (res?.attributes ?? {}) as Record<string, unknown>;
	return {
		cluster: (attrs.cluster as ClusterInfo) ?? {
			name: '',
			version: '',
			nodeCount: 0,
			available: false,
		},
		nodes: (attrs.nodes as TopologyNode[]) ?? [],
		edges: (attrs.edges as TopologyEdge[]) ?? [],
	};
}

export async function fetchTopology(): Promise<Topology> {
	return toTopology(await call('/topology'));
}

export interface WorkloadTarget {
	kind: string;
	namespace: string;
	name: string;
}

function workloadPath(t: WorkloadTarget): string {
	return `/workloads/${encodeURIComponent(t.namespace)}/${encodeURIComponent(t.kind)}/${encodeURIComponent(t.name)}`;
}

export async function restartWorkload(t: WorkloadTarget): Promise<Topology> {
	return toTopology(await call(`${workloadPath(t)}/restart`, { method: 'POST' }));
}

export async function pauseWorkload(t: WorkloadTarget): Promise<Topology> {
	return toTopology(await call(`${workloadPath(t)}/pause`, { method: 'POST' }));
}

export async function resumeWorkload(t: WorkloadTarget): Promise<Topology> {
	return toTopology(await call(`${workloadPath(t)}/resume`, { method: 'POST' }));
}

export async function deleteWorkload(t: WorkloadTarget): Promise<Topology> {
	return toTopology(await call(workloadPath(t), { method: 'DELETE' }));
}

export async function scaleWorkload(t: WorkloadTarget, replicas: number): Promise<Topology> {
	const body = { data: { type: 'topologies', attributes: { replicas } } };
	return toTopology(await call(`${workloadPath(t)}/scale`, { method: 'POST', body }));
}

export async function nodeAction(
	name: string,
	action: 'cordon' | 'uncordon' | 'drain',
): Promise<Topology> {
	return toTopology(await call(`/nodes/${encodeURIComponent(name)}/${action}`, { method: 'POST' }));
}
