// ⚡ Core Mage Tasks.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/k8s"
	"github.com/bitfield/script"

	//mage:import
	_ "github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/kind"
	//mage:import
	_ "github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/cert"
	//mage:import
	_ "github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/helm"
	// This breaks the app because the new version of google.golang.org/grpc is not compatible with the old version of grpc v1.16.0.
	// "github.com/DelineaXPM/dsv-k8s/v2/magefiles/helm".

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/fancy"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"

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
	var err error
	fancy.IntroScreen(ci.IsCI())
	pterm.Success.Println("running Init()...")

	mg.SerialDeps(
		Clean,
		createDirectories,
		(gotools.Go{}.Tidy),
	)

	if ci.IsCI() {
		pterm.Debug.Println("CI detected, installing remaining CI required tools")
		pterm.DefaultSection.Println("aqua install of CI tooling")
		if err := sh.RunV("aqua", "install", "--tags", "ci"); err != nil {
			pterm.Error.Printfln("aqua install not successful, is the aqua installed?")
			return fmt.Errorf("aqua install not successful, is the aqua installed? %w", err)
		}
		pterm.Success.Println("Init() complete")
		return nil
	}
	// These can run in parallel as different toolchains.
	// Mg.Deps(
	//
	// ).
	pterm.DefaultSection.Println("Setup Project Specific Tools")
	if err := tooling.SilentInstallTools(toolList); err != nil {
		return err
	}

	// if err := sh.Run("docker", "pull", "alpine:latest"); err != nil {
	// 	return err
	// }

	mg.Deps(
		k8s.K8s{}.Init,
	)

	if runtime.GOOS == "windows" {
		pterm.Warning.Printfln("Trunk is not supported on windows, must run in WSL2, skipping trunk install")
	} else {
		if err = InstallTrunk(); err != nil {
			pterm.Error.Printfln("failed to install trunk (try installing manually from: https://trunk.io/): %v", err)
			return err
		}
		mg.Deps(TrunkInit)
	}

	// Aqua install is run in devcontainer/codespace automatically.
	// If this environment isn't being used, try to jump start, but if failure, output warning and let the developer choose if they want to go install or not.
	pterm.DefaultSection.Println("aqua install of tooling")
	if err := sh.RunV("aqua", "install"); err != nil {
		pterm.Warning.Printfln("aqua install not successful.\n" +
			"This is optional, but will ensure every tool for the project is installed and matching version." +
			"To install see developer docs or go to https://aquaproj.github.io/docs/reference/install")
	}
	pterm.Success.Println("Init() complete")
	return nil
}

// Clean up after yourself.
func Clean() {
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

// InstallTrunk installs trunk.io tooling if it isn't already found.
func InstallTrunk() error {
	magetoolsutils.CheckPtermDebug()
	_, err := exec.LookPath("trunk")
	if err != nil && os.IsNotExist(err) {
		pterm.Warning.Printfln("unable to resolve aqua cli tool, please install for automated project tooling setup: https://aquaproj.github.io/docs/tutorial-basics/quick-start#install-aqua")
		_, err := script.Exec("curl https://get.trunk.io -fsSL").Exec("bash -s -- -y").Stdout()
		if err != nil {
			return err
		}
	} else {
		pterm.Success.Printfln("trunk.io already installed, skipping")
	}
	return nil
}

// TrunkInit ensures the required runtimes are installed.
func TrunkInit() error {
	return sh.RunV("trunk", "install")
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
