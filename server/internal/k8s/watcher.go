package k8s

import (
	"context"
	"io"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Watcher is the live layer: SharedInformers on the workload/pod/node types
// notify subscribers on any change so the stream pushes a fresh topology
// snapshot. It also opens following pod-log streams. Implements
// app.TopologyWatcher + app.LogStreamer.
type Watcher struct {
	cs   kubernetes.Interface
	mu   sync.Mutex
	subs map[int]chan struct{}
	next int
	stop chan struct{}
}

func NewWatcher(cs kubernetes.Interface) *Watcher {
	return &Watcher{cs: cs, subs: map[int]chan struct{}{}}
}

// Start wires the informers; a no-op when no cluster is reachable.
func (w *Watcher) Start() {
	if w.cs == nil {
		return
	}
	w.stop = make(chan struct{})
	factory := informers.NewSharedInformerFactory(w.cs, 30*time.Second)
	h := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(any) { w.notify() },
		UpdateFunc: func(any, any) { w.notify() },
		DeleteFunc: func(any) { w.notify() },
	}
	for _, inf := range []cache.SharedIndexInformer{
		factory.Apps().V1().Deployments().Informer(),
		factory.Apps().V1().StatefulSets().Informer(),
		factory.Core().V1().Pods().Informer(),
		factory.Core().V1().Nodes().Informer(),
	} {
		_, _ = inf.AddEventHandler(h)
	}
	factory.Start(w.stop)
}

func (w *Watcher) Stop() {
	if w.stop != nil {
		close(w.stop)
	}
}

func (w *Watcher) notify() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, ch := range w.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Subscribe returns a coalescing tick channel (buffer 1) and a cancel func.
func (w *Watcher) Subscribe() (<-chan struct{}, func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	id := w.next
	w.next++
	ch := make(chan struct{}, 1)
	w.subs[id] = ch
	return ch, func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		if c, ok := w.subs[id]; ok {
			delete(w.subs, id)
			close(c)
		}
	}
}

// OpenPodLogs returns a following log stream for one pod container.
func (w *Watcher) OpenPodLogs(ctx context.Context, namespace, pod, container string) (io.ReadCloser, error) {
	tail := int64(200)
	return w.cs.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{
		Container: container,
		Follow:    true,
		TailLines: &tail,
	}).Stream(ctx)
}
