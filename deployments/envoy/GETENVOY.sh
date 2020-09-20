#!/bin/sh

# curl -L https://getenvoy.io/cli | bash -s -- -b /usr/local/bin 

getenvoy run standard:1.14.4 -- -c envoy.yaml
