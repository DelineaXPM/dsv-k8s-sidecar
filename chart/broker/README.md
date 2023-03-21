helm install --dry-run --debug broker -f ./broker/values.yaml --generate-name

helm install broker broker -f ./broker/values.yaml

helm uninstall broker broker
