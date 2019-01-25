# Basic Component Example

This directory provides a basic example of a component building process,
somewhat arbitrarily picking `etcd` as the concrete example.

The `etcd-component.yaml` was generated frome the `etcd-component-builder.yaml` using

`go run ../../cmd/bundlectl/main.go build --input-file=etcd-component-builder.yaml > etcd-component.yaml`
