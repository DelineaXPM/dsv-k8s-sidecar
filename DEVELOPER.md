# Dev Tooling

## Core Requirements

- Docker
- Tooling defined in aqua, such as Go.

## Code & Task Docs

While you can get help via mage such as `mage -h job:setup` for detailed help (when available), there's also [docs](./docs/) that is generated from the Go files in the project.

## Artifacts

_IMPORTANT_: All automated tasks and builds publish to either `.artifacts` or `.cache` and these are in the `.gitignore`.

### First Time Setup

> **_note_**
> Note much of this is automatically setup if using devcontainer.
> You'll still need to run `aqua install && mage init` though.

- Run `mage init` to install tooling.
- Install [trunk](https://trunk.io/products/check) (quick install script: `curl https://get.trunk.io -fsSL | bash`)
- Install [aqua](https://aquaproj.github.io/docs/tutorial-basics/quick-start#install-aqua) and have it configured in your path per directions.
- This will allow faster installs of project tooling by grabbing binaries for your platform.
- Run `aqua install` for tooling such as changie or others for the project.
- At this time, it expects you have to Go pre-installed.
- Already Have Mage Installed?: `mage init`
- Need other tools?

## Tasks

Get a list of currenet automation tasks via: `mage -l`

Start with `mage init` and then to setup a local kind cluster to experiment with, run `mage kind:init`.

## Devcontainer Usage

- Devcontainer configuration included for Codespaces or [Remote Container](https://code.visualstudio.com/docs/remote/containers)

### Prerequisites For Devcontainer

- Docker
- Visual Studio Code
  - Run `code --install-extension ms-vscode-remote.remote-containers`
  - For supporting Codespaces: `code --install-extension GitHub.codespaces`

### Spin It Up

> 🐎 PERFORMANCE TIP: Using the directions provided for named container volume will optimize performance over trying to just "open in container" as there is no mounting files to your local filesystem.
> Use command pallet with vscode (Control+Shift+P or F1) and type to find the command `Remote Containers: Clone Repository in Named Container`.

- Put the git clone url in.
  Some extra features are included such as:

- Extensions for VSCode defined in `.devcontainers`, such as Go, Kubernetes & Docker, and some others.
- Initial placeholder `.zshrc` file included to help initialize usage of `direnv` for automatically loading default `.envrc` which contains local developement default environment variables.

#### After Devcontainer Loads

1. Accept "Install Recommended Extensions" from popup, to automatically get all the preset tools, and you can choose do this without syncing so it's just for this development environment.
2. Open a new `zsh-login` terminal and allow the automatic setup to finish, as this will ensure all other required tools are setup.
   - Make sure to run `direnv allow` as it prompts you, to ensure all project and your personal environment variables (optional).
3. Run setup task:
   - Using CLI: Run `mage init`

### Troubleshooting

#### Mismatch With Checksum for Go Modules

- Run `go clean -modcache && go mod tidy`.

#### Connecting to Services Outside of devcontainer

You are in an isolated, self-contained Docker setup.
The ports internally aren't the same as externally in your host OS.
If the port forward isn't discovered automatically, enable it yourself, by using the port forward tab (next to the terminal tab).

1. You should see a port forward once the services are up (next to the terminal button in the bottom pane).
   1. If the click to open url doesn't work, try accessing the path manually, and ensure it is `https`.
      Example: `https://127.0.0.1:9999`

You can choose the external port to access, or even click on it in the tab and it will open in your host for you.
