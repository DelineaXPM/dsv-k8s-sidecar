package pods

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	syncTimeout = time.Minute
	// Watch events keep the store current, so periodic resync is disabled.
	noResync time.Duration = 0
)

var errSyncTimeout = errors.New("pod informer did not sync")

type podRegistry struct {
	tenant string
	store  cache.Store
	stop   chan struct{}
}

type PodRegistry interface {
	Get(name string) *corev1.Pod
	Done()
}

func NewPodRegistry(tenant, namespace string) (PodRegistry, error) { //nolint:ireturn // PodRegistry is the seam auth tests mock.
	log.Info("Creating Pod Registry")

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return newInformerRegistry(client, tenant, namespace, syncTimeout)
}

// newInformerRegistry watches pods through an informer, whose store is safe
// for concurrent reads while the informer applies watch events and re-lists
// after a dropped watch.
func newInformerRegistry(client kubernetes.Interface, tenant, namespace string, timeout time.Duration) (*podRegistry, error) {
	factory := informers.NewSharedInformerFactoryWithOptions(client, noResync, informers.WithNamespace(namespace))
	informer := factory.Core().V1().Pods().Informer()

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { logPodEvent(tenant, "ADDED", obj) },
		UpdateFunc: func(_, obj any) { logPodEvent(tenant, "MODIFIED", obj) },
		DeleteFunc: func(obj any) { logPodEvent(tenant, "DELETED", obj) },
	})
	if err != nil {
		return nil, err
	}

	stop := make(chan struct{})
	factory.Start(stop)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		close(stop)
		return nil, fmt.Errorf("%w within %s", errSyncTimeout, timeout)
	}

	return &podRegistry{tenant, informer.GetStore(), stop}, nil
}

// Get returns the pod with the given "namespace/name" key, or nil unless the
// pod carries this tenant's "dsv" annotation.
func (r *podRegistry) Get(name string) *corev1.Pod {
	obj, exists, err := r.store.GetByKey(name)
	if err != nil || !exists {
		return nil
	}

	pod, ok := obj.(*corev1.Pod)
	if !ok || pod.Annotations["dsv"] != r.tenant {
		return nil
	}

	return pod
}

func (r *podRegistry) Done() {
	close(r.stop)
}

func logPodEvent(tenant, eventType string, obj any) {
	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		obj = tombstone.Obj
	}

	pod, ok := obj.(*corev1.Pod)
	if !ok || pod.Annotations["dsv"] != tenant {
		return
	}

	log.WithFields(log.Fields{
		"event":     eventType,
		"name":      pod.Name,
		"namespace": pod.Namespace,
		"message":   pod.Status.Message,
	}).Info("Received Pod Event")
}
