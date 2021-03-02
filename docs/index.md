## The Registry API

We've found that many organizations, including ours, are challenged by the increasing number of APIs that they make and use. The Registry API is an experimental approach to organizing information about APIs. Machine-readable API specs are key, but we also want to track APIs that lack specs, and we also want a way to store and track API-related metadata that doesn't fit well in specs.

The Registry API ([protocol documentation](/registry/api.html)) presents a simple resource hierarchy for tracking API information. All **APIs** are tracked in a container called a **Project**. APIs contain **Versions**, and Versions contain **Specs**. To support this, we use the following convention for naming resources:

```
projects/{project_id}/apis/{api_id}/versions/{version_id}/specs/{spec_id}
```

Specs can be of any format, and spec formats are specified with a `mime_type` field in the Spec record. Metadata can be associated with APIs,
Versions, and Specs using maps of key-value pairs attached to these resources, and larger metadata can be stored in **Artifacts** associated
with Projects, APIs, Versions, and Specs. Artifacts are binary blobs with associated `mime_type` values that can optionally refer to Protocol
Buffer message types. Depending on the entities they are associated with, artifacts can have any of the following resource names:

```
projects/{project_id}/artifacts/{artifact_id}
projects/{project_id}/apis/{api_id}/artifacts/{artifact_id}
projects/{project_id}/apis/{api_id}/versions/{version_id}/artifacts/{artifact_id}
projects/{project_id}/apis/{api_id}/versions/{version_id}/specs/{spec_id}/artifacts/{artifact_id}
```

The Registry API is a [gRPC](https://grpc.io) service that closely follows the guidelines in the Google [API Improvement Proposals](https://aip.dev). This includes following standards for [pagination](https://google.aip.dev/158), [reading across collections](https://google.aip.dev/159), and [filtering](https://google.aip.dev/160) for all collections, optional [partial responses](https://google.aip.dev/157), [resource revisions](https://google.aip.dev/162) for Specs, and support for [generated client libraries](https://google.aip.dev/client-libraries/4210) and [HTTP transcoding](https://aip.dev/127).

For more information, see the [README](https://github.com/apigee/registry/blob/main/README.md) at the root of the project and other README files scattered throughout the repository. For questions or concerns, feel free to contact the repo owners directly or to use the [Issues](https://github.com/apigee/registry/issues) area. Thanks for reading! 
