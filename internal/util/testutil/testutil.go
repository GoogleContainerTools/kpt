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

package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
)

func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	if !assert.NoError(t, err, msgAndArgs) {
		if err, ok := err.(*errors.Error); ok {
			fmt.Fprintf(os.Stderr, "%s", err.Stack())
		}
		t.FailNow()
	}
}

func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	if !assert.Equal(t, expected, actual, msgAndArgs) {
		t.FailNow()
	}
}

func PrintPackage(paths ...string) error {
	path := filepath.Join(paths...)
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(path, "/.git") {
			return nil
		}
		fmt.Println(path)
		return nil
	})
}

func PrintFile(paths ...string) error {
	path := filepath.Join(paths...)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
