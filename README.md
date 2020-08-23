<p align="center">🚧🚧 this is WIP software 🚧🚧</p>

<h1 align="center">Gentle</h1>

<p align="center">
  <strong>Fast</strong>, <strong>simple</strong>, <strong>type-safe</strong>, <strong>code-first</strong> <strong>GraphQL framework</strong>.
</p>

### Quick Start

it's recommended to installation Gentle using [Go modules](https://github.com/golang/go/wiki/Modules#quick-start).

#### Install Gentle

Run `go get github.com/sijad/gentle/cmd/gentc` to install Gentle

#### Create a GraphQL project

Run `go run github.com/sijad/gentle/cmd/gentc init` to initialize a GraphQL project

#### Change Schema

Make changes to GraphQL schema at `./graph/schema/`

#### Generate GraphQL codes

Run `go generate graph/generate.go`

#### Run Server

Run `go run server.go`

### Credits

inspired by [Nexus](https://www.nexusjs.org/), [PostGraphile](https://www.graphile.org/postgraphile/) and [gqlgen](https://gqlgen.com/)
