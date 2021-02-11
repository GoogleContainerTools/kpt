// Copyright 2021 Google LLC
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

// Package pipeline provides struct definitions for Pipeline and utility
// methods to read and write a pipeline resource.
package pipeline

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/GoogleContainerTools/kpt/internal/pipeline/runtime"
	"github.com/GoogleContainerTools/kpt/pkg/api/kptfile/v1alpha2"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// validateFunction will validate all fields in function.
func validateFunction(f *v1alpha2.Function) error {
	err := validateFunctionName(f.Image)
	if err != nil {
		return fmt.Errorf("'image' is invalid: %w", err)
	}

	var configFields []string
	if f.ConfigPath != "" {
		if err := validatePath(f.ConfigPath); err != nil {
			return fmt.Errorf("'configPath' %q is invalid: %w", f.ConfigPath, err)
		}
		configFields = append(configFields, "configPath")
	}
	if len(f.ConfigMap) != 0 {
		configFields = append(configFields, "configMap")
	}
	if !isNodeZero(&f.Config) {
		configFields = append(configFields, "config")
	}
	if len(configFields) > 1 {
		return fmt.Errorf("following fields are mutually exclusive: 'config', 'configMap', 'configPath'. Got %q",
			strings.Join(configFields, ", "))
	}

	return nil
}

// newFnRunner returns a fnRunner from the image and configs of
// this function.
func newFnRunner(f *v1alpha2.Function, pkgPath string) (kio.Filter, error) {
	config, err := newFnConfig(f, pkgPath)
	if err != nil {
		return nil, err
	}
	return &fnRunner{
		fn: &runtime.ContainerFn{
			Image: f.Image,
		},
		fnConfig: config,
	}, nil
}

func newFnConfig(f *v1alpha2.Function, pkgPath string) (*yaml.RNode, error) {
	var node *yaml.RNode
	switch {
	case f.ConfigPath != "":
		path := path.Join(pkgPath, f.ConfigPath)
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file path %s: %w", f.ConfigPath, err)
		}
		b, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read content from config file: %w", err)
		}
		node, err = yaml.Parse(string(b))
		if err != nil {
			return nil, fmt.Errorf("failed to parse config %s: %w", string(b), err)
		}
		// directly use the config from file
		return node, nil
	case !isNodeZero(&f.Config):
		// directly use the inline config
		return yaml.NewRNode(&f.Config), nil
	case len(f.ConfigMap) != 0:
		node = yaml.NewMapRNode(&f.ConfigMap)
		if node == nil {
			return nil, nil
		}
		// create a ConfigMap only for configMap config
		configNode, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: function-input
data: {}
`)
		if err != nil {
			return nil, fmt.Errorf("failed to parse function config skeleton: %w", err)
		}
		err = configNode.PipeE(yaml.SetField("data", node))
		if err != nil {
			return nil, fmt.Errorf("failed to set 'data' field: %w", err)
		}
		return configNode, nil
	}
	// no need to reutrn ConfigMap if no config given
	return nil, nil
}

// validateFunctionName validates the function name.
// According to Docker implementation
// https://github.com/docker/distribution/blob/master/reference/reference.go. A valid
// name definition is:
//	name                            := [domain '/'] path-component ['/' path-component]*
//	domain                          := domain-component ['.' domain-component]* [':' port-number]
//	domain-component                := /([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])/
//	port-number                     := /[0-9]+/
//	path-component                  := alpha-numeric [separator alpha-numeric]*
// 	alpha-numeric                   := /[a-z0-9]+/
//	separator                       := /[_.]|__|[-]*/
func validateFunctionName(name string) error {
	pathComponentRegexp := `(?:[a-z0-9](?:(?:[_.]|__|[-]*)[a-z0-9]+)*)`
	domainComponentRegexp := `(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])`
	domainRegexp := fmt.Sprintf(`%s(?:\.%s)*(?:\:[0-9]+)?`, domainComponentRegexp, domainComponentRegexp)
	nameRegexp := fmt.Sprintf(`^(?:%s\/)?%s(?:\/%s)*$`, domainRegexp,
		pathComponentRegexp, pathComponentRegexp)
	matched, err := regexp.MatchString(nameRegexp, name)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("function name %q is invalid", name)
	}
	return nil
}

// IsNodeZero returns true if all the public fields in the Node are empty.
// Which means it's not initialized and should be omitted when marshal.
// The Node itself has a method IsZero but it is not released
// in yaml.v3. https://pkg.go.dev/gopkg.in/yaml.v3#Node.IsZero
// TODO: Use `IsYNodeZero` method from kyaml when kyaml has been updated to
// >= 0.10.5
func isNodeZero(n *yaml.Node) bool {
	return n != nil && n.Kind == 0 && n.Style == 0 && n.Tag == "" && n.Value == "" &&
		n.Anchor == "" && n.Alias == nil && n.Content == nil &&
		n.HeadComment == "" && n.LineComment == "" && n.FootComment == "" &&
		n.Line == 0 && n.Column == 0
}

// validatePath validates input path and return an error if it's invalid
func validatePath(p string) error {
	if path.IsAbs(p) {
		return fmt.Errorf("path is not relative")
	}
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("path cannot have only white spaces")
	}
	if p != sourceAllSubPkgs && strings.Contains(p, "*") {
		return fmt.Errorf("path contains asterisk, asterisk is only allowed in './*'")
	}
	// backslash (\\), alert bell (\a), backspace (\b), form feed (\f), vertical tab(\v) are
	// unlikely to be in a valid path
	for _, c := range "\\\a\b\f\v" {
		if strings.Contains(p, string(c)) {
			return fmt.Errorf("path cannot have character %q", c)
		}
	}
	return nil
}
