# Registry GraphQL

## A GraphQL wrapper for the Registry API.

This project contains a small proxy for the Registry API that provides a
GraphQL interface.

It uses the [graphql-go](https://github.com/graphql-go/graphql) package.

## Limitations

Currently the GraphQL interface is read-only (no mutations) and does not
provide access to spec revisions or property values.

## Invocation

Just run the `registry-graphql` program. It currently takes no options and uses
the `APG_*` environment variables to connect to a Registry API server. Because
it serves static files, it should be run in the same directory as its source
files.

## Usage

After you've started the `registry-graphql` server, visit http://localhost:8088
to open the GraphiQL browser. Then use standard GraphQL to explore the schema
and make queries. For example, to lookup a project by name, enter:

```
{
  project (id: "projects/test") {
    id
    display_name
  }
}
```

## Pagination

List results are paginated. As an example, here is a fully-specified request
for a page of projects:

```
{
  projects (first: 2, after:$cursor) {
    edges {
      node {
        id
      }
    }
    pageInfo {
      endCursor
    }
  }
}
```

## Schema

[registry.graphql](registry.graphql) is an SDL schema that was produced with
[prisma-labs/get-graphql-schema](https://github.com/prisma-labs/get-graphql-schema).

```
$ get-graphql-schema http://localhost:8088/graphql > registry.graphql
```

## Credits

Contents of the `static` directory are manually vendored from
[github.com/graphql/graphiql](https://github.com/graphql/graphiql).
