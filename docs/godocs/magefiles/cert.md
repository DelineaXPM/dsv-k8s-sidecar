<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# cert

```go
import "github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/cert"
```

## Index

- [type Cert](<#type-cert>)
  - [func (Cert) Generate() error](<#func-cert-generate>)


## type [Cert](<https://github.com/DelineaXPM/dsv-k8s-sidecar/blob/main/magefiles/cert/cert.magefile.go#L22>)

Cert contains tasks to generate cert.

```go
type Cert mg.Namespace
```

### func \(Cert\) [Generate](<https://github.com/DelineaXPM/dsv-k8s-sidecar/blob/main/magefiles/cert/cert.magefile.go#L25>)

```go
func (Cert) Generate() error
```

Generate certs using cffsl \(cloudflare toolkit\). Requires aqua to have installed already.



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
