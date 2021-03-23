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

package testing

import (
	"testing"

	"github.com/GoogleContainerTools/kpt/internal/pkg"
	"github.com/stretchr/testify/assert"
)

// CreatePkgOrFail creates a new package from the provided path. Unlike the
// pkg.New function, it fails the test instead of returning an error.
func CreatePkgOrFail(t *testing.T, path string) *pkg.Pkg {
	p, err := pkg.New(path)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return p
}
