#!/bin/sh

if [ ! -d "api-common-protos" ]
then
  git clone https://github.com/googleapis/api-common-protos
else
  echo "Using previous download of third_party/api-common-protos."
fi