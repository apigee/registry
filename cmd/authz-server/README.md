# authz-server

This directory contains a lightweight server that can be used as an
[Envoy authorization filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter)
to provide simple access controls for the Registry API. Written as a gRPC
service, the filter expects API calls to have an authorization header that
passes a Google access or identity token as a Bearer token.

- Google access tokens are OAuth2 tokens that can be obtained from Google
  signin or by running
  `gcloud auth print-identity-token ${APG_REGISTRY_CLIENT_EMAIL}`. Access
  tokens are verified by calling the
  https://www.googleapis.com/oauth2/v1/userinfo API with the provided token.
- Google identity tokens are JWTs that can be obtained by running
  `gcloud auth print-identity-token ${APG_REGISTRY_CLIENT_EMAIL}`. Identity
  tokens are verified by calling the https://oauth2.googleapis.com/tokeninfo
  API.

In either case, verified tokens are cached in-memory for the lifetime of the
`authz-server` process.

Read-only access is allowed for all authenticated users, and write access is
allowed for users on a hard-coded list of approved users (set this before
deploying).
