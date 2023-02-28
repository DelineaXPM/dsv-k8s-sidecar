package pods

import (
	"context"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	log "github.com/sirupsen/logrus"
)

type podRegistry struct {
	tenant         string
	registeredPods map[string]*corev1.Pod
	create         chan *corev1.Pod
	update         chan *corev1.Pod
	remove         chan *corev1.Pod
}

type PodRegistry interface {
	Get(name string) *corev1.Pod
	Done()
}

func NewPodRegistry(tenant, namespace string) PodRegistry {

	log.Info("Creating Pod Registry")
	ctx := context.Background()

	client, err := k8s.NewInClusterClient()

	if err != nil {
		log.Fatal(err)
	}

	create := make(chan *corev1.Pod)
	update := make(chan *corev1.Pod)
	remove := make(chan *corev1.Pod)

	pods := make(map[string]*corev1.Pod)
	registry := &podRegistry{
		tenant,
		pods,
		create,
		update,
		remove,
	}

	var pod corev1.Pod
	watcher, err := client.Watch(ctx, namespace, &pod)

	if err != nil {
		log.WithField("error", err.Error()).Fatal("cannot create Pod event watcher")
	}

	go func(tenant string, watcher *k8s.Watcher, create, update, remove chan<- *corev1.Pod) {
		//TODO: This should be change from pull to push model , otherwise this will create a problem
		for {
			p := new(corev1.Pod)
			eventType, err := watcher.Next(p)
			log.WithField("eventType", eventType).Info("watcher new eventType")
			if err != nil {
				log.WithField("error", err.Error()).Error("Error getting pod")
				watcher, err = client.Watch(ctx, namespace, &pod)
				if err != nil {
					log.WithField("error", err.Error()).Error("not able to recreate pod watcher")
					panic(err)
				}
				continue
			}

			t, exists := p.Metadata.Annotations["dsv"]
			if !exists || tenant != t {
				continue
			}

			log.WithFields(log.Fields{
				"event":     eventType,
				"name":      *p.Metadata.Name,
				"namespace": *p.Metadata.Namespace,
				"message":   *p.Status.Message,
			}).Info("Received Pod Event")

			switch eventType {
			case "ADDED":
				create <- p
			case "MODIFIED":
				update <- p
			case "DELETED":
				remove <- p
			default:
				log.WithField("event", eventType).Error("Unable to find event")
			}
		}
	}(tenant, watcher, create, update, remove)

	go registry.addRegistry(create)
	go registry.updateRegistry(update)
	go registry.removeRegistry(remove)

	return registry
}

func (r *podRegistry) Get(name string) *corev1.Pod {
	return r.registeredPods[name]
}

func (r *podRegistry) Done() {
	close(r.create)
	close(r.update)
	close(r.remove)
}

func (r *podRegistry) addRegistry(pods <-chan *corev1.Pod) {
	for p := range pods {
		name := *p.Metadata.Name
		nameSpace := *p.Metadata.Namespace
		r.registeredPods[nameSpace+"/"+name] = p
		log.WithFields(log.Fields{
			"name": name,
		}).Info("Pod Added")
	}
}

func (r *podRegistry) updateRegistry(pods <-chan *corev1.Pod) {
	for p := range pods {
		name := *p.Metadata.Name
		nameSpace := *p.Metadata.Namespace
		r.registeredPods[nameSpace+"/"+name] = p
		log.WithFields(log.Fields{
			"name": name,
		}).Info("Pod Updated")
	}
}

func (r *podRegistry) removeRegistry(pods <-chan *corev1.Pod) {
	for p := range pods {
		name := *p.Metadata.Name
		nameSpace := *p.Metadata.Namespace
		delete(r.registeredPods, nameSpace+"/"+name)
		log.WithFields(log.Fields{
			"name": name,
		}).Info("Pod Removed")
	}
}
