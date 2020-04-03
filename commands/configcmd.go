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

package commands

import (
	"github.com/GoogleContainerTools/kpt/internal/docs/generated/cfgdocs"
	"github.com/GoogleContainerTools/kpt/internal/util/functions"
	"github.com/GoogleContainerTools/kpt/internal/util/setters"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/configcobra"
)

func GetConfigCommand(name string) *cobra.Command {
	cfgCmd := &cobra.Command{
		Use:     "cfg",
		Short:   cfgdocs.CfgShort,
		Long:    cfgdocs.CfgShort + "\n" + cfgdocs.CfgLong,
		Example: cfgdocs.CfgExamples,
		Aliases: []string{"config"},
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

	an := configcobra.Annotate(name)
	an.Short = cfgdocs.AnnotateShort
	an.Long = cfgdocs.AnnotateShort + "\n" + cfgdocs.AnnotateLong
	an.Example = cfgdocs.AnnotateExamples

	cat := configcobra.Cat(name)
	cat.Short = cfgdocs.CatShort
	cat.Long = cfgdocs.CatShort + "\n" + cfgdocs.CatLong
	cat.Example = cfgdocs.CatExamples

	count := configcobra.Count(name)
	count.Short = cfgdocs.CountShort
	count.Long = cfgdocs.CountShort + "\n" + cfgdocs.CountLong
	count.Example = cfgdocs.CountExamples

	createSetter := configcobra.CreateSetter(name)
	createSetter.Short = cfgdocs.CreateSetterShort
	createSetter.Long = cfgdocs.CreateSetterShort + "\n" + cfgdocs.CreateSetterLong
	createSetter.Example = cfgdocs.CreateSetterExamples

	createSubstitution := configcobra.CreateSubstitution(name)
	createSubstitution.Short = cfgdocs.CreateSubstShort
	createSubstitution.Long = cfgdocs.CreateSubstShort + "\n" + cfgdocs.CreateSubstLong
	createSubstitution.Example = cfgdocs.CreateSubstExamples

	fmt := configcobra.Fmt(name)
	fmt.Short = cfgdocs.FmtShort
	fmt.Long = cfgdocs.FmtShort + "\n" + cfgdocs.FmtLong
	fmt.Example = cfgdocs.FmtExamples

	grep := configcobra.Grep(name)
	grep.Short = cfgdocs.GrepShort
	grep.Long = cfgdocs.GrepShort + "\n" + cfgdocs.GrepLong
	grep.Example = cfgdocs.GrepExamples

	listSetters := configcobra.ListSetters(name)
	listSetters.Short = cfgdocs.ListSettersShort
	listSetters.Long = cfgdocs.ListSettersShort + "\n" + cfgdocs.ListSettersLong
	listSetters.Example = cfgdocs.ListSettersExamples

	set := SetCommand(name)
	set.Short = cfgdocs.SetShort
	set.Long = cfgdocs.SetShort + "\n" + cfgdocs.SetLong
	set.Example = cfgdocs.SetExamples

	tree := configcobra.Tree(name)
	tree.Short = cfgdocs.TreeShort
	tree.Long = cfgdocs.TreeShort + "\n" + cfgdocs.TreeLong
	tree.Example = cfgdocs.TreeExamples

	cfgCmd.AddCommand(an, cat, count, createSetter, createSubstitution, fmt,
		grep, listSetters, set, tree)
	return cfgCmd
}

// SetCommand wraps the kustomize set command in order to automatically update
// a project number if a project id is set.
func SetCommand(parent string) *cobra.Command {
	kustomizeCmd := configcobra.Set(parent)
	setCmd := *kustomizeCmd
	setCmd.RunE = func(c *cobra.Command, args []string) error {
		kustomizeCmd.SetArgs(args)
		if err := kustomizeCmd.Execute(); err != nil {
			return err
		}

		if err := functions.ReconcileFunctions(args[0]); err != nil {
			return err
		}

		if len(args) != 3 || args[1] != "gcloud.core.project" {
			return nil
		}
		projectNumber, err := setters.GetProjectNumberFromProjectID(args[2])
		if err != nil {
			return nil
		}
		kustomizeCmd.SetArgs([]string{args[0], "gcloud.project.projectNumber", projectNumber})
		return kustomizeCmd.Execute()
	}
	return &setCmd
}
