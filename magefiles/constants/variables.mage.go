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
		ReleaseName: "dsv-broker",
		ChartPath:   filepath.Join(ChartsDirectory, "dsv-broker"),
		Namespace:   "dsv",
		Values:      filepath.Join(CacheDirectory, "dsv-broker", "values.yaml"),
	},
	{
		ReleaseName: "dsv-sidecar",
		ChartPath:   filepath.Join(ChartsDirectory, "dsv-sidecar"),
		Namespace:   "dsv",
		Values:      filepath.Join(CacheDirectory, "dsv-sidecar", "values.yaml"),
	},
}
