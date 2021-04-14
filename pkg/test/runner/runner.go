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

package runner

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Runner runs an e2e test
type Runner struct {
	pkgName       string
	testCase      TestCase
	cmd           string
	t             *testing.T
	initialCommit string
	kptBin        string
}

func getKptBin() (string, error) {
	if p := os.Getenv(kptBinEnv); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("must specify env '%s' for kpt binary path", kptBinEnv)
}

const (
	// If this env is set to "true", this e2e test framework will update the
	// expected diff and results if they already exist. If will not change
	// config.yaml.
	updateExpectedEnv string = "KPT_E2E_UPDATE_EXPECTED"
	kptBinEnv         string = "KPT_E2E_BIN"

	expectedDir         string = ".expected"
	expectedResultsFile string = "results.yaml"
	expectedDiffFile    string = "diff.patch"
	expectedConfigFile  string = "config.yaml"
	CommandFnEval       string = "eval"
	CommandFnRender     string = "render"
)

// NewRunner returns a new runner for pkg
func NewRunner(t *testing.T, testCase TestCase, c string) (*Runner, error) {
	info, err := os.Stat(testCase.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot open path %s: %w", testCase.Path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", testCase.Path)
	}
	kptBin, err := getKptBin()
	if err != nil {
		return nil, fmt.Errorf("failed to find kpt binary: %w", err)
	}
	t.Logf("Using kpt binary: %s", kptBin)
	return &Runner{
		pkgName:  filepath.Base(testCase.Path),
		testCase: testCase,
		cmd:      c,
		t:        t,
		kptBin:   kptBin,
	}, nil
}

// Run runs the test.
func (r *Runner) Run() error {
	switch r.cmd {
	case CommandFnEval:
		return r.runFnEval()
	case CommandFnRender:
		return r.runFnRender()
	default:
		return fmt.Errorf("invalid command %s", r.cmd)
	}
}

func (r *Runner) runFnEval() error {
	r.t.Logf("Running test against package %s\n", r.pkgName)
	tmpDir, err := ioutil.TempDir("", "kpt-fn-e2e-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	pkgPath := filepath.Join(tmpDir, r.pkgName)
	// create result dir
	resultsPath := filepath.Join(tmpDir, "results")
	err = os.Mkdir(resultsPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create results dir %s: %w", resultsPath, err)
	}

	// copy package to temp directory
	err = copyDir(r.testCase.Path, pkgPath)
	if err != nil {
		return fmt.Errorf("failed to copy package: %w", err)
	}

	// init and commit package files
	err = r.preparePackage(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to prepare package: %w", err)
	}

	// run function
	kptArgs := []string{"fn", "eval", pkgPath, "--results-dir", resultsPath}
	if r.testCase.Config.EvalConfig.Network {
		kptArgs = append(kptArgs, "--network")
	}
	if r.testCase.Config.EvalConfig.Image != "" {
		kptArgs = append(kptArgs, "--image", r.testCase.Config.EvalConfig.Image)
	} else if r.testCase.Config.EvalConfig.ExecPath != "" {
		kptArgs = append(kptArgs, "--exec-path", r.testCase.Config.EvalConfig.ExecPath)
	}
	if r.testCase.Config.EvalConfig.FnConfig != "" {
		fnConfigPath, err := filepath.Abs(filepath.Join(r.testCase.Path, r.testCase.Config.EvalConfig.FnConfig))
		if err != nil {
			return fmt.Errorf("failed to get absolute path to function config file: %w", err)
		}
		kptArgs = append(kptArgs, "--fn-config", fnConfigPath)
	}
	if r.testCase.Config.EvalConfig.IncludeMetaResources {
		kptArgs = append(kptArgs, "--include-meta-resources")
	}
	// args must be appended last
	if len(r.testCase.Config.EvalConfig.Args) > 0 {
		kptArgs = append(kptArgs, "--")
		for k, v := range r.testCase.Config.EvalConfig.Args {
			kptArgs = append(kptArgs, fmt.Sprintf("%s=%s", k, v))
		}
	}
	for i := 0; i < r.testCase.Config.RunCount(); i++ {
		output, fnErr := runCommand("", r.kptBin, kptArgs)
		// Update the diff file or results file if updateExpectedEnv is set.
		if strings.ToLower(os.Getenv(updateExpectedEnv)) == "true" {
			return r.updateExpected(pkgPath, resultsPath, filepath.Join(r.testCase.Path, expectedDir))
		}

		// compare results
		err = r.compareResult(fnErr, pkgPath, resultsPath)
		if err != nil {
			return fmt.Errorf("%w\nkpt output:\n%s", err, output)
		}
	}

	return nil
}

func (r *Runner) runFnRender() error {
	r.t.Logf("Running test against package %s\n", r.pkgName)
	tmpDir, err := ioutil.TempDir("", "kpt-pipeline-e2e-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary dir: %w", err)
	}
	if r.testCase.Config.Debug {
		fmt.Printf("Running test against package %s in dir %s \n", r.pkgName, tmpDir)
	}
	if !r.testCase.Config.Debug {
		// if debug is true, keep the test directory around for debugging
		defer os.RemoveAll(tmpDir)
	}
	pkgPath := filepath.Join(tmpDir, r.pkgName)
	// create dir to store untouched pkg to compare against
	origPkgPath := filepath.Join(tmpDir, "original")
	err = os.Mkdir(origPkgPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create original dir %s: %w", origPkgPath, err)
	}

	// copy package to temp directory
	err = copyDir(r.testCase.Path, pkgPath)
	if err != nil {
		return fmt.Errorf("failed to copy package: %w", err)
	}
	err = copyDir(r.testCase.Path, origPkgPath)
	if err != nil {
		return fmt.Errorf("failed to copy package: %w", err)
	}

	// init and commit package files
	err = r.preparePackage(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to prepare package: %w", err)
	}

	// run function
	kptArgs := []string{"fn", "render", pkgPath}
	for i := 0; i < r.testCase.Config.RunCount(); i++ {
		output, fnErr := runCommand("", r.kptBin, kptArgs)
		// Update the diff file or results file if updateExpectedEnv is set.
		if strings.ToLower(os.Getenv(updateExpectedEnv)) == "true" {
			// TODO: `fn render` doesn't support result file now
			// use empty string to skip update results
			return r.updateExpected(pkgPath, "", filepath.Join(r.testCase.Path, expectedDir))
		}
		if fnErr != nil {
			r.t.Logf("kpt error output: %s", output)
		}
		// compare results
		// TODO: `fn render` doesn't support result file now
		// use empty string for results dir
		err = r.compareResult(fnErr, pkgPath, "")
	}
	return err
}

func (r *Runner) preparePackage(pkgPath string) error {
	err := gitInit(pkgPath)
	if err != nil {
		return err
	}

	err = gitAddAll(pkgPath)
	if err != nil {
		return err
	}

	err = gitCommit(pkgPath, "first")
	if err != nil {
		return err
	}

	r.initialCommit, err = getCommitHash(pkgPath)
	return err
}

func (r *Runner) compareResult(exitErr error, tmpPkgPath, resultsPath string) error {
	expected, err := newExpected(tmpPkgPath)
	if err != nil {
		return err
	}
	// get exit code
	exitCode := 0
	if e, ok := exitErr.(*exec.ExitError); ok {
		exitCode = e.ExitCode()
	} else if exitErr != nil {
		return fmt.Errorf("cannot get exit code, received error '%w'", exitErr)
	}

	if exitCode != r.testCase.Config.ExitCode {
		return fmt.Errorf("actual exit code %d doesn't match expected %d", exitCode, r.testCase.Config.ExitCode)
	}

	// compare results
	actual, err := readActualResults(resultsPath)
	if err != nil {
		return fmt.Errorf("failed to read actual results: %w", err)
	}
	diffOfResult, err := diffStrings(actual, expected.Results)
	if err != nil {
		return fmt.Errorf("error when run diff of results: %w: %s", err, diffOfResult)
	}
	if actual != expected.Results {
		return fmt.Errorf("actual results doesn't match expected\nActual\n===\n%s\nDiff of Results\n===\n%s",
			actual, diffOfResult)
	}

	// compare diff
	actual, err = readActualDiff(tmpPkgPath, r.initialCommit)
	if err != nil {
		return fmt.Errorf("failed to read actual diff: %w", err)
	}
	if actual != expected.Diff {
		diffOfDiff, err := diffStrings(actual, expected.Diff)
		if err != nil {
			return fmt.Errorf("error when run diff of diff: %w: %s", err, diffOfDiff)
		}
		return fmt.Errorf("actual diff doesn't match expected\nActual\n===\n%s\nDiff of Diff\n===\n%s",
			actual, diffOfDiff)
	}
	return nil
}

func (r *Runner) Skip() bool {
	return r.testCase.Config.Skip
}

func readActualResults(resultsPath string) (string, error) {
	// no results
	if resultsPath == "" {
		return "", nil
	}
	l, err := ioutil.ReadDir(resultsPath)
	if err != nil {
		return "", fmt.Errorf("failed to get files in results dir: %w", err)
	}
	if len(l) > 1 {
		return "", fmt.Errorf("unexpected results files number %d, should be 0 or 1", len(l))
	}
	if len(l) == 0 {
		// no result file
		return "", nil
	}
	resultsFile := l[0].Name()
	actualResults, err := ioutil.ReadFile(filepath.Join(resultsPath, resultsFile))
	if err != nil {
		return "", fmt.Errorf("failed to read actual results: %w", err)
	}
	return strings.TrimSpace(string(actualResults)), nil
}

func readActualDiff(path, origHash string) (string, error) {
	err := gitAddAll(path)
	if err != nil {
		return "", err
	}
	err = gitCommit(path, "second")
	if err != nil {
		return "", err
	}
	// diff with first commit
	actualDiff, err := gitDiff(path, origHash, "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(actualDiff), nil
}

// expected contains the expected result for the function running
type expected struct {
	Results string
	Diff    string
}

func newExpected(path string) (expected, error) {
	e := expected{}
	// get expected results
	expectedResults, err := ioutil.ReadFile(filepath.Join(path, expectedDir, expectedResultsFile))
	switch {
	case os.IsNotExist(err):
		e.Results = ""
	case err != nil:
		return e, fmt.Errorf("failed to read expected results: %w", err)
	default:
		e.Results = strings.TrimSpace(string(expectedResults))
	}

	// get expected diff
	expectedDiff, err := ioutil.ReadFile(filepath.Join(path, expectedDir, expectedDiffFile))
	switch {
	case os.IsNotExist(err):
		e.Diff = ""
	case err != nil:
		return e, fmt.Errorf("failed to read expected diff: %w", err)
	default:
		e.Diff = strings.TrimSpace(string(expectedDiff))
	}

	return e, nil
}

func (r *Runner) updateExpected(tmpPkgPath, resultsPath, sourceOfTruthPath string) error {
	// We update results directory only when a result file already exists.
	l, err := ioutil.ReadDir(resultsPath)
	if err != nil {
		return err
	}
	if len(l) == 0 {
		actualDiff, err := readActualDiff(tmpPkgPath, r.initialCommit)
		if err != nil {
			return err
		}
		if actualDiff != "" {
			if err := ioutil.WriteFile(filepath.Join(sourceOfTruthPath, expectedDiffFile), []byte(actualDiff+"\n"), 0666); err != nil {
				return err
			}
		}
	} else {
		actualResults, err := readActualResults(resultsPath)
		if err != nil {
			return err
		}
		if actualResults != "" {
			if err := ioutil.WriteFile(filepath.Join(sourceOfTruthPath, expectedResultsFile), []byte(actualResults+"\n"), 0666); err != nil {
				return err
			}
		}
	}

	return nil
}
