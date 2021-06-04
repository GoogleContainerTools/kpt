// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/GoogleContainerTools/kpt/internal/util/openapi/augments"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"io/ioutil"
	"k8s.io/kubectl/pkg/cmd/util"
	"net/http"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi"
	"sigs.k8s.io/kustomize/kyaml/openapi/kustomizationapi"
)

const (
	SchemaSourceBuiltin = "builtin"
	SchemaSourceFile    = "file"
	SchemaSourceCluster = "cluster"

	BuiltinSchemaVersion = "v1204"
	KubernetesAssetName  = "kubernetesapi/v1204/swagger.json"
	KustomizeAssetName   = "kustomizationapi/swagger.json"
)

var SchemaSources = fmt.Sprintf("{%q, %q, %q}", SchemaSourceBuiltin, SchemaSourceCluster, SchemaSourceFile)

// ConfigureOpenAPI sets the openAPI schema in kyaml. It can either
// fetch the schema from a cluster, read it from file, or just the
// schema built into kyaml.
func ConfigureOpenAPI(factory util.Factory, k8sSchemaSource, k8sSchemaPath string) error {
	switch k8sSchemaSource {
	case SchemaSourceCluster:
		openAPISchema, err := FetchOpenAPISchemaFromCluster(factory)
		if err != nil {
			return fmt.Errorf("error fetching schema from cluster: %v", err)
		}
		return ConfigureOpenAPISchema(openAPISchema)
	case SchemaSourceFile:
		openAPISchema, err := ReadOpenAPISchemaFromDisk(k8sSchemaPath)
		if err != nil {
			return fmt.Errorf("error reading file at path %s: %v",
				k8sSchemaPath, err)
		}
		return ConfigureOpenAPISchema(openAPISchema)
	case SchemaSourceBuiltin:
		openAPISchema := kubernetesapi.OpenAPIMustAsset[BuiltinSchemaVersion](KubernetesAssetName)
		return ConfigureOpenAPISchema(openAPISchema)
	default:
		return fmt.Errorf("unknown schema source %s. Must be one of %s",
			k8sSchemaSource, SchemaSources)
	}
}

func FetchOpenAPISchemaFromCluster(f util.Factory) ([]byte, error) {
	restClient, err := f.RESTClient()
	if err != nil {
		return nil, err
	}
	data, err := restClient.Get().AbsPath("/openapi/v2").
		SetHeader("Accept", "application/json").Do(context.Background()).Raw()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ReadOpenAPISchemaFromDisk(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func ConfigureOpenAPISchema(openAPISchema []byte) error {
	openapi.SuppressBuiltInSchemaUse()
	openAPISchema, err := addExtensionsToBuiltinTypes(openAPISchema)
	if err != nil {
		return err
	}
	if err := openapi.AddSchema(openAPISchema); err != nil {
		return err
	}
	// Kustomize schema should always be added
	return openapi.AddSchema(kustomizationapi.MustAsset(KustomizeAssetName))
}

// GetJSONSchema returns the JSON OpenAPI schema being used in kyaml
func GetJSONSchema() ([]byte, error) {
	schema := openapi.Schema()
	if schema == nil {
		return nil, nil
	}
	output, err := openapi.Schema().MarshalJSON()
	if err != nil {
		return nil, err
	}
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal(output, &jsonSchema); err != nil {
		return nil, err
	}
	if output, err = json.MarshalIndent(jsonSchema, "", "  "); err != nil {
		return nil, err
	}
	return output, nil
}

func StartLocalServer() error {
	http.HandleFunc("/OpenAPI", func(w http.ResponseWriter, r *http.Request){
		schema, err := GetJSONSchema()
		if err != nil {
			fmt.Fprintf(w, "error getting schema: %w", err.Error())
		}
		fmt.Println("endpoint hit: /OpenAPI")
		fmt.Fprintf(w, string(schema))
	})

	var err error
	go func () {
		fmt.Println("starting server at port 8080\n")
		err = http.ListenAndServe(":8080", nil) // set listen port
	}()

	return err
}

func addExtensionsToBuiltinTypes(openAPISchema []byte) ([]byte, error) {
	patch, err := jsonpatch.DecodePatch([]byte(augments.JsonPatchBuiltin))
	if err != nil {
		return nil, err
	}
	modified, err := patch.Apply(openAPISchema)
	if err != nil {
		return nil, err
	}
	return modified, nil
}
