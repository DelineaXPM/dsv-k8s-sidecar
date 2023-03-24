# dsv-broker

![Version: 0.2.3](https://img.shields.io/badge/Version-0.2.3-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

A Helm chart for the Delinea DevOps Secrets Vault (DSV) k8 sidecar broker.

Install directly with helm via:

```shell
NAMESPACE='dsv'
CREDENTIALS_JSON_FILE='.cache/credentials.json'
IMAGE_REPOSITORY='docker.io/delineaxpm/dsv-broker'

helm install
    --namespace $NAMESPACE
    --create-namespace \
    --set-file credentialsJson=${CREDENTIALS_JSON_FILE} \
      --set image.repository=${IMAGE_REPOSITORY} \
    dsv-broker ./charts/dsv-broker
```

## Maintainers

| Name             | Email | Url |
| ---------------- | ----- | --- |
| Delinea DSV Team |       |     |

## Values

| Key                                   | Type   | Default                                 | Description                                                    |
| ------------------------------------- | ------ | --------------------------------------- | -------------------------------------------------------------- |
| configmap.authType                    | string | `"client_credentials"`                  |                                                                |
| configmap.clientID                    | string | `nil`                                   |                                                                |
| configmap.clientSecret                | string | `nil`                                   |                                                                |
| configmap.dsvAPIURL                   | string | `"https://%s.secretsvaultcloud.com/v1"` |                                                                |
| configmap.logLevel                    | string | `"info"`                                |                                                                |
| configmap.refreshTime                 | string | `"5m"`                                  |                                                                |
| configmap.tenant                      | string | `nil`                                   |                                                                |
| fullnameOverride                      | string | `""`                                    |                                                                |
| image.pullPolicy                      | string | `"IfNotPresent"`                        |                                                                |
| image.repository                      | string | `"dsv-k8s-controller"`                  |                                                                |
| image.tag                             | string | `"latest"`                              | Overrides the image tag whose default is the chart appVersion. |
| nameOverride                          | string | `""`                                    |                                                                |
| podAnnotations                        | object | `{}`                                    |                                                                |
| podSecurityContext                    | object | `{}`                                    |                                                                |
| replicaCount                          | int    | `1`                                     | replicate count @default - 1                                   |
| resources                             | object | `{}`                                    |                                                                |
| securityContext                       | object | `{}`                                    |                                                                |
| service.brokerauth.tcpport            | int    | `80`                                    |                                                                |
| service.brokerauth.tcptargetPort      | int    | `8080`                                  |                                                                |
| service.brokerauth.tlsport            | int    | `443`                                   |                                                                |
| service.brokerauth.tlstargetPort      | int    | `443`                                   |                                                                |
| service.brokergrpc.port               | int    | `80`                                    |                                                                |
| service.brokergrpc.targetPort         | int    | `3000`                                  |                                                                |
| strategy.rollingUpdate.maxSurge       | int    | `1`                                     |                                                                |
| strategy.rollingUpdate.maxUnavailable | int    | `1`                                     |                                                                |
| strategy.type                         | string | `"RollingUpdate"`                       |                                                                |

---

Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
