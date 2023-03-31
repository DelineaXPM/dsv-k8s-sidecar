package constants

import (
	"path/filepath"
)

type HelmCharts struct {
	ReleaseName string
	ChartPath   string
	Namespace   string
	// Values is the customized yaml file placed in the CacheDirectory to run install with.
	// Copy the values.yaml from the helm chart to start here.
	Values string
}

var HelmChartsList = []HelmCharts{
	{
		ReleaseName: "dsv-k8s-controller",
		ChartPath:   filepath.Join(ChartsDirectory, "dsv-k8s-controller"),
		Namespace:   "dsv",
		Values:      filepath.Join(CacheDirectory, "dsv-k8s-controller", "values.yaml"),
	},
	{
		ReleaseName: "dsv-k8s-sidecar",
		ChartPath:   filepath.Join(ChartsDirectory, "dsv-k8s-sidecar"),
		Namespace:   "dsv",
		Values:      filepath.Join(CacheDirectory, "dsv-k8s-sidecar", "values.yaml"),
	},
}
