module github.com/GoogleContainerTools/kpt

go 1.14

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/spec v0.19.5
	github.com/olekukonko/tablewriter v0.0.4
	// TODO: find a library that have proper releases or just implement
	// topsort in kpt.
	github.com/philopon/go-toposort v0.0.0-20170620085441-9be86dbd762f
	github.com/pkg/errors v0.9.1
	github.com/posener/complete/v2 v2.0.1-alpha.12
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	gotest.tools v2.2.0+incompatible
	k8s.io/apimachinery v0.18.10
	k8s.io/cli-runtime v0.18.10
	k8s.io/client-go v0.18.10
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.18.10
	sigs.k8s.io/cli-utils v0.22.1-0.20201117031003-fd39030f0508
	sigs.k8s.io/kustomize/cmd/config v0.8.5
	sigs.k8s.io/kustomize/kyaml v0.9.4
)
