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

Access is configured with the `authz.yaml` file. If `trustJWTs` is true, email
addresses in JWT tokens are trusted without verification (only enable this in
environments where tokens are already verified). The `readers` and `writers`
arrays contain glob patterns that, if matched against user emails, allow read
and write access, respectively. "Read" methods correspond to RPCs with names
that begin with "Get" and "List". An optional `tokens` map can be used to map
specified test tokens to user IDs. For example, the following allows the
`test@example.com` user ID to be authenticated with the bearer token
`1234ABCDWXYZ`.

```
tokens:
  1234ABCDWXYZ: test@example.com
```
