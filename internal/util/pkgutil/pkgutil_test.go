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

package pkgutil

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/GoogleContainerTools/kpt/internal/testutil/pkgbuilder"
	kptfilev1alpha2 "github.com/GoogleContainerTools/kpt/pkg/api/kptfile/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestWalkPackage(t *testing.T) {
	testCases := map[string]struct {
		pkg      *pkgbuilder.RootPkg
		expected []string
	}{
		"walks subdirectories of a package": {
			pkg: pkgbuilder.NewRootPkg().
				WithFile("abc.yaml", "42").
				WithFile("test.txt", "Hello, World!").
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithFile("def.yaml", "123"),
				),
			expected: []string{
				".",
				"abc.yaml",
				"foo",
				"foo/def.yaml",
				"test.txt",
			},
		},
		"ignores .git folder": {
			pkg: pkgbuilder.NewRootPkg().
				WithFile("abc.yaml", "42").
				WithSubPackages(
					pkgbuilder.NewSubPkg(".git").
						WithFile("INDEX", "ABC123"),
				),
			expected: []string{
				".",
				"abc.yaml",
			},
		},
		"ignores subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithFile("abc.yaml", "42").
				WithFile("test.txt", "Hello, World!").
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile().
						WithFile("def.yaml", "123"),
				),
			expected: []string{
				".",
				"abc.yaml",
				"test.txt",
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			pkgPath := pkgbuilder.ExpandPkg(t, tc.pkg, map[string]string{})

			var visited []string
			if err := WalkPackage(pkgPath, func(s string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(pkgPath, s)
				if err != nil {
					return err
				}
				visited = append(visited, relPath)
				return nil
			}); !assert.NoError(t, err) {
				t.FailNow()
			}

			sort.Strings(visited)

			assert.Equal(t, tc.expected, visited)
		})
	}
}

func TestFindAllDirectSubpackages(t *testing.T) {
	testCases := map[string]struct {
		pkg      *pkgbuilder.RootPkg
		expected []string
	}{
		"includes remote subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile(
							pkgbuilder.NewKptfile().
								WithUpstream("github.com/GoogleContainerTools/kpt",
									"/", "main", string(kptfilev1alpha2.ResourceMerge)),
						).
						WithResource(pkgbuilder.ConfigMapResource),
				),
			expected: []string{
				"foo",
			},
		},
		"includes local subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile().
						WithResource(pkgbuilder.ConfigMapResource),
				),
			expected: []string{
				"foo",
			},
		},
		"does not include root package": {
			pkg: pkgbuilder.NewRootPkg().
				WithKptfile().
				WithResource(pkgbuilder.DeploymentResource),
			expected: []string{},
		},
		"does not include nested remote subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile(
							pkgbuilder.NewKptfile().
								WithUpstream("github.com/GoogleContainerTools/kpt",
									"/", "main", string(kptfilev1alpha2.ResourceMerge)),
						).
						WithResource(pkgbuilder.ConfigMapResource).
						WithSubPackages(
							pkgbuilder.NewSubPkg("bar").
								WithSubPackages(
									pkgbuilder.NewSubPkg("zork").
										WithKptfile(
											pkgbuilder.NewKptfile().
												WithUpstream("github.com/GoogleContainerTools/kpt",
													"/", "main", string(kptfilev1alpha2.ResourceMerge)),
										).
										WithResource(pkgbuilder.ConfigMapResource),
								),
						),
				),
			expected: []string{
				"foo",
			},
		},
		"does not include nested local subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile().
						WithResource(pkgbuilder.ConfigMapResource).
						WithSubPackages(
							pkgbuilder.NewSubPkg("zork").
								WithKptfile().
								WithResource(pkgbuilder.ConfigMapResource),
						),
					pkgbuilder.NewSubPkg("subpkg").
						WithKptfile(),
				),
			expected: []string{
				"foo",
				"subpkg",
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			pkgPath := pkgbuilder.ExpandPkg(t, tc.pkg, map[string]string{})

			paths, err := FindAllDirectSubpackages(pkgPath)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			relPaths := toRelativePaths(t, paths, pkgPath)
			sort.Strings(relPaths)

			assert.Equal(t, tc.expected, relPaths)
		})
	}
}

func TestFindLocalRecursiveSubpackages(t *testing.T) {
	testCases := map[string]struct {
		pkg      *pkgbuilder.RootPkg
		expected []string
	}{
		"does not include remote subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile(
							pkgbuilder.NewKptfile().
								WithUpstream("github.com/GoogleContainerTools/kpt",
									"/", "main", string(kptfilev1alpha2.ResourceMerge)),
						).
						WithResource(pkgbuilder.ConfigMapResource),
				),
			expected: []string{},
		},
		"includes local subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile().
						WithResource(pkgbuilder.ConfigMapResource),
				),
			expected: []string{
				"foo",
			},
		},
		"does not include root package": {
			pkg: pkgbuilder.NewRootPkg().
				WithKptfile().
				WithResource(pkgbuilder.DeploymentResource),
			expected: []string{},
		},
		"does not include nested remote subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile(
							pkgbuilder.NewKptfile().
								WithUpstream("github.com/GoogleContainerTools/kpt",
									"/", "main", string(kptfilev1alpha2.ResourceMerge)),
						).
						WithResource(pkgbuilder.ConfigMapResource).
						WithSubPackages(
							pkgbuilder.NewSubPkg("bar").
								WithSubPackages(
									pkgbuilder.NewSubPkg("zork").
										WithKptfile(
											pkgbuilder.NewKptfile().
												WithUpstream("github.com/GoogleContainerTools/kpt",
													"/", "main", string(kptfilev1alpha2.ResourceMerge)),
										).
										WithResource(pkgbuilder.ConfigMapResource),
								),
						),
				),
			expected: []string{},
		},
		"includes nested local subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile().
						WithResource(pkgbuilder.ConfigMapResource).
						WithSubPackages(
							pkgbuilder.NewSubPkg("zork").
								WithKptfile().
								WithResource(pkgbuilder.ConfigMapResource),
						),
					pkgbuilder.NewSubPkg("subpkg").
						WithKptfile(),
				),
			expected: []string{
				"foo",
				"foo/zork",
				"subpkg",
			},
		},
		"does not include local subpackages within remote subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile(
							pkgbuilder.NewKptfile().
								WithUpstream("github.com/GoogleContainerTools/kpt",
									"/", "main", string(kptfilev1alpha2.ResourceMerge)),
						).
						WithResource(pkgbuilder.ConfigMapResource).
						WithSubPackages(
							pkgbuilder.NewSubPkg("zork").
								WithKptfile().
								WithResource(pkgbuilder.ConfigMapResource),
						),
				),
			expected: []string{},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			pkgPath := pkgbuilder.ExpandPkg(t, tc.pkg, map[string]string{})

			paths, err := FindLocalRecursiveSubpackages(pkgPath)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			relPaths := toRelativePaths(t, paths, pkgPath)
			sort.Strings(relPaths)

			assert.Equal(t, tc.expected, relPaths)
		})
	}
}

func TestFindRemoteDirectSubpackages(t *testing.T) {
	testCases := map[string]struct {
		pkg      *pkgbuilder.RootPkg
		expected []string
	}{
		"includes remote subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile(
							pkgbuilder.NewKptfile().
								WithUpstream("github.com/GoogleContainerTools/kpt",
									"/", "main", string(kptfilev1alpha2.ResourceMerge)),
						).
						WithResource(pkgbuilder.ConfigMapResource),
				),
			expected: []string{
				"foo",
			},
		},
		"does not include local subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile().
						WithResource(pkgbuilder.ConfigMapResource),
				),
			expected: []string{},
		},
		"does not include root package": {
			pkg: pkgbuilder.NewRootPkg().
				WithKptfile().
				WithResource(pkgbuilder.DeploymentResource),
			expected: []string{},
		},
		"does not include nested remote subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile(
							pkgbuilder.NewKptfile().
								WithUpstream("github.com/GoogleContainerTools/kpt",
									"/", "main", string(kptfilev1alpha2.ResourceMerge)),
						).
						WithResource(pkgbuilder.ConfigMapResource).
						WithSubPackages(
							pkgbuilder.NewSubPkg("bar").
								WithSubPackages(
									pkgbuilder.NewSubPkg("zork").
										WithKptfile(
											pkgbuilder.NewKptfile().
												WithUpstream("github.com/GoogleContainerTools/kpt",
													"/", "main", string(kptfilev1alpha2.ResourceMerge)),
										).
										WithResource(pkgbuilder.ConfigMapResource),
								),
						),
				),
			expected: []string{
				"foo",
			},
		},
		"does not include nested local subpackages": {
			pkg: pkgbuilder.NewRootPkg().
				WithResource(pkgbuilder.DeploymentResource).
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithKptfile().
						WithResource(pkgbuilder.ConfigMapResource).
						WithSubPackages(
							pkgbuilder.NewSubPkg("zork").
								WithKptfile().
								WithResource(pkgbuilder.ConfigMapResource),
						),
					pkgbuilder.NewSubPkg("subpkg").
						WithKptfile(),
				),
			expected: []string{},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			pkgPath := pkgbuilder.ExpandPkg(t, tc.pkg, map[string]string{})

			paths, err := FindRemoteDirectSubpackages(pkgPath)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			relPaths := toRelativePaths(t, paths, pkgPath)
			sort.Strings(relPaths)

			assert.Equal(t, tc.expected, relPaths)
		})
	}
}

func TestFindLocalRecursiveSubpackagesForPaths(t *testing.T) {
	testCases := map[string]struct {
		pkgs     []*pkgbuilder.RootPkg
		expected []string
	}{
		"does not include remote subpackages": {
			pkgs: []*pkgbuilder.RootPkg{
				pkgbuilder.NewRootPkg().
					WithResource(pkgbuilder.DeploymentResource).
					WithSubPackages(
						pkgbuilder.NewSubPkg("foo").
							WithKptfile(
								pkgbuilder.NewKptfile().
									WithUpstream("github.com/GoogleContainerTools/kpt",
										"/", "main", string(kptfilev1alpha2.ResourceMerge)),
							).
							WithResource(pkgbuilder.ConfigMapResource),
					),
			},
			expected: []string{
				".",
			},
		},
		"includes local subpackages": {
			pkgs: []*pkgbuilder.RootPkg{
				pkgbuilder.NewRootPkg().
					WithResource(pkgbuilder.DeploymentResource).
					WithSubPackages(
						pkgbuilder.NewSubPkg("foo").
							WithKptfile().
							WithResource(pkgbuilder.ConfigMapResource),
					),
			},
			expected: []string{
				".",
				"foo",
			},
		},
		"includes root package": {
			pkgs: []*pkgbuilder.RootPkg{
				pkgbuilder.NewRootPkg().
					WithKptfile().
					WithResource(pkgbuilder.DeploymentResource),
			},
			expected: []string{
				".",
			},
		},
		"does not include nested remote subpackages": {
			pkgs: []*pkgbuilder.RootPkg{
				pkgbuilder.NewRootPkg().
					WithResource(pkgbuilder.DeploymentResource).
					WithSubPackages(
						pkgbuilder.NewSubPkg("foo").
							WithKptfile(
								pkgbuilder.NewKptfile().
									WithUpstream("github.com/GoogleContainerTools/kpt",
										"/", "main", string(kptfilev1alpha2.ResourceMerge)),
							).
							WithResource(pkgbuilder.ConfigMapResource).
							WithSubPackages(
								pkgbuilder.NewSubPkg("bar").
									WithSubPackages(
										pkgbuilder.NewSubPkg("zork").
											WithKptfile(
												pkgbuilder.NewKptfile().
													WithUpstream("github.com/GoogleContainerTools/kpt",
														"/", "main", string(kptfilev1alpha2.ResourceMerge)),
											).
											WithResource(pkgbuilder.ConfigMapResource),
									),
							),
					),
			},
			expected: []string{
				".",
			},
		},
		"includes nested local subpackages": {
			pkgs: []*pkgbuilder.RootPkg{
				pkgbuilder.NewRootPkg().
					WithResource(pkgbuilder.DeploymentResource).
					WithSubPackages(
						pkgbuilder.NewSubPkg("foo").
							WithKptfile().
							WithResource(pkgbuilder.ConfigMapResource).
							WithSubPackages(
								pkgbuilder.NewSubPkg("zork").
									WithKptfile().
									WithResource(pkgbuilder.ConfigMapResource),
							),
						pkgbuilder.NewSubPkg("subpkg").
							WithKptfile(),
					),
			},
			expected: []string{
				".",
				"foo",
				"foo/zork",
				"subpkg",
			},
		},
		"multiple packages": {
			pkgs: []*pkgbuilder.RootPkg{
				pkgbuilder.NewRootPkg().
					WithResource(pkgbuilder.DeploymentResource).
					WithSubPackages(
						pkgbuilder.NewSubPkg("foo").
							WithKptfile().
							WithResource(pkgbuilder.ConfigMapResource).
							WithSubPackages(
								pkgbuilder.NewSubPkg("zork").
									WithKptfile().
									WithResource(pkgbuilder.ConfigMapResource),
							),
						pkgbuilder.NewSubPkg("subpkg").
							WithKptfile(),
					),
				pkgbuilder.NewRootPkg().
					WithKptfile().
					WithSubPackages(
						pkgbuilder.NewSubPkg("foo").
							WithKptfile(),
					),
				pkgbuilder.NewRootPkg().
					WithKptfile().
					WithSubPackages(
						pkgbuilder.NewSubPkg("bar").
							WithKptfile(),
						pkgbuilder.NewSubPkg("remotebar").
							WithKptfile(
								pkgbuilder.NewKptfile().
									WithUpstream("github.com/GoogleContainerTools/kpt",
										"/", "main", string(kptfilev1alpha2.ResourceMerge)),
							),
					),
			},
			expected: []string{
				".",
				"bar",
				"foo",
				"foo/zork",
				"subpkg",
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			var pkgPaths []string
			for _, p := range tc.pkgs {
				pkgPaths = append(pkgPaths, pkgbuilder.ExpandPkg(t, p, map[string]string{}))
			}

			paths, err := FindLocalRecursiveSubpackagesForPaths(pkgPaths...)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			sort.Strings(paths)

			assert.Equal(t, tc.expected, paths)
		})
	}
}

func TestCopyPackage(t *testing.T) {

}

func toRelativePaths(t *testing.T, paths []string, base string) []string {
	relPaths := []string{}
	for _, p := range paths {
		r, err := filepath.Rel(base, p)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		relPaths = append(relPaths, r)
	}
	return relPaths
}
