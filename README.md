# Cluster Bundle

[![GoDoc](https://godoc.org/github.com/GoogleCloudPlatform/k8s-cluster-bundle?status.svg)](https://godoc.org/github.com/GoogleCloudPlatform/k8s-cluster-bundle)

**Note: This is not an officially supported Google product**

The Cluster Bundle is a Kubernetes-related project developed to improve the way
we package and release Kubernetes software. It was created by the Google
Kubernetes team, and was developed based on our experience managing Kubernetes
clusters and applications at scale in both GKE and GKE On-Prem.

It is currently experimental (Pre-Alpha) and will have frequent, breaking
changes until the API settles down.

## Installation

Most users will interact with Cluster Bundle objects via the command line
interface `bundlectl`.

To install, run:

```shell
go install github.com/GoogleCloudPlatform/k8s-cluster-bundle/cmd/bundlectl
```

## Packaging in the Cluster Bundle

Packaging in Cluster Bundle revolves around a new type, called the `Component`:

* **Component**: A versioned collection of Kubernetes objects. This should
  correspond to a logical application. For example, we might imagine
  components for each of Etcd, Istio, KubeProxy, and CoreDNS.
* **ComponentBuilder**: An intermediate type that allows for easier creation of
  component objects.
* **ComponentSet**: A set of references to components.

The Cluster Bundle APIs are minimal and focused on the problem of packaging
Kubernetes objects into a single unit. It is assumed that some external
deployment mechanisms like a command line interface or an in-cluster controller
will consume the components and apply the Kubernetes objects in the component to
a cluster.

## Components

Here's how you might use the Component type to represent Etcd:

```yaml
apiVersion: bundle.gke.io/v1alpha1
kind: Component
spec:

  # A human readable name for the component.
  componentName: etcd-component

  # SemVer version for the component representing any and all changes to the
  # component or included object manifests. This should not be tied to the
  # application version.
  version: 30.0.2

  # A list of included Kubernetes objects.
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      name: etcd-server
      namespace: kube-system
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - |
          exec /usr/local/bin/etcd
    # ... and so on
```

Some notes about components:

- Components have a name and version. The pair of these should be a unique
  identifier for the component.
- Objects must be Kubernetes objects -- they must have `apiVersion`, `kind`, and
  `metadata`. Other than that, there are no restrictions on what objects can be
  contained in a component.

You can see more about examples of components in the [examples directory](https://github.com/GoogleCloudPlatform/k8s-cluster-bundle/tree/master/examples).

### Building

To make it easier to create components, you can use the `ComponentBuilder` type:

```yaml
apiVersion: bundle.gke.io/v1alpha1
kind: ComponentBuilder

componentName: etcd-component

version: 30.0.2

# File-references to kubernetes objects that make up this component.
objectFiles:
- url: file:///etcd-server.yaml
```

ComponentBuilders can be built into Components with:

```shell
bundlectl build --input-file=my-component.yaml
```

During `build`, objects are inlined, which means they are imported directly
into the component.

Raw-text can also be imported into a component. When inlined, this text is
converted into a ConfigMap and then added to the objects list:

```yaml
apiVersion: bundle.gke.io/v1alpha1
kind: Component
spec:
  canonicalName: data-blob
  version: 0.1.0
  rawTextFiles:
  - name: data-blob
    files:
    - url: file:///some-data.txt
```

### Patching

The Cluster Bundle library provides a new type called the `PatchTemplate` to
perform [Kustomize-style](https://github.com/kubernetes-sigs/kustomize) patching
to objects in component.

For example, if you had the following PatchTemplate in the component

```yaml
apiVersion: bundle.gke.io/v1alpha1
kind: PatchTemplate
template: |
  apiVersion: v1
  kind: Pod
  metadata:
    namespace: {{.namespace}}
```

and the following Kubernetes object:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: etcd-server
spec:
  containers:
  - command:
    - /bin/sh
    - -c
    - |
      exec /usr/local/bin/etcd
  # etc...
```

and the following options YAML file:

```yaml
namespace: foo-namespace
```

Running the `bundlectl` command like so:

```shell
bundlectl patch --input-file=etcd-component.yaml --options-file=options.yaml
```

would produce:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: etcd-server
  namespace: foo-namespace
spec:
  containers:
  - command:
    - /bin/sh
    - -c
    - |
      exec /usr/local/bin/etcd
  # etc...
```

For more examples of Patching and Patch-Building, see the
[examples directory](https://github.com/GoogleCloudPlatform/k8s-cluster-bundle/tree/master/examples).

### Filtering

Filtering removes objects from a component.

```shell
# Filter all config map objects
bundlectl filter -f my-component.yaml --filterType=objects --kind=ConfigMap

# Keep only config map objects.
bundlectl filter -f my-component.yaml --filterType=objects --kind=ConfigMap --keep-only
```

## Public APIs and Compatibility.

During pre-Alpha and Alpha, breaking changes may happen in any of the types in
`pkg/apis`, the `bundlectl` CLI, or the Go APIs (methods, structs, interfaces
not in `pkg/apis`).

During Beta, backwards incompatible breaking changes in the `bundlectl` and the
`pkg/apis` directory will only happy during Major version boundaries.

The Go APIs are not covered by any compatibility guarantee and can break in
backwards incompatible ways during any release.

## Development

### Directory Structure

This directory follows a layout similar to other Kubernetes projects.

*   `pkg/`: Library code.
*   `pkg/apis`: APIs and schema for the Cluster Bundle.
*   `pkg/client`: Generated client for the Cluster Bundle types.
*   `config/crds`: Generated CRDs for the ClusterBundle types.
*   `examples/`: Examples of components
*   `cmd/`: Binaries. This contains the `bundlectl` CLI tool.

### Building and Testing

The Cluster Bundle relies on [Bazel](https://bazel.build/) for building and
testing, but the library is go-gettable and it works equally well to use `go
build` and `go test` for building and testing.

### Testing

To run the unit tests, run

```shell
bazel test ...
```

It should also work to use the `go` command:

```shell
go test ./...
```

### Regenerate BUILD files

To make using Bazel easier, we use Gazelle to automatically write Build targets.
To automatically write the Build targets, run:

```shell
bazel run //:gazelle
```

### Regenerate Generated Code

We rely heavily on Kubernetes code generation. To regenerate the code,
run:

```shell
hacke/update-codegen.sh
```

If new files are added, will to re-run Gazelle (see above).
