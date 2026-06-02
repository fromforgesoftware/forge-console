// Topology domain types mirror the forge backend's JSON:API attributes
// (apps/forge/server/internal/api/topology.go).

export type NodeKind = 'service' | 'lib' | 'database' | 'gateway' | 'external' | 'worker';

export type NodeStatus = 'running' | 'pending' | 'degraded' | 'paused' | 'not-deployed' | 'unknown';

export type EdgeKind = 'depends-on' | 'connects-to' | 'routes-to';

export interface Replicas {
	desired: number;
	ready: number;
}

export interface WorkloadRef {
	kind?: string;
	name?: string;
	namespace?: string;
}

export interface TopologyNode {
	id: string;
	kind: NodeKind;
	name: string;
	namespace?: string;
	status: NodeStatus;
	project?: string;
	projectType?: string;
	language?: string;
	image?: string;
	engine?: string;
	placement?: 'in-cluster' | 'external';
	host?: string;
	replicas: Replicas;
	workload?: WorkloadRef;
	tags?: string[];
	pods?: string[];
	meta?: Record<string, string>;
}

export interface TopologyEdge {
	id: string;
	source: string;
	target: string;
	kind: EdgeKind;
	status?: NodeStatus;
	label?: string;
}

export interface ClusterInfo {
	name: string;
	version: string;
	nodeCount: number;
	available: boolean;
}

export interface Topology {
	cluster: ClusterInfo;
	nodes: TopologyNode[];
	edges: TopologyEdge[];
}

// statusTone maps a node/edge status to the design-system colour token used for
// borders, pills and edge strokes — the "real feedback" the operator reads.
export function statusTone(
	status: NodeStatus | undefined,
): 'success' | 'info' | 'destructive' | 'muted' {
	switch (status) {
		case 'running':
			return 'success';
		case 'pending':
			return 'info';
		case 'degraded':
			return 'destructive';
		default:
			return 'muted';
	}
}
