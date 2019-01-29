# Basic Component Example

This directory provides a basic example of a component building process,
somewhat arbitrarily picking `etcd` as the concrete example.

The `etcd-component.yaml` was generated frome the `etcd-component-builder.yaml` using

`go run ../../cmd/bundlectl/main.go build --input-file=etcd-component-builder.yaml > etcd-component.yaml`

Patching can be done with:

`go run ../../cmd/bundlectl/main.go patch --input-file=etcd-component.yaml --options-file=options.yaml`

Alternatively, a subset of patches can be applied with:

`go run ../../cmd/bundlectl/main.go patch --input-file=etcd-component.yaml --options-file=options.yaml --patch-annotations=build-label,test`
