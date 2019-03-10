package generate

import "io/ioutil"
import "os"
import "path"

const componentBuilder = `apiVersion: bundle.gke.io/v1alpha1
kind: ComponentBuilder
componentName: etcd-component
version: 30.0.2
objectFiles:
- url: /etcd-server.yaml
- url: /patches.yaml
`

const etcdServer = `apiVersion: v1
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
      --name etcd-server-primary
      --listen-peer-urls $(LISTEN_PEER_URLS)
      --initial-advertise-peer-urls $(INITIAL_ADVERTISE_PEER_URLS)
      --advertise-client-urls $(CLIENT_URLS)
      --listen-client-urls $(CLIENT_URLS)
      --quota-backend-bytes=6442450944
      --data-dir /var/etcd/data
      --initial-cluster-state new
      --initial-cluster etcd-server-primary=$(CLIENT_URLS)
      1>>/var/log/etcd.log
      2>&1
    env:
    - name: TARGET_STORAGE
      value: etcd3
    - name: TARGET_VERSION
      value: 3.1.11
    - name: DATA_DIRECTORY
      value: /var/etcd/data
    - name: INITIAL_CLUSTER
      value: etcd-server-primary=http://10.1.1.1:2380
    - name: LISTEN_PEER_URLS
      value: http://10.1.1.1:2380
    - name: INITIAL_ADVERTISE_PEER_URLS
      value: http://10.1.1.1:2380
    - name: CLIENT_URLS
      value: http://127.0.0.1:2379
    - name: ETCD_CREDS
      value: ""
    image: k8s.gcr.io/etcd:3.1.11
    livenessProbe:
      httpGet:
        host: 127.0.0.1
        path: /health
        port: 2379
      initialDelaySeconds: 15
      timeoutSeconds: 15
    name: etcd-container
    ports:
    - containerPort: 2380
      hostPort: 2380
      name: serverport
    - containerPort: 2379
      hostPort: 2379
      name: clientport
    resources:
      requests:
        cpu: 200m
    volumeMounts:
    - mountPath: /var/etcd
      name: varetcd
      readOnly: false
    - mountPath: /var/log/etcd.log
      name: varlogetcd
      readOnly: false
    - mountPath: /etc/srv/kubernetes
      name: etc
      readOnly: false
  hostNetwork: true
  volumes:
  - hostPath:
      path: /mnt/disks/master-pd/var/etcd
    name: varetcd
  - hostPath:
      path: /var/log/etcd.log
      type: FileOrCreate
    name: varlogetcd
  - hostPath:
      path: /etc/srv/kubernetes
    name: etc
`

const patches = `apiVersion: bundle.gke.io/v1alpha1
kind: PatchTemplate
template: |
  apiVersion: v1
  kind: Pod
  metadata:
    namespace: {{.namespace}}
optionsSchema:
  required:
  - namespace
  properties:
    namespace:
      type: string
      pattern: '^[a-z0-9-]+$'
---
apiVersion: bundle.gke.io/v1alpha1
kind: PatchTemplate
metadata:
  annotations:
    build-label-experiment: test
optionsSchema:
  properties:
    buildLabel:
      type: string
      default: dev-env
      pattern: '^[a-z0-9-]+$'
template: |
  apiVersion: v1
  kind: Pod
  metadata:
    annotations:
      build-label: {{.buildLabel}}
`
const options = `namespace: foo-namespace
buildLabel: test-build
`

const testSuite = `componentFile: etcd-component-builder.yaml
rootDirectory: './'

testCases:
- description: Success
  apply:
    options:
      namespace: default-ns
      buildLabel: test-env
  expect:
    objects:
    - kind: Pod
      name: etcd-server
      findSubstrs:
      - 'build-label: test-env'
      - 'image: k8s.gcr.io/etcd:3.1.11'
      - 'namespace: default-ns'
      notFindSubstrs:
      - 'build-label: dev-env'

- description: 'Success: default build label'
  apply:
    options:
      namespace: default-ns
  expect:
    objects:
    - kind: Pod
      name: etcd-server
      findSubstrs:
      - 'build-label: dev-env'
      - 'namespace: default-ns'
      notFindSubstrs:
      - 'build-label: test-env'

- description: 'Fail: parameter missing'
  expect:
    applyErrSubstr: 'namespace in body is required'

- description: 'Fail: parameter does not match regex'
  apply:
    options:
      namespace: default-ns
      buildLabel: '1263[]){'
  expect:
    applyErrSubstr: "buildLabel in body should match '^[a-z0-9-]+$'"
`

func GenerateComponent(filepath string, includeTestSuite bool) error {
	os.Mkdir(filepath, 0777)
	ioutil.WriteFile(path.Join(filepath, "etcd-component-builder.yaml"), []byte(componentBuilder), 0666)
	ioutil.WriteFile(path.Join(filepath, "etcd-server.yaml"), []byte(etcdServer), 0666)
	ioutil.WriteFile(path.Join(filepath, "patches.yaml"), []byte(patches), 0666)
	ioutil.WriteFile(path.Join(filepath, "options.yaml"), []byte(options), 0666)
	if includeTestSuite {
		ioutil.WriteFile(path.Join(filepath, "test-suite.yaml"), []byte(testSuite), 0666)
	}
	return nil
}
