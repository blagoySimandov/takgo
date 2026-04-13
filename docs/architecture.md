# Architecture

TakGo uses **hexagonal architecture** (ports & adapters). Business logic lives in `internal/domain/` and is completely isolated from infrastructure. Adapters wire infrastructure to domain ports - the domain never imports an adapter.

## Layer map

```
+-----------------------------------------------------+
|                    cmd/                             |
|   server / client / migrate / asyncapi             |
|            (composition roots)                      |
+--------------------+--------------------------------+
                     | wires
+--------------------V--------------------------------+
|              internal/adapters/                     |
|                                                     |
|  http/      Echo + Huma handlers, JWT middleware    |
|  ws/        WsNotifier  -> game.Notifier port        |
|  sqlite/    UserRepo    -> auth.IUserRepository port |
|  jwt/       TokenSvc    -> auth.ITokenService port   |
+--------------------+--------------------------------+
                     | implements
+--------------------V--------------------------------+
|              internal/domain/                       |
|                                                     |
|  auth/        AuthService, User model, ports        |
|  game/        GameService, Hub, Game model, ports   |
|  matchmaking/ Queue (in-memory matchmaker)          |
+-----------------------------------------------------+
```

## Domains

### `auth`

Handles registration and login. Pure bcrypt + UUID logic. No HTTP, no DB.

**Ports defined by the domain:**

| Port | Methods |
|------|---------|
| `IUserRepository` | `Create`, `FindByUsername` |
| `ITokenService` | `Generate`, `Validate` |

**Adapters that implement them:**

| Port | Adapter |
|------|---------|
| `IUserRepository` | `adapters/sqlite.UserRepo` |
| `ITokenService` | `adapters/jwt.TokenService` |

### `game`

Owns tic-tac-toe rules: move validation, win/draw detection, turn rotation. State is persisted via `GameRepository`. After every move the domain calls `Notifier` - it does not know or care that the transport is WebSocket.

**Ports defined by the domain:**

| Port | Methods |
|------|---------|
| `GameRepository` | `Create`, `FindByID`, `Save` |
| `Notifier` | `NotifyMove` |

**Adapters that implement them:**

| Port | Adapter |
|------|---------|
| `GameRepository` | `adapters/sqlite` (TODO - currently Hub holds state) |
| `Notifier` | `adapters/ws.WsNotifier` |

### `matchmaking`

In-memory `Queue` pairs two waiting players into a `game.Game`. No external dependencies.

## Dependency rule

> Nothing inside `internal/domain/` imports anything outside `internal/domain/`.

All infrastructure dependencies flow inward via interfaces. This means you can swap SQLite for Postgres, or WebSocket for SSE, by writing a new adapter and updating the composition root in `cmd/server/main.go`.
