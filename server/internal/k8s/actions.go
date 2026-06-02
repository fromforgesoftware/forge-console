package k8s

import (
	"context"
	"fmt"
	"io"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"

	apierrors "github.com/fromforgesoftware/go-kit/errors"

	"github.com/fromforgesoftware/forge/server/internal/app"
)

// Actions implements app.ClusterActions against the live Kubernetes API using
// strategic/merge patches and the scale + eviction surfaces.
type Actions struct {
	cs  kubernetes.Interface
	now func() time.Time
}

func NewActions(cs kubernetes.Interface) *Actions {
	return &Actions{cs: cs, now: time.Now}
}

func (a *Actions) requireCluster() error {
	if a.cs == nil {
		return apierrors.ServiceUnavailable("no kubernetes cluster is reachable")
	}
	return nil
}

func (a *Actions) patchWorkload(ctx context.Context, cmd app.WorkloadRefCommand, patch []byte) error {
	if err := a.requireCluster(); err != nil {
		return err
	}
	switch cmd.Kind {
	case "Deployment":
		_, err := a.cs.AppsV1().Deployments(cmd.Namespace).Patch(ctx, cmd.Name, types.MergePatchType, patch, metav1.PatchOptions{})
		return err
	case "StatefulSet":
		_, err := a.cs.AppsV1().StatefulSets(cmd.Namespace).Patch(ctx, cmd.Name, types.MergePatchType, patch, metav1.PatchOptions{})
		return err
	default:
		return apierrors.InvalidArgument(fmt.Sprintf("unsupported workload kind %q", cmd.Kind))
	}
}

func (a *Actions) RestartWorkload(ctx context.Context, cmd app.WorkloadRefCommand) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`,
		a.now().UTC().Format(time.RFC3339))
	return a.patchWorkload(ctx, cmd, []byte(patch))
}

func (a *Actions) ScaleWorkload(ctx context.Context, cmd app.WorkloadScaleCommand) error {
	if err := a.requireCluster(); err != nil {
		return err
	}
	if cmd.Replicas < 0 {
		return apierrors.InvalidArgument("replicas must be >= 0")
	}
	switch cmd.Kind {
	case "Deployment":
		sc, err := a.cs.AppsV1().Deployments(cmd.Namespace).GetScale(ctx, cmd.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		sc.Spec.Replicas = cmd.Replicas
		_, err = a.cs.AppsV1().Deployments(cmd.Namespace).UpdateScale(ctx, cmd.Name, sc, metav1.UpdateOptions{})
		return err
	case "StatefulSet":
		sc, err := a.cs.AppsV1().StatefulSets(cmd.Namespace).GetScale(ctx, cmd.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		sc.Spec.Replicas = cmd.Replicas
		_, err = a.cs.AppsV1().StatefulSets(cmd.Namespace).UpdateScale(ctx, cmd.Name, sc, metav1.UpdateOptions{})
		return err
	default:
		return apierrors.InvalidArgument(fmt.Sprintf("unsupported workload kind %q", cmd.Kind))
	}
}

func (a *Actions) PauseWorkload(ctx context.Context, cmd app.WorkloadRefCommand) error {
	if cmd.Kind != "Deployment" {
		return apierrors.InvalidArgument("only Deployments support pause/resume")
	}
	return a.patchWorkload(ctx, cmd, []byte(`{"spec":{"paused":true}}`))
}

func (a *Actions) ResumeWorkload(ctx context.Context, cmd app.WorkloadRefCommand) error {
	if cmd.Kind != "Deployment" {
		return apierrors.InvalidArgument("only Deployments support pause/resume")
	}
	return a.patchWorkload(ctx, cmd, []byte(`{"spec":{"paused":false}}`))
}

func (a *Actions) DeleteWorkload(ctx context.Context, cmd app.WorkloadRefCommand) error {
	if err := a.requireCluster(); err != nil {
		return err
	}
	switch cmd.Kind {
	case "Deployment":
		return a.cs.AppsV1().Deployments(cmd.Namespace).Delete(ctx, cmd.Name, metav1.DeleteOptions{})
	case "StatefulSet":
		return a.cs.AppsV1().StatefulSets(cmd.Namespace).Delete(ctx, cmd.Name, metav1.DeleteOptions{})
	default:
		return apierrors.InvalidArgument(fmt.Sprintf("unsupported workload kind %q", cmd.Kind))
	}
}

func (a *Actions) CordonNode(ctx context.Context, cmd app.NodeCommand) error {
	return a.cordon(ctx, cmd.Name, true)
}

func (a *Actions) UncordonNode(ctx context.Context, cmd app.NodeCommand) error {
	return a.cordon(ctx, cmd.Name, false)
}

func (a *Actions) cordon(ctx context.Context, name string, desired bool) error {
	if err := a.requireCluster(); err != nil {
		return err
	}
	node, err := a.cs.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return drain.RunCordonOrUncordon(a.drainHelper(ctx), node, desired)
}

// DrainNode cordons then evicts the node's pods using the same drain helper
// kubectl uses (PDB-aware eviction, daemonset/mirror-pod aware).
func (a *Actions) DrainNode(ctx context.Context, cmd app.NodeCommand) error {
	if err := a.cordon(ctx, cmd.Name, true); err != nil {
		return err
	}
	return drain.RunNodeDrain(a.drainHelper(ctx), cmd.Name)
}

func (a *Actions) drainHelper(ctx context.Context) *drain.Helper {
	return &drain.Helper{
		Ctx:                 ctx,
		Client:              a.cs,
		Force:               true,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
		GracePeriodSeconds:  -1,
		Timeout:             2 * time.Minute,
		Out:                 io.Discard,
		ErrOut:              io.Discard,
	}
}
