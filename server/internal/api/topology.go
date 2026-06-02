package api

import (
	"github.com/fromforgesoftware/go-kit/resource"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// TopologyDTO is the jsonapi representation of the cluster graph. Nodes/edges
// are plain json-tagged value objects carried as attribute values.
type TopologyDTO struct {
	resource.RestDTO

	RCluster ClusterDTO `jsonapi:"attr,cluster"`
	RNodes   []NodeDTO  `jsonapi:"attr,nodes"`
	REdges   []EdgeDTO  `jsonapi:"attr,edges"`
}

type ClusterDTO struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	NodeCount int    `json:"nodeCount"`
	Available bool   `json:"available"`
}

type NodeDTO struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	Status      string            `json:"status"`
	Project     string            `json:"project,omitempty"`
	ProjectType string            `json:"projectType,omitempty"`
	Language    string            `json:"language,omitempty"`
	Image       string            `json:"image,omitempty"`
	Engine      string            `json:"engine,omitempty"`
	Placement   string            `json:"placement,omitempty"`
	Host        string            `json:"host,omitempty"`
	Replicas    ReplicasDTO       `json:"replicas"`
	Workload    WorkloadRefDTO    `json:"workload,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Pods        []string          `json:"pods,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

type ReplicasDTO struct {
	Desired int32 `json:"desired"`
	Ready   int32 `json:"ready"`
}

type WorkloadRefDTO struct {
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type EdgeDTO struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Kind   string `json:"kind"`
	Status string `json:"status,omitempty"`
	Label  string `json:"label,omitempty"`
}

func TopologyToDTO(t app.Topology) *TopologyDTO {
	if t == nil {
		return nil
	}
	c := t.Cluster()
	dto := &TopologyDTO{
		RestDTO: resource.ToRestDTO(t),
		RCluster: ClusterDTO{
			Name:      c.Name,
			Version:   c.Version,
			NodeCount: c.NodeCount,
			Available: c.Available,
		},
		RNodes: nodesToDTO(t.Nodes()),
		REdges: edgesToDTO(t.Edges()),
	}
	dto.RType = app.ResourceTypeTopology
	return dto
}

func nodesToDTO(nodes []app.Node) []NodeDTO {
	out := make([]NodeDTO, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, NodeDTO{
			ID:          n.ID,
			Kind:        string(n.Kind),
			Name:        n.Name,
			Namespace:   n.Namespace,
			Status:      string(n.Status),
			Project:     n.Project,
			ProjectType: n.ProjectType,
			Language:    n.Language,
			Image:       n.Image,
			Engine:      n.Engine,
			Placement:   string(n.Placement),
			Host:        n.Host,
			Replicas:    ReplicasDTO{Desired: n.Replicas.Desired, Ready: n.Replicas.Ready},
			Workload:    WorkloadRefDTO{Kind: n.WorkloadRef.Kind, Name: n.WorkloadRef.Name, Namespace: n.WorkloadRef.Namespace},
			Tags:        n.Tags,
			Pods:        n.PodNames,
			Meta:        n.Meta,
		})
	}
	return out
}

func edgesToDTO(edges []app.Edge) []EdgeDTO {
	out := make([]EdgeDTO, 0, len(edges))
	for _, e := range edges {
		out = append(out, EdgeDTO{
			ID:     e.ID,
			Source: e.Source,
			Target: e.Target,
			Kind:   string(e.Kind),
			Status: string(e.Status),
			Label:  e.Label,
		})
	}
	return out
}

// TopologyPayload is the plain-JSON shape pushed over the websocket — the same
// attributes the REST resource exposes, so the frontend maps it identically.
type TopologyPayload struct {
	Cluster ClusterDTO `json:"cluster"`
	Nodes   []NodeDTO  `json:"nodes"`
	Edges   []EdgeDTO  `json:"edges"`
}

func TopologyToPayload(t app.Topology) TopologyPayload {
	c := t.Cluster()
	return TopologyPayload{
		Cluster: ClusterDTO{Name: c.Name, Version: c.Version, NodeCount: c.NodeCount, Available: c.Available},
		Nodes:   nodesToDTO(t.Nodes()),
		Edges:   edgesToDTO(t.Edges()),
	}
}

// ScaleCommandDTO is the request body for the scale action.
type ScaleCommandDTO struct {
	resource.RestDTO

	RReplicas int32 `jsonapi:"attr,replicas"`
}

func (d *ScaleCommandDTO) Replicas() int32 { return d.RReplicas }
