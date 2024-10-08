// Minikube package contains all the tasks for automation of kind cluster creation and tear down, and the required kubectl commands to correctly use this.
package minikube

import (
	"fmt"
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	mtu "github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

// _minikube is the binary name for minikube.
const _minikube = "minikube"

// Minikube contains the kind cli commands.
type Minikube mg.Namespace

// checkKubeConfig outputs a warning if the kubeconfig is not set to the local project, as this can cause confusion.
func checkKubeConfig() {
	mtu.CheckPtermDebug()
	if os.Getenv("KUBECONFIG") != constants.Kubeconfig {
		pterm.Warning.Printfln("KUBECONFIG is not set to %s, this can cause confusion, you probably haven't loaded direnv which should take care of this", constants.Kubeconfig)
	}
}
func createCluster() error {
	mtu.CheckPtermDebug()
	minikubeArgs := []string{
		"start",
		"--profile", constants.KindClusterName,
		"--namespace", constants.KubectlNamespace,
	}
	// if os.Getenv("KIND_SETUP_CONFIG") != "" {
	// 	pterm.Info.Printfln("KIND_SETUP_CONFIG: %s", os.Getenv("KIND_SETUP_CONFIG"))
	// 	minikubeArgs = append(minikubeArgs, "--config", os.Getenv("KIND_SETUP_CONFIG"))
	// }
	if err := sh.RunV(
		"minikube",
		minikubeArgs...,
	); err != nil {
		return err
	}
	return nil
}

func updateKubeconfig() error {
	mtu.CheckPtermDebug()
	if _, err := os.Stat(constants.Kubeconfig); os.IsNotExist(err) {
		if _, err := os.Create(constants.Kubeconfig); err != nil {
			pterm.Error.Printfln("unable to create empty placeholder file: %v", err)
		}
	}
	_, err := sh.Output(_minikube, "update-context", "--profile", constants.KindClusterName)
	if err != nil {
		pterm.Error.Println("unable to get minikube cluster info, maybe you need to run mage minikube:init first?")
		return err
	}

	// if err := os.WriteFile(constants.Kubeconfig, []byte(kc), constants.PermissionUserReadWriteExecute); err != nil {
	// 	pterm.Error.Printfln("unable to write kubeconfig to file: %v", err)
	// 	return err
	// }
	pterm.Info.Printfln("kubeconfig updated: %s", constants.Kubeconfig)
	// for now this is only going to be run against Kind cluster.
	// if err := sh.Run(
	// 	"kubectl",
	// 	"cluster-info", "--context", constants.KindContextName,
	// 	"--cluster", constants.KindContextName,
	// ); err != nil {
	// 	return err
	// }
	return nil
}

// 💾 LoadImages loads the images into the minikube cluster.
func (Minikube) LoadImages() {
	mtu.CheckPtermDebug()
	for _, chart := range constants.HelmChartsList {
		// Load image into minikube
		if err := sh.Run(_minikube, "image", "load", "--overwrite", "--profile", constants.KindClusterName, fmt.Sprintf("%s/%s:latest", constants.DockerImageLocalRegistry, chart.ReleaseName)); err != nil { //nolint:revive // ok to have string constant for the minikube profile command.
			pterm.Error.Printfln("unable to load image into minikube: %v", err)
		}
		pterm.Success.Printfln("image loaded into minikube: %s", chart.ReleaseName)
	}
}

// ➕ Create creates a new Minikube cluster and populates a kubeconfig in cachedirectory.
func (Minikube) Init() error {
	mtu.CheckPtermDebug()
	checkKubeConfig()
	if err := createCluster(); err != nil {
		return err
	}
	dspin, _ := pterm.DefaultSpinner.
		WithDelay(time.Second).
		WithRemoveWhenDone(true).
		WithShowTimer(true).
		WithText("Init()\n").
		WithSequence("|", "/", "-", "|", "/", "-", "\\").Start()
	dspin.SuccessPrinter.Println("ensuring it's in kubeconfig")
	if err := updateKubeconfig(); err != nil {
		pterm.Error.Printfln("updateKubeconfig(): %v", err)
	}
	dspin.UpdateText("setting context")
	if err := sh.Run("kubectl", "config", "use-context", constants.KindContextName); err != nil {
		dspin.WarningPrinter.Printfln("default context might not be setup correct to new context: %v", err)
	}
	if err := sh.Run("kubectl", "config", "set-context", "--context", constants.KindContextName, "--current", "--namespace", constants.KubectlNamespace); err != nil {
		dspin.WarningPrinter.Printfln("default namespace might not be setup correct to new namespace: %v", err)
	}
	// Create the namespace if it doesn't exist.
	dspin.UpdateText("creating namespace if not exists")
	if _, err := sh.Output("kubectl", "get", "namespace", constants.KubectlNamespace); err != nil {
		dspin.UpdateText(fmt.Sprintf("namespace does not exist, creating namespace: %s...", constants.KubectlNamespace))

		if err := sh.Run("kubectl", "create", "namespace", constants.KubectlNamespace); err != nil {
			dspin.FailPrinter.Printfln("unable to create namespace: %v", err)
			return fmt.Errorf("kubectl create namespace %s: %w", constants.KubectlNamespace, err)
		}
		dspin.SuccessPrinter.Printfln("namespace created: %s", constants.KubectlNamespace)
	}
	// dspin.UpdateText("pulling docker images")
	// if err := sh.Run("docker", "pull", constants.DockerImageQualified); err != nil {
	// 	dspin.WarningPrinter.Printfln("docker pull: %v", err)
	// 	return fmt.Errorf("docker pull: %w", err)
	// }
	// dspin.SuccessPrinter.Println("docker image for " + constants.DockerImageQualified)
	// Not working right now, can't find nodes for Kind to preload. Not critical so commenting out for now - sheldon.
	// Sp.UpdateText("preloading docker image into kind cluster")
	// if err := sh.Run("kind", "load", "docker-image", "quay.io/delinea/dsv-k8s:latest"); err != nil {
	// 	return fmt.Errorf("kind load docker-image: %w", err)
	// }.
	dspin.SuccessPrinter.Println("(Minikube) Init()")
	_ = dspin.Stop()
	return nil
}

// 🗑️ Destroy tears down the Kind cluster.
func (Minikube) Destroy() error {
	mtu.CheckPtermDebug()
	checkKubeConfig()
	if err := sh.Run(_minikube, "delete", "--profile", constants.KindClusterName); err != nil {
		pterm.Error.Printfln("minikube delete error: %v", err)
		return err
	}
	if err := sh.Run("kubectl", "config", "unset", fmt.Sprintf("clusters.%s", constants.KindContextName)); err != nil {
		pterm.Warning.Printfln("default context might not be setup correct to new context: %v", err)
	}

	pterm.Success.Println("(Minikube) Destroy()")
	return nil
}

// Dashboard opens the Minikube dashboard.
func (Minikube) Dashboard() error {
	mtu.CheckPtermDebug()
	checkKubeConfig()
	if err := sh.Run("minikube", "dashboard", "--profile", constants.KindClusterName, "addons", "enable", "metrics-service"); err != nil {
		return err
	}
	if err := sh.Run(_minikube, "dashboard", "--profile", "dsvtest"); err != nil {
		return err
	}
	return nil
}

// 🔍 ListImages provides a list of the minikube loaded images
func (Minikube) ListImages() {
	mtu.CheckPtermDebug()
	pterm.DefaultSection.Println("(Minikube) ListImages()")
	if err := sh.RunV("minikube", "image", "ls", "--profile", constants.KindClusterName); err != nil {
		pterm.Error.Printfln("images not listed from minikube: %v", err)
	}
	pterm.Success.Printfln("images listed from minikube")
}

// 💾 RemoveImages removes the images both local and docker registered from the minikube cluster.
func (Minikube) RemoveImages() {
	mtu.CheckPtermDebug()
	var output string
	// var err error
	var elapsed time.Duration
	for _, chart := range constants.HelmChartsList {
		for {
			// Run the docker rmi command and capture the output

			cmd := exec.Command( //nolint:gosec // this is a local command being built
				"minikube",
				"image",
				"rm",
				"--profile", constants.KindClusterName,
				fmt.Sprintf("%s", chart.ReleaseName),
			)
			out, err := cmd.CombinedOutput()
			output = string(out)
			if err != nil {
				pterm.Error.Printfln("image not rm from minikube: %v", err)
			}
			// Check if the output contains the image name
			if !strings.Contains(output, chart.ReleaseName) ||
				strings.Contains(output, fmt.Sprintf("No such image: %s", chart.ReleaseName)) {
				pterm.Success.Printfln("image unloaded")
				break
			}

			// If the image is still being unloaded, print a progress message
			pterm.Info.Printf("Still waiting for image [%s] to unload (elapsed time: %s)\n", chart.ReleaseName, elapsed.Round(time.Second))

			// Wait for 3 seconds before trying again
			time.Sleep(3 * time.Second) //nolint:gomnd // no need to make a constant
			elapsed += 3 * time.Second
		}
		pterm.Success.Printfln("image removed from minikube: %s", chart.ReleaseName)
	}
}
