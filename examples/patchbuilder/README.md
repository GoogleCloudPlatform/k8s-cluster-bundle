# Patch template builder example

This example shows how to package up the
[hello-app](https://github.com/GoogleCloudPlatform/kubernetes-engine-samples/tree/master/hello-app)
as a Component and store that in the API server. Then, we deploy multiple
instances of that Component with different options.

In this directory, you'll find the following:

* `builder.yaml` is the ComponentBuilder that pulls everything together
* `helloweb-deployment.yaml` is a manifest for the helloweb deployment
* `helloweb-service-clusterip.yaml` a manifest for the helloweb service
* `deployment-patch-builder.yaml` is a manifest building the deployment patch
  template, which will include a build-time parameter.
* `service-patch.yaml` is the manifest for the service patch template, which
  includes only run-time parameters

```
$ bundlectl build --input-file builder.yaml --options-file build-options.yaml
```

which results in a Component that includes a PatchTemplate. This we can load
this directly into the API server, since it is a custom resource:

```
$ bundlectl build --input-file builder.yaml --options-file build-options.yaml | kubectl apply -f -
I0208 14:21:43.065640  147647 bundleio.go:101] Reading input file builder.yaml
component.bundle.gke.io "helloweb-0.1.0" configured
```

This can now be patched and the Component deployed using the `patch` and
`export` commands.

```
$ kubectl get component helloweb-0.1.0 -o yaml \
    | bundlectl patch --options-file deploy-options-1.yaml \
    | bundlectl export \
    | kubectl apply -f -
I0208 15:01:38.117065  165280 bundleio.go:107] No component data file, reading from stdin
I0208 15:01:38.118502  165281 bundleio.go:107] No component data file, reading from stdin
I0208 15:01:38.350527  165280 patch.go:77] Patching component
deployment.apps "helloweb" created
service "helloweb" created
```

To deploy in a different namepace, you will have to create that namespace first:

```
$ kubectl create ns foo
namespace "foo" created
$ kubectl get component helloweb-0.1.0 -o yaml \
    | bundlectl patch --options-file deploy-options-2.yaml \
    | bundlectl export \
    | kubectl apply -f -
I0208 15:03:12.239470  165666 bundleio.go:107] No component data file, reading from stdin
I0208 15:03:12.243220  165667 bundleio.go:107] No component data file, reading from stdin
I0208 15:03:12.468814  165666 patch.go:77] Patching component
deployment.apps "helloweb" created
service "helloweb" created
```

Now, we can see the two different deployment from the same Component source.

```
$ kubectl get deploy,svc helloweb
NAME                             DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/helloweb   1         1         1            1           1m

NAME               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/helloweb   ClusterIP   10.7.254.157   <none>        80/TCP    1m
$ kubectl get deploy,svc helloweb -n foo
NAME                             DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/helloweb   1         1         1            1           20s

NAME               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/helloweb   ClusterIP   10.7.249.118   <none>        80/TCP    20s
```
