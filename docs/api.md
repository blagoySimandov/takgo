# API Reference

## REST - OpenAPI

The server auto-generates an OpenAPI 3.1 spec via [Huma](https://huma.rocks). Two endpoints are available without any extra setup:

| URL                 | Description             |
| ------------------- | ----------------------- |
| `GET /openapi.json` | Raw OpenAPI spec (JSON) |
| `GET /docs`         | Swagger UI              |

All application routes are prefixed `/api/v1`.

### Auth endpoints

| Method | Path                    | Auth | Description                      |
| ------ | ----------------------- | ---- | -------------------------------- |
| `POST` | `/api/v1/auth/register` | -    | Register a new user, returns JWT |
| `POST` | `/api/v1/auth/login`    | -    | Login, returns JWT               |

All other endpoints require `Authorization: Bearer <token>`.

---

## WebSocket - AsyncAPI

Real-time game communication uses WebSocket. The spec is in [`asyncapi.yaml`](../asyncapi.yaml) at the repo root.

### Viewing the AsyncAPI docs

**Option A - AsyncAPI CLI (recommended)**

```bash
npm install -g @asyncapi/cli
make async-docs
```

Opens a live preview in your browser.

**Option B - AsyncAPI Studio (no install)**

1. Copy the contents of `asyncapi.yaml`
2. Paste into [studio.asyncapi.com](https://studio.asyncapi.com)

**Option C - raw YAML**

```bash
cat asyncapi.yaml
```

### Regenerating the spec

The spec is generated from Go types in `adapters/ws`. After changing message structs, run:

```bash
make asyncapi-generate
```

### WebSocket flow

```
Client                          Server
  |                               |
  |-- GET /api/v1/game/connect -->|  (JWT in header or query param)
  |<----------- 101 Upgrade ------|
  |                               |  server pairs you via matchmaking queue
  |                               |  once matched, game starts
  |-- { "position": 4 } --------->|  WsMoveMsg  (your move)
  |<- { board, state, ... } ------|  WsGameStateMsg (new state)
  |                               |
  |  ... moves continue ...       |
  |                               |
  |<- WsGameStateMsg (finished) --|
  |<- WS Close (game over) -------|
```

### Message schemas

**Client -> Server: `WsMoveMsg`**

```json
{
  "position": 4
}
```

`position` is 0-8, row-major (top-left = 0, bottom-right = 8).

**Server -> Client: `WsGameStateMsg`**

```json
{
  "board": [0, 0, 0, 0, 1, 0, 0, 0, 0],
  "state": 1,
  "currentTurn": "uuid-of-player",
  "winnerId": null,
  "lastMove": { "position": 4 }
}
```

| Field      | Values                                       |
| ---------- | -------------------------------------------- |
| `board[i]` | `0` empty / `1` X / `2` O                    |
| `state`    | `0` waiting / `1` playing / `2` finished     |
| `winnerId` | UUID of winner, `null` on draw or unfinished |

After `state: 2` the server sends a WebSocket close frame.
