// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package converter

import (
	"strings"
	"testing"
)

const reschedulerManifest = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    scheduler.alpha.kubernetes.io/critical-pod: ""
  labels:
    k8s-app: rescheduler
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: Rescheduler
    version: v0.3.1
  name: rescheduler-v0.3.1
  namespace: kube-system
spec:
  containers:
  - command:
    - sh
    - -c
    - exec /rescheduler --running-in-cluster=false 1>>/var/log/rescheduler.log 2>&1
    image: k8s.gcr.io/rescheduler:v0.3.1
    name: rescheduler
    resources:
      requests:
        cpu: 10m
        memory: 100Mi
    volumeMounts:
    - mountPath: /var/log/rescheduler.log
      name: logfile
      readOnly: false
  hostNetwork: true
  volumes:
  - hostPath:
      path: /var/log/rescheduler.log
      type: FileOrCreate
    name: logfile
`

func TestConvertUnstructured(t *testing.T) {
	firstIttObj, err := FromYAMLString(reschedulerManifest).ToUnstructured()
	if err != nil {
		t.Fatalf("unexpected error parsing manifest: %v", err)
	}
	// Check if one of the original YAML keys are reachable and value match
	apiversionCheck, ok := firstIttObj.Object["apiVersion"].(string)
	if !ok || apiversionCheck != "v1" {
		t.Fatalf("manifest apiVersion key doesn't match original one:expected=v1, got=%v", apiversionCheck)
	}
	// convert back for sanity
	b, err := FromObject(firstIttObj).ToYAML()
	if err != nil {
		t.Fatalf("unexpected error serializing manifest: %v", err)
	}
	str := string(b)

	// Even though the YAMLs have maps, YAML generation sorts the keys based on
	// name, and so it should be stable.
	if strings.Trim(str, " \n\t") != strings.Trim(reschedulerManifest, " \n\t") {
		t.Errorf("Got serilaized manifest:\n\n%s\nexpected it to equal:\n\n%s", str, reschedulerManifest)
	}
	// Convert YAML to Object one more time
	secondIttObj, err := FromYAMLString(str).ToUnstructured()
	apiversionCheck, ok = secondIttObj.Object["apiVersion"].(string)
	if !ok || apiversionCheck != "v1" {
		t.Fatalf("manifest apiVersion key doesn't match original one:expected=v1, got=%v", apiversionCheck)
	}
}
