package pods

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func pod(namespace, name, tenant, ip string) *corev1.Pod {
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name, UID: types.UID("uid-" + name)},
		Status:     corev1.PodStatus{PodIP: ip},
	}
	if tenant != "" {
		p.Annotations = map[string]string{"dsv": tenant}
	}
	return p
}

func TestPodRegistryGet(t *testing.T) {
	client := fake.NewSimpleClientset(
		pod("app", "mine", "tenant-a", "10.0.0.1"),
		pod("app", "other-tenant", "tenant-b", "10.0.0.2"),
		pod("app", "unannotated", "", "10.0.0.3"),
	)

	r, err := newPodRegistry(client, "tenant-a", "", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Done()

	if p := r.Get("app/mine"); p == nil || p.Status.PodIP != "10.0.0.1" {
		t.Errorf("Get(app/mine) = %v, want the tenant-a pod", p)
	}
	for _, key := range []string{"app/other-tenant", "app/unannotated", "app/missing", "mine"} {
		if p := r.Get(key); p != nil {
			t.Errorf("Get(%q) = %s, want nil", key, p.Name)
		}
	}
}

func TestPodRegistryFollowsWatchEvents(t *testing.T) {
	client := fake.NewSimpleClientset()

	r, err := newPodRegistry(client, "tenant-a", "app", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Done()

	pods := client.CoreV1().Pods("app")
	ctx := context.Background()

	if _, err := pods.Create(ctx, pod("app", "p", "tenant-a", "10.0.0.1"), metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "pod added", func() bool { return r.Get("app/p") != nil })

	// Removing the annotation must revoke the pod, not leave it registered.
	if _, err := pods.Update(ctx, pod("app", "p", "", "10.0.0.1"), metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "annotation removed", func() bool { return r.Get("app/p") == nil })

	if _, err := pods.Update(ctx, pod("app", "p", "tenant-a", "10.0.0.9"), metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "pod updated", func() bool { p := r.Get("app/p"); return p != nil && p.Status.PodIP == "10.0.0.9" })

	if err := pods.Delete(ctx, "p", metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	eventually(t, "pod deleted", func() bool { return r.Get("app/p") == nil })
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
