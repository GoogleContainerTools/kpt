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

package v1alpha2

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func (kf *KptFile) Validate() error {
	if err := kf.Pipeline.validate(); err != nil {
		return fmt.Errorf("invalid pipeline: %w", err)
	}
	// TODO: validate other fields
	return nil
}

// validate will validate all fields in the Pipeline
// 'mutators' and 'validators' share same schema and
// they are valid if all functions in them are ALL valid.
func (p *Pipeline) validate() error {
	if p == nil {
		return nil
	}
	for i := range p.Mutators {
		f := p.Mutators[i]
		err := f.validate("mutators", i)
		if err != nil {
			return fmt.Errorf("function %q: %w", f.Image, err)
		}
	}
	for i := range p.Validators {
		f := p.Validators[i]
		err := f.validate("validators", i)
		if err != nil {
			return fmt.Errorf("function %q: %w", f.Image, err)
		}
	}
	return nil
}

func (f *Function) validate(fnType string, idx int) error {
	err := validateFunctionName(f.Image)
	if err != nil {
		return &ValidateError{
			Field:  fmt.Sprintf("pipeline.%s[%d].image", fnType, idx),
			Value:  f.Image,
			Reason: err.Error(),
		}
	}

	if len(f.ConfigMap) != 0 && f.ConfigPath != "" {
		return &ValidateError{
			Field:  fmt.Sprintf("pipeline.%s[%d]", fnType, idx),
			Reason: "functionConfig must not specify both `configMap` and `configPath` at the same time",
		}
	}

	if f.ConfigPath != "" {
		if err := validateFnConfigPath(f.ConfigPath); err != nil {
			return &ValidateError{
				Field:  fmt.Sprintf("pipeline.%s[%d].configPath", fnType, idx),
				Value:  f.ConfigPath,
				Reason: err.Error(),
			}
		}
	}

	return nil
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
	nameRegexp := fmt.Sprintf(`(?:%s\/)?%s(?:\/%s)*`, domainRegexp,
		pathComponentRegexp, pathComponentRegexp)
	tagRegexp := `(?:[\w][\w.-]{0,127})`
	r := fmt.Sprintf(`^(?:%s(?:\:%s)?)$`, nameRegexp, tagRegexp)

	matched, err := regexp.MatchString(r, name)
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
func IsNodeZero(n *yaml.Node) bool {
	return n != nil && n.Kind == 0 && n.Style == 0 && n.Tag == "" && n.Value == "" &&
		n.Anchor == "" && n.Alias == nil && n.Content == nil &&
		n.HeadComment == "" && n.LineComment == "" && n.FootComment == "" &&
		n.Line == 0 && n.Column == 0
}

// validateFnConfigPath validates syntactic correctness of given functionConfig path
// and return an error if it's invalid.
func validateFnConfigPath(p string) error {
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("path must not be empty")
	}
	p = filepath.Clean(p)
	if filepath.IsAbs(p) {
		return fmt.Errorf("path must be relative")
	}
	if strings.Contains(p, "..") {
		// fn config must not live outside the package directory
		// Allowing outside path opens up an attack vector that allows
		// reading any YAML file on package consumer's machine.
		return fmt.Errorf("path must not be outside the package")
	}
	return nil
}

// ValidateError is the error returned when validation fails.
type ValidateError struct {
	// Field is the field that causes error
	Field string
	// Value is the value of invalid field
	Value string
	// Reason is the reson for the error
	Reason string
}

func (e *ValidateError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Kptfile is invalid:\nField: `%s`\n", e.Field))
	if e.Value != "" {
		sb.WriteString(fmt.Sprintf("Value: %q\n", e.Value))
	}
	sb.WriteString(fmt.Sprintf("Reason: %s\n", e.Reason))
	return sb.String()
}
