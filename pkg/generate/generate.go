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
import "text/template"
import "bytes"

const componentBuilder = `apiVersion: bundle.gke.io/v1alpha1
kind: ComponentBuilder
componentName: etcd-component
version: 1.0.0
objectFiles:
- url: /sample-deployment.yaml
- url: /sample-service.yaml
- url: /sample-patch.yaml
`
const sampleDeployment = `apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    app: {{.Name}}
  name: helloworld
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: {{.Name}}
      name: {{.Name}}
    spec:
      containers:
        - name: {{.Name}}
          image: gcr.io/google-samples/hello-app:2.0
`
const sampleService = `apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  labels: 
    app: {{.Name}}
spec:
  type: NodePort
  ports:
  - name: {{.Name}}
    port: 8080
    targetPort: 8080
  selector:
    app: {{.Name}}
`

const patchTemplate = `apiVersion: bundle.gke.io/v1alpha1
kind: PatchTemplate
template: |
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    namespace: {{.namespace}}
optionsSchema:
  required:
  - namespace
  properties:
    namespace:
      type: string
      pattern: '^[a-z0-9-]+$'
`
const patchOptions = `# Options for applying to the component via patch-templates
namespace: foo-namespace
buildLabel: test-build
`

// Create scaffolds basic set of files to the filesystem
func Create(filepath string, name string) error {
	var writeErr error
	if err := os.Mkdir(filepath, 0777); err != nil {
		writeErr = err
		return writeErr
	}

	replacement := struct{ Name string }{Name: name}

	deploymentTemplate, _ := template.New("deployment").Parse(sampleDeployment)
	var deploymentText bytes.Buffer
	deploymentTemplate.Execute(&deploymentText, replacement)

	serviceTemplate, _ := template.New("service").Parse(sampleService)
	var serviceText bytes.Buffer
	serviceTemplate.Execute(&serviceText, replacement)

	if err := ioutil.WriteFile(path.Join(filepath, "sample-component-builder.yaml"), []byte(componentBuilder), 0666); err != nil {
		writeErr = err
	}
	if err := ioutil.WriteFile(path.Join(filepath, "sample-deployment.yaml"), deploymentText.Bytes(), 0666); err != nil {
		writeErr = err
		goto handleError
	}
	if err := ioutil.WriteFile(path.Join(filepath, "sample-service.yaml"), serviceText.Bytes(), 0666); err != nil {
		writeErr = err
		goto handleError
	}
	if err := ioutil.WriteFile(path.Join(filepath, "sample-patch.yaml"), []byte(patchTemplate), 0666); err != nil {
		writeErr = err
		goto handleError
	}
	if err := ioutil.WriteFile(path.Join(filepath, "sample-patch.yaml"), []byte(patchTemplate), 0666); err != nil {
		writeErr = err
		goto handleError
	}
	if err := ioutil.WriteFile(path.Join(filepath, "sample-options.yaml"), []byte(patchOptions), 0666); err != nil {
		writeErr = err
		goto handleError
	}

handleError:
	if writeErr != nil {
		os.Remove(filepath)
	}
	return writeErr
}
