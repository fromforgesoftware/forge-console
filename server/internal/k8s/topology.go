package k8s

import (
	"context"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/fromforgesoftware/go-kit/resource"
	"github.com/fromforgesoftware/go-kit/search"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// TopologyRepository reads the live cluster into the graph. Nodes come from
// workload labels; edges from forge.dev/connects-to annotations the Helm charts
// set — there is no declared catalog (live-only).
type TopologyRepository struct {
	cs kubernetes.Interface
}

func NewTopologyRepository(cs kubernetes.Interface) *TopologyRepository {
	return &TopologyRepository{cs: cs}
}

// List returns the singleton topology as a one-element collection so the kit
// list handler works unchanged.
func (r *TopologyRepository) List(ctx context.Context, _ ...search.Option) (resource.ListResponse[app.Topology], error) {
	t, err := r.Get(ctx)
	if err != nil {
		return nil, err
	}
	return resource.NewListResponse([]app.Topology{t}, 1), nil
}

// Get assembles the whole graph. With no reachable cluster it returns an empty
// topology flagged unavailable rather than erroring.
func (r *TopologyRepository) Get(ctx context.Context, _ ...search.Option) (app.Topology, error) {
	if r.cs == nil {
		return app.NewTopology(app.ClusterInfo{Available: false}, nil, nil), nil
	}

	b := &builder{
		nodesByID: map[string]*app.Node{},
		byProject: map[string]string{},
	}

	cluster := app.ClusterInfo{Available: true, Name: "cluster"}
	if v, err := r.cs.Discovery().ServerVersion(); err == nil {
		cluster.Version = v.GitVersion
	}

	nodeList, err := r.cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err == nil {
		cluster.NodeCount = len(nodeList.Items)
		for i := range nodeList.Items {
			b.addWorkerNode(&nodeList.Items[i])
		}
	}

	pods, _ := r.cs.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	deploys, _ := r.cs.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	statefulsets, _ := r.cs.AppsV1().StatefulSets(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	services, _ := r.cs.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	ingresses, _ := r.cs.NetworkingV1().Ingresses(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})

	if deploys != nil {
		for i := range deploys.Items {
			d := &deploys.Items[i]
			if isSystemNamespace(d.Namespace) {
				continue
			}
			spec := d.Spec.Template.Spec
			b.addWorkload("Deployment", d.Namespace, d.Name, d.Labels, d.Annotations,
				d.Spec.Selector, podSpecImage(spec), dbHostFromPodSpec(spec),
				replicas(d.Spec.Replicas, d.Status.ReadyReplicas), d.Spec.Paused, pods)
		}
	}
	if statefulsets != nil {
		for i := range statefulsets.Items {
			s := &statefulsets.Items[i]
			if isSystemNamespace(s.Namespace) {
				continue
			}
			spec := s.Spec.Template.Spec
			b.addWorkload("StatefulSet", s.Namespace, s.Name, s.Labels, s.Annotations,
				s.Spec.Selector, podSpecImage(spec), dbHostFromPodSpec(spec),
				replicas(s.Spec.Replicas, s.Status.ReadyReplicas), false, pods)
		}
	}
	if services != nil {
		for i := range services.Items {
			b.addService(&services.Items[i])
		}
	}
	if ingresses != nil {
		for i := range ingresses.Items {
			b.addIngress(&ingresses.Items[i])
		}
	}

	b.applyConnectsTo()

	// Deterministic order so every snapshot is identical when the cluster is
	// unchanged — the frontend keys its layout on node identity and must not see
	// nodes reorder between live updates.
	nodes := b.nodes()
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	edges := b.edges
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })

	return app.NewTopology(cluster, nodes, edges), nil
}

func isSystemNamespace(ns string) bool {
	switch ns {
	case "kube-system", "kube-public", "kube-node-lease", "local-path-storage":
		return true
	}
	return false
}

func podSpecImage(s corev1.PodSpec) string {
	if len(s.Containers) > 0 {
		return s.Containers[0].Image
	}
	return ""
}

// dbHostFromPodSpec finds a literal datastore host in the workload's container
// env so the builder can draw the service→datastore edge without deploy-time
// annotations. valueFrom (secret) hosts are skipped — they can't be resolved
// here.
func dbHostFromPodSpec(s corev1.PodSpec) string {
	wanted := map[string]bool{"DB_HOST": true, "DATABASE_HOST": true, "POSTGRES_HOST": true, "PGHOST": true, "REDIS_HOST": true}
	for _, c := range s.Containers {
		for _, e := range c.Env {
			if wanted[e.Name] && e.Value != "" {
				return e.Value
			}
		}
	}
	return ""
}

func replicas(desired *int32, ready int32) app.Replicas {
	var d int32
	if desired != nil {
		d = *desired
	}
	return app.Replicas{Desired: d, Ready: ready}
}
