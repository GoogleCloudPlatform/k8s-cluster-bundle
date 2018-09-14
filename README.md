# Cluster Bundle

**Note: This is not an officially supported Google product**

The Cluster Bundle is a Kubernetes-related project developed by the Google
Kubernetes team, that comes from Google's experience creating Kubernetes
clusters in GKE and GKE On-Prem.

The Cluster Bundle is a Kubernetes type and has the goal of packaging all the
necessary **software** that bootstraps and forms a functioning Kubernetes
Cluster. It is currently *highly* experimental and will have likely have
frequent, unannounced, breaking changes for a while until the API settles down.

The Cluster Bundle has three components:

*   **Type**: A declarative, hermetic expression of the Cluster, and is designed
    for building automation related to managing Kubernetes clusters.
*   **Library**: Go-library code for interacting with the Cluster Bundle
*   **CLI**: A minimal CLI for interacting with Cluster Bundles.

## Directory Structure

This directory should follow the structure the standard Go package layout
specified in https://github.com/golang-standards/project-layout

*   `pkg/`: Library code.
*   `pkg/apis`: APIs and schema for the Cluster Bundle.
*   `cmd/`: Binaries. In particular, this contains the `bundler` CLI which
    assists in modifying and inspecting Bundles.

## Building and Testing

The Cluster Bundle relies on [Bazel](https://bazel.build/) for building and
testing.

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

If new files are added, you may need to re-run Gazelle.
