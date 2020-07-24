#!/bin/sh 
#
# This script runs in a container and starts the registry-server in the
# background before running envoy.
#

# run the registry server on a fixed port.
PORT=8081 /registry-server &

# update envoy.yaml to point to the registry-server port.
sed -i "s/8080/${PORT}/g" /etc/envoy/envoy.yaml

# run envoy in the foreground.
/usr/local/bin/envoy -c /etc/envoy/envoy.yaml