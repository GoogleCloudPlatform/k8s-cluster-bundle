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

	if found.Key != key {
		t.Errorf("Got found.Key %v, but wanted key %v", found.Key, key)
	}
	if len(found.Images) != 1 {
		t.Fatalf("Got found.Images %v, but wanted exactly 1", found.Images)
	}
	wantImage := "gcr.io/google_containers/kube-scheduler:v1.9.7"
	if found.Images[0] != wantImage {
		t.Errorf("Got found.Images[0] %q, but wanted %q", found.Images[0], wantImage)
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

	if found.Key != key {
		t.Errorf("Got found.Key %v, but wanted key %v", found.Key, key)
	}
	if len(found.Images) != 0 {
		t.Fatalf("Got found.Images %v, but wanted none", found.Images)
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

	if found.Key != key {
		t.Errorf("Got found.Key %v, but wanted key %v", found.Key, key)
	}
	if len(found.Images) != 2 {
		t.Fatalf("Got found.Images %v, but wanted exactly 1", found.Images)
	}
	wantImage := "gcr.io/floof/logger"
	if found.Images[0] != wantImage {
		t.Errorf("Got found.Images[0] %q, but wanted %q", found.Images[0], wantImage)
	}
	wantImage = "gcr.io/floof/chopper"
	if found.Images[1] != wantImage {
		t.Errorf("Got found.Images[P] %q, but wanted %q", found.Images[1], wantImage)
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
	found, err := finder.ClusterObjectImages()
	if err != nil {
		t.Fatalf("error finding images: %v", err)
	}
	expected := []*ClusterObjectImages{
		&ClusterObjectImages{
			core.ClusterObjectKey{"logger", "logger-pod"},
			[]string{"gcr.io/floof/logger", "gcr.io/floof/chopper"},
		},
		&ClusterObjectImages{
			core.ClusterObjectKey{"dap", "dap"},
			[]string{"gcr.io/floof/dapper"},
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
