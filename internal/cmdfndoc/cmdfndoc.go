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

// Package cmddesc contains the desc command
package cmdfndoc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/GoogleContainerTools/kpt/internal/docs/generated/fndocs"
	"github.com/GoogleContainerTools/kpt/internal/util/cmdutil"
	"github.com/spf13/cobra"
)

func NewRunner(parent string) *Runner {
	r := &Runner{}
	c := &cobra.Command{
		Use:     "doc [PKG_PATH] [flags]",
		Args:    cobra.MaximumNArgs(1),
		Short:   fndocs.DocShort,
		Long:    fndocs.DocShort + "\n" + fndocs.DocLong,
		Example: fndocs.DocExamples,
		RunE:    r.runE,
	}
	r.Command = c
	c.Flags().StringVar(&r.Image, "image", "", "kpt function image name")
	cmdutil.FixDocs("kpt", parent, c)
	return r
}

func NewCommand(parent string) *cobra.Command {
	return NewRunner(parent).Command
}

type Runner struct {
	Image   string
	Command *cobra.Command
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	var out, errout bytes.Buffer
	dockerRunArgs := []string{
		"run",
		"--rm",                         // delete the container afterward
		"-a", "STDOUT", "-a", "STDERR", // attach stdin, stdout, stderr
		r.Image,
		"--help",
	}
	cmd := exec.Command("docker", dockerRunArgs...)
	cmd.Stdout = &out
	cmd.Stderr = &errout
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, errout.String())
		return fmt.Errorf("please ensure the container has an entrypoint and it supports --help flag: %w", err)
	}
	fmt.Fprintln(os.Stdout, out.String())
	return nil
}
