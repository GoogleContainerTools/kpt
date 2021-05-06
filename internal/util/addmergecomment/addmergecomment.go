package addmergecomment

import (
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt/internal/util/merge"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// AddMergeComment adds merge comments with format "kpt-merge: namespace/name"
// to all resources in the package
type AddMergeComment struct{}

// Process invokes AddMergeComment kyaml filter on the resources in input packages paths
func Process(paths ...string) error {
	for _, path := range paths {
		inout := &kio.LocalPackageReadWriter{PackagePath: path}
		amc := &AddMergeComment{}
		err := kio.Pipeline{
			Inputs:  []kio.Reader{inout},
			Filters: []kio.Filter{kio.FilterAll(amc)},
			Outputs: []kio.Writer{inout},
		}.Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

// Filter implements kyaml.Filter
// this filter adds merge comment with format "kpt-merge: namespace/name" to
// the input resource, if the namespace field doesn't exist on the resource,
// it uses "default" namespace
func (amc *AddMergeComment) Filter(object *kyaml.RNode) (*kyaml.RNode, error) {
	rm, err := object.GetMeta()
	if err != nil {
		// skip adding merge comment if no metadata
		return object, nil
	}
	mf := object.Field(kyaml.MetadataField)
	if mf.IsNilOrEmpty() {
		// skip adding merge comment if empty metadata
		return object, nil
	}
	if strings.Contains(mf.Key.YNode().LineComment, merge.MergeCommentPrefix) {
		// skip adding merge comment if merge comment is already present
		return object, nil
	}
	mf.Key.YNode().LineComment = fmt.Sprintf("%s %s/%s", merge.MergeCommentPrefix, rm.Namespace, rm.Name)
	return object, nil
}
