# http-client

This directory contains an example that makes a single "raw" HTTP call to the
Registry API using [gRPC JSON Transcoding](https://google.aip.dev/127).

It requires that calls be made to a proxy that supports transcoding. Typically
this is
[Envoy](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/grpc_json_transcoder_filter).
