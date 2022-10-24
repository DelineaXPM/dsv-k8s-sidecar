// ⚡ Core Mage Tasks.
package main

import (
	"os"

	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"
	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/k8s"

	//mage:import
	_ "github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/kind"
	//mage:import
	_ "github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/cert"
	// This breaks the app because the new version of google.golang.org/grpc is not compatible with the old version of grpc v1.16.0.
	// "github.com/DelineaXPM/dsv-k8s/v2/magefiles/helm".

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/fancy"
	"github.com/sheldonhull/magetools/pkg/req"
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
	fancy.IntroScreen(ci.IsCI())
	pterm.Success.Println("running Init()...")

	mg.SerialDeps(
		Clean,
		createDirectories,
		(gotools.Go{}.Tidy),
	)

	if ci.IsCI() {
		pterm.Debug.Println("CI detected, installing remaining CI required tools")
		if err := tooling.SilentInstallTools(CITools); err != nil {
			return err
		}
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

	if err := sh.Run("docker", "pull", "alpine:latest"); err != nil {
		return err
	}

	mg.Deps(
		k8s.K8s{}.Init,
	)
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

// 🔨 Build builds the project for the current platform.
func Build() error {
	binary, err := req.ResolveBinaryByInstall("goreleaser", "github.com/goreleaser/goreleaser@latest")
	if err != nil {
		return err
	}

	return sh.RunV(binary,
		"build",
		"--rm-dist",
		"--snapshot",
		"--single-target",
	)
}
