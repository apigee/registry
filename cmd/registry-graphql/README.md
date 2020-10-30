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
    id
    display_name
  }
}
```
