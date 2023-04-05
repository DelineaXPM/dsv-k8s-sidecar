# Local Kubernetes

For easier development workflow, the project has prebuilt tasks for minikube.

## Tilt

Project includes tilt.
Once you run `aqua install && mage init` you can run `tilt up` to bring up a local tilt UI for running tasks and streaming kubernetes logs.
This is not required, but might make things a little more friendly over running many terminals.

> NOTE: zsh is expected as the default terminal to work correctly.

## Working With Kubernetes & Stack Locally

> **_NOTE_**
> For any tasks get more help with `-h`, for example, run `mage -h k8s:init`

For local development, Mage tasks have been created to automate most of the setup and usage for local testing.

- run `mage job:setup` to setup a local k8s cluster, initial local copies of the helm chart and kubernetes manifest files.
- Modify the `.cache/dsv-k8s-sidecar/values.yaml` with your list of secrets to map and base image to attach sidecar to.
- Modify the `.cache/charts/dsv-k8s-controller/values.yaml` configuration to include the client credentials to connect to.
- To deploy (or redeploy after changes) all the helm charts and kuberenetes manifests run `mage job:redeploy`.

## Loading Images

You can build the docker images locally and then load into minikube with the following command.

```shell
mage buildall minikube:loadimages
```

If you don't want to build locally and instead use the docker images, you can run the following.
Make sure to update your `values.yaml` files in `.cache/charts/**/values.yaml` to point to this image instead of the local image.

```shell
minikube image load --profile dsvtest 'docker.io/delineaxpm/dsv-k8s-controller:latest'
minikube image load --profile dsvtest 'docker.io/delineaxpm/dsv-k8s-sidecar:latest'
```
