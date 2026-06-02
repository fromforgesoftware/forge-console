package k8s_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/fromforgesoftware/forge/server/internal/app"
	"github.com/fromforgesoftware/forge/server/internal/k8s"
)

func ptr[T any](v T) *T { return &v }

func TestGet_BuildsGraphFromAnnotations(t *testing.T) {
	cs := fake.NewSimpleClientset(
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "kind-control-plane",
				Labels: map[string]string{"topology.kubernetes.io/region": "local"},
			},
			Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			}},
		},
		deployment("petstore", "default", map[string]string{
			k8s.AnnProject:    "petstore",
			k8s.AnnType:       "service",
			k8s.AnnConnectsTo: "aegis,db:postgres,ext:https://db.supabase.co",
		}, 1, 1, "ghcr.io/forge/petstore:dev"),
		deployment("aegis", "default", map[string]string{k8s.AnnProject: "aegis"}, 3, 3, "ghcr.io/forge/aegis:dev"),
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "postgres",
				Namespace: "default",
				Labels:    map[string]string{k8s.LabelComponent: k8s.ComponentDatabase},
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr[int32](1),
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "postgres"}},
				Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: "postgres:16"}}}},
			},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 1},
		},
	)

	repo := k8s.NewTopologyRepository(cs)
	topo, err := repo.Get(context.Background())
	require.NoError(t, err)

	assert.True(t, topo.Cluster().Available)
	assert.Equal(t, 1, topo.Cluster().NodeCount)

	nodes := indexNodes(topo.Nodes())

	petstore := nodes["deployment/default/petstore"]
	require.NotNil(t, petstore)
	assert.Equal(t, app.NodeKindService, petstore.Kind)
	assert.Equal(t, app.StatusRunning, petstore.Status)

	pg := nodes["statefulset/default/postgres"]
	require.NotNil(t, pg)
	assert.Equal(t, app.NodeKindDatabase, pg.Kind, "component=database label classifies it as a datastore")

	ext := nodes["external/https://db.supabase.co"]
	require.NotNil(t, ext)
	assert.Equal(t, app.PlacementExternal, ext.Placement)

	assert.True(t, hasEdge(topo.Edges(), "deployment/default/petstore", "deployment/default/aegis", app.EdgeDependsOn))
	assert.True(t, hasEdge(topo.Edges(), "deployment/default/petstore", "statefulset/default/postgres", app.EdgeConnectsTo))
	assert.True(t, hasEdge(topo.Edges(), "deployment/default/petstore", "external/https://db.supabase.co", app.EdgeConnectsTo))
}

func TestGet_NoClusterReportsUnavailable(t *testing.T) {
	repo := k8s.NewTopologyRepository(nil)
	topo, err := repo.Get(context.Background())
	require.NoError(t, err)
	assert.False(t, topo.Cluster().Available)
	assert.Empty(t, topo.Nodes())
}

func deployment(name, ns string, ann map[string]string, desired, ready int32, image string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns, Annotations: ann},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr(desired),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: image}}}},
		},
		Status: appsv1.DeploymentStatus{ReadyReplicas: ready},
	}
}

func indexNodes(nodes []app.Node) map[string]*app.Node {
	m := map[string]*app.Node{}
	for i := range nodes {
		m[nodes[i].ID] = &nodes[i]
	}
	return m
}

func hasEdge(edges []app.Edge, source, target string, kind app.EdgeKind) bool {
	for _, e := range edges {
		if e.Source == source && e.Target == target && e.Kind == kind {
			return true
		}
	}
	return false
}
