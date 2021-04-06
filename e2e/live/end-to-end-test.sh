#!/bin/bash
###########################################################################
# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Description:
#   Validates multiple scenarios for kpt live commands.
#
# How to use this script:
#   FROM KPT ROOT DIR: ./e2e/live/end-to-end-test.sh
#
# Example KPT ROOT DIR:
#   ~/go/src/github.com/GoogleContainerTools/kpt
#
# Prerequisites (must be in $PATH):
#   kind - Kubernetes in Docker
#   kubectl - version of kubectl should be within +/- 1 version of cluster.
#     CHECK: kubectl version
#
###########################################################################

set +e

# Change from empty string to build the kpt binary from the downloaded
# repositories at HEAD, including dependencies cli-utils and kustomize.
BUILD_DEPS_AT_HEAD=""

###########################################################################
#  Setup for test
###########################################################################

# Setup temporary directory for src, bin, and output.
TMP_DIR=$(mktemp -d -t kpt-e2e-XXXXXXXXXX)
SRC_DIR="${TMP_DIR}/src"
mkdir -p $SRC_DIR
BIN_DIR="${TMP_DIR}/bin"
mkdir -p ${BIN_DIR}
OUTPUT_DIR="${TMP_DIR}/output"
mkdir -p $OUTPUT_DIR

# Build the kpt binary and copy it to the temp dir. If BUILD_DEPS_AT_HEAD
# is set, then copy the kpt repository AND dependency directories into
# TMP_DIR and build from there.
echo "kpt end-to-end test"
echo
echo "Temp Dir: $TMP_DIR"
echo

if [ -z $BUILD_DEPS_AT_HEAD ]; then
    echo "Building kpt locally..."
    go build -o $BIN_DIR -v . > $OUTPUT_DIR/kptbuild 2>&1
    echo "Building kpt locally...SUCCESS"

else
    echo "Building kpt using dependencies at HEAD..."
    echo
    # Clone kpt repository into kpt source directory
    KPT_SRC_DIR="${SRC_DIR}/github.com/GoogleContainerTools/kpt"
    mkdir -p $KPT_SRC_DIR
    echo "Downloading kpt repository at HEAD..."
    git clone https://github.com/GoogleContainerTools/kpt ${KPT_SRC_DIR} > ${OUTPUT_DIR}/kptbuild 2>&1
    echo "Downloading kpt repository at HEAD...SUCCESS"
    # Clone cli-utils repository into source directory
    CLI_UTILS_SRC_DIR="${SRC_DIR}/sigs.k8s.io/cli-utils"
    mkdir -p $CLI_UTILS_SRC_DIR
    echo "Downloading cli-utils repository at HEAD..."
    git clone https://github.com/kubernetes-sigs/cli-utils ${CLI_UTILS_SRC_DIR} > ${OUTPUT_DIR}/kptbuild 2>&1
    echo "Downloading cli-utils repository at HEAD...SUCCESS"
    # Clone kustomize respository into source directory
    KUSTOMIZE_SRC_DIR="${SRC_DIR}/sigs.k8s.io/kustomize"
    mkdir -p $KUSTOMIZE_SRC_DIR
    echo "Downloading kustomize repository at HEAD..."
    git clone https://github.com/kubernetes-sigs/kustomize ${KUSTOMIZE_SRC_DIR} > ${OUTPUT_DIR}/kptbuild 2>&1
    echo "Downloading kustomize repository at HEAD...SUCCESS"
    # Tell kpt to build using the locally downloaded dependencies
    echo "Updating kpt/go.mod to reference locally downloaded repositories..."
    echo -e "\n\nreplace sigs.k8s.io/cli-utils => ../../../sigs.k8s.io/cli-utils" >> ${KPT_SRC_DIR}/go.mod
    echo -e "replace sigs.k8s.io/kustomize/kyaml => ../../../sigs.k8s.io/kustomize/kyaml\n" >> ${KPT_SRC_DIR}/go.mod
    echo "Updating kpt/go.mod to reference locally downloaded repositories...SUCCESS"
    # Build kpt using the cloned directories
    export GOPATH=${TMP_DIR}
    echo "Building kpt..."
    (cd -- ${KPT_SRC_DIR} && go build -o $BIN_DIR -v . > ${OUTPUT_DIR}/kptbuild 2>&1)
    echo "Building kpt...SUCCESS"
    echo
    echo "Building kpt using dependencies at HEAD...SUCCESS"
fi

echo

###########################################################################
#  Helper functions
###########################################################################

# createTestSuite deletes then creates the kind cluster.
function createTestSuite {
    echo "Setting Up Test Suite..."
    echo
    # Create the k8s cluster
    echo "Deleting kind cluster..."
    kind delete cluster > /dev/null 2>&1
    echo "Deleting kind cluster...COMPLETED"
    echo "Creating kind cluster..."
    kind create cluster > /dev/null 2>&1    
    echo "Creating kind cluster...COMPLETED"
    echo
    echo "Setting Up Test Suite...COMPLETED"
    echo
}

function waitForDefaultServiceAccount {
    # Necessary to ensure default service account is created before pods.
    echo -n "Waiting for default service account..."
    echo -n ' '
    sp="/-\|"
    n=1
    until ((n >= 300)); do
	kubectl -n default get serviceaccount default -o name > $OUTPUT_DIR/status 2>&1
	test 1 == $(grep "serviceaccount/default" $OUTPUT_DIR/status | wc -l)
	if [ $? == 0 ]; then
	    echo
	    break
	fi
	printf "\b${sp:n++%${#sp}:1}"
	sleep 0.2
    done
    ((n < 300))
    echo "Waiting for default service account...CREATED"
    echo
}

# assertContains checks that the passed string is a substring of
# the $OUTPUT_DIR/status file.
ERROR=""
function assertContains {
  test 1 == \
  $(grep "$@" $OUTPUT_DIR/status | wc -l); \
  if [ $? == 0 ]; then
      echo -n '.'
  else
      echo -n 'E'
      ERROR+="ERROR: assertContains $@, but missing\n"
  fi
}

# assertCMInventory checks that a ConfigMap inventory object exists in
# the passed namespace with the passed number of inventory items.
# Assumes the inventory object name begins with "inventory-".
function assertCMInventory {
    local ns=$1
    local numInv=$2
    
    inv=$(kubectl get cm -n $ns --selector='cli-utils.sigs.k8s.io/inventory-id' --no-headers)
    echo $inv | awk '{print $1}' > $OUTPUT_DIR/invname
    echo $inv | awk '{print $2}' > $OUTPUT_DIR/numinv

    test 1 == $(grep "inventory-" $OUTPUT_DIR/invname | wc -l);
    if [ $? == 0 ]; then
	echo -n '.'
    else
	echo -n 'E'
	ERROR+="ERROR: expected ConfigMap inventory to exist\n"
    fi

    test 1 == $(grep $numInv $OUTPUT_DIR/numinv | wc -l);
    if [ $? == 0 ]; then
	echo -n '.'
    else
	echo -n 'E'
	ERROR+="ERROR: expected ConfigMap inventory to have $numInv inventory items\n"
    fi
}

# assertRGInventory checks that a ResourceGroup inventory object exists
# in the passed namespace. Assumes the inventory object name begins
# with "inventory-".
function assertRGInventory {
    local ns=$1
    
    kubectl get resourcegroups.kpt.dev -n $ns --selector='cli-utils.sigs.k8s.io/inventory-id' --no-headers | awk '{print $1}' > $OUTPUT_DIR/invname

    test 1 == $(grep "inventory-" $OUTPUT_DIR/invname | wc -l);
    if [ $? == 0 ]; then
	echo -n '.'
    else
	echo -n 'E'
    fi
}

# assertPodExists checks that a pod with the passed podName and passed
# namespace exists in the cluster.
TIMEOUT_SECS=30
function assertPodExists {
    local podName=$1
    local namespace=$2

    kubectl wait --for=condition=Ready -n $namespace pod/$podName --timeout=${TIMEOUT_SECS}s > /dev/null 2>&1
    kubectl get po -n $namespace $podName -o name | awk '{print $1}' > $OUTPUT_DIR/podname

    test 1 == $(grep $podName $OUTPUT_DIR/podname | wc -l);
    if [ $? == 0 ]; then
	echo -n '.'
    else
	echo -n 'E'
	ERROR+="ERROR: expected pod $namespace/$podName to exist\n"
    fi    
}

# assertPodNotExists checks that a pod with the passed podName and passed
# namespace DOES NOT exist in the cluster. Waits 20 seconds for pod
# termination if pod has not finished deleting.
function assertPodNotExists {
    local podName=$1
    local namespace=$2

    kubectl wait --for=delete -n $namespace pod/$podName --timeout=${TIMEOUT_SECS}s > /dev/null 2>&1
    kubectl get po -n $namespace $podName -o name > $OUTPUT_DIR/podname 2>&1
    
    test 1 == $(grep "(NotFound)" $OUTPUT_DIR/podname | wc -l);
    if [ $? == 0 ]; then
	echo -n '.'
    else
	echo -n 'E'
	ERROR+="ERROR: expected pod $namespace/$podName to not exist\n"
    fi    
}

# printResult prints the results of the previous assert statements
function printResult {
    if [ -z $ERROR ]; then
	echo "SUCCESS"
    else
	echo "ERROR"
    fi
    echo
    ERROR=""
}

# wait sleeps for the passed number of seconds.
function wait {
    local numSecs=$1

    sleep $numSecs
}

###########################################################################
#  Run tests
###########################################################################

unset RESOURCE_GROUP_INVENTORY

createTestSuite
waitForDefaultServiceAccount

# Test 1: Basic ConfigMap init
# Creates ConfigMap inventory-template.yaml in "test-case-1a" directory
echo "Testing basic ConfigMap init"
echo "kpt live init e2e/live/testdata/test-case-1a"
${BIN_DIR}/kpt live init e2e/live/testdata/test-case-1a > $OUTPUT_DIR/status 2>&1
assertContains "namespace: test-namespace is used for inventory object"
assertContains "testdata/test-case-1a/inventory-template.yaml"
printResult

# Copy the ConfigMap inventory template to the test-case-1b directory.
cp -f e2e/live/testdata/test-case-1a/inventory-template.yaml e2e/live/testdata/test-case-1b

# Test 2: Basic kpt live preview
# Preview run for "test-case-1a" directory
echo "Testing initial preview"
echo "kpt live preview e2e/live/testdata/test-case-1a"
${BIN_DIR}/kpt live preview e2e/live/testdata/test-case-1a > $OUTPUT_DIR/status
assertContains "namespace/test-namespace created (preview)"
assertContains "pod/pod-a created (preview)"
assertContains "pod/pod-b created (preview)"
assertContains "pod/pod-c created (preview)"
assertContains "4 resource(s) applied. 4 created, 0 unchanged, 0 configured"
assertContains "0 resource(s) pruned, 0 skipped"
printResult

# Test 3: Basic kpt live apply
# Apply run for "test-case-1a" directory
echo "Testing basic apply"
echo "kpt live apply e2e/live/testdata/test-case-1a"
${BIN_DIR}/kpt live apply e2e/live/testdata/test-case-1a > $OUTPUT_DIR/status
assertContains "namespace/test-namespace"
assertContains "pod/pod-a created"
assertContains "pod/pod-b created"
assertContains "pod/pod-c created"
assertContains "4 resource(s) applied. 3 created, 1 unchanged, 0 configured"
assertContains "0 resource(s) pruned, 0 skipped"
wait 2
# Validate resources in the cluster
# ConfigMap inventory with four inventory items.
assertCMInventory "test-namespace" "4"
printResult

# Test 4: kpt live preview of apply/prune
# "test-case-1b" directory is "test-case-1a" directory with "pod-a" removed and "pod-d" added.
echo "Testing basic preview"
echo "kpt live preview e2e/live/testdata/test-case-1b"
${BIN_DIR}/kpt live preview e2e/live/testdata/test-case-1b > $OUTPUT_DIR/status
assertContains "namespace/test-namespace configured (preview)"
assertContains "pod/pod-b configured (preview)"
assertContains "pod/pod-c configured (preview)"
assertContains "pod/pod-d created (preview)"
assertContains "4 resource(s) applied. 1 created, 0 unchanged, 3 configured (preview)"
assertContains "pod/pod-a pruned (preview)"
assertContains "1 resource(s) pruned, 0 skipped (preview)"
wait 2
# Validate resources in the cluster
# ConfigMap inventory with four inventory items.
assertCMInventory "test-namespace" "4"
assertPodExists "pod-a" "test-namespace"
assertPodExists "pod-b" "test-namespace"
assertPodExists "pod-c" "test-namespace"
printResult

# Test 5: Basic kpt live apply/prune
# "test-case-1b" directory is "test-case-1a" directory with "pod-a" removed and "pod-d" added.
echo "Testing basic prune"
echo "kpt live apply e2e/live/testdata/test-case-1b"
${BIN_DIR}/kpt live apply e2e/live/testdata/test-case-1b > $OUTPUT_DIR/status
assertContains "namespace/test-namespace unchanged"
assertContains "pod/pod-b unchanged"
assertContains "pod/pod-c unchanged"
assertContains "pod/pod-d created"
assertContains "4 resource(s) applied. 1 created, 3 unchanged, 0 configured"
assertContains "pod/pod-a pruned"
assertContains "1 resource(s) pruned, 0 skipped"
wait 2
# Validate resources in the cluster
# ConfigMap inventory with four inventory items.
assertCMInventory "test-namespace" "4"
assertPodExists "pod-b" "test-namespace"
assertPodExists "pod-c" "test-namespace"
assertPodExists "pod-d" "test-namespace"
assertPodNotExists "pod-a" "test-namespace"
printResult

# Test 6: Basic kpt live destroy
# "test-case-1b" directory is "test-case-1a" directory with "pod-a" removed and "pod-d" added.
echo "Testing basic destroy"
echo "kpt live destroy e2e/live/testdata/test-case-1b"
${BIN_DIR}/kpt live destroy e2e/live/testdata/test-case-1b > $OUTPUT_DIR/status
assertContains "pod/pod-d deleted"
assertContains "pod/pod-c deleted"
assertContains "pod/pod-b deleted"
assertContains "namespace/test-namespace deleted"
assertContains "4 resource(s) deleted, 0 skipped"
# Validate resources NOT in the cluster
assertPodNotExists "pod-b" "test-namespace"
assertPodNotExists "pod-c" "test-namespace"
assertPodNotExists "pod-d" "test-namespace"
printResult

# Creates new inventory-template.yaml for "migrate-case-1a" directory.
echo "kpt live init e2e/live/testdata/migrate-case-1a"
rm -f e2e/live/testdata/migrate-case-1a/inventory-template.yaml
${BIN_DIR}/kpt live init e2e/live/testdata/migrate-case-1a > $OUTPUT_DIR/status
assertContains "namespace: test-rg-namespace is used for inventory object"
assertContains "live/testdata/migrate-case-1a/inventory-template.yaml"
printResult


###########################################################################
#  Tests with RESOURCE_GROUP_INVENTORY env var set
###########################################################################

export RESOURCE_GROUP_INVENTORY=1

# Test 7: kpt live apply ConfigMap inventory with RESOURCE_GROUP_INVENTORY set
# Applies resources in "migrate-case-1a" directory.
echo "Testing kpt live apply with ConfigMap inventory"
echo "kpt live apply e2e/live/testdata/migrate-case-1a"
# Copy Kptfile into "migrate-case-1a" WITHOUT inventory information. This ensures
# the apply uses the ConfigMap inventory-template.yaml during the apply.
cp -f e2e/live/testdata/Kptfile e2e/live/testdata/migrate-case-1a
${BIN_DIR}/kpt live apply e2e/live/testdata/migrate-case-1a > $OUTPUT_DIR/status
assertContains "namespace/test-rg-namespace unchanged"
assertContains "pod/pod-a created"
assertContains "pod/pod-b created"
assertContains "pod/pod-c created"
assertContains "4 resource(s) applied. 3 created, 1 unchanged, 0 configured"
assertContains "0 resource(s) pruned, 0 skipped"
# Validate resources in the cluster
assertCMInventory "test-rg-namespace" "4"
assertPodExists "pod-a" "test-rg-namespace"
assertPodExists "pod-b" "test-rg-namespace"
assertPodExists "pod-c" "test-rg-namespace"
printResult

# Test 8: kpt live migrate from ConfigMap to ResourceGroup inventory
# Migrates resources in "migrate-case-1a" directory.
echo "Testing migrate from ConfigMap to ResourceGroup inventory"
echo "kpt live migrate e2e/live/testdata/migrate-case-1a"
${BIN_DIR}/kpt live migrate e2e/live/testdata/migrate-case-1a > $OUTPUT_DIR/status
assertContains "ensuring ResourceGroup CRD exists in cluster...success"
assertContains "updating Kptfile inventory values...success"
assertContains "retrieve the current ConfigMap inventory...success (4 inventory objects)"
assertContains "migrate inventory to ResourceGroup...success"
assertContains "deleting old ConfigMap inventory object...success"
assertContains "deleting inventory template file"
assertContains "inventory migration...success"
# Validate resources in the cluster
assertPodExists "pod-a" "test-rg-namespace"
assertPodExists "pod-b" "test-rg-namespace"
assertPodExists "pod-c" "test-rg-namespace"
assertRGInventory "test-rg-namespace"
# Run it again, and validate the output
${BIN_DIR}/kpt live migrate e2e/live/testdata/migrate-case-1a > $OUTPUT_DIR/status
assertContains "ensuring ResourceGroup CRD exists in cluster...already installed...success"
assertContains "updating Kptfile inventory values...values already exist...success"
assertContains "retrieve the current ConfigMap inventory...no ConfigMap inventory...completed"
assertContains "inventory migration...success"
printResult

# Test 9: kpt live preview with ResourceGroup inventory
# Previews resources in the "migrate-case-1a" directory.
echo "Testing kpt live preview with ResourceGroup inventory"
echo "kpt live preview e2e/live/testdata/migrate-case-1a"
${BIN_DIR}/kpt live preview e2e/live/testdata/migrate-case-1a > $OUTPUT_DIR/status
assertContains "namespace/test-rg-namespace configured (preview)"
assertContains "pod/pod-a configured (preview)"
assertContains "pod/pod-b configured (preview)"
assertContains "pod/pod-c configured (preview)"
assertContains "4 resource(s) applied. 0 created, 0 unchanged, 4 configured (preview)"
assertContains "0 resource(s) pruned, 0 skipped (preview)"
# Validate resources in the cluster
assertRGInventory "test-rg-namespace"
assertPodExists "pod-a" "test-rg-namespace"
assertPodExists "pod-b" "test-rg-namespace"
assertPodExists "pod-c" "test-rg-namespace"
printResult

# Test 10: kpt live apply/prune with ResourceGroup inventory
# "migrate-case-1b" directory is the same as "migrate-case-1a" with "pod-a" missing, and "pod-d" added.
echo "Testing kpt live apply/prune with ResourceGroup inventory"
echo "kpt live apply e2e/live/testdata/migrate-case-1b"
cp -f e2e/live/testdata/migrate-case-1a/Kptfile e2e/live/testdata/migrate-case-1b
${BIN_DIR}/kpt live apply e2e/live/testdata/migrate-case-1b > $OUTPUT_DIR/status
assertContains "namespace/test-rg-namespace unchanged"
assertContains "pod/pod-a pruned"
assertContains "pod/pod-b unchanged"
assertContains "pod/pod-c unchanged"
assertContains "pod/pod-d created"
assertContains "4 resource(s) applied. 1 created, 3 unchanged, 0 configured"
assertContains "1 resource(s) pruned, 0 skipped"
# Validate resources in the cluster
assertRGInventory "test-rg-namespace"
assertPodExists "pod-b" "test-rg-namespace"
assertPodExists "pod-c" "test-rg-namespace"
assertPodExists "pod-d" "test-rg-namespace"
assertPodNotExists "pod-a" "test-rg-namespace"
printResult

# Test 11: kpt live destroy with ResourceGroup inventory
echo "Testing kpt destroy with ResourceGroup inventory"
echo "kpt live destroy e2e/live/testdata/migrate-case-1b"
${BIN_DIR}/kpt live destroy e2e/live/testdata/migrate-case-1b > $OUTPUT_DIR/status
assertContains "pod/pod-d deleted"
assertContains "pod/pod-c deleted"
assertContains "pod/pod-b deleted"
assertContains "namespace/test-rg-namespace deleted"
assertContains "4 resource(s) deleted, 0 skipped"
assertPodNotExists "pod-b" "test-rg-namespace"
assertPodNotExists "pod-c" "test-rg-namespace"
assertPodNotExists "pod-d" "test-rg-namespace"
printResult

# Test 12: kpt live init for Kptfile (ResourceGroup inventory)
# initial Kptfile does NOT have inventory info
cp -f e2e/live/testdata/Kptfile e2e/live/testdata/migrate-error
echo "Testing kpt live init for Kptfile (ResourceGroup inventory)"
echo "kpt live init e2e/live/testdata/migrate-error"
${BIN_DIR}/kpt live init e2e/live/testdata/migrate-error > $OUTPUT_DIR/status 2>&1
# Difference in Kptfile should have inventory data
diff e2e/live/testdata/Kptfile e2e/live/testdata/migrate-error/Kptfile > $OUTPUT_DIR/status 2>&1
assertContains "inventory:"
assertContains "namespace: test-rg-namespace"
assertContains "name: inventory-"
assertContains "inventoryID:"
printResult

# Test 14: kpt live migrate with no objects in cluster
# Add inventory-template.yaml to "migrate-error", but there are no objects in cluster.
cp -f e2e/live/testdata/inventory-template.yaml e2e/live/testdata/migrate-error
echo "Testing kpt live migrate with no objects in cluster"
echo "kpt live migrate e2e/live/testdata/migrate-error"
${BIN_DIR}/kpt live migrate e2e/live/testdata/migrate-error > $OUTPUT_DIR/status 2>&1
assertContains "ensuring ResourceGroup CRD exists in cluster...already installed...success"
assertContains "updating Kptfile inventory values...values already exist...success"
assertContains "retrieve the current ConfigMap inventory...success (0 inventory objects)"
assertContains "deleting inventory template file:"
assertContains "e2e/live/testdata/migrate-error/inventory-template.yaml...success"
assertContains "inventory migration...success"
printResult

# Test 15: kpt live initial apply ResourceGroup inventory
echo "Testing kpt apply ResourceGroup inventory"
echo "kpt live apply e2e/live/testdata/migrate-error"
${BIN_DIR}/kpt live apply e2e/live/testdata/migrate-error > $OUTPUT_DIR/status
assertContains "pod/pod-a created"
assertContains "pod/pod-b created"
assertContains "pod/pod-c created"
assertContains "0 resource(s) pruned, 0 skipped"
# Validate resources in the cluster
assertPodExists "pod-a" "test-rg-namespace"
assertPodExists "pod-b" "test-rg-namespace"
assertPodExists "pod-c" "test-rg-namespace"
printResult

echo

# Cleanup
cp -f e2e/live/testdata/Kptfile e2e/live/testdata/migrate-case-1a
cp -f e2e/live/testdata/Kptfile e2e/live/testdata/migrate-case-1b
cp -f e2e/live/testdata/Kptfile e2e/live/testdata/migrate-error


###########################################################################
#  Tests for ResourceGroup CRD installation
###########################################################################

# Delete/Create cluster for new test suite
createTestSuite

export RESOURCE_GROUP_INVENTORY=1

echo "Testing kpt live install-resource-group"
echo "kpt live install-resource-group"
# First, check that the ResourceGroup CRD does NOT exist
kubectl get resourcegroups.kpt.dev > $OUTPUT_DIR/status 2>&1
assertContains "error: the server doesn't have a resource type \"resourcegroups\""
# Next, add the ResourceGroup CRD
${BIN_DIR}/kpt live install-resource-group > $OUTPUT_DIR/status
assertContains "installing ResourceGroup custom resource definition...success"
kubectl get resourcegroups.kpt.dev > $OUTPUT_DIR/status 2>&1
assertContains "No resources found"
# Add a simple ResourceGroup custom resource, and verify it exists in the cluster.
kubectl apply -f e2e/live/testdata/install-rg-crd/example-resource-group.yaml > $OUTPUT_DIR/status
assertContains "resourcegroup.kpt.dev/example-inventory created"
kubectl get resourcegroups.kpt.dev --no-headers > $OUTPUT_DIR/status
assertContains "example-inventory"
# Finally, add the ResourceGroup CRD again, and check it says it already exists.
${BIN_DIR}/kpt live install-resource-group > $OUTPUT_DIR/status 2>&1
assertContains "...already installed...success"
printResult

# Clean-up the k8s cluster
echo "Cleaning up cluster"
kind delete cluster
echo "FINISHED"

