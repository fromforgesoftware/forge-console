package app

import (
	"context"
	"io"

	"github.com/fromforgesoftware/go-kit/application/repository"
	"github.com/fromforgesoftware/go-kit/resource"
)

const ResourceTypeTopology resource.Type = "topologies"

// TopologyID is the id of the singleton topology resource (one cluster per
// Foundry deployment in this phase).
const TopologyID = "cluster"

// NodeKind classifies a topology node for icon + layout selection.
type NodeKind string

const (
	NodeKindService    NodeKind = "service"  // a deployed forge service/app workload
	NodeKindLib        NodeKind = "lib"      // a declared library (never deployed)
	NodeKindDatabase   NodeKind = "database" // a datastore, in-cluster or external
	NodeKindGateway    NodeKind = "gateway"  // an ingress / load-balancer edge
	NodeKindExternal   NodeKind = "external" // an external managed dependency
	NodeKindWorkerNode NodeKind = "worker"   // a kubernetes worker node
)

// NodeStatus is the live health of a node, mapped to the graph's status colors.
type NodeStatus string

const (
	StatusRunning     NodeStatus = "running"      // ready/healthy
	StatusPending     NodeStatus = "pending"      // progressing / rolling out
	StatusDegraded    NodeStatus = "degraded"     // crash-looping / failed / not ready
	StatusPaused      NodeStatus = "paused"       // rollout paused / scaled to zero
	StatusNotDeployed NodeStatus = "not-deployed" // declared in forge.json, absent in cluster
	StatusUnknown     NodeStatus = "unknown"
)

// Placement says where a node physically runs.
type Placement string

const (
	PlacementInCluster Placement = "in-cluster"
	PlacementExternal  Placement = "external"
)

// Node is one box on the topology canvas. Identity/wiring come from deploy-time
// annotations + forge.json (the declared layer); Status/Replicas come from the
// live cluster (the runtime layer).
type Node struct {
	ID          string
	Kind        NodeKind
	Name        string
	Namespace   string
	Status      NodeStatus
	Project     string
	ProjectType string
	Language    string
	WorkloadRef WorkloadRef
	Replicas    Replicas
	Image       string
	Engine      string
	Placement   Placement
	Host        string
	Tags        []string
	PodNames    []string
	Meta        map[string]string
}

// WorkloadRef points at the Kubernetes object that backs a node, so actions
// (restart/scale/...) know what to operate on.
type WorkloadRef struct {
	Kind      string // Deployment | StatefulSet | Node | ""
	Name      string
	Namespace string
}

// Replicas is the live readiness summary shown on workload nodes.
type Replicas struct {
	Desired int32
	Ready   int32
}

// EdgeKind labels the relationship an edge represents.
type EdgeKind string

const (
	EdgeDependsOn  EdgeKind = "depends-on"  // declared @forge/* dependency
	EdgeConnectsTo EdgeKind = "connects-to" // service → datastore/external (annotation)
	EdgeRoutesTo   EdgeKind = "routes-to"   // service/ingress → backend
)

// Edge is a directed relationship between two nodes.
type Edge struct {
	ID     string
	Source string
	Target string
	Kind   EdgeKind
	Status NodeStatus
	Label  string
}

// ClusterInfo is the summary of the observed cluster shown as the root node.
type ClusterInfo struct {
	Name      string
	Version   string
	NodeCount int
	Available bool
}

// Topology is the singleton resource: the whole declared+observed graph.
type Topology interface {
	resource.Resource
	Cluster() ClusterInfo
	Nodes() []Node
	Edges() []Edge
}

type topologyRes struct {
	resource.Resource

	cluster ClusterInfo
	nodes   []Node
	edges   []Edge
}

// NewTopology builds the singleton topology aggregate.
func NewTopology(cluster ClusterInfo, nodes []Node, edges []Edge) Topology {
	return &topologyRes{
		Resource: resource.New(resource.WithType(ResourceTypeTopology), resource.WithID(TopologyID)),
		cluster:  cluster,
		nodes:    nodes,
		edges:    edges,
	}
}

func (t *topologyRes) Cluster() ClusterInfo { return t.cluster }
func (t *topologyRes) Nodes() []Node        { return t.nodes }
func (t *topologyRes) Edges() []Edge        { return t.edges }

// TopologyRepository reads the live cluster graph. It is the kit generic read
// surface; the only backing implementation is the k8s adapter.
type TopologyRepository interface {
	repository.Getter[Topology]
	repository.Lister[Topology]
}

// WorkloadScaleCommand sets the desired replica count of a workload.
type WorkloadScaleCommand struct {
	Kind      string
	Namespace string
	Name      string
	Replicas  int32
}

// WorkloadRefCommand targets a single workload for restart/pause/resume/delete.
type WorkloadRefCommand struct {
	Kind      string
	Namespace string
	Name      string
}

// NodeCommand targets a worker node for cordon/uncordon/drain.
type NodeCommand struct {
	Name string
}

// TopologyWatcher notifies subscribers when the cluster changes, so the stream
// can push a fresh snapshot. Subscribe returns a tick channel and a cancel func.
type TopologyWatcher interface {
	Subscribe() (<-chan struct{}, func())
}

// LogStreamer opens a following pod log stream for the realtime logs tab.
type LogStreamer interface {
	OpenPodLogs(ctx context.Context, namespace, pod, container string) (io.ReadCloser, error)
}

// ClusterActions mutates the cluster. Each method maps to a kubernetes API call
// and is audited by the usecase.
type ClusterActions interface {
	RestartWorkload(ctx context.Context, cmd WorkloadRefCommand) error
	ScaleWorkload(ctx context.Context, cmd WorkloadScaleCommand) error
	PauseWorkload(ctx context.Context, cmd WorkloadRefCommand) error
	ResumeWorkload(ctx context.Context, cmd WorkloadRefCommand) error
	DeleteWorkload(ctx context.Context, cmd WorkloadRefCommand) error
	CordonNode(ctx context.Context, cmd NodeCommand) error
	UncordonNode(ctx context.Context, cmd NodeCommand) error
	DrainNode(ctx context.Context, cmd NodeCommand) error
}
