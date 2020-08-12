# gapic-insecure

This directory contains an alternate implementation of the
automatically-generated `doc.go` produced by `gapic-go-generator`. This
modified implementation directly adds auth tokens to gRPC API calls and is only
needed when using GAPIC libraries to access insecure (non-HTTPS) services that
expect authentication. Typically this is only needed during testing of
configurations that require auth tokens, such as when testing locally-running
Envoy setups that use `authz-server`. It is needed because the Go gRPC
implementation removes credentials from calls made to insecure services as a
safety precaution.

This code should not be used in production.
