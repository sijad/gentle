<p align="center">🚧🚧 this is pre-alpha software 🚧🚧</p>

<h1 align="center">Gentle</h1>

<p align="center">
  <strong>Fast</strong>, <strong>simple</strong>, <strong>type-safe</strong>, <strong>code-first</strong> <strong>GraphQL framework</strong>.
</p>

### Quick Start

it's recommended to install Gentle using [Go modules](https://github.com/golang/go/wiki/Modules#quick-start).

1. `go get github.com/sijad/gentle/cmd/gentc` to install Gentle
2. `go run github.com/sijad/gentle/cmd/gentc init` to initialize a GraphQL project
3. Make changes to GraphQL schema at `./graph/schema/`
4. `go generate graph/generate.go` to generate Graph codes
5. Run `go run server.go`

### Credits

inspired by [Nexus](https://www.nexusjs.org/), [PostGraphile](https://www.graphile.org/postgraphile/) and [gqlgen](https://gqlgen.com/)
