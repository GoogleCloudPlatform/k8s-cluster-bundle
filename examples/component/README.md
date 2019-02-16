# Basic Component Example

This directory provides a basic example of a component building process,
somewhat arbitrarily picking `etcd` as the concrete example.

The `etcd-component.yaml` was generated frome the `etcd-component-builder.yaml` using

`go run ../../cmd/bundlectl/main.go build --input-file=etcd-component-builder.yaml > etcd-component.yaml`

Patching can be done with:

```sh
go run ../../cmd/bundlectl/main.go patch \
  --input-file=etcd-component.yaml \
  --options-file=options.yaml
```

Alternatively, a subset of patches can be applied with:

```sh
go run ../../cmd/bundlectl/main.go patch \
  --input-file=etcd-component.yaml \
  --options-file=options.yaml \
  --patch-annotations=build-label-experiment=test
```

## Patch Templates

Patch Templates are a way to customize a kubernetes file. Here is an example
Patch Template:

```yaml
apiVersion: bundle.gke.io/v1alpha1
kind: PatchTemplate
template: |
  apiVersion: v1
  kind: Pod
  metadata:
    namespace:
      {{.namespace}}
```

Patch Templates apply to objects that match the `apiVersion`, `kind`, and
`metadata.name` specified in the `template`. Thus, these fields cannot be
modified with PatchTemplates.
