# dsv-k8s-sidecar

## Overview

There are two applications that are built in this repo:

| Application        | Description                                                                                                                                                                               |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| dsv-k8s-sidecar    | Responsible for fetching and periodically updating a configuration file stored at a shared volume that is used by the pods.                                                               |
| dsv-k8s-controller | The controller watches for new pods with the specific annotation `dsv` to come online with the value of the `tenant` intended to be used, it then adds this pod to the internal registry. |

> **_Note_**
> There is no guarantee that the file has been created by the time the companion containers are online.

## Installing

Both the sidecar & controller have helm charts located in [charts](charts/) with `README.md` files containing a reference for the input values required.

## How It Works

See [Architecture](docs/architecture.md) for more detail.

The general concept is:

- DSV Controller retrieves and caches secrets from DSV.
- Authenticated sidecar pods communicate with a unique JWT to the DSV Controller requesting the desired secrets.
- The secret is either read from the in-memory cache or retrieved if non-existent.

## FAQ

- Do I need more than one controller?
  - One controller can do the job required.
  - If you want to scope the controller to a specific namespace and/or client credential for more isolation, then you could consider installing more.

## Development

- See [developer](docs/developer-quick-start.md)

## Possible Future Improvements

- Push token instead of pull
- Certificate auth instead of JWT

If there are needs missing for your usage, feel free to open a GitHub issue describing your challenges and any suggestions for improvement.

### Sample Applications for QA Testing

- [example/app1](examples/app1)

## Running Project Against Local Kind Cluster

> **Note**
> Further directions on development setup are in [Developer - Tasks](DEVELOPER.md#tasks).

- `mage cert:generate` to create the local certs in `.cache` directory.

> Currently if you copy these into artifacts, it will apply.

- For creation of the secret in development mode: `mage k8s:createsecret`
- For a customer: `kubectl create secret generic keys --from-file=mysecretpath/server.key --from-file=mysecretpath/server.crt`

Dev Deployment:

- Deploy the manifests individually: `mage k8s:apply ./.cache/charts/k8s/controller.yml`.
- Deploy all locally: `mage helm:install`.

## Optional Running Locally with TLS

There are two communication between sidecard and controller

1. Getting JWT token via http/https
2. Secrets via GRPC

Optionally we can encrypt these communications at container level.

## Generate self signed certificate

- run `mage cert:generate` and choose `Sidecar To Controller`: This will generate certs and keys in .cache folder.
- create kubernetes secret: `mage k8s:createsecret` or manually: kubectl create secret generic keys --from-file=key.pem --from-file=cert.pem --from-file=ca.pem

### SideCard to Controller GRPC

Add above k8 secret as volume in Controller's k8 deployment and add the name of cert and private key name env in k8 values.yml.
`KEY_DIR` => the volume directory.
`SERVER_CRT` => this will be certs.
`SERVER_KEY` => this will be private key.

### SideCard to Controller Token

Add above k8 secret as volume in sidecar's k8 deployment and add the name of cert env in k8 values.yml.
`KEY_DIR` => the volume directory.
`SERVER_CRT` => this will be certs.

Once the above setup is done all kubernetes secret will mapped to volume and both the sidecard and controller will pick certificates up from volume.

## Additional Resources

- [Developer](DEVELOPER.md): instructions on running tests, local tooling, and other resources.
- [DSV Documentation](https://docs.delinea.com/dsv/current?ref=githubrepo)
- [DSV-K8S](https://github.com/DelineaXPM/dsv-k8s) is another approach using a Kubernetes syncing and injector hook to directly update Kubernetes secrets.
  This alternative approach does not leverage a sidecar.
