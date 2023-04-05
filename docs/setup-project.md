# Setup Project

## Compatibility

Linux & MacOS is supported out of the box for local development.
Windows with WSL2 should also work fine.

While the majority of this is cross-platform, the automatically linting and some other commands are only compatible with Linux/MacOS.

## Overview

- Mage: Mage is a Go based automation alternative to Make and provides newer functionality for local Minikube cluster setup, Go development tooling/linting, and more.
  Use [aqua](#aqua) to automatically install, or run `go install github.com/magefile/mage@latest`.
- `mage init` to setup developer tooling.

> Get more detail on a task, if it's available by running `mage -h init`.

## Initial Setup

Most of the setup is automated via Mage, but there are some initial assumptions such as Go/Aqua expected to help automate the remaining setup.

## When Using Without Devcontainer/Codespaces

- Install [Aqua](#aqua)
  - Alternative: Manually ensure Go is installed.
- Run `mage init` to install tooling.
  - Done automatically by Mage -> Install [trunk](https://trunk.io/products/check) (quick install script: `curl https://get.trunk.io -fsSL | bash`)
  - This will allow faster installs of project tooling by grabbing binaries for your platform more quickly (most of the time release binaries instead of building from source).

## For someone creating a release

## Aqua

Install [aqua](https://aquaproj.github.io/docs/tutorial-basics/quick-start#install-aqua) and have it configured in your path per directions.

Run `aqua install` for tooling such as changie or others for the project.

Ensure your profile has this in it:

```shell
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH" # for those using aqua this will ensure it's in the path with all tools if loading from home
```

## Direnv

Important: If you don't set this up the minikube and other tasks might not work as expected as this loads the project default environment variables found in `.envrc`.

Direnv: Default test values are loaded on macOS/Linux based system using [direnv](https://direnv.net/docs/installation.html).
Ensure you add the hook to your shell per installation directions.

Run `direnv allow` in the directory to load default env configuration for testing.
