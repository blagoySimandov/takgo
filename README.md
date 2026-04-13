# TakGo

Multiplayer tic-tac-toe. Go HTTP/WebSocket server with a terminal UI client.

- **REST API** — [Echo](https://echo.labstack.com) + [Huma](https://huma.rocks), OpenAPI spec and Swagger UI served out of the box
- **Real-time** — WebSocket game channel, [AsyncAPI](https://www.asyncapi.com) spec generated from Go types
- **TUI client** — [Bubbletea](https://github.com/charmbracelet/bubbletea) terminal interface
- **Auth** — JWT (HS256) via [golang-jwt](https://github.com/golang-jwt/jwt)
- **Storage** — [Bun](https://bun.uptrace.dev) ORM + SQLite
- **Architecture** — hexagonal (ports & adapters)

---

## Quick start

**Prerequisites:** Go 1.25+, `gotestdox` (for tests), AsyncAPI CLI (for WS docs preview)

```bash
# 1. Apply migrations
make migrate-up

# 2. Start the server
make run
```

Server listens on `http://localhost:8080`.

Open **`http://localhost:8080/docs`** for the interactive Swagger UI.

---

## Stack

| Concern             | Library                    |
| ------------------- | -------------------------- |
| HTTP routing        | Echo v4 + Huma v2          |
| ORM / migrations    | Bun + SQLite               |
| Auth                | JWT HS256 — golang-jwt/jwt |
| WebSocket           | gorilla/websocket          |
| TUI                 | Bubbletea + Bubbles        |
| AsyncAPI generation | swaggest/go-asyncapi       |
| Testing             | gotestdox                  |

---

## Project structure

```
cmd/          Entrypoints — one binary per concern (server, client, migrate, …)
internal/
  domain/     Business logic and port interfaces — no external imports
  adapters/   Infrastructure implementations (HTTP, WebSocket, DB, JWT, …)
  migrations/ Database migrations
```

Each domain package owns its models, service, and the port interfaces it requires. Adapters implement those interfaces and are wired together in `cmd/`. See [docs/architecture.md](docs/architecture.md) for the full breakdown.

---

## Documentation

| Topic                                             | Doc                                          |
| ------------------------------------------------- | -------------------------------------------- |
| Hexagonal architecture — ports, adapters, domains | [docs/architecture.md](docs/architecture.md) |
| REST + WebSocket API reference                    | [docs/api.md](docs/api.md)                   |

---

## API

### REST — OpenAPI / Swagger

| URL                 | Description          |
| ------------------- | -------------------- |
| `GET /docs`         | Swagger UI           |
| `GET /openapi.json` | Raw OpenAPI 3.1 spec |

All routes are prefixed `/api/v1`. Most require `Authorization: Bearer <jwt>`.

See [docs/api.md](docs/api.md) for the full route table.

### WebSocket — AsyncAPI

Connect to `/api/v1/game/connect` with a valid JWT. The server queues you and starts a game when a second player connects. Send moves as JSON; receive board state updates after every move.

Full message schemas and flow diagram: [docs/api.md#websocket--asyncapi](docs/api.md#websocket--asyncapi)

**Preview the AsyncAPI docs:**

```bash
make async-docs          # requires: npm install -g @asyncapi/cli
```

**Regenerate the spec after changing message types:**

```bash
make asyncapi-generate
```

---

## Architecture

TakGo is structured around hexagonal architecture. Domain packages (`internal/domain/`) contain pure business logic and declare port interfaces. Adapter packages (`internal/adapters/`) implement those interfaces using real infrastructure. Nothing in `domain/` imports outside itself.

```
cmd/  (composition roots)
  └── wires adapters → domain ports

internal/adapters/  (infrastructure)
  └── implements port interfaces

internal/domain/  (business logic + port definitions)
  └── no external imports
```

Full details: [docs/architecture.md](docs/architecture.md)

---

## Commits

Commits follow [Conventional Commits](https://www.conventionalcommits.org) (`feat:`, `fix:`, `chore:`, etc.).

---

## Development

```bash
make build        # compile bin/server and bin/takgo
make run          # run the server
make test         # run all tests (requires gotestdox)
make migrate-up   # apply pending migrations
make migrate-down # roll back last migration
```

Run `make help` for the full target list.

---

## Testing

Integration tests use in-memory SQLite — no mocks or external services.
We also use [gotestdox](https://github.com/gotestdox/gotestdox) for test coverage.
It allows us to have really readable test output.

```bash
make test
```
