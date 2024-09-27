# dsv-k8s-sidecar

![Version: v1.0.2](https://img.shields.io/badge/Version-v1.0.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

A Helm chart for the Delinea DevOps Secrets Vault (DSV) k8 sidecar.
This is installed per service as a sidecar.

Install directly with helm via:

```shell
NAMESPACE='dsv'
IMAGE_REPOSITORY='docker.io/delineaxpm/dsv-k8s-sidecar'

helm install
    --namespace $NAMESPACE
    --create-namespace \
    --atomic \
    --timeout "5m" \
    --set image.repository=${IMAGE_REPOSITORY} \
    dsv-k8s-sidecar ./charts/dsv-k8s-sidecar
```

## Maintainers

| Name             | Email | Url |
| ---------------- | ----- | --- |
| Delinea DSV Team |       |     |

## Values

| Key                                    | Type   | Default                                                                                      | Description                                                                                                                                                                                                                                 |
| -------------------------------------- | ------ | -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| containername                          | string | `nil`                                                                                        | containername is the name of the user provided container. _REQUIRED_                                                                                                                                                                        |
| env.brokerNamespace                    | string | ""                                                                                           | brokerNamespace is the namespace of the broker, which is only required if the broker is not in the same namespace as the sidecar                                                                                                            |
| env.logLevel                           | string | info                                                                                         | logLevel is the log level for the sidecar                                                                                                                                                                                                   |
| env.refreshTime                        | string | 8543                                                                                         | refreshTime is the time interval between refreshes of the secrets                                                                                                                                                                           |
| env.secrets                            | string | `nil`                                                                                        | secrets is a list of secrets to be injected into the pod. This should be comma delimited if providing more than one. For example `dev:mysecret:mypath,dev:mysecret2:mypath2`. _REQUIRED FIELD_                                              |
| fullnameOverride                       | string | `""`                                                                                         |                                                                                                                                                                                                                                             |
| image                                  | object | `{"pullPolicy":"IfNotPresent","repository":null,"tag":null}`                                 | IMPORTANT: This image is the configuration for the user provided service, not the sidecar. The sidecar will be part of this deployment. For example, you could use docker.io/nginx:last and name it nginx as the container name. _REQUIRED_ |
| image.pullPolicy                       | string | IfNotPresent                                                                                 | pullPolicy is the image pull policy                                                                                                                                                                                                         |
| image.repository                       | string | `nil`                                                                                        | repository is the name of the fully qualified image, and is the user image, not the sidecar image. _REQUIRED_                                                                                                                               |
| image.tag                              | string | '' (empty)                                                                                   | Overrides the image tag whose default is the chart appVersion.                                                                                                                                                                              |
| nameOverride                           | string | `""`                                                                                         |                                                                                                                                                                                                                                             |
| podAnnotations                         | object | `{"dsv":null}`                                                                               | podAnnotations is the pod annotations for the controller _REQUIRED_                                                                                                                                                                         |
| podAnnotations.dsv                     | string | `nil`                                                                                        | dsv is the tenant name in DSV _REQUIRED FIELD_                                                                                                                                                                                              |
| podSecurityContext                     | object | `{}`                                                                                         |                                                                                                                                                                                                                                             |
| replicaCount                           | int    | 1                                                                                            | replicaCount                                                                                                                                                                                                                                |
| resources                              | object | `{}`                                                                                         |                                                                                                                                                                                                                                             |
| securityContext                        | object | `{"readOnlyRootFilesystem":true,"runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532}`   | securityContext is the security context for the controller. This uses chainguard static nonroot based image. Reference: https://edu.chainguard.dev/chainguard/chainguard-images/reference/static/overview/                                  |
| securityContext.readOnlyRootFilesystem | bool   | true                                                                                         | readOnlyRootFilesystem is the read only root file system flag.                                                                                                                                                                              |
| securityContext.runAsGroup             | int    | 65532 (from chainguard static image)                                                         | runAsGroup is the run as group.                                                                                                                                                                                                             |
| securityContext.runAsNonRoot           | bool   | true                                                                                         | runAsNonRoot is the run as non root flag.                                                                                                                                                                                                   |
| securityContext.runAsUser              | int    | 65532 (from chainguard static image)                                                         | runAsUser is the run as user.                                                                                                                                                                                                               |
| sidecarimage                           | object | `{"pullPolicy":"IfNotPresent","repository":"docker.io/delineaxpm/dsv-k8s-sidecar","tag":""}` | sidecarimage is DSV Sidecar Image attached to target pod                                                                                                                                                                                    |
| sidecarimage.pullPolicy                | string | IfNotPresent                                                                                 | pullPolicy is the image pull policy                                                                                                                                                                                                         |
| sidecarimage.repository                | string | docker.io/delineaxpm/dsv-k8s-sidecar                                                         | repository is the name of the fully qualified image                                                                                                                                                                                         |
| sidecarimage.tag                       | string | `""`                                                                                         | @default '' (empty)                                                                                                                                                                                                                         |

---

Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
