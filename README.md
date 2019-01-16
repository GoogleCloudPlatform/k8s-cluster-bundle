# Cluster Bundle

[![GoDoc](https://godoc.org/github.com/GoogleCloudPlatform/k8s-cluster-bundle?status.svg)](https://godoc.org/github.com/GoogleCloudPlatform/k8s-cluster-bundle)

**Note: This is not an officially supported Google product**

The Cluster Bundle is a Kubernetes-related project developed to improve the way
we package and release Kubernetes software. It was created by the Google
Kubernetes team, and was developed based on our experience managing Kubernetes
clusters and applications at scale in both GKE and GKE On-Prem.

It is currently experimental and will have likely have frequent, breaking
changes for a while until the API settles down.

## An Introduction to Packaging in the Cluster Bundle

Packaging in the cluster bundle revolves around a new type, called the
Component:

* **Component**: A versioned collection of Kubernetes objects. This
  should correspond to a logical application.
* **Component Set**: A set of references to Components.

The Cluster Bundle APIs are minimal and focused. In particular, they are
designed to represent Kubernetes components without worrying about deployment.
It is assumed that external deployment mechanisms like a command line interface
or a deployment controller will consume the components and apply the components
to a cluster.

### Components

In the wild, component packages look something like the following:

```yaml
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Component
spec:
  # A human readable name for the component. The combination of name + version
  # should be unique in a cluster.
  componentName: etcd-component

  # Version of the component, representing changes to the manifest (like a flag
  # change) or the to container image[s].
  version: 30.0.2

  # Kubernetes objects that make up this component.
  objectFiles:
  - url: 'file://etcd-server.yaml'
```

All cluster objects can be *inlined*, which means they are imported directly
into the component, which allows the component package to hermetically describe
the component. After inlining, the component might look like:

```yaml
apiVersion: 'bundle.gke.io/v1alpha1'
kind: Component
spec:
  componentName: etcd-component
  version: 30.0.2
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
    # and so on
```

Additionally, raw text can be imported into a component package. When inlined,
this text is converted into a config map and then added to the objects list:

```yaml
apiVersion: bundle.gke.io/v1alpha1
kind: Component
spec:
  canonicalName: data-blob
  version: 0.1.0
  rawTextFiles:
  - url: 'file://some-data.txt'
```

### ComponentSets

Component Sets are collections of components, which makes it possible publish,
validate, and deploy components as a single unit.

```yaml
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentSet
spec:
  version: '2.3.4'

  components:
  - etcd-server-30.0.2
  - data-blob-0.1.0
```

## Bundle CLI (bundlectl)

`bundlectl` is the name for the command line interface and is the
standard way for interacting with components and component sets

Install with `go install
github.com/GoogleCloudPlatform/k8s-cluster-bundle/cmd/bundlectl`

### Validation

Components and ComponentSets have various constraints that must be validated.
For functions in the library to work, the components are generalled assumed to
have already been validated. For the ease of validating collections of
components, the Bundle CLI knows about a type called `ComponentData` and has the
form:

```yaml
componentFiles:
- url: <component>
```

And so, in the examples directory, you will see:

```yaml
componentFiles:
- url: 'file://etcd/etcd-component.yaml'
- url: 'file://nodes/nodes-component.yaml'
- url: 'file://kubernetes/kubernetes-component.yaml'
- url: 'file://kubedns/kubedns-component.yaml'
- url: 'file://kubeproxy/kube-proxy-component.yaml'
- url: 'file://datablob/data-blob-component.yaml'
```

To validate component data, run:

```
bundlectl validate -f <component-data>
```

### Inlining

Inlining pulls files from the various URLs and imports the contents directly
into the component. To inline component data, run:

```
bundlectl inline -f <component-data>
```

### Filtering

Filtering allows removal of components and objects from the bundle.

```
# Filter all config map objects
bundlectl filter -f <component-data> --kind=ConfigMap

# Keep only config map objects.
bundlectl filter -f <component-data> --kind=ConfigMap --keep-only
```

## Development

### Directory Structure

This directory should follow the structure the standard Go package layout
specified in https://github.com/golang-standards/project-layout

*   `pkg/`: Library code.
*   `examples/`: Examples of components and component packages.
*   `pkg/apis`: APIs and schema for the Cluster Bundle.
*   `cmd/`: Binaries. In particular, this contains the `bundler` CLI which
    assists in modifying and inspecting Bundles.

### Building and Testing

The Cluster Bundle relies on [Bazel](https://bazel.build/) for building and
testing, but the library is go-gettable and it works equally well to use `go
build` and `go test` for building and testing.

### Testing

To run the unit tests, run

```shell
bazel test ...
```

Or, it should work fine to use the `go` command

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
