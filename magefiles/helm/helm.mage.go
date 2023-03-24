package helm

import (
	"fmt"
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
