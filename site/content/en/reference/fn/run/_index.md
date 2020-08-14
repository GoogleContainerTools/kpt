---
title: "Run"
linkTitle: "run"
type: docs
description: >
   Locally execute one or more functions in containers
---

<!--mdtogo:Short
    Locally execute one or more functions in containers
-->

Generate, transform, or validate configuration files using locally run
functions.

Functions are packaged as container images, starlark scripts, or binary
executables which are run against the contents of a package.

Get an overview on how to use `kpt fn run` from the [Running Functions] guide.

This page dives into details of the `kpt fn run` command flow and serves as a
reference for advanced usecases.

## Synopsis

<!--mdtogo:Long-->

```sh
kpt fn run DIR [flags]
```

If the container exits with non-zero status code, run will fail and print the
container `STDERR`.

```sh
DIR:
  Path to a package directory.  Defaults to stdin if unspecified.
```

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

```sh
# read the Resources from DIR, provide them to a container my-fun as input,
# write my-fn output back to DIR
kpt fn run DIR/ --image gcr.io/example.com/my-fn
```

```sh
# provide the my-fn with an input ConfigMap containing `data: {foo: bar}`
kpt fn run DIR/ --image gcr.io/example.com/my-fn:v1.0.0 -- foo=bar
```

```sh
# run the functions in FUNCTIONS_DIR against the Resources in DIR
kpt fn run DIR/ --fn-path FUNCTIONS_DIR/
```

```sh
# discover functions in DIR and run them against Resource in DIR.
# functions may be scoped to a subset of Resources -- see `kpt help fn run`
kpt fn run DIR/
```

<!--mdtogo-->

## Structured Results

Functions may emit results using the structure defined in the
[typescript result] interface as an alternative to exiting with a non-zero
status code. Users may want to store these results separately from
configuration files. Kpt provides the `--results-dir` flag for users to
specify a destination to write results to.

**Example**: Run `validate-rolebinding` on an example package

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-sdk.git/example-configs example-configs
mkdir results/
kpt fn run example-configs/ --results-dir results/ --image gcr.io/kpt-functions/validate-rolebinding:results -- subject_name=bob@foo-corp.com
```

## Network Access

By default, container functions cannot access network. `kpt` may enable network
access using the `--network` flag, and specifying that a network is required in
the functionConfig.

**Example**: Run `kubeval` on a package

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/kubeval
        network:
          required: true
    config.kubernetes.io/local-config: 'true'
```

```sh
kpt pkg get https://github.com/instrumenta/kubeval.git/fixtures .
kpt fn source fixtures/*invalid.yaml |
  kpt fn run --fn-path fc.yaml --network 2>error.txt || true
```

## Mounting Directories

By default, container functions cannot access the local file system. `kpt` may
enable functions to mount volumes using the `--mount` flag passing the same
arguments as for `docker run`.

**Example**: Run `kustomize-build` on a helloWorld package

```sh
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples/helloWorld helloWorld
kpt fn source helloWorld |
  kpt fn run --mount type=bind,src="$(pwd)/helloWorld",dst=/source --image gcr.io/kpt-functions/kustomize-build -- path=/source |
  kpt fn sink .
```

All volumes are mounted readonly by default. Specify `rw=true` to mount volumes
in read-write mode.

```sh
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples/helloWorld helloWorld
kpt fn source helloWorld |
  kpt fn run --mount type=bind,src="$(pwd)/helloWorld",dst=/source,rw=true --image gcr.io/kpt-functions/kustomize-build -- path=/source |
  kpt fn sink .
```

`kpt` accepts the same options to `--mount` specified on the [Docker Volumes]
page.

Depending on the container image, the configuration function may not have
permissions to access mounted volumes. Check how the function is running inside
the container in case of permissions issues.

## Environment Variables

`kpt` exports all local environment variables except the `TMPDIR` variable
when launching a docker container. Check the value of any environment
variables that are declared locally as they may be affecting container
behavior.

```sh
# Check the value of the GOPATH environment variable
echo $GOPATH

# List all environment variables
env
```

## Deferring Failure

When running multiple validation functions, it may be desired to defer failures
until all functions have been run so that the results of all functions are
written to the results directory. Functions can specify that failures should be
deferred by specifying `deferFailure` in the declaration.

```yaml
apiVersion: example.com/v1alpha1
kind: ExampleFunction
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/example.com/image:v1.0.0
      # continue running functions if this fails, and fail at the end
      deferFailure: true
```

## Imperative Run Specifics

### Generating FunctionConfig for Imperative Runs

When functions are run imperatively, the functionConfig will be generated from
command line arguments.

When functions are run using the below command, the key-value pairs following
`--` are parsed into a ConfigMap which is set as the functionConfig. The
arguments are passed as `data` elements in the ConfigMap.

**Example**: Run `validate-rolebinding` on an example package

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-sdk.git/example-configs example-configs
mkdir results/
kpt fn run example-configs/ --results-dir results/ --image gcr.io/kpt-functions/validate-rolebinding:results -- subject_name=bob@foo-corp.com
```

Function Input:

```yaml
kind: ResourceList
functionConfig:
  apiVersion: v1
  kind: ConfigMap
  data:
    subject_name: bob@foo-corp.com
    ...
items:
  ...
```

If the first argument after `--` is _not_ a key=value pair, it will be used as
the functionConfig type.

Run the function:

```sh
kpt fn run DIR/ --image foo:v1 -- Foo a=b c=d
```

Function Input:

```yaml
kind: ResourceList
functionConfig:
  kind: Foo
  data:
    a: b
    c: d
    ...
items:
  ...
```

### Caveats to Running Imperatively

kpt does not handle imperatively running functions which use the following
types of arguments.

#### Complex arguments

Functions may take complex arguements such as lists and maps that have nested
fields. It's recommended to run such functions declaratively.

#### Arguments interpreted as flags

Some functions like `helm-template`, `istioctl-analyze`, and `kustomize-build`
take arbitrary command line flags as arguments. Passing arguments such as
`--use-kube=false` imperatively results in parsing issues. See more details in
the following:

* [Issue 823]
* [Issue 824]

When passing flags as arguments, it's recommended to run functions
declaratively.

#### Functions expecting spec field

`kpt fn run` provides any arguments passed imperatively to the container image
in a `ConfigMap` containing a `data` field. Some config functions may expect
arguemnts passed in a `spec` field instead. It's recommended to run such
functions declaratively by passing a `ConfigMap` with a `spec` field. See more
details in the following:

* [Issue 757]

## Declarative Run Specifics

### Scoping Rules

Functions which are nested under some sub directory are scoped only to
Resources under that same sub directory. This allows fine grain control over
how functions are executed.

`kpt fn run DIR/` will recursively traverse DIR/ looking for declared functions
and invoking them -- passing in only those resources scoped to the function.

**Example:** Function declared in `stuff/my-function.yaml` is scoped to
Resources in `stuff/` and is NOT scoped to Resources in `apps/`

```sh
.
├── stuff
│   ├── inscope-deployment.yaml
│   ├── stuff2
│   │     └── inscope-deployment.yaml
│   └── my-function.yaml # functions is scoped to stuff/...
└── apps
    ├── not-inscope-deployment.yaml
    └── not-inscope-service.yaml
```

Alternatively, you can also place function configurations in a special
directory named `functions`.

**Example**: This is equivalent to previous example

```sh
.
├── stuff
│   ├── inscope-deployment.yaml
│   ├── stuff2
│   │     └── inscope-deployment.yaml
│   └── functions
│         └── my-function.yaml
└── apps
    ├── not-inscope-deployment.yaml
    └── not-inscope-service.yaml
```

Alternatively, you can also use `--fn-path` to explicitly provide the directory
containing function configurations:

```sh
kpt fn run DIR/ --fn-path FUNCTIONS_DIR/
```

Alternatively, scoping can be disabled using `--global-scope` flag.

### Declaring Multiple Functions

You may declare multiple functions. If they are specified in the same file
(multi-object YAML file separated by `---`), they will
be run sequentially in the order that they are specified.

If multiple functions are specified in different yaml files, there is no
guarantee of the order in which kpt will run the functions.

### Custom `functionConfig`

Functions may define their own API input types - these may be client-side
equivalents of CRDs:

**Example**: Declare two functions in `DIR/functions/my-functions.yaml`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/example.com/my-fn
data:
  foo: bar
---
apiVersion: v1
kind: MyType
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/example.com/other-fn
spec:
  field:
    nestedField: value
  --flag-arg: flag-value
```

Because the first function uses the default `ConfigMap` kind with arguments as
key-value pairs in the `data` field, kpt is able to run it both imperatively
and declaratively. kpt is only able to run the second function declaratively
because:

* it expects arguments in the `spec` field
* it takes complex arguments with nested values
* it takes flags as arguments

## Next Steps

* See more examples of functions in the [functions catalog].
* Get a quickstart on writing functions from the [function producer docs].
* Find out how to structure a pipeline of functions from the
  [functions concepts] page.

[Running Functions]: ../../../guides/consumer/function/
[typescript result]: https://github.com/GoogleContainerTools/kpt-functions-sdk/blob/master/ts/kpt-functions/src/types.ts
[Docker Volumes]: https://docs.docker.com/storage/volumes/
[Issue 823]: https://github.com/GoogleContainerTools/kpt/issues/823/
[Issue 824]: https://github.com/GoogleContainerTools/kpt/issues/824/
[Issue 757]: https://github.com/GoogleContainerTools/kpt/issues/757/
[functions catalog]: ../../../guides/consumer/function/catalog/
[function producer docs]: ../../../guides/producer/functions/
[functions concepts]: ../../../concepts/functions/
