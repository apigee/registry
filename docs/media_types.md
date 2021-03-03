## Media Types

In the Registry API, specs and artifacts may contain arbitrary data with types
specified in their `mime_type` fields. We recommend that these be valid
[Media Types](https://en.wikipedia.org/wiki/Media_type) and further suggest the
following values, which we use ourselves in our Registry tools.

### API Specs

API specifications currently have little coverage in the
[IANA Media Types](https://www.iana.org/assignments/media-types/media-types.xhtml);
the only general-purpose API description listed there is
[application/vnd.api+json](https://www.iana.org/assignments/media-types/application/vnd.api+json).
We hope to collaborate with others to register official types, particularly for
common API description formats like OpenAPI and Protocol Buffers. Rather than
preemptively (and permanently) register types in the
[Vendor Tree](https://tools.ietf.org/html/rfc6838#section-3.2), we instead use
the [Unregistered x. Tree](https://tools.ietf.org/html/rfc6838#section-3.4)
with at least the following values for API specification types:

- `application/x.openapi` for [OpenAPI](https://www.openapis.org/) descriptions
  with an optional `version` parameter that provides the specification version.
  Values that we accept for `version` include `2`, `2.0`, `3`, `3.X`, and
  `3.X.Y`, where `X` and `Y` are optional integers. `application/x.openapi` can
  also be followed by an optional `+gzip` suffix indicating that the API
  description is compressed with gzip encoding.

- `application/x.discovery` for the
  [Google API Discovery Service Format](https://developers.google.com/discovery).
  This also can be followed by an optional `+gzip` suffix indicating that the
  API description is compressed with gzip encoding.

* `application/x.protobuf` for API descriptions in the
  [Protocol Buffer Language](https://developers.google.com/protocol-buffers).
  Since Protocol Buffer API descriptions are frequently stored in multiple
  files, this is usually followed by a `+zip` suffix indicating that the API
  description is a multifile zip archive.

Until registered media types exist for other API description formats, we
recommend that they also be specified with names in the `application/x.` tree.

### Artifacts

Artifacts allow relatively large and potentially-structured information to be
attached to other resources in the registry model. We recommend but currently
do not require that valid
[Media Types](https://en.wikipedia.org/wiki/Media_type) be used, and we
currently use the following types:

- `text/plain` for strings and string-encoded numbers.

- `application/octet-stream;type=<type>` for general binary data, which we
  typically serialize using Protocol Buffers. In these cases, the `type`
  argument is the fully-qualified name of the Protocol Buffer message type used
  to encode the data. For example,
  `type=google.cloud.apigee.registry.applications.v1alpha1.Lint` describes
  encoded linter output in the
  [Lint message](https://github.com/apigee/registry/blob/7cd962bf79ac51d9d0601aee0f3cce45bc2ec170/google/cloud/apigee/registry/applications/v1alpha1/registry_lint.proto#L27)
  defined in
  [google/cloud/apigee/registry/applications/v1alpha1/registry_lint.proto](https://github.com/apigee/registry/blob/7cd962bf79ac51d9d0601aee0f3cce45bc2ec170/google/cloud/apigee/registry/applications/v1alpha1/registry_lint.proto).
  We may also use the `+gzip` suffix on type arguments to indicate that these
  values are compressed with gzip compression.
