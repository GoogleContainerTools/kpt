// Copyright 2020 Google LLC
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

package commands

import (
	"os"

	"github.com/GoogleContainerTools/kpt/internal/cmdfetchk8sschema"
	"github.com/GoogleContainerTools/kpt/internal/docs/generated/livedocs"
	"github.com/GoogleContainerTools/kpt/internal/util/setters"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/cli-utils/cmd/apply"
	"sigs.k8s.io/cli-utils/cmd/destroy"
	"sigs.k8s.io/cli-utils/cmd/diff"
	"sigs.k8s.io/cli-utils/cmd/initcmd"
	"sigs.k8s.io/cli-utils/cmd/preview"
)

func GetLiveCommand(name string, f util.Factory) *cobra.Command {
	liveCmd := &cobra.Command{
		Use:   "live",
		Short: livedocs.LiveShort,
		Long:  livedocs.LiveShort + "\n" + livedocs.LiveLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			h, err := cmd.Flags().GetBool("help")
			if err != nil {
				return err
			}
			if h {
				return cmd.Help()
			}
			return cmd.Usage()
		},
	}

	ioStreams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	initCmd := initcmd.NewCmdInit(ioStreams)
	initCmd.Short = livedocs.InitShort
	initCmd.Long = livedocs.InitShort + "\n" + livedocs.InitLong
	initCmd.Example = livedocs.InitExamples

	applyCmd := ApplyCommand(f, ioStreams)
	_ = applyCmd.Flags().MarkHidden("no-prune")
	applyCmd.Short = livedocs.ApplyShort
	applyCmd.Long = livedocs.ApplyShort + "\n" + livedocs.ApplyLong
	applyCmd.Example = livedocs.ApplyExamples

	previewCmd := PreviewCommand(f, ioStreams)
	previewCmd.Short = livedocs.PreviewShort
	previewCmd.Long = livedocs.PreviewShort + "\n" + livedocs.PreviewLong
	previewCmd.Example = livedocs.PreviewExamples

	diffCmd := diff.NewCmdDiff(f, ioStreams)
	diffCmd.Short = livedocs.DiffShort
	diffCmd.Long = livedocs.DiffShort + "\n" + livedocs.DiffLong
	diffCmd.Example = livedocs.DiffExamples

	destroyCmd := destroy.NewCmdDestroy(f, ioStreams)
	destroyCmd.Short = livedocs.DestroyShort
	destroyCmd.Long = livedocs.DestroyShort + "\n" + livedocs.DestroyLong
	destroyCmd.Example = livedocs.DestroyExamples

	fetchOpenAPICmd := cmdfetchk8sschema.NewCommand(name, f, ioStreams)

	liveCmd.AddCommand(initCmd, applyCmd, previewCmd, diffCmd, destroyCmd,
		fetchOpenAPICmd)

	return liveCmd
}

func ApplyCommand(f util.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	liveCmd := apply.ApplyCommand(f, ioStreams)
	applyCmd := *liveCmd
	applyCmd.RunE = func(c *cobra.Command, args []string) error {
		if err := setters.CheckRequiredSettersSet(args[0]); err != nil {
			return err
		}
		liveCmd.SetArgs(args)
		if err := liveCmd.Execute(); err != nil {
			return err
		}
		return nil
	}
	return &applyCmd
}

func PreviewCommand(f util.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	liveCmd := preview.NewCmdPreview(f, ioStreams)
	previewCmd := *liveCmd
	previewCmd.RunE = func(c *cobra.Command, args []string) error {
		if err := setters.CheckRequiredSettersSet(args[0]); err != nil {
			return err
		}
		liveCmd.SetArgs(args)
		if err := liveCmd.Execute(); err != nil {
			return err
		}
		return nil
	}
	return &previewCmd
}
