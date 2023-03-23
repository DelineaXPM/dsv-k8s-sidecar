helm install --dry-run --debug sidecar -f ./sidecar/values.yaml --generate-name

helm install sidecar sidecar -f ./sidecar/values.yaml

helm uninstall sidecar sidecar
