// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmdsink

import (
	"context"
	"os"

	"github.com/GoogleContainerTools/kpt/internal/docs/generated/fndocs"
	"github.com/GoogleContainerTools/kpt/internal/pkg"
	"github.com/GoogleContainerTools/kpt/internal/printer"
	"github.com/GoogleContainerTools/kpt/thirdparty/cmdconfig/commands/runner"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// GetSinkRunner returns a command for Sink.
func GetSinkRunner(ctx context.Context, name string) *SinkRunner {
	r := &SinkRunner{
		Ctx: ctx,
	}
	c := &cobra.Command{
		Use:     "sink DIR [flags]",
		Short:   fndocs.SinkShort,
		Long:    fndocs.SinkShort + "\n" + fndocs.SinkLong,
		Example: fndocs.SinkExamples,
		RunE:    r.runE,
	}
	r.Command = c
	return r
}

func NewCommand(ctx context.Context, name string) *cobra.Command {
	return GetSinkRunner(ctx, name).Command
}

// SinkRunner contains the run function
type SinkRunner struct {
	Command *cobra.Command
	Ctx     context.Context
}

func (r *SinkRunner) runE(c *cobra.Command, args []string) error {
	dir := pkg.CurDir
	if len(args) > 0 {
		dir = args[0]
	}

	pr := printer.FromContextOrDie(r.Ctx)

	dirAlreadyExists := true
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		dirAlreadyExists = false
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}
	}

	outputs := []kio.Writer{&kio.LocalPackageWriter{PackagePath: dir}}

	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: c.InOrStdin()}},
		Outputs: outputs}.Execute()
	if !dirAlreadyExists {
		pr.Printf("directory %q doesn't exist, creating the directory...\n", dir)
	}
	return runner.HandleError(r.Ctx, err)
}
