---
name: go-folder-structure
description: Folder and package structure for all our Go projects. Defines where code goes: domain packages under internal/<domain> with model.go, an optional <domain>.go for errors and constants, and a service/ subfolder (package <domain>svc) with one service file that owns its SQL; transport (HTTP handlers), infrastructure, config, migration, utility, and cmd wiring. No repository layer. Use this skill whenever creating or moving packages, files, or folders, adding a new domain, adding a service, handler, query, migration, worker, or integration, deciding where code belongs, or splitting a large file. Coding style is in go-code-standard; platform APIs are in the platform skill.
---

# Go Folder Structure

Where code goes in our Go projects. Companion skills: **go-code-standard** (how code is written) and the **platform skill** (`platform-be`: the reusable technical infrastructure).

This structure **replaces** the DDD and Clean Architecture layout suggested in the platform README (domain, application, infrastructure, interfaces layers with repository interfaces). The platform library is still the foundation; this skill only governs our application code.

## Principles

1. **Organize by business domain**, not by technical layer. A developer looking for "client" behavior opens `internal/client/`.
2. **Platform first.** The platform already provides the database, transactions, router, responses, errors, validation, pagination metadata, logging, auth, workers, and lifecycle. Do not rebuild them inside an application.
3. **No repository layer.** A domain's service owns its business rules **and** its SQL, using the platform database type directly.
4. **One cohesive concept per file.** Split because of complexity, not because of function count.
5. **Create only what the project needs.** Do not create empty folders, empty files, or layers "for later".
6. **Business code never depends on transport.** HTTP is a delivery mechanism, not part of the domain.

## Default structure

```text
project/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── auth.go
│   │   ├── model.go
│   │   └── service/
│   │       └── auth_service.go
│   ├── client/
│   │   ├── client.go
│   │   ├── model.go
│   │   └── service/
│   │       └── client_service.go
│   └── user/
│       ├── user.go
│       ├── model.go
│       └── service/
│           └── user_service.go
├── transport/
│   └── http/
│       ├── server.go
│       ├── routes.go
│       ├── middleware.go
│       └── handler/
│           ├── client.go
│           └── user.go
├── infrastructure/
│   └── ...
├── migration/
│   ├── 001_create_users.sql
│   └── 002_create_clients.sql
├── config/
│   └── config.go
├── utility/
├── go.mod
├── go.sum
└── README.md
```

This is a starting point, not a mandatory tree. Only create what the project needs (for example no `infrastructure/` until there is a third-party integration, no `utility/` until there is a real helper).

## Folder by folder

### `cmd/<binary>/`

Executable entry points (`cmd/app/main.go`, or one folder per binary). `main.go` only: loads configuration, creates platform dependencies (logger, database, router, server), constructs services and handlers, registers routes, starts the lifecycle runner, and handles shutdown. No business logic. If wiring grows long, move it to a `wire.go` file in the same `main` package. See `references/boundaries-and-wiring.md`.

### `internal/<domain>/`

One folder per business domain, owning its data contracts and rules.

| File | Holds |
|---|---|
| `model.go` | Entities, commands, queries, search/result types, enums, and their `Validate()` methods |
| `<domain>.go` | The domain's named errors and constants (and nothing else). Create only when it has content. |
| `service/<domain>_service.go` | The domain's service: business rules, SQL queries, transactions |

The package name is the domain (`package client`). Details and a full example are in `references/domain-package.md`.

### `internal/<domain>/service/`

The directory is `service`, the package name is `<domain>svc` (for example `package clientsvc`). Import it with an explicit alias:

```go
import (
	"github.com/your-org/project/internal/client"
	clientsvc "github.com/your-org/project/internal/client/service"
)
```

- **One service file per domain** (`client_service.go`) holding the service struct, constructor, SQL query constants, and all related operations. CRUD and closely related operations stay together.
- The service depends on the platform database type, not on a repository.
- Split only when complexity demands it (see "When to split" below).

### `transport/http/`

HTTP-specific code: `server.go` (builds the platform router and server), `routes.go` (registers routes and middleware groups), `middleware.go` (project-specific middleware only; use the platform's for the rest), and `handler/` with one file per domain (`client.go`). Handlers are thin: bind, call the service, return the result or error. Other transports follow the same shape (`transport/grpc/`).

### `infrastructure/`

Technical adapters that are **not** domain logic and **not** already provided by the platform: for example a client for a third-party API, a notification sender, an object-storage adapter. Do **not** put database connections, pools, health checks, or SQL here: the platform provides them, and SQL belongs to the domain service. Services reach these adapters through small interfaces declared in the service package. See `references/boundaries-and-wiring.md`.

### `migration/`

Database schema migrations only, numbered and ordered (`001_create_users.sql`). Use the migration tool the project/platform already establishes; follow its naming if it differs. Migrations are not repositories and never live under `infrastructure/` or domain folders.

### `config/`

Typed configuration loaded and validated once at startup (`config.go`, split by concern such as `server.go`, `database.go` when it grows). Use the platform's environment helpers. Business code receives values through constructors; it does not read environment variables.

### `utility/`

The only place for application-level helpers that belong to no domain and are not provided by the platform. Rules: purpose-named subpackages only (`utility/phone`, not `utility/helpers`); stateless; no business rules; check the platform's utilities first. If a helper belongs to one domain, it lives in that domain.

## Dependency direction

```text
transport/http/handler ──▶ internal/<domain>/service (<domain>svc) ──▶ internal/<domain> (model)
                                      │
                                      ├──▶ platform (db, apperror, logger, ...)
                                      └──▶ small interfaces (implemented by infrastructure/)

cmd/<binary>/main.go ──▶ wires everything above
```

- `internal/<domain>` (model) imports only the standard library, small third-party types (such as `uuid`), and platform types it needs (errors, pagination meta). It never imports services, transport, or infrastructure.
- A service imports its own domain model, other domains' **models**, and the platform. It never imports transport.
- A service that needs another domain's **behavior** declares a small interface in its own package and receives the other service through the constructor. Do not import another domain's service package directly (this prevents import cycles and tangled domains).
- Handlers import services and models. Nothing imports `transport/`.
- `config` is imported by `main` only.

## When to split service files

Split only when the service becomes genuinely hard to navigate. The split is driven by complexity, not by how many functions exist.

```text
service/
└── client_service.go          start here: struct, constructor, queries, all CRUD
```

Split when:

- a workflow is large and self-contained (an import, a multi-step approval);
- search/query logic is substantial and independent;
- the SQL constants have grown large (move them to `client_query.go`);
- the file has become difficult to scan.

```text
service/
├── client_service.go
├── client_search_service.go    substantial search logic
├── client_query.go             SQL constants moved out because there are many
└── client_import.go            one large, self-contained workflow
```

Never create one file per operation (`create.go`, `get.go`, `update.go`, `delete.go`). Same rule for `model.go`: keep related models together until the file is hard to maintain; then split by concept (`client_search_model.go`), not per struct.

A very large domain may get sub-packages when a part is a genuinely independent concept (`internal/tournament/bracket/`), each with its own `model.go` and service. Do this rarely.

## Naming

- Folders and packages: lowercase, singular, no underscores (`client`, `user`). Service packages: `<domain>svc`.
- Files: `snake_case.go` named for the main concept. Tests end in `_test.go` and sit next to the code; integration tests use the `integration` build tag.
- No `utils`, `helpers`, `common`, `base`, `misc`, `shared`, `models`, `services`, `controllers` folders or packages.

## Adding a new domain

1. Create `internal/<domain>/model.go` with the entity, commands, queries, result types, enums, and `Validate()` methods.
2. Create `internal/<domain>/<domain>.go` if the domain has named errors or constants.
3. Create `internal/<domain>/service/<domain>_service.go` (package `<domain>svc`) with the service, constructor, and SQL constants.
4. Add the migration in `migration/` for the tables.
5. Add `transport/http/handler/<domain>.go` and register routes in `routes.go`.
6. Construct the service and handler in `main.go` and pass the platform database in.
7. Add tests (unit next to the code; integration for SQL).

## What not to create

- A repository layer or `repositories/` folders.
- `infrastructure/postgres/` wrapping the platform database, or SQL files kept apart from the service.
- Per-operation service files, or per-struct model files.
- Global technical folders (`services/`, `models/`, `controllers/`, `utils/`).
- Interfaces next to their only implementation "for mocking".
- Layers that only forward calls.

## References

| Read when | File |
|---|---|
| Writing `model.go`, `<domain>.go`, or a service file; deciding what goes in each; splitting | `references/domain-package.md` |
| Writing handlers, routes, `main.go` wiring, config, infrastructure adapters, workers, cross-domain calls | `references/boundaries-and-wiring.md` |
| Want a complete example across files | `references/examples.md` |
