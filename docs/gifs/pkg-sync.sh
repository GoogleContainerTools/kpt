#!/bin/bash
# Copyright 2019 Google LLC
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

########################
# include the magic
########################
. ../../demos/demo-magic/demo-magic.sh

cd $(mktemp -d)
git init

stty rows 80 cols 30

export SRC_REPO=git@github.com:GoogleContainerTools/kpt.git

# start demo
clear
echo "#"
echo "# init a package"
echo "#"
pe "kpt pkg init . --description 'my package'"
echo "$ export SRC_REPO=git@github.com:GoogleContainerTools/kpt.git"
pe "kpt pkg sync set \$SRC_REPO/package-examples/helloworld-set@v0.2.0 hello-world --strategy=resource-merge"

echo "#"
echo "# print the package contents"
echo "#"
pe "kpt config count helloworld # Resource counts"
pe "kpt config tree helloworld --name --image --replicas # Structured output"
pe "kpt config cat helloworld | less # Raw configuration"

pe "clear"