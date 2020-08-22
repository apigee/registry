#!/bin/sh

# curl -L https://getenvoy.io/cli | bash -s -- -b /usr/local/bin 

getenvoy run standard:1.11.2 -- -c envoy.yaml
