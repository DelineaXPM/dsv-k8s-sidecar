package pods

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const syncTimeout = time.Minute

type podRegistry struct {
	tenant string
	store  cache.Store
	stop   chan struct{}
}

type PodRegistry interface {
	Get(name string) *corev1.Pod
	Done()
}

func NewPodRegistry(tenant, namespace string) PodRegistry {
	log.Info("Creating Pod Registry")

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	registry, err := newPodRegistry(client, tenant, namespace, syncTimeout)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("cannot create Pod informer")
	}

	return registry
}

// newPodRegistry watches pods through an informer, whose store is safe for
// concurrent reads while the informer applies watch events and re-lists
// after a dropped watch.
func newPodRegistry(client kubernetes.Interface, tenant, namespace string, timeout time.Duration) (*podRegistry, error) {
	factory := informers.NewSharedInformerFactoryWithOptions(client, 0, informers.WithNamespace(namespace))
	informer := factory.Core().V1().Pods().Informer()

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { logPodEvent(tenant, "ADDED", obj) },
		UpdateFunc: func(_, obj interface{}) { logPodEvent(tenant, "MODIFIED", obj) },
		DeleteFunc: func(obj interface{}) { logPodEvent(tenant, "DELETED", obj) },
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
		return nil, context.DeadlineExceeded
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

func logPodEvent(tenant, eventType string, obj interface{}) {
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
