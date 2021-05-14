#!/bin/sh
#
# Copyright 2020 Google LLC. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#
# This script is used in docker builds to download platform-appropriate builds of protoc.
#
case "$(arch)" in
  "x86_64") export ARCH="x86_64"
  ;;
  "aarch64") export ARCH="aarch_64"
  ;;
  "arm64") export ARCH="aarch_64"
  ;;
esac

export VERSION="3.17.0"

export SOURCE="https://github.com/protocolbuffers/protobuf/releases/download/v$VERSION/protoc-$VERSION-linux-$ARCH.zip"

echo $SOURCE

curl -L $SOURCE > /tmp/protoc.zip

unzip /tmp/protoc.zip -d /usr/local