<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# env

```go
import "github.com/DelineaXPM/dsv-k8s-sidecar/pkg/env"
```

## Index

- [Constants](#constants)
- [Variables](#variables)
- [type EnvironmentAgent](#type-environmentagent)
  - [func CreateEnvironmentAgent(client secrets.DsvClient) EnvironmentAgent](#func-createenvironmentagent)

## Constants

```go
const SecretEnvName = "DSV_SECRETS"
```

## Variables

```go
var (
    SecretFile = configDir + util.EnvString("SECRET_FILE", "dsv.json")
)
```

## type [EnvironmentAgent](https://github.com/DelineaXPM/dsv-k8s-sidecar/blob/main/pkg/env/service.go#L25-L29)

```go
type EnvironmentAgent interface {
    Run() <-chan error
    UpdateEnv()
    Close()
}
```

### func [CreateEnvironmentAgent](https://github.com/DelineaXPM/dsv-k8s-sidecar/blob/main/pkg/env/service.go#L37)

```go
func CreateEnvironmentAgent(client secrets.DsvClient) EnvironmentAgent
```

Generated by [gomarkdoc](https://github.com/princjef/gomarkdoc)
