package main

import (
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

// D2 is the namespace for mage tasks related to D2.
type D2 mg.Namespace

// 🧹 Fmt runs D2 fmt for formatting a D2 diagram.
func (D2) Fmt() error {
	magetoolsutils.CheckPtermDebug()
	args := []string{"fmt", constants.D2OverviewDiagram}
	return sh.RunV("d2", args...)
}

// ⬆️ Serve runs d2 serve for interactively building and previewing a D2 diagram.
func (D2) Serve() error {
	magetoolsutils.CheckPtermDebug()
	args := []string{"--dark-theme=200", "-l", "elk", "--watch", "--sketch", "--pad", "0", constants.D2OverviewDiagram, constants.D2OverviewDiagramSVG}
	return sh.RunV("d2", args...)
}
