package main

import (
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/k8s"
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/kind"

	"github.com/magefile/mage/mg"
	"github.com/pterm/pterm"
)

// Job is a namespace to contain chained sets of automation actions, to reduce the need to chain many commands together for common workflows.
type Job mg.Namespace

// Setup initializes all the required steps for the cluster creation, initial helm chart copies, and kubeconfig copies.
func (Job) Setup() {
	pterm.DefaultSection.Println("(Job) Setup()")
	mg.SerialDeps(
		kind.Kind{}.Init,
		k8s.K8s{}.Init,
		mg.F(k8s.K8s{}.Apply, constants.CacheManifestDirectory),
		k8s.K8s{}.Logs,
	)
}

// Redeploy removes kubernetes resources and helm charts and then redeploys with log streaming by default.
func (Job) Redeploy() {
	pterm.DefaultSection.Println("(Job) Redeploy()")
	mg.SerialDeps(
		mg.F(k8s.K8s{}.Delete, constants.CacheManifestDirectory),
		mg.F(k8s.K8s{}.Apply, constants.CacheManifestDirectory),
		k8s.K8s{}.Logs,
	)
}
