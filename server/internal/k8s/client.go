// Package k8s adapts the live Kubernetes API into Foundry's topology domain:
// a read-only graph builder (the TopologyRepository) and the audited cluster
// actions. It talks to the API server directly — in-cluster config when running
// as a pod, falling back to KUBECONFIG for local out-of-cluster development.
package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClientSet resolves a Kubernetes client. It returns (nil, nil) when no
// cluster is reachable so Foundry still boots; the topology surface then reports
// the cluster as unavailable instead of failing.
func NewClientSet() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cfg, err = outOfClusterConfig()
		if err != nil {
			return nil, nil
		}
	}
	return kubernetes.NewForConfig(cfg)
}

func outOfClusterConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	if _, err := os.Stat(kubeconfig); err != nil {
		return nil, err
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
