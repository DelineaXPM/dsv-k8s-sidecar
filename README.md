# dsv-k8s-sidecar

## Getting Started

- [Developer](DEVELOPER.md): instructions on running tests, local tooling, and other resources.
- [DSV Documentation](https://docs.delinea.com/dsv/current?ref=githubrepo)

There are two applications that are built in this repo: sidecar and broker

The sidecar container is a responsible for fetching and periodically update a configuration file stored at a shared volume.
This is defined as a shared volume by the pods within the container (see `example.yml`).
Note: there is no guarantee that the file has been created by the time the companion containers are online.
The sidecar must have the following ENV variables defined:

```yaml
- name: DSV_SECRETS
  value: foo bar us-east-1/baz
- name: POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
```

The top value defines the paths of the secrets intended to be used by the container.
This is a list separated by spaces.
The broker watches for new pods with a the specific annotation `dsv` to come online with the value of the `tenant` intended to be used, it then adds this pod to the internal registry.
When the sidecar comes online it must first call the auth endpoint using it's podname and ip address.
The broker requires the following environmental variables to be present:

```yaml
- name: TENANT
  value: tenant_name
- name: CLIENT_ID
  value:
- name: CLIENT_SECRET
  value:
```

The role definition at the beginning of the `broker.yml` file is required for the broker pod to run.
The services below are also required because the sidecar uses the name to make internal calls based on the service name.
Customers should use a similar file to run the services in their cluster.

In order to run the following flags are required.

### Controller

| Flags             | Description                |
| ----------------- | -------------------------- |
| `tenant`          | abbreviation for tenant    |
| `client-id`       | - client credential id     |
| `client-secret`   | - client credential secret |
| `port` (optional) | port to run on             |

### Client

Client uses OS environment variables for configuration.

| Environment Variables | Description                                                                                                                                                 |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CONTROLLER_SERVICE`  | name of the controller service that is running on the kubernetes instance                                                                                   |
| `CONFIG_DIRECTORY`    | location of where to find configuration files                                                                                                               |
| `REFRESH_TIME`        | how often the client should look for updates to configuration (default 15m) NOTE: uses [golang duration format](https://golang.org/pkg/time/#ParseDuration) |

## Build and Test

- Build: `mage build` and view artifacts in `.artifacts/platform/<binaryname>`
- Test: `mage go:testsum ./...`

> Tip: Chain commands like `mage build go:testsum ./...`

## Default Ports

| Port | App         |
| ---- | ----------- |
| 3000 | Server      |
| 8080 | Auth Server |

## Kubernetes

Examples of kubernetes files can be found in the `k8s` folder

## Todos

- Push token instead of pull
- Certificate auth instead of JWT

### Sample Applications for QA Testing

- [example/app1](example/app1)
- [example/app2](example/app2)

## Running Project Against Local Kind Cluster

> **Note**
> Further directions on development setup are in [Developer - Tasks](DEVELOPER.md#tasks).

- `mage cert:generate` to create the local certs in `.cache` directory.

> Pending, mage tasks for creation of broker/example. Currently if you copy these into artifacts, it will apply

```shell
kubectl create secret generic keys --from-file=../certs/server.key --from-file=../certs/server.crt
```

Deploy the manifests

```shell
mage k8s:apply ./k8s/broker.yml
mage k8s:apply ./k8s/example.yml
```

## Optional Running Locally with TLS

There are two communication between sidecard and broker

1. Getting JWT token via http/https
2. Secrets via GRPC

Optionally we can encrypt these communications at container level.

### SideCard to Broker GRPC

- Generate certificates
- run generate-certs in the certs folder
- create kubernetes secret

```shell
kubectl delete secret keys
kubectl create secret generic keys --from-file=key.pem --from-file=cert.pem --from-file=ca.pem
```

### SideCard to Broker Token

- run generate-certs in the certstoken folder
- Create kubernetes secret

```shell
kubectl delete secret keys
kubectl create secret generic keys --from-file=keytoken.pem --from-file=certtoken.pem --from-file=catoken.pem
```

Once the above setup is done all kubernetes secret will mapped to volume and both the sidecard and broker will pick certificates up from volume.
