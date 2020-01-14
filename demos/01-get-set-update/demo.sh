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
. demo-magic/demo-magic.sh

cd $(mktemp -d)
git init

# hide the evidence
clear

bold=$(tput bold)
normal=$(tput sgr0)
stty rows 50 cols 180

# start demo
clear
echo "# fetch the package..."
pe "kpt pkg get git@github.com:GoogleContainerTools/kpt.git/package-examples/helloworld-set@v0.1.0 helloworld"

echo "# print its contents..."
pe "kpt config tree helloworld --image --ports --name --replicas  --field 'metadata.labels'"

echo "# add to git..."
pe "git add helloworld && git commit -m 'fetch helloworld package at v0.1.0'"

pe "clear"
echo "# print setters..."
pe "kpt config set helloworld"

echo "# change a value..."
pe "kpt config set helloworld replicas 3 --set-by phil --description 'minimal HA mode'"

echo "# print setters again..."
pe "kpt config set helloworld"

echo "# print its contents..."
pe "kpt config tree helloworld --name --replicas"

echo "# view the diff..."
pe "git diff"

echo "# commit changes..."
pe "git add helloworld && git commit -m 'set replicas to 3'"

pe "clear"
echo "# update the package to a new version..."
pe "kpt pkg update helloworld@v0.2.0 --strategy=resource-merge"

echo "# view the diff..."
pe "git diff"

echo "# print its contents..."
pe "kpt config tree helloworld --name --replicas --field 'metadata.labels'"

echo "# update git..."
pe "git add helloworld && git commit -m 'update helloworld package to v0.2.0'"
