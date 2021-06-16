// Code generated by "mdtogo"; DO NOT EDIT.
package fndocs

var FnShort = `Transform and validate packages using containerized functions.`
var FnLong = `
The ` + "`" + `fn` + "`" + ` command group contains subcommands for transforming and validating ` + "`" + `kpt` + "`" + ` packages
using containerized functions.
`

var DocShort = `Display the documentation for a function`
var DocLong = `
` + "`" + `kpt fn doc` + "`" + ` invokes the function container with ` + "`" + `--help` + "`" + ` flag.

  kpt fn doc --image=IMAGE

--image is a required flag. If the function supports --help, it will print the
documentation to STDOUT. Otherwise, it will exit with non-zero exit code and
print the error message to STDERR.
`
var DocExamples = `
  # diplay the documentation for image gcr.io/kpt-fn/set-namespace:v0.1.1
  $ kpt fn doc --image gcr.io/kpt-fn/set-namespace:v0.1.1
`

var EvalShort = `Execute function on resources`
var EvalLong = `
  kpt fn eval [DIR|-] [flags] [-- fn-args]

Args:

  DIR|-:
    Path to the local directory containing resources. Defaults to the current
    working directory. Using '-' as the directory path will cause ` + "`" + `eval` + "`" + ` to
    read resources from ` + "`" + `stdin` + "`" + ` and write the output to ` + "`" + `stdout` + "`" + `. When resources are
    read from ` + "`" + `stdin` + "`" + `, they must be in one of the following input formats:
  
    1. Multi object YAML where resources are separated by ` + "`" + `---` + "`" + `.
  
    2. KRM Function Specification wire format where resources are wrapped in an object
       of kind ResourceList.
  
    If the output is written to ` + "`" + `stdout` + "`" + `, resources are written in multi object YAML
    format where resources are separated by ` + "`" + `---` + "`" + `.

  fn-args:
    function arguments to be provided as input to the function. These must be
    provided in the ` + "`" + `key=value` + "`" + ` format and come after the separator ` + "`" + `--` + "`" + `.

Flags:

  --as-current-user:
    Use the ` + "`" + `uid` + "`" + ` and ` + "`" + `gid` + "`" + ` of the kpt process for container function execution.
    By default, container function is executed as ` + "`" + `nobody` + "`" + ` user. You may want to use
    this flag to run higher privilege operations such as mounting the local filesystem.
  
  --env, e:
    List of local environment variables to be exported to the container function.
    By default, none of local environment variables are made available to the
    container running the function. The value can be in ` + "`" + `key=value` + "`" + ` format or only
    the key of an already exported environment variable.
  
  --exec:
    Path to the local executable binary to execute as a function. ` + "`" + `eval` + "`" + ` executes
    only one function, so do not use ` + "`" + `--image` + "`" + ` flag with this flag. This is useful
    for testing function locally during development. It enables faster dev iterations
    by avoiding the function to be published as container image.
  
  --fn-config:
    Path to the file containing ` + "`" + `functionConfig` + "`" + ` for the function.
  
  --image, i:
    Container image of the function to execute e.g. ` + "`" + `gcr.io/kpt-fn/set-namespace:v0.1` + "`" + `.
    ` + "`" + `eval` + "`" + ` executes only one function, so do not use ` + "`" + `--exec` + "`" + ` flag with this flag.
  
  --image-pull-policy:
    If the image should be pulled before rendering the package(s). It can be set
    to one of always, ifNotPresent, never. If unspecified, always will be the
    default.
    If using always, kpt will ensure the function images to run are up-to-date
    with the remote container registry. This can be useful for tags like v1.
    If using ifNotPresent, kpt will only pull the image when it can't find it in
    the local cache.
    If using never, kpt will only use images from the local cache.
  
  --include-meta-resources:
    If enabled, meta resources (i.e. ` + "`" + `Kptfile` + "`" + ` and ` + "`" + `functionConfig` + "`" + `) are included
    in the input to the function. By default it is disabled.
  
  --mount:
    List of storage options to enable reading from the local filesytem. By default,
    container functions can not access the local filesystem. It accepts the same options
    as specified on the [Docker Volumes] for ` + "`" + `docker run` + "`" + `. All volumes are mounted
    readonly by default. Specify ` + "`" + `rw=true` + "`" + ` to mount volumes in read-write mode.
  
  --network:
    If enabled, container functions are allowed to access network.
    By default it is disabled.
  
  --output, o:
    If specified, the output resources are written to provided location,
    if not specified, resources are modified in-place.
    Allowed values: stdout|unwrap|<OUT_DIR_PATH>
    1. stdout: output resources are wrapped in ResourceList and written to stdout.
    2. unwrap: output resources are written to stdout, in multi-object yaml format.
    3. OUT_DIR_PATH: output resources are written to provided directory, the directory is created if it doesn't already exist.
  
  --results-dir:
    Path to a directory to write structured results. Directory will be created if
    it doesn't exist. Structured results emitted by the functions are aggregated and saved
    to ` + "`" + `results.yaml` + "`" + ` file in the specified directory.
    If not specified, no result files are written to the local filesystem.
`
var EvalExamples = `
  # execute container my-fn on the resources in DIR directory and
  # write output back to DIR
  $ kpt fn eval DIR -i gcr.io/example.com/my-fn

  # execute container my-fn on the resources in DIR directory with
  # ` + "`" + `functionConfig` + "`" + ` my-fn-config
  $ kpt fn eval DIR -i gcr.io/example.com/my-fn --fn-config my-fn-config

  # execute container my-fn with an input ConfigMap containing ` + "`" + `data: {foo: bar}` + "`" + `
  $ kpt fn eval DIR -i gcr.io/example.com/my-fn:v1.0.0 -- foo=bar

  # execute executable my-fn on the resources in DIR directory and
  # write output back to DIR
  $ kpt fn eval DIR --exec ./my-fn

  # execute container my-fn on the resources in DIR directory,
  # save structured results in /tmp/my-results dir and write output back to DIR
  $ kpt fn eval DIR -i gcr.io/example.com/my-fn --results-dir /tmp/my-results-dir

  # execute container my-fn on the resources in DIR directory with network access enabled,
  # and write output back to DIR
  $ kpt fn eval DIR -i gcr.io/example.com/my-fn --network

  # execute container my-fn on the resource in DIR and export KUBECONFIG
  # and foo environment variable
  $ kpt fn eval DIR -i gcr.io/example.com/my-fn --env KUBECONFIG -e foo=bar

  # execute kubeval function by mounting schema from a local directory on wordpress package
  $ kpt fn eval -i gcr.io/kpt-fn/kubeval:v0.1 \
    --mount type=bind,src="/path/to/schema-dir",dst=/schema-dir \
    --as-current-user wordpress -- additional_schema_locations=/schema-dir

  # chaining functions using the unix pipe to set namespace and set labels on
  # wordpress package
  $ kpt fn source wordpress \
    | kpt fn eval - -i gcr.io/kpt-fn/set-namespace:v0.1 -- namespace=mywordpress \
    | kpt fn eval - -i gcr.io/kpt-fn/set-labels:v0.1 -- label_name=color label_value=orange \
    | kpt fn sink wordpress

  # execute container 'set-namespace' on the resources in current directory and write
  # the output resources to another directory
  $ kpt fn eval -i gcr.io/kpt-fn/set-namespace:v0.1 -o path/to/dir -- namespace=mywordpress

  # execute container 'set-namespace' on the resources in current directory and write
  # the output resources to stdout which are piped to 'kubectl apply'
  $ kpt fn eval -i gcr.io/kpt-fn/set-namespace:v0.1 -o unwrap -- namespace=mywordpress \
  | kubectl apply -f -

  # execute container 'set-namespace' on the resources in current directory and write
  # the wrapped output resources to stdout which are passed to 'set-annotations' function
  # and the output resources after setting namespace and annotation is written to another directory
  $ kpt fn eval -i gcr.io/kpt-fn/set-namespace:v0.1 -o stdout -- namespace=staging \
  | kpt fn eval - -i gcr.io/kpt-fn/set-annotations:v0.1.3 -o path/to/dir -- foo=bar
`

var ExportShort = `Auto-generating function pipelines for different workflow orchestrators`
var ExportLong = `
  kpt fn export DIR/ [--fn-path FUNCTIONS_DIR/] --workflow ORCHESTRATOR [--output OUTPUT_FILENAME]
  
  DIR:
    Path to a package directory.
  FUNCTIONS_DIR:
    Read functions from the directory instead of the DIR/.
  ORCHESTRATOR:
    Supported orchestrators are:
      - github-actions
      - cloud-build
      - gitlab-ci
      - jenkins
      - tekton
      - circleci
  OUTPUT_FILENAME:
    Specifies the filename of the generated pipeline. If omitted, the default
    output is stdout
`
var ExportExamples = `

  # read functions from DIR, run them against it as one step.
  # write the generated GitHub Actions pipeline to main.yaml.
  kpt fn export DIR/ --output main.yaml --workflow github-actions


  # discover functions in FUNCTIONS_DIR and run them against resource in DIR.
  # write the generated Cloud Build pipeline to stdout.
  kpt fn export DIR/ --fn-path FUNCTIONS_DIR/ --workflow cloud-build
`

var RenderShort = `Render a package.`
var RenderLong = `
  kpt fn render [PKG_PATH] [flags]

Args:

  PKG_PATH:
    Local package path to render. Directory must exist and contain a Kptfile
    to be updated. Defaults to the current working directory.

Flags:

  --image-pull-policy:
    If the image should be pulled before rendering the package(s). It can be set
    to one of always, ifNotPresent, never. If unspecified, always will be the
    default.
  
  --output, o:
    If specified, the output resources are written to provided location,
    if not specified, resources are modified in-place.
    Allowed values: stdout|unwrap|<OUT_DIR_PATH>
    1. stdout: output resources are wrapped in ResourceList and written to stdout.
    2. unwrap: output resources are written to stdout, in multi-object yaml format.
    3. OUT_DIR_PATH: output resources are written to provided directory, the directory is created if it doesn't already exist.
  
  --results-dir:
    Path to a directory to write structured results. Directory will be created if
    it doesn't exist. Structured results emitted by the functions are aggregated and saved
    to ` + "`" + `results.yaml` + "`" + ` file in the specified directory.
    If not specified, no result files are written to the local filesystem.
`
var RenderExamples = `
  # Render the package in current directory
  $ kpt fn render

  # Render the package in current directory and save results in my-results-dir
  $ kpt fn render --results-dir my-results-dir

  # Render my-package-dir
  $ kpt fn render my-package-dir

  # Render the package in current directory and write output resources to another DIR
  $ kpt fn render -o path/to/dir

  # Render resources in current directory and write unwrapped resources to stdout
  # which can be piped to kubectl apply
  $ kpt fn render -o unwrap | kubectl apply -f -

  # Render resources in current directory, write the wrapped resources
  # to stdout which are piped to 'set-annotations' function,
  # the transformed resources are written to another directory
  $ kpt fn render -o stdout \
  | kpt fn eval - -i gcr.io/kpt-fn/set-annotations:v0.1.3 -o path/to/dir  -- foo=bar
`

var SinkShort = `Write resources to a local directory`
var SinkLong = `
  kpt fn sink [DIR] [flags]
  
  DIR:
    Path to a local directory to write resources to. Defaults to the current
    working directory. Directory must exist.
`
var SinkExamples = `
  # read resources from DIR directory, execute my-fn on them and write the
  # output to DIR directory.
  $ kpt fn source DIR |
    kpt fn eval - --image gcr.io/example.com/my-fn - |
    kpt fn sink DIR
`

var SourceShort = `Source resources from a local directory`
var SourceLong = `
  kpt fn source [DIR] [flags]

Args:

  DIR:
    Path to the local directory containing resources. Defaults to the current
    working directory.

Flags:

  --fn-config:
    Path to the file containing ` + "`" + `functionConfig` + "`" + `.
  
  --include-meta-resources:
    If enabled, meta resources (i.e. ` + "`" + `Kptfile` + "`" + ` and ` + "`" + `functionConfig` + "`" + `) are included
    in the output of the command. By default it is disabled.
  
  --output, o:
    If specified, the output resources are written to stdout in provided format.
    Allowed values:
    1. stdout(default): output resources are wrapped in ResourceList and written to stdout.
    2. unwrap: output resources are written to stdout, in multi-object yaml format.
`
var SourceExamples = `
  # read resources from DIR directory and write the output on stdout.
  $ kpt fn source DIR

  # read resources from DIR directory, execute my-fn on them and write the
  # output to DIR directory.
  $ kpt fn source DIR |
    kpt fn eval - --image gcr.io/example.com/my-fn - |
    kpt fn sink DIR
`
