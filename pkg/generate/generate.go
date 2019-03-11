// Copyright 2019 Google LLC
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

// Package generate scaffolds components to guide a user for component creation
package generate

import "io/ioutil"
import "os"
import "path"
import "strings"

const componentBuilder = `apiVersion: bundle.gke.io/v1alpha1
kind: ComponentBuilder
componentName: etcd-component
version: 1.0.0
objectFiles:
- url: /sample-deployment.yaml
- url: /sample-service.yaml
`
const sampleDeployment = `apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    app: $DEPLOYMENT_NAME
  name: helloworld
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: $DEPLOYMENT_NAME
      name: $DEPLOYMENT_NAME
    spec:
      containers:
        - name: $DEPLOYMENT_NAME
          image: gcr.io/google-samples/hello-app:2.0
`
const sampleService = `apiVersion: v1
kind: Service
metadata:
  name: $SERVICE_NAME
  labels: 
    app: $SERVICE_NAME
spec:
  type: NodePort
  ports:
  - name: $SERVICE_NAME
    port: 8080
    targetPort: 8080
  selector:
    app: $SERVICE_NAME
`

// Create scaffolds basic set of files to the filesystem
func Create(filepath string, name string) error {
	var writeErr error
  if err := os.Mkdir(filepath, 0777); err != nil { 
    writeErr = err 
    return writeErr
  }

	if err := ioutil.WriteFile(path.Join(filepath, "etcd-component-builder.yaml"), []byte(componentBuilder), 0666); err != nil {
    writeErr = err
  }
	if err := ioutil.WriteFile(path.Join(filepath, "sample-deployment.yaml"), []byte(strings.Replace(sampleDeployment, "$DEPLOYMENT_NAME", name, -1)), 0666); err != nil {
    writeErr = err
  }
  if err := ioutil.WriteFile(path.Join(filepath, "sample-service.yaml"), []byte(strings.Replace(sampleService, "$SERVICE_NAME", name, -1)), 0666); err != nil {
    writeErr = err
  }

  if writeErr != nil {
    os.Remove(filepath)
  }
	return writeErr
}
