## Architecture

Hexagonal architecture. Domain defines ports, adapters implement them. Nothing in `domain/` imports from outside itself.

## Migrations

Migrations are done via [bun](https://github.com/uptrace/bun) and [bun/migrate](https://github.com/uptrace/bun/tree/master/migrate).

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
```

## Running

```bash
go run ./cmd/migrate up
go run ./cmd/server
go run ./cmd/client   # in two separate terminals
```
