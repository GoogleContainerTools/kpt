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

// Code generated by "mdtogo"; DO NOT EDIT.
package pkgdocs

var PkgShort = `Fetch, update, and sync configuration files using git`
var PkgLong = `
|              Reads From | Writes To                |
|-------------------------|--------------------------|
| git repository          | local directory          |

The ` + "`" + `pkg` + "`" + ` command group contains subcommands which read remote upstream
git repositories, and write local directories.  They are focused on
providing porcelain on top of workflows which would otherwise require
wrapping git to pull clone subdirectories and perform updates by merging
resources rather than files.
`
var PkgExamples = `
  # create your workspace
  $ mkdir hello-world-workspace
  $ cd hello-world-workspace
  $ git init
  
  # get the package
  $ export SRC_REPO=https://github.com/GoogleContainerTools/kpt.git
  $ kpt pkg get $SRC_REPO/package-examples/helloworld-set@v0.3.0 helloworld
  
  # add helloworld to your workspace
  $ git add .
  $ git commit -am "Add hello world to my workspace."
  
  # pull in upstream updates by merging Resources
  $ kpt pkg update helloworld@v0.5.0 --strategy=resource-merge
`

var DescShort = `Display upstream package metadata`
var DescLong = `
  kpt pkg desc DIR
  
  DIR:
    Path to a package directory
`
var DescExamples = `
<!-- @pkgDesc @verifyExamples-->
  # display description for the local hello-world package
  kpt pkg desc hello-world/
`

var DiffShort = `Diff a local package against upstream`
var DiffLong = `
  kpt pkg diff [DIR@VERSION]

Args:

  DIR:
    Local package to compare. Command will fail if the directory doesn't exist, or does not
    contain a Kptfile.  Defaults to the current working directory.
  
  VERSION:
    A git tag, branch, ref or commit. Specified after the local_package with @ -- pkg_dir@version.
    Defaults to the local package version that was last fetched.

Flags:

  --diff-type:
    The type of changes to view (local by default). Following types are
    supported:
  
    local: shows changes in local package relative to upstream source package
           at original version
    remote: shows changes in upstream source package at target version
            relative to original version
    combined: shows changes in local package relative to upstream source
              package at target version
    3way: shows changes in local package and source package at target version
          relative to original version side by side
  
  --diff-tool:
    Commandline tool (diff by default) for showing the changes.
    Note that it overrides the KPT_EXTERNAL_DIFF environment variable.
    
    # Show changes using 'meld' commandline tool
    kpt pkg diff @master --diff-tool meld
  
  --diff-opts:
    Commandline options to use with the diffing tool.
    Note that it overrides the KPT_EXTERNAL_DIFF_OPTS environment variable.
    # Show changes using "diff" with recurive options
    kpt pkg diff @master --diff-tool meld --diff-opts "-r"

Environment Variables:

  KPT_EXTERNAL_DIFF:
     Commandline diffing tool (diff by default) that will be used to show
     changes.
     # Use meld to show changes
     KPT_EXTERNAL_DIFF=meld kpt pkg diff
  
  KPT_EXTERNAL_DIFF_OPTS:
     Commandline options to use for the diffing tool. For ex.
     # Using "-a" diff option
     KPT_EXTERNAL_DIFF_OPTS="-a" kpt pkg diff --diff-tool meld
`
var DiffExamples = `
<!-- @pkgDiff @verifyExamples-->
  # Show changes in current package relative to upstream source package
  kpt pkg diff

  # Show changes in current package relative to upstream source package
  # using meld tool with auto compare option.
  kpt pkg diff --diff-tool meld --diff-tool-opts "-a"

<!-- @pkgDiff @verifyExamples-->
  # Show changes in upstream source package between current version and
  # target version
  kpt pkg diff @v0.4.0 --diff-type remote

<!-- @pkgDiff @verifyExamples-->
  # Show changes in current package relative to target version
  kpt pkg diff @v0.4.0 --diff-type combined

  # Show 3way changes between the local package, upstream package at original
  # version and upstream package at target version using meld
  kpt pkg diff @v0.4.0 --diff-type 3way --diff-tool meld --diff-tool-opts "-a"
`

var FixShort = `Fix a local package which is using deprecated features.`
var FixLong = `
  kpt pkg fix LOCAL_PKG_DIRECTORY [flags]
  
  Args:
    LOCAL_PKG_DIRECTORY:
      Local directory with kpt package. Directory must exist and
      contain a Kptfile.
  
  Flags:
    --dry-run
      if set, the fix command shall only print the fixes which will be made to the
      package without actually fixing/modifying the resources.
  
`

var GetShort = `Fetch a package from a git repo.`
var GetLong = `
  kpt pkg get REPO_URI[.git]/PKG_PATH[@VERSION] LOCAL_DEST_DIRECTORY [flags]
  
  REPO_URI:
    URI of a git repository containing 1 or more packages as subdirectories.
    In most cases the .git suffix should be specified to delimit the REPO_URI
    from the PKG_PATH, but this is not required for widely recognized repo
    prefixes.  If get cannot parse the repo for the directory and version,
    then it will print an error asking for '.git' to be specified as part of
    the argument.
    e.g. https://github.com/kubernetes/examples.git
    Specify - to read Resources from stdin and write to a LOCAL_DEST_DIRECTORY
  
  PKG_PATH:
    Path to remote subdirectory containing Kubernetes resource configuration
    files or directories. Defaults to the root directory.
    Uses '/' as the path separator (regardless of OS).
    e.g. staging/cockroachdb
  
  VERSION:
    A git tag, branch, ref or commit for the remote version of the package
    to fetch.  Defaults to the repository master branch.
    e.g. @master
  
  LOCAL_DEST_DIRECTORY:
    The local directory to write the package to.
    e.g. ./my-cockroachdb-copy
  
      * If the directory does NOT exist: create the specified directory
        and write the package contents to it
      * If the directory DOES exist: create a NEW directory under the
        specified one, defaulting the name to the Base of REPO/PKG_PATH
      * If the directory DOES exist and already contains a directory with
        the same name of the one that would be created: fail
`
var GetExamples = `
<!-- @pkgGet @verifyExamples-->
  # fetch package cockroachdb from github.com/kubernetes/examples/staging/cockroachdb
  # creates directory ./cockroachdb/ containing the package contents
  kpt pkg get https://github.com/kubernetes/examples.git/staging/cockroachdb@master ./

<!-- @pkgGet @verifyExamples-->
  # fetch a cockroachdb
  # if ./my-package doesn't exist, creates directory ./my-package/ containing
  # the package contents
  kpt pkg get https://github.com/kubernetes/examples.git/staging/cockroachdb@master ./my-package/

<!-- @pkgGet @verifyExamples-->
  # fetch package examples from github.com/kubernetes/examples
  # creates directory ./examples fetched from the provided commit hash
  kpt pkg get https://github.com/kubernetes/examples.git/@6fe2792 ./
`

var InitShort = `Initialize an empty package`
var InitLong = `
  kpt pkg init DIR [flags]

Args:

  DIR:
    Init fails if DIR does not already exist

Flags:

  --description
    short description of the package. (default "sample description")
  
  --name
    package name.  defaults to the directory base name.
  
  --tag
    list of tags for the package.
  
  --url
    link to page with information about the package.
`
var InitExamples = `
<!-- @pkgInit @verifyStaleExamples-->
  # writes Kptfile package meta if not found
  mkdir my-pkg
  kpt pkg init my-pkg --tag kpt.dev/app=cockroachdb \
      --description "my cockroachdb implementation"
`

var SyncShort = `Fetch and update packages declaratively`
var SyncLong = `
  kpt pkg sync LOCAL_PKG_DIR [flags]
  
  LOCAL_PKG_DIR:
    Local package with dependencies to sync.  Directory must exist and
    contain a Kptfile.

Env Vars:

  KPT_CACHE_DIR:
    Controls where to cache remote packages during updates.
    Defaults to ~/.kpt/repos/
`
var SyncExamples = `
<!-- @pkgSync @verifyStaleExamples-->
  # print the dependencies that would be modified
  kpt pkg sync . --dry-run

<!-- @pkgSync @verifyExamples-->
  # sync the dependencies
  kpt pkg sync .
`

var SetShort = `Add a sync dependency to a Kptfile`
var SetLong = `
  kpt pkg set REPO_URI[.git]/PKG_PATH[@VERSION] LOCAL_DEST_DIRECTORY [flags]
  
  REPO_URI:
    URI of a git repository containing 1 or more packages as subdirectories.
    In most cases the .git suffix should be specified to delimit the REPO_URI
    from the PKG_PATH, but this is not required for widely recognized repo
    prefixes.  If get cannot parse the repo for the directory and version,
    then it will print an error asking for '.git' to be specified as part of
    the argument.
    e.g. https://github.com/kubernetes/examples.git
    Specify - to read Resources from stdin and write to a LOCAL_DEST_DIRECTORY
  
  PKG_PATH:
    Path to remote subdirectory containing Kubernetes Resource configuration
    files or directories.  Defaults to the root directory.
    Uses '/' as the path separator (regardless of OS).
    e.g. staging/cockroachdb
  
  VERSION:
    A git tag, branch, ref or commit for the remote version of the package to
    fetch.  Defaults to the repository master branch.
    e.g. @master
  
  LOCAL_DEST_DIRECTORY:
    The local directory to write the package to. e.g. ./my-cockroachdb-copy
  
      * If the directory does NOT exist: create the specified directory and write
        the package contents to it
      * If the directory DOES exist: create a NEW directory under the specified one,
        defaulting the name to the Base of REPO/PKG_PATH
      * If the directory DOES exist and already contains a directory with the same name
        of the one that would be created: fail

Flags:

  --strategy:
    Controls how changes to the local package are handled.
    Defaults to fast-forward.
  
      * resource-merge: perform a structural comparison of the original /
        updated Resources, and merge the changes into the local package.
        See ` + "`" + `kpt help apis merge3` + "`" + ` for details on merge.
      * fast-forward: fail without updating if the local package was modified
        since it was fetched.
      * alpha-git-patch: use 'git format-patch' and 'git am' to apply a
        patch of the changes between the source version and destination
        version.
        REQUIRES THE LOCAL PACKAGE TO HAVE BEEN COMMITTED TO A LOCAL GIT REPO.
      * force-delete-replace: THIS WILL WIPE ALL LOCAL CHANGES TO
        THE PACKAGE.  DELETE the local package at local_pkg_dir/ and replace
        it with the remote version.
`
var SetExamples = `
Create a new package and add a dependency to it:

<!-- @pkgSyncSet @verifyExamples-->
  # init a package so it can be synced
  kpt pkg init .
  
  # add a dependency to the package
  kpt pkg sync set https://github.com/GoogleContainerTools/kpt.git/package-examples/helloworld-set \
      hello-world
  
  # sync the dependencies
  kpt pkg sync .

Update an existing package dependency:

<!-- @pkgSyncSet @verifyStaleExamples-->
  # add a dependency to an existing package
  kpt pkg sync set https://github.com/GoogleContainerTools/kpt.git/package-examples/helloworld-set@v0.2.0 \
      hello-world --strategy=resource-merge
`

var TreeShort = `Render resources using a tree structure`
var TreeLong = `
  kpt pkg tree [DIR] [flags]

Args:

  DIR:
    Path to a package directory.  Defaults to STDIN if not specified.

Flags:

  --args:
    if true, print the container args field
  
  --command:
    if true, print the container command field
  
  --env:
    if true, print the container env field
  
  --field:
    dot-separated path to a field to print
  
  --image:
    if true, print the container image fields
  
  --name:
    if true, print the container name fields
  
  --ports:
    if true, print the container port fields
  
  --replicas:
    if true, print the replica field
  
  --resources:
    if true, print the resource reservations
`
var TreeExamples = `
<!-- @pkgTree @verifyExamples-->
  # print Resources using directory structure
  kpt pkg tree my-dir/

<!-- @pkgTree @verifyExamples-->
  # print replicas, container name, and container image and fields for Resources
  kpt pkg tree my-dir --replicas --image --name

<!-- @pkgTree @verifyExamples-->
  # print all common Resource fields
  kpt pkg tree my-dir/ --all

<!-- @pkgTree @verifyExamples-->
  # print the "foo"" annotation
  kpt pkg tree my-dir/ --field "metadata.annotations.foo"

<!-- @pkgTree @verifyStaleExamples-->
  # print the status of resources with status.condition type of "Completed"
  kubectl get all -o yaml | kpt pkg tree \
    --field="status.conditions[type=Completed].status"

<!-- @pkgTree @verifyStaleExamples-->
  # print live Resources from a cluster using owners for graph structure
  kubectl get all -o yaml | kpt pkg tree --replicas --name --image

<!-- @pkgTree @verifyStaleExamples-->
  # print live Resources with status condition fields
  kubectl get all -o yaml | kpt pkg tree \
    --name --image --replicas \
    --field="status.conditions[type=Completed].status" \
    --field="status.conditions[type=Complete].status" \
    --field="status.conditions[type=Ready].status" \
    --field="status.conditions[type=ContainersReady].status"
`

var UpdateShort = `Apply upstream package updates`
var UpdateLong = `
  kpt pkg update LOCAL_PKG_DIR[@VERSION] [flags]

Args:

  LOCAL_PKG_DIR:
    Local package to update.  Directory must exist and contain a Kptfile
    to be updated.
  
  VERSION:
    A git tag, branch, ref or commit.  Specified after the local_package
    with @ -- pkg@version.
    Defaults the local package version that was last fetched.
  
    Version types:
      * branch: update the local contents to the tip of the remote branch
      * tag: update the local contents to the remote tag
      * commit: update the local contents to the remote commit

Flags:

  --strategy:
    Controls how changes to the local package are handled.  Defaults to fast-forward.
  
      * resource-merge: perform a structural comparison of the original /
        updated Resources, and merge the changes into the local package.
      * fast-forward: fail without updating if the local package was modified
        since it was fetched.
      * alpha-git-patch: use 'git format-patch' and 'git am' to apply a
        patch of the changes between the source version and destination
        version.
      * force-delete-replace: WIPE ALL LOCAL CHANGES TO THE PACKAGE.
        DELETE the local package at local_pkg_dir/ and replace it
        with the remote version.
  
  -r, --repo:
    Git repo url for updating contents.  Defaults to the repo the package
    was fetched from.
  
  --dry-run
    Print the 'alpha-git-patch' strategy patch rather than merging it.

Env Vars:

  KPT_CACHE_DIR:
    Controls where to cache remote packages when fetching them to update
    local packages.
    Defaults to ~/.kpt/repos/
`
var UpdateExamples = `
  # update my-package-dir/
  git add . && git commit -m 'some message'
  kpt pkg update my-package-dir/

  # update my-package-dir/ to match the v1.3 branch or tag
  git add . && git commit -m 'some message'
  kpt pkg update my-package-dir/@v1.3

  # update applying a git patch
  git add . && git commit -m "package updates"
  kpt pkg  update my-package-dir/@master --strategy alpha-git-patch
`
