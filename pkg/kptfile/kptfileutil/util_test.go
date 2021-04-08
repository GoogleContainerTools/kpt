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

package kptfileutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	kptfilev1alpha2 "github.com/GoogleContainerTools/kpt/pkg/api/kptfile/v1alpha2"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// TestValidateInventory tests the ValidateInventory function.
func TestValidateInventory(t *testing.T) {
	// nil inventory should not validate
	isValid, err := ValidateInventory(nil)
	if isValid || err == nil {
		t.Errorf("nil inventory should not validate")
	}
	// Empty inventory should not validate
	inv := &kptfilev1alpha2.Inventory{}
	isValid, err = ValidateInventory(inv)
	if isValid || err == nil {
		t.Errorf("empty inventory should not validate")
	}
	// Empty inventory parameters strings should not validate
	inv = &kptfilev1alpha2.Inventory{
		Namespace:   "",
		Name:        "",
		InventoryID: "",
	}
	isValid, err = ValidateInventory(inv)
	if isValid || err == nil {
		t.Errorf("empty inventory parameters strings should not validate")
	}
	// Inventory with non-empty namespace, name, and id should validate.
	inv = &kptfilev1alpha2.Inventory{
		Namespace:   "test-namespace",
		Name:        "test-name",
		InventoryID: "test-id",
	}
	isValid, err = ValidateInventory(inv)
	if !isValid || err != nil {
		t.Errorf("inventory with non-empty namespace, name, and id should validate")
	}
}

// TestReadFile tests the ReadFile function.
func TestReadFile(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("%s-pkgfile-read", "test-kpt"))
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(dir, kptfilev1alpha2.KptFileName), []byte(`apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: cockroachdb
upstreamLock:
  type: git
  gitLock:
    commit: dd7adeb5492cca4c24169cecee023dbe632e5167
    directory: staging/cockroachdb
    ref: refs/heads/owners-update
    repo: https://github.com/kubernetes/examples
`), 0600)
	assert.NoError(t, err)

	f, err := ReadFile(dir)
	assert.NoError(t, err)
	assert.Equal(t, kptfilev1alpha2.KptFile{
		ResourceMeta: yaml.ResourceMeta{
			ObjectMeta: yaml.ObjectMeta{
				NameMeta: yaml.NameMeta{
					Name: "cockroachdb",
				},
			},
			TypeMeta: yaml.TypeMeta{
				APIVersion: kptfilev1alpha2.TypeMeta.APIVersion,
				Kind:       kptfilev1alpha2.TypeMeta.Kind},
		},
		UpstreamLock: &kptfilev1alpha2.UpstreamLock{
			Type: "git",
			GitLock: &kptfilev1alpha2.GitLock{
				Commit:    "dd7adeb5492cca4c24169cecee023dbe632e5167",
				Directory: "staging/cockroachdb",
				Ref:       "refs/heads/owners-update",
				Repo:      "https://github.com/kubernetes/examples",
			},
		},
	}, f)
}

// TestReadFile_failRead verifies an error is returned if the file cannot be read
func TestReadFile_failRead(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("%s-pkgfile-read", "test-kpt"))
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(dir, " KptFileError"), []byte(`apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: cockroachdb
upstream:
  type: git
  git:
    commit: dd7adeb5492cca4c24169cecee023dbe632e5167
    directory: staging/cockroachdb
    ref: refs/heads/owners-update
    repo: https://github.com/kubernetes/examples
`), 0600)
	assert.NoError(t, err)

	f, err := ReadFile(dir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
	assert.Equal(t, kptfilev1alpha2.KptFile{}, f)
}

// TestReadFile_failUnmarshal verifies an error is returned if the file contains any unrecognized fields.
func TestReadFile_failUnmarshal(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("%s-pkgfile-read", "test-kpt"))
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(dir, kptfilev1alpha2.KptFileName), []byte(`apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: cockroachdb
upstreamBadField:
  type: git
  git:
    commit: dd7adeb5492cca4c24169cecee023dbe632e5167
    directory: staging/cockroachdb
    ref: refs/heads/owners-update
    repo: https://github.com/kubernetes/examples
`), 0600)
	assert.NoError(t, err)

	f, err := ReadFile(dir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "upstreamBadField not found")
	assert.Equal(t, kptfilev1alpha2.KptFile{}, f)
}

func TestUpdateKptfile(t *testing.T) {
	writeKptfileToTemp := func(name string, content string) string {
		dir, err := ioutil.TempDir("", name)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		err = ioutil.WriteFile(filepath.Join(dir, kptfilev1alpha2.KptFileName), []byte(content), 0600)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		return dir
	}

	testCases := map[string]struct {
		origin         string
		updated        string
		local          string
		updateUpstream bool
		expected       string
	}{
		"no pipeline and no upstream info": {
			origin: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: base
`,
			updated: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: base
`,
			local: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
`,
			updateUpstream: false,
			expected: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
`,
		},

		"upstream information is not copied from upstream unless updateUpstream is true": {
			origin: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
upstream:
  type: git
  git:
    repo: github.com/GoogleContainerTools/kpt
    directory: /
    ref: v1
`,
			updated: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
upstream:
  type: git
  git:
    repo: github.com/GoogleContainerTools/kpt
    directory: /
    ref: v2
`,
			local: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
`,
			updateUpstream: false,
			expected: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
`,
		},

		"upstream information is copied from upstream when updateUpstream is true": {
			origin: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
upstream:
  type: git
  git:
    repo: github.com/GoogleContainerTools/kpt
    directory: /
    ref: v1
`,
			updated: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
upstream:
  type: git
  git:
    repo: github.com/GoogleContainerTools/kpt
    directory: /
    ref: v2
upstreamLock:
  type: git
  gitLock:
    repo: github.com/GoogleContainerTools/kpt
    directory: /
    ref: v2
    commit: abc123
`,
			local: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
`,
			updateUpstream: true,
			expected: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
upstream:
  type: git
  git:
    repo: github.com/GoogleContainerTools/kpt
    directory: /
    ref: v2
upstreamLock:
  type: git
  gitLock:
    repo: github.com/GoogleContainerTools/kpt
    directory: /
    ref: v2
    commit: abc123
`,
		},

		"pipeline in upstream replaces local": {
			origin: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: foo:bar
`,
			updated: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: foo:bar
  - image: some:image
`,
			local: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: my:image
  - image: foo:bar
`,
			updateUpstream: true,
			expected: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: foo:bar
  - image: some:image
`,
		},

		"pipeline in local remains if there are no changes in upstream": {
			origin: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: foo:bar
`,
			updated: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: foo:bar
`,
			local: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: my:image
  - image: foo:bar
`,
			updateUpstream: true,
			expected: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: my:image
  - image: foo:bar
`,
		},

		"pipeline remains if it is only added locally": {
			origin: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
`,
			updated: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
`,
			local: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: my:image
  - image: foo:bar
`,
			updateUpstream: true,
			expected: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: my:image
  - image: foo:bar
`,
		},

		"pipeline in local is emptied if it is gone from upstream": {
			origin: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: foo:bar
`,
			updated: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
`,
			local: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline:
  mutators:
  - image: my:image
  - image: foo:bar
`,
			updateUpstream: false,
			expected: `
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: foo
pipeline: {}
`,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			files := map[string]string{
				"origin":  tc.origin,
				"updated": tc.updated,
				"local":   tc.local,
			}
			dirs := make(map[string]string)
			for n, content := range files {
				dir := writeKptfileToTemp(n, content)
				dirs[n] = dir
			}
			defer func() {
				for _, p := range dirs {
					_ = os.RemoveAll(p)
				}
			}()

			err := UpdateKptfile(dirs["local"], dirs["updated"], dirs["origin"], tc.updateUpstream)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			c, err := ioutil.ReadFile(filepath.Join(dirs["local"], kptfilev1alpha2.KptFileName))
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			assert.Equal(t, strings.TrimSpace(tc.expected)+"\n", string(c))
		})
	}

}
