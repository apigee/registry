# Registry GraphQL

## A GraphQL wrapper for the Registry API.

This project contains a small proxy for the Registry API that provides a
GraphQL interface.

It uses the [graphql-go](https://github.com/graphql-go/graphql) package.

## Credits

Contents of the `static` directory are manually vendored from
[github.com/graphql/graphiql](https://github.com/graphql/graphiql).

## Invocation

Just run the `registry-graphql` program. It currently takes no options and uses
the `APG_*` environment variables to connect to a Registry API server. Because
it serves static files, it should be run in the same directory as its source
files.

## Usage

After you've started the `registry-graphql` server, visit http://localhost:8088
to open the GraphiQL browser. Then use standard GraphQL to explore the schema
and make queries. For example, to see a list of projects, enter:

```
{
  projects {
    values {
      id
      display_name
    }
  }
}
```

## Pagination

List results are paginated. As an example, here is a fully-specified request
for a page of projects:

```
{
  projects (page_size: 2, page_token:"<previously-returned-page-token>") {
    values {
      id
    }
    next_page_token
  }
}
```

## Schema

[registry.graphql](registry.graphql) is an SDL schema that was produced with
[prisma-labs/get-graphql-schema](https://github.com/prisma-labs/get-graphql-schema).

```
$ get-graphql-schema http://localhost:8088/graphql > registry.graphql
```
