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

// Package cmdget contains the get command
package cmdrender

import (
	"context"
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt/internal/errors"
	"github.com/GoogleContainerTools/kpt/internal/util/cmdutil"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner
func NewRunner(ctx context.Context, parent string) *Runner {
	r := &Runner{ctx: ctx}
	c := &cobra.Command{
		Use:     "render [DIR]",
		Short:   "render",
		Long:    "render",
		Example: "render",
		RunE:    r.runE,
		PreRunE: r.preRunE,
	}
	c.Flags().StringVar(&r.resultsDirPath, "results-dir", "",
		"path to a directory to save function results")
	cmdutil.FixDocs("kpt", parent, c)
	r.Command = c
	return r
}

func NewCommand(ctx context.Context, parent string) *cobra.Command {
	return NewRunner(ctx, parent).Command
}

// Runner contains the run function pipeline run command
type Runner struct {
	pkgPath        string
	resultsDirPath string
	Command        *cobra.Command
	ctx            context.Context
}

func (r *Runner) preRunE(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		// no pkg path specified, default to current working dir
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		r.pkgPath = wd
	} else {
		// resolve and validate the provided path
		r.pkgPath = args[0]
	}
	if r.resultsDirPath != "" {
		if _, err := os.Stat(r.resultsDirPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("results-dir %q must exist", r.resultsDirPath)
			}
			return fmt.Errorf("results-dir %q check failed: %w", r.resultsDirPath, err)
		}
	}
	return nil
}

func (r *Runner) runE(c *cobra.Command, _ []string) error {
	err := cmdutil.DockerCmdAvailable()
	if err != nil {
		return err
	}
	executor := Executor{
		PkgPath:        r.pkgPath,
		ResultsDirPath: r.resultsDirPath,
	}
	if err = executor.Execute(r.ctx); err != nil {
		return err
	}
	return nil
}
