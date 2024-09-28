// ⚡ Core Mage Tasks.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"

	//mage:import
	_ "github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/minikube"
	//mage:import
	_ "github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/cert"
	//mage:import
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/helm"
	//mage:import
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/k8s"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/fancy"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
	"github.com/sheldonhull/magetools/trunk"

	"github.com/sheldonhull/magetools/tooling"

	// mage:import
	_ "github.com/sheldonhull/magetools/docgen"
	// mage:import
	"github.com/sheldonhull/magetools/gotools"
)

// createDirectories creates the local working directories for build artifacts and tooling.
func createDirectories() error {
	for _, dir := range []string{constants.ArtifactDirectory, constants.CacheDirectory} {
		if err := os.MkdirAll(dir, constants.PermissionUserReadWriteExecute); err != nil {
			pterm.Error.Printf("failed to create dir: [%s] with error: %v\n", dir, err)

			return err
		}
		pterm.Success.Printf("✅ [%s] dir created\n", dir)
	}

	return nil
}

// Init runs multiple tasks to initialize all the requirements for running a project for a new contributor.
func Init() error { //nolint:deadcode // Not dead, it's alive.
	magetoolsutils.CheckPtermDebug()

	fancy.IntroScreen(ci.IsCI())
	pterm.Success.Println("running Init()...")

	mg.SerialDeps(
		Clean,
		createDirectories,
		(gotools.Go{}.Tidy),
	)

	if ci.IsCI() {
		pterm.Success.Println("Init() complete, exiting early per CI detected")
		return nil
	}
	pterm.DefaultSection.Println("Setup Project Specific Tools")
	if err := tooling.SilentInstallTools(toolList); err != nil {
		return err
	}
	mg.Deps(
		k8s.K8s{}.Init,
		helm.Helm{}.Init,
	)

	if runtime.GOOS == "windows" {
		pterm.Warning.Printfln("Trunk is not supported on windows, must run in WSL2, skipping trunk install")
	} else {

		mg.SerialDeps(
			trunk.Trunk{}.Init,
			trunk.Trunk{}.Install,
		)

		if _, err := exec.LookPath("direnv"); err != nil {
			pterm.Warning.Printfln("non-terminating] direnv not detected. recommend setup and shell integration to automatically load .envrc project configuration")
		}
	}

	// Aqua install is run in devcontainer/codespace automatically.
	// If this environment isn't being used, try to jump start, but if failure, output warning and let the developer choose if they want to go install or not.
	pterm.DefaultSection.Println("aqua install of tooling")
	if err := sh.RunV("aqua", "policy", "allow", ".aqua/aqua-policy.yaml"); err != nil {
		pterm.Warning.Printfln("aqua policy not successful.\n" +
			"This is optional, but will ensure every tool for the project is installed and matching version." +
			"To install see developer docs or go to https://aquaproj.github.io/docs/reference/install")
	}
	pterm.Success.Println("aqua policy allow")
	if err := sh.RunV("aqua", "install"); err != nil {
		pterm.Warning.Printfln("aqua install not successful.\n" +
			"This is optional, but will ensure every tool for the project is installed and matching version." +
			"To install see developer docs or go to https://aquaproj.github.io/docs/reference/install")
	}
	pterm.Success.Printfln("aqua install")
	pterm.Success.Println("Init() complete")
	return nil
}

// 🧹 Clean up after yourself.
func Clean() {
	magetoolsutils.CheckPtermDebug()
	pterm.Success.Println("Cleaning...")
	for _, dir := range []string{constants.ArtifactDirectory, constants.CacheDirectory} {
		err := os.RemoveAll(dir)
		if err != nil {
			pterm.Error.Printf("failed to removeall: [%s] with error: %v\n", dir, err)
		}
		pterm.Success.Printf("🧹 [%s] dir removed\n", dir)
	}
	mg.Deps(createDirectories)
}

// getVersion returns the version and path for the changefile to use for the semver and release notes.
func getVersion() (releaseVersion, cleanPath string, err error) { //nolint:unparam // leaving as optional parameter for future release tasks.
	releaseVersion, err = sh.Output("changie", "latest")
	if err != nil {
		pterm.Error.Printfln("changie pulling latest release note version failure: %v", err)
		return "", "", err
	}
	cleanVersion := strings.TrimSpace(releaseVersion)
	cleanPath = filepath.Join(".changes", cleanVersion+".md")
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		cleanPath = filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".changes", cleanVersion+".md")
	}
	return cleanVersion, cleanPath, nil
}
