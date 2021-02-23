// Copyright 2019 Google LLC
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

package cmddesc_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleContainerTools/kpt/internal/cmddesc"
	"github.com/GoogleContainerTools/kpt/internal/testutil"
	kptfilev1alpha2 "github.com/GoogleContainerTools/kpt/pkg/api/kptfile/v1alpha2"
	"gotest.tools/assert"
)

// TestDesc_Execute tests happy path for Describe command.
func TestDesc_Execute(t *testing.T) {
	d, err := ioutil.TempDir("", "kptdesc")
	testutil.AssertNoError(t, err)

	defer func() {
		_ = os.RemoveAll(d)
	}()

	// write the KptFile
	err = ioutil.WriteFile(filepath.Join(d, kptfilev1alpha2.KptFileName), []byte(`
apiVersion: kpt.dev/v1alpha1
kind: Kptfile
metadata:
  name: cockroachdb_perf
upstreamLock:
  gitLock:
    commit: 9b6aeba0f9c2f8c44c712848b6f147f15ca3344f
    directory: cloud/kubernetes/performance
    ref: master
    repo: https://github.com/cockroachdb/cockroach
  type: git
`), 0600)
	testutil.AssertNoError(t, err)

	b := &bytes.Buffer{}
	cmd := cmddesc.NewRunner("kpt")
	cmd.Description.PrintBasePath = true
	cmd.Command.SetArgs([]string{d})
	cmd.Command.SetOut(b)
	err = cmd.Command.Execute()
	testutil.AssertNoError(t, err)

	exp := fmt.Sprintf(`    PACKAGE NAME           DIR                           REMOTE                            REMOTE PATH            REMOTE REF   REMOTE COMMIT  
  cockroachdb_perf   %s   https://github.com/cockroachdb/cockroach   cloud/kubernetes/performance   master       9b6aeba        
`, filepath.Base(d))
	assert.Equal(t, exp, b.String())
}

// TestCmd_defaultPkg tests describe command execution with no directory
// specified.
func TestCmd_defaultPkg(t *testing.T) {
	b := &bytes.Buffer{}
	cmd := cmddesc.NewRunner("kpt")
	cmd.Command.SetOut(b)
	err := cmd.Command.Execute()
	testutil.AssertNoError(t, err)

	exp := `  PACKAGE NAME   DIR   REMOTE   REMOTE PATH   REMOTE REF   REMOTE COMMIT  
`
	assert.Equal(t, exp, b.String())
}
