package find

import (
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

const schedulerExample = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    component: kube-scheduler
    tier: control-plane
  name: kube-scheduler
  namespace: kube-system
spec:
  containers:
  - command:
    - /bin/sh
    - -c
    - exec /usr/local/bin/kube-scheduler --v=2  --kubeconfig=/etc/srv/kubernetes/kube-scheduler/kubeconfig
      1>>/var/log/kube-scheduler.log 2>&1
    image: gcr.io/google_containers/kube-scheduler:v1.9.7
    name: kube-scheduler
    resources:
      requests:
        cpu: 75m
    volumeMounts:
    - mountPath: /var/log/kube-scheduler.log
      name: logfile
      readOnly: false
    - mountPath: /etc/srv/kubernetes
      name: srvkube
      readOnly: true
  hostNetwork: true`

func TestImageFinder_Found(t *testing.T) {
	s, err := converter.Struct.YAMLToProto([]byte(schedulerExample))
	if err != nil {
		t.Fatalf("error converting obj: %v", err)
	}
	key := core.ClusterObjectKey{"foo", "bar"}
	found := findImagesInKubeObj(key, converter.ToStruct(s))

	if len(found) != 1 {
		t.Fatalf("Got found %v, but wanted exactly 1", len(found))
	}
	if found[0].Key != key {
		t.Errorf("Got found[0].Key %v, but wanted key %v", found[0].Key, key)
	}
	wantImage := "gcr.io/google_containers/kube-scheduler:v1.9.7"
	if found[0].Image != wantImage {
		t.Errorf("Got found[0].Images %q, but wanted %q", found[0].Image, wantImage)
	}
}

const kubeDNSServiceExample = `apiVersion: v1
kind: Service
metadata:
  name: kube-dns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    addonmanager.kubernetes.io/mode: Reconcile
    kubernetes.io/name: "KubeDNS"
spec:
  selector:
    k8s-app: kube-dns
  clusterIP: 10.0.0.10
  ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP`

func TestImageFinder_NotFound(t *testing.T) {
	s, err := converter.Struct.YAMLToProto([]byte(kubeDNSServiceExample))
	if err != nil {
		t.Fatalf("error converting obj: %v", err)
	}
	key := core.ClusterObjectKey{"foo", "biff"}
	found := findImagesInKubeObj(key, converter.ToStruct(s))
	if len(found) != 0 {
		t.Fatalf("Got len(found)=%d, but wanted none", len(found))
	}
}

var loggerExample = `
apiVersion: v1
kind: Pod
metadata:
  name: logger
spec:
  dnsPolicy: Default
  containers:
  - name: logger
    image: gcr.io/floof/logger
    command:
       - /logger
       - --logtostderr
  - name: chopper
    image: gcr.io/floof/chopper
    command:
       - /chopper
       - --logtostderr`

func TestImageFinder_MultipleImages(t *testing.T) {
	s, err := converter.Struct.YAMLToProto([]byte(loggerExample))
	if err != nil {
		t.Fatalf("error converting obj: %v", err)
	}
	key := core.ClusterObjectKey{"gloo", "logger"}
	found := findImagesInKubeObj(key, converter.ToStruct(s))
	if len(found) != 2 {
		t.Fatalf("Got len(found)=%v, but wanted exactly 2", len(found))
	}

	if found[0].Key != key {
		t.Errorf("Got found[0].Key %v, but wanted key %v", found[0].Key, key)
	}
	wantImage := "gcr.io/floof/logger"
	if found[0].Image != wantImage {
		t.Errorf("Got found[0].Image %q, but wanted %q", found[0].Image, wantImage)
	}

	if found[1].Key != key {
		t.Errorf("Got found[1].Key %v, but wanted key %v", found[0].Key, key)
	}
	wantImage = "gcr.io/floof/chopper"
	if found[1].Image != wantImage {
		t.Errorf("Got found[1].Image %q, but wanted %q", found[1].Image, wantImage)
	}
}

var bundleExample = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  components:
  - name: logger
    clusterObjects:
    - name: logger-pod
      inlined:
        apiVersion: v1
        kind: Pod
        metadata:
          name: logger-pod
        spec:
          dnsPolicy: Default
          containers:
          - name: logger
            image: gcr.io/floof/logger
            command:
               - /logger
               - --logtostderr
          - name: chopper
            image: gcr.io/floof/chopper
            command:
               - /chopper
               - --logtostderr
  - name: zap
    clusterObjects:
    - name: zap
      inlined:
        apiVersion: v1
        kind: Pod
        metadata:
          name: zap-pod
  - name: dap
    clusterObjects:
    - name: dap
      inlined:
        apiVersion: v1
        kind: Pod
        metadata:
          name: dap-pod
        spec:
          containers:
          - name: dapper
            image: gcr.io/floof/dapper`

func TestImageFinder_Bundle(t *testing.T) {
	s, err := converter.Bundle.YAMLToProto([]byte(bundleExample))
	if err != nil {
		t.Fatalf("error converting bundle: %v", err)
	}

	finder := &ImageFinder{converter.ToBundle(s)}
	found, err := finder.ContainerImages()
	if err != nil {
		t.Fatalf("error finding images: %v", err)
	}
	expected := []*ContainerImage{
		&ContainerImage{
			core.ClusterObjectKey{"logger", "logger-pod"},
			"gcr.io/floof/logger",
		},
		&ContainerImage{
			core.ClusterObjectKey{"logger", "logger-pod"},
			"gcr.io/floof/chopper",
		},
		&ContainerImage{
			core.ClusterObjectKey{"dap", "dap"},
			"gcr.io/floof/dapper",
		},
	}

	if !reflect.DeepEqual(found, expected) {
		t.Errorf("For finding cluster object images, got %v, but wanted %v", found, expected)
	}
}

var bundleExampleNodeConfig = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  nodeConfigs:
  - name: 'ubuntu-control-plane'
    initFile: "echo 'I'm a script'"
    osImage:
      url: 'gs://base-os-images/ubuntu/ubuntu-1604-xenial-20180509-1'
    envVars:
      - name: FOO_VAR
        value: 'foo-val'
  - name: 'ubuntu-cluster-node'
    initFile: "echo 'I'm another script'"
    osImage:
      url: 'gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1'
  - name: 'ubuntu-cluster-node-no-image'
    initFile: "echo 'I'm another script'"`

func TestImageFinder_NodeImages(t *testing.T) {
	s, err := converter.Bundle.YAMLToProto([]byte(bundleExampleNodeConfig))
	if err != nil {
		t.Fatalf("error converting bundle: %v", err)
	}

	finder := &ImageFinder{converter.ToBundle(s)}
	found, err := finder.NodeImages()
	if err != nil {
		t.Fatalf("error finding images: %v", err)
	}
	expected := []*NodeImage{
		&NodeImage{"ubuntu-control-plane", "gs://base-os-images/ubuntu/ubuntu-1604-xenial-20180509-1"},
		&NodeImage{"ubuntu-cluster-node", "gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1"},
	}

	if !reflect.DeepEqual(found, expected) {
		t.Errorf("For finding node images, got %v, but wanted %v", found, expected)
	}
}

var bundleExampleAll = `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: '1.9.7.testbundle-zork'
spec:
  nodeConfigs:
  - name: 'ubuntu-control-plane'
    initFile: "echo 'I'm a script'"
    osImage:
      url: 'gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1'
    envVars:
      - name: FOO_VAR
        value: 'foo-val'
  - name: 'ubuntu-cluster-node'
    initFile: "echo 'I'm another script'"
    osImage:
      url: 'gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1'
  - name: 'ubuntu-cluster-node-no-image'
    initFile: "echo 'I'm another script'"
  components:
  - name: logger
    clusterObjects:
    - name: logger-pod
      inlined:
        apiVersion: v1
        kind: Pod
        metadata:
          name: logger-pod
        spec:
          dnsPolicy: Default
          containers:
          - name: logger
            image: gcr.io/floof/logger
            command:
               - /logger
               - --logtostderr
          - name: chopper
            image: gcr.io/floof/chopper
            command:
               - /chopper
               - --logtostderr
  - name: zap
    clusterObjects:
    - name: zap
      inlined:
        apiVersion: v1
        kind: Pod
        metadata:
          name: zap-pod
  - name: dap
    clusterObjects:
    - name: dap
      inlined:
        apiVersion: v1
        kind: Pod
        metadata:
          name: dap-pod
        spec:
          containers:
          - name: dapper
            image: gcr.io/floof/dapper
          - name: verydapper
            image: gcr.io/floof/dapper`

func TestImageFinder_AllFlattened(t *testing.T) {
	s, err := converter.Bundle.YAMLToProto([]byte(bundleExampleAll))
	if err != nil {
		t.Fatalf("error converting bundle: %v", err)
	}

	finder := &ImageFinder{converter.ToBundle(s)}
	found, err := finder.AllImages()
	if err != nil {
		t.Fatalf("error finding images: %v", err)
	}
	expected := &AllImagesFlattened{
		NodeImages:      []string{"gs://google-images/ubuntu/ubuntu-1604-xenial-20180509-1"},
		ContainerImages: []string{"gcr.io/floof/logger", "gcr.io/floof/chopper", "gcr.io/floof/dapper"},
	}

	flat := found.Flattened()
	if !reflect.DeepEqual(flat, expected) {
		t.Errorf("For finding all images, got %v, but wanted %v", flat, expected)
	}
}
