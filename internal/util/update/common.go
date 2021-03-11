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

package update

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/GoogleContainerTools/kpt/internal/util/pkgutil"
)

// findAllSubpackages traverses the provided package paths
// and finds all subpackages. A subpackage is a subdirectory underneath the
// root that has a Kptfile in it.
// The list is sorted in increasing order based on the depth of the subpackage
// relative to the root package.
func findAllSubpackages(pkgPaths ...string) ([]string, error) {
	uniquePaths := make(map[string]bool)
	uniquePaths["."] = true
	for _, path := range pkgPaths {
		paths, err := pkgutil.FindLocalSubpackages(path)
		if err != nil {
			return []string{}, err
		}
		for _, p := range paths {
			relPath, err := filepath.Rel(path, p)
			if err != nil {
				return []string{}, err
			}
			uniquePaths[relPath] = true
		}
	}
	var paths []string
	for p := range uniquePaths {
		paths = append(paths, p)
	}
	sort.Slice(paths, pkgutil.RootPkgFirstSorter(paths))
	return paths, nil
}

// exists returns true if a file or directory exists on the provided path,
// and false otherwise.
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	return !os.IsNotExist(err), nil
}
