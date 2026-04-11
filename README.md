# takgo

Multiplayer tic-tac-toe with a TUI client and a Go HTTP/WebSocket server.

## Stack

| Concern          | Library                                                         |
| ---------------- | --------------------------------------------------------------- |
| HTTP / routing   | [Echo](https://echo.labstack.com) + [Huma](https://huma.rocks)  |
| ORM / migrations | [Bun](https://bun.uptrace.dev) + SQLite                         |
| Auth             | JWT (HS256) via [golang-jwt](https://github.com/golang-jwt/jwt) |
| TUI              | _(in progress)_                                                 |

## Architecture

Hexagonal. `internal/domain/` holds pure business logic and defines port interfaces. Adapters in `internal/adapters/` implement them. Nothing in `domain/` imports outside itself.

```
cmd/
  server/       HTTP server entrypoint
  migrate/      Migration CLI (up / down)
  client/       TUI client entrypoint (WIP)
internal/
  domain/auth/  User model, AuthService, port interfaces
  adapters/
    http/       Echo + Huma handlers, JWT middleware
    jwt/        TokenService implementation
    sqlite/     UserRepo + Bun DB setup
  migrations/   Bun migration files
```

## API

OpenAPI spec and Swagger UI served automatically at `/openapi.json` and `/docs`. All routes prefixed `/api/v1`.

## Running locally

```bash
go run ./cmd/migrate up
go run ./cmd/server
```

## Testing

Integration tests use in-memory SQLite — no mocks.

```bash
go test -v -json ./... | gotestdox
```
