#!/bin/sh 
#
# This script runs in a container and starts the registry-server in the
# background before running envoy.
#

# run the authz server on its default port.
/authz-server -c authz.yaml &

# run the registry server on a fixed port.
REGISTRY_SERVER_PORT=8081
PORT=$REGISTRY_SERVER_PORT /registry-server -c registry.yaml &

# update envoy.yaml to look for the registry-server on the port we just set.
sed -i "s/8080/${REGISTRY_SERVER_PORT}/g" /etc/envoy/envoy.yaml

# update envoy.yaml to point to the container-assigned port.
sed -i "s/9999/${PORT}/g" /etc/envoy/envoy.yaml

# run envoy in the foreground.
/usr/local/bin/envoy -c /etc/envoy/envoy.yaml
