# CRD To OpenAPI

Read CRD from stdin and write OpenAPI spec to stdout.
```
Usage:
   [flags]

Flags:
      --crd-version string   set crd version
  -h, --help                 help for this command
      --output-openapi-v2    output in OpenAPI v2 format, the default format is OpenAPI v3
```

Example:

```shell
go install github.com/negz/crd-to-openapi@latest
kubectl get crd applications.app.k8s.io -o yaml | crd-to-openapi
```
