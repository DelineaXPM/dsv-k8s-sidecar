package helm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/req"
)

type Helm mg.Namespace

// Docs generates helm documentation using `helm-doc` tool.
func (Helm) Docs() error {
	binary, err := req.ResolveBinaryByInstall("helm-docs", "github.com/norwoodj/helm-docs/cmd/helm-docs@latest")
	if err != nil {
		return err
	}
	for _, chart := range constants.HelmChartsList {
		pterm.DefaultSection.Printfln("Generating docs for %s", chart.ReleaseName)
		err := sh.Run(binary,
			"--chart-search-root", chart.ChartPath,
			"--output-file", "README.md",
			// NOTE: using default layout, but can change here if we wanted.
			// "--template-files", filepath.Join("magefiles", "helm", "README.md.gotmpl"),
		)
		if err != nil {
			return fmt.Errorf("helm-docs failed: %w", err)
		}
		pterm.Success.Printfln("generated file: %s", filepath.Join(chart.ChartPath, "README.md"))
	}
	pterm.Success.Println("(Helm) Docs() - Successfully generated readmes for charts")

	return nil
}

// invokeHelm is a wrapper for running the helm binary.
func invokeHelm(args ...string) error {
	binary, err := req.ResolveBinaryByInstall("helm", "helm.sh/helm/v3@latest")
	if err != nil {
		return err
	}
	return sh.Run(binary, args...)
}

// 🚀 InstallCharts installs the helm charts for any charts listed in constants.HelmChartsList.
func (Helm) InstallCharts() error {
	if os.Getenv("KUBECONFIG") != ".cache/config" {
		pterm.Warning.Printfln("KUBECONFIG is not set to .cache/config. Make sure direnv/env variables loading if you want to keep the project changes from changing your user KUBECONFIG.")
	}
	for _, chart := range constants.HelmChartsList {
		if err :=
			invokeHelm("install",
				chart.ReleaseName,
				chart.ChartPath,
				"--namespace", constants.KubectlNamespace,
				"--atomic",  // if set, the installation process deletes the installation on failure. The --wait flag will be set automatically if --atomic is used
				"--replace", // re-use the given name, only if that name is a deleted release which remains in the history. This is unsafe in production
				"--wait",    // waits, those atomic already runs this
				"--values", filepath.Join(chart.ChartPath, "values.yaml"),
				"--timeout", constants.HelmTimeout,
				"--force",             // force resource updates through a replacement strategy
				"--wait-for-jobs",     // will wait until all Jobs have been completed before marking the release as successful
				"--dependency-update", // update dependencies if they are missing before installing the chart
				// NOTE: Can pass credentials/certs etc in. NOT ADDED YET - "--set-file", "sidecar.configFile=config.yaml",
			); err != nil {
			return err
		}
	}
	return nil
}
