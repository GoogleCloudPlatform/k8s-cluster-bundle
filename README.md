# Cluster Bundle

[![GoDoc](https://godoc.org/github.com/GoogleCloudPlatform/k8s-cluster-bundle?status.svg)](https://godoc.org/github.com/GoogleCloudPlatform/k8s-cluster-bundle)

**Note: This is not an officially supported Google product**

The Cluster Bundle is a Kubernetes-related project developed to improve the way
we package and release Kubernetes software. It was created by the Google
Kubernetes team, and was developed based on our experience managing Kubernetes
clusters and applications at scale in both GKE and GKE On-Prem.

It is currently experimental and will have likely have frequent, breaking
changes for a while until the API settles down.

## An Introduction to Cluster Bundle

![Cluster Bundle](https://raw.githubusercontent.com/GoogleCloudPlatform/k8s-cluster-bundle/master/cluster_bundle.png)

The core of the Cluster Bundle are the Kubernetes types that are used to
represent the software components (or packages) and the collections of these
components (or bundles).

* **Cluster Object**: A Kubernetes object, such as a pod, deployment, and so on.
* **Component Package**: A versioned collection of Kubernetes objects.
* **Bundle**: A versioned collection of Kubernetes component packages.

The Cluster Bundle APIs are minimal and focused. In particular, they are
focused only on representing Kubernetes components without worrying about
deployment. It is assumed that external deployment mechanisms (like a CLI) will
consume the components and apply the components to a cluster.

### Component Packages

In the wild, component packages look something like the following:

```yaml
apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ComponentPackage
metadata:
  name: etcd-component
spec:
  # Version of the component, representing changes to the manifest (like a flag
  # change) or the to container image[s].
  version: '30.0.2'

  # Kubernetes objects that make up this component.
  clusterObjectFiles:
  - url: 'file://etcd-server.yaml'
```

All cluster objects can be 'inlined', which means they are imported directly
into the component, which allows the component object to hermetically describe
the component. After inlining, the component might look like:

```yaml
apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ComponentPackage
metadata:
  name: etcd-component
spec:
  version: '30.0.2'
  clusterObjects:
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
this text is converted into a config map and then added to the ClusterObjects:

```yaml
apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
kind: ComponentPackage
metadata:
  name: data-blob
spec:
  version: '0.1'
  rawTextFiles:
  - url: 'file://some-data.txt'
```

### Bundles

Bundles are collections of components, which makes it possible publish,
validate, and deploy components as a single unit. Like components, Bundles can
contain inlined or externalized components:

```ymal

apiVersion: 'gke.io/k8s-cluster-bundle/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  version: '2.3.4'

  componentFiles:
  - url: 'file://etcd-component.yaml'

  components:
  - apiVersion: gke.io/k8s-cluster-bundle/v1alpha1
    kind: ComponentPackage
    metadata:
      name: kubernetes-control-plane
    spec:
      version: '11.0.0'
      componentApiVersion: '1.9.7'
      clusterObjectFiles:
      - url: 'file://kube-apiserver.yaml'
      - url: 'file://kube-scheduler.yaml'
      - url: 'file://kube-controller-manager.yaml'
```

## Bundle CLI (bundlectl)

`bundlectl` is the name for the Bundle command line interface and is the
standard way for interacting with Bundles.

Install with `go install
github.com/GoogleCloudPlatform/k8s-cluster-bundle/cmd/bundlectl`

### Validation

Bundles have various constraints that must be validated. For functions in the
Bundle library to work, the Bundle is generalled assumed to have already been
validated. To validate a Bundle, run:

```
bundle validate -f <bundle>
```

### Inlining

Inlining pulls files from the various URLs and imports the contents directly
into the component. To inline a bundle, run:

```
bundle inline -f <bundle>
```

Additionally, most commands will allow you to inline the bundle by specifying
`--inline`

### Filtering

Filtering allows removal of components and objects from the bundle.

```
# Filter all config map objects
bundle filter -f <bundle> --kind=ConfigMap 

# Keep only config map objects.
bundle filter -f <bundle> --kind=ConfigMap --keep-only
```

## Development

### Directory Structure

This directory should follow the structure the standard Go package layout
specified in https://github.com/golang-standards/project-layout

*   `pkg/`: Library code.
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

### Generating Proto Go Files

Currently, the schema for the Cluster Bundle is specified as Proto files. To
generate the Go client libraries, first install
[protoc-gen-go](https://github.com/golang/protobuf#installation) and run:

```shell
pkg/apis/bundle/v1alpha1/regenerate-sources.sh
```

If new files are added, will to re-run Gazelle (see above).
