package k8s

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// builder accumulates nodes/edges as the cluster is walked, then resolves the
// declared-wiring (connects-to) layer from annotations.
type builder struct {
	nodesByID map[string]*app.Node
	order     []string
	byProject map[string]string
	edges     []app.Edge
	edgeIDs   map[string]bool

	pending []pendingEdge
}

type pendingEdge struct {
	source string
	refs   []connectsRef
}

func (b *builder) put(n app.Node) *app.Node {
	if existing, ok := b.nodesByID[n.ID]; ok {
		return existing
	}
	cp := n
	b.nodesByID[n.ID] = &cp
	b.order = append(b.order, n.ID)
	return &cp
}

func (b *builder) nodes() []app.Node {
	out := make([]app.Node, 0, len(b.order))
	for _, id := range b.order {
		out = append(out, *b.nodesByID[id])
	}
	return out
}

func (b *builder) addWorkerNode(n *corev1.Node) {
	b.put(app.Node{
		ID:        "worker/" + n.Name,
		Kind:      app.NodeKindWorkerNode,
		Name:      n.Name,
		Status:    workerStatus(n),
		Placement: app.PlacementInCluster,
		Meta: map[string]string{
			"region": n.Labels["topology.kubernetes.io/region"],
			"zone":   n.Labels["topology.kubernetes.io/zone"],
		},
	})
}

func (b *builder) addWorkload(kind, ns, name string, labels, ann map[string]string, sel *metav1.LabelSelector, image, dbHost string, rep app.Replicas, paused bool, pods *corev1.PodList) {
	project := firstNonEmpty(ann[AnnProject], labels[LabelPartOf], name)
	engine := dbEngineFromImage(image)
	nodeKind := app.NodeKindService
	if labels[LabelComponent] == ComponentDatabase || engine != "" {
		nodeKind = app.NodeKindDatabase
	}

	podNames, crashing := matchingPods(sel, ns, pods)
	id := strings.ToLower(kind) + "/" + ns + "/" + name

	n := b.put(app.Node{
		ID:          id,
		Kind:        nodeKind,
		Name:        name,
		Namespace:   ns,
		Status:      workloadStatus(rep, paused, crashing),
		Project:     project,
		ProjectType: ann[AnnType],
		Replicas:    rep,
		Image:       image,
		Engine:      engine,
		Placement:   app.PlacementInCluster,
		WorkloadRef: app.WorkloadRef{Kind: kind, Name: name, Namespace: ns},
		PodNames:    podNames,
	})
	if _, ok := b.byProject[project]; !ok {
		b.byProject[project] = n.ID
	}
	// Also index by workload name so dependency lookups resolve even when many
	// workloads share a generic part-of label (e.g. all platform services
	// labelled part-of=forge).
	if _, ok := b.byProject[name]; !ok {
		b.byProject[name] = n.ID
	}

	// Declared connects-to annotations win; otherwise infer the datastore edge
	// from the live DB_HOST env (resolved/classified at read time). A database
	// node never connects to itself.
	refs := parseConnectsTo(ann[AnnConnectsTo])
	if dbHost != "" && nodeKind != app.NodeKindDatabase && !hasDBRef(refs) {
		refs = append(refs, classifyHost(dbHost))
	}
	if len(refs) > 0 {
		b.pending = append(b.pending, pendingEdge{source: id, refs: refs})
	}
}

func hasDBRef(refs []connectsRef) bool {
	for _, r := range refs {
		if r.kind == "db" || r.kind == "ext" {
			return true
		}
	}
	return false
}

// classifyHost decides whether a configured datastore host is an in-cluster
// service (bare name or *.svc) or an external managed dependency (FQDN).
func classifyHost(host string) connectsRef {
	if strings.Contains(host, ".") && !strings.HasSuffix(host, ".svc.cluster.local") && !strings.HasSuffix(host, ".svc") {
		return connectsRef{kind: "ext", target: host}
	}
	bare := host
	if i := strings.IndexByte(bare, '.'); i >= 0 {
		bare = bare[:i]
	}
	return connectsRef{kind: "db", target: bare}
}

func (b *builder) addService(s *corev1.Service) {
	if s.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return
	}
	b.put(app.Node{
		ID:        "gateway/svc/" + s.Namespace + "/" + s.Name,
		Kind:      app.NodeKindGateway,
		Name:      s.Name,
		Namespace: s.Namespace,
		Status:    app.StatusRunning,
		Placement: app.PlacementInCluster,
		Meta:      map[string]string{"via": "LoadBalancer"},
	})
}

func (b *builder) addIngress(ing *networkingv1.Ingress) {
	gwID := "gateway/ing/" + ing.Namespace + "/" + ing.Name
	b.put(app.Node{
		ID:        gwID,
		Kind:      app.NodeKindGateway,
		Name:      ing.Name,
		Namespace: ing.Namespace,
		Status:    app.StatusRunning,
		Placement: app.PlacementInCluster,
		Meta:      map[string]string{"via": "Ingress"},
	})
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, p := range rule.HTTP.Paths {
			if p.Backend.Service == nil {
				continue
			}
			if target, ok := b.byProject[p.Backend.Service.Name]; ok {
				b.edges = append(b.edges, app.Edge{
					ID:     gwID + "->" + target,
					Source: gwID,
					Target: target,
					Kind:   app.EdgeRoutesTo,
				})
			}
		}
	}
}

func (b *builder) applyConnectsTo() {
	for _, pe := range b.pending {
		for _, ref := range pe.refs {
			switch ref.kind {
			case "project":
				if target, ok := b.byProject[ref.target]; ok {
					b.addEdge(pe.source, target, app.EdgeDependsOn, "")
				}
			case "db":
				target := b.byProject[ref.target]
				if target == "" {
					n := b.put(app.Node{
						ID:        "database/" + ref.target,
						Kind:      app.NodeKindDatabase,
						Name:      ref.target,
						Status:    app.StatusUnknown,
						Placement: app.PlacementInCluster,
					})
					target = n.ID
				}
				b.addEdge(pe.source, target, app.EdgeConnectsTo, "")
			case "ext":
				n := b.put(app.Node{
					ID:        "external/" + ref.target,
					Kind:      app.NodeKindExternal,
					Name:      hostLabel(ref.target),
					Status:    app.StatusUnknown,
					Placement: app.PlacementExternal,
					Host:      ref.target,
				})
				b.addEdge(pe.source, n.ID, app.EdgeConnectsTo, "")
			}
		}
	}
}

func (b *builder) addEdge(source, target string, kind app.EdgeKind, label string) {
	id := source + "->" + target
	if b.edgeIDs[id] {
		return
	}
	if b.edgeIDs == nil {
		b.edgeIDs = map[string]bool{}
	}
	b.edgeIDs[id] = true
	b.edges = append(b.edges, app.Edge{
		ID:     id,
		Source: source,
		Target: target,
		Kind:   kind,
		Label:  label,
	})
}

func workerStatus(n *corev1.Node) app.NodeStatus {
	if n.Spec.Unschedulable {
		return app.StatusPaused
	}
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady {
			if c.Status == corev1.ConditionTrue {
				return app.StatusRunning
			}
			return app.StatusDegraded
		}
	}
	return app.StatusUnknown
}

func workloadStatus(rep app.Replicas, paused, crashing bool) app.NodeStatus {
	switch {
	case crashing:
		return app.StatusDegraded
	case paused, rep.Desired == 0:
		return app.StatusPaused
	case rep.Ready >= rep.Desired && rep.Desired > 0:
		return app.StatusRunning
	default:
		return app.StatusPending
	}
}

// matchingPods returns the names of pods backing a workload and whether any is
// in a crash/error wait state (used for the node's live status + logs target).
func matchingPods(sel *metav1.LabelSelector, ns string, pods *corev1.PodList) (names []string, crashing bool) {
	if sel == nil || pods == nil || len(sel.MatchLabels) == 0 {
		return nil, false
	}
	for i := range pods.Items {
		p := &pods.Items[i]
		if p.Namespace != ns || !labelsMatch(sel.MatchLabels, p.Labels) {
			continue
		}
		names = append(names, p.Name)
		for _, cs := range p.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				switch cs.State.Waiting.Reason {
				case "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull", "CreateContainerConfigError":
					crashing = true
				}
			}
		}
	}
	return names, crashing
}

func labelsMatch(want, have map[string]string) bool {
	for k, v := range want {
		if have[k] != v {
			return false
		}
	}
	return true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func hostLabel(host string) string {
	h := strings.TrimPrefix(host, "https://")
	h = strings.TrimPrefix(h, "http://")
	if i := strings.IndexByte(h, '/'); i >= 0 {
		h = h[:i]
	}
	return h
}
