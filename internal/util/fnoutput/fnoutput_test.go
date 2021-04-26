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

package fnoutput

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFnFailureString(t *testing.T) {
	testcases := []struct {
		name      string
		fnFailure FnFailure
		truncate  bool
		expected  string
	}{
		{
			name:      "no truncate - empty stderr",
			fnFailure: FnFailure{},
			expected: `Stderr:
  
Exit Code: 0
`,
		},
		{
			name: "no truncate - normal failure",
			fnFailure: FnFailure{
				Stderr: `error message1
error message2`,
				ExitCode: 1,
			},
			expected: `Stderr:
  error message1
  error message2
Exit Code: 1
`,
		},
		{
			name: "no truncate - long stderr",
			fnFailure: FnFailure{
				Stderr: `error message
error message
error message
error message
error message`,
				ExitCode: 1,
			},
			expected: `Stderr:
  error message
  error message
  error message
  error message
  error message
Exit Code: 1
`,
		},
		{
			name: "truncate - normal failure",
			fnFailure: FnFailure{
				Stderr: `error message
error message
error message
error message`,
				ExitCode: 1,
			},
			truncate: true,
			expected: `Stderr:
  error message
  error message
  error message
  error message
Exit Code: 1
`,
		},
		{
			name: "truncate - long stderr 1",
			fnFailure: FnFailure{
				Stderr: `error message
error message
error message
error message
error message`,
				ExitCode: 1,
			},
			truncate: true,
			expected: `Stderr:
  error message
  error message
  error message
  error message
  ...(1 line(s) truncated)
Exit Code: 1
`,
		},
		{
			name: "truncate - long stderr 2",
			fnFailure: FnFailure{
				Stderr: `error message
error message
error message
error message
error message
error message
error message
error message`,
				ExitCode: 1,
			},
			truncate: true,
			expected: `Stderr:
  error message
  error message
  error message
  error message
  ...(4 line(s) truncated)
Exit Code: 1
`,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s, err := tc.fnFailure.String(tc.truncate)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.expected, s)
		})
	}
}
