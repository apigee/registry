#!/bin/sh 

PORT=8081 /registry-server &

sed -i "s/8080/${PORT}/g" /etc/envoy/envoy.yaml
/usr/local/bin/envoy -c /etc/envoy/envoy.yaml
