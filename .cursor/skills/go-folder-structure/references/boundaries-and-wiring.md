# Boundaries and Wiring

Read when writing handlers, routes, `main.go`, configuration, infrastructure adapters, workers, or when deciding how packages talk to each other. Platform API names follow the platform README; confirm exact signatures in the platform skill.

## Contents
1. Handlers
2. Routes and middleware
3. `main.go` wiring
4. Configuration
5. Infrastructure adapters
6. Background workers
7. Utility
8. Migrations
9. Dependency checklist

## 1. Handlers

`transport/http/handler/<domain>.go`, one file per domain, one handler struct per domain.

- Receive the service through the constructor; hold it in an unexported field.
- Each method: bind input into a command or parse path/query values, call the service, return the error unchanged or write the success response through the platform's response helpers.
- No business rules, no SQL, no transactions.
- Errors from services are returned as they are; the platform router maps application errors to HTTP responses.
- Never bind a request into an entity; bind into a command.

A handler depends on the concrete service type, or on a small interface declared in the handler package if it needs to be faked in tests.

## 2. Routes and middleware

- `routes.go` registers all routes for the application in one place, grouped by path prefix and by the middleware each group needs (authentication, authorization).
- Use the platform's authentication and authorization middleware; resource-ownership checks follow the platform's pattern.
- `middleware.go` holds only project-specific middleware that the platform does not provide.
- `server.go` builds the platform router and server (timeouts, CORS, rate limiting, security headers) and exposes a `Run(ctx) error` method suitable for the lifecycle runner.

## 3. `main.go` wiring

`main.go` constructs dependencies from the outside in and passes them down; nothing below `main` reaches back up.

```go
func main() {
	log, err := logger.NewLogger("app")

	if err != nil {
		panic(err)
	}

	configuration, err := config.Load()

	if err != nil {
		log.Error("load configuration", logger.F("err", err))
		os.Exit(1)
	}

	database, err := dbimpl.NewPostgresDB(context.Background(), configuration.Database)

	if err != nil {
		log.Error("connect to postgres", logger.F("err", err))
		os.Exit(1)
	}

	defer database.Close()

	clientService := clientsvc.NewClientService(database)
	clientHandler := handler.NewClientHandler(clientService)

	httpServer, err := transporthttp.NewServer(configuration.Server, log, clientHandler)

	if err != nil {
		log.Error("create http server", logger.F("err", err))
		os.Exit(1)
	}

	runner := lifecycle.New(lifecycle.WithLogger(log))
	runner.Add(httpServer.Run)

	err = runner.Run(context.Background())

	if err != nil {
		log.Error("runner exited", logger.F("err", err))
		os.Exit(1)
	}
}
```

Rules:

- Create the logger first, then configuration, then the database, then services, handlers, and servers.
- Use the platform lifecycle runner to start servers and workers and to coordinate shutdown.
- `os.Exit` and `panic` appear only in `main`.
- If wiring grows long, move it to `wire.go` in the same package; do not create a global container or a service locator.

## 4. Configuration

- `config/config.go` defines one typed struct per concern (server, database, auth) and a `Load()` that reads the environment through the platform's helpers and validates it.
- Fail fast at startup with a message that lists everything missing.
- No secrets in the repository; no machine-specific paths.
- Services receive plain values or small structs through constructors; they never read the environment.

## 5. Infrastructure adapters

Use `infrastructure/` only for a technical adapter the platform does not already provide: a third-party API client, a notification sender, an object-storage adapter.

- Name by what it does: `infrastructure/messenger/`, `infrastructure/storage/`.
- The service that needs it declares a small interface in its own package, for example:

```go
type notificationSender interface {
	SendClientSuspendedNotice(ctx context.Context, clientID uuid.UUID) error
}
```

- The adapter implements that interface; `main.go` passes it in.
- An adapter contains no business rules. It translates between the service's interface and the external system, and applies timeouts.
- Do not create `infrastructure/postgres/`, `infrastructure/redis/`, `infrastructure/router/`, or similar wrappers around platform capabilities.

## 6. Background workers

A worker belongs to the domain that owns the behavior, and uses the same service and database as the rest of the domain.

- Put the work in the domain service (or a clearly named file in `service/`, for example `client_expiry_service.go`).
- Expose a `Run(ctx context.Context) error` method and register it with the lifecycle runner in `main.go`.
- The worker owns its loop, stops on `ctx.Done()`, and returns errors; it does not start unmanaged goroutines.
- Operations that run on several instances use database constraints, transactions, or atomic updates so two instances cannot do the same work twice.

## 7. Utility

`utility/` is the single home for small, stateless, domain-agnostic helpers that the platform does not provide.

- Purpose-named subpackages (`utility/phone`, `utility/slug`); never a grab-bag file at the root.
- No business rules, no I/O, no dependencies on domains.
- Check the platform's utilities first; if a helper is generic enough for all projects, it belongs in the platform, not here.
- If a helper belongs to one domain, keep it in that domain.

## 8. Migrations

- Files in `migration/`, numbered and ordered, one change per file, forward-safe (additive where possible).
- Use the project's migration tool; follow its naming and format.
- Constraints (unique, foreign keys, checks) live in migrations and are relied upon by services for correctness.
- Migrations never contain application logic and are not applied by services at request time.

## 9. Dependency checklist

Before merging a structural change, confirm:

- [ ] The new code is in the domain that owns the behavior.
- [ ] No import points from `internal/` to `transport/` or `cmd/`.
- [ ] No service imports another domain's service package; cross-domain behavior goes through a consumer-side interface.
- [ ] No repository layer, no `infrastructure/postgres/`, no SQL outside the service.
- [ ] No duplicated platform functionality.
- [ ] No per-operation service files or per-struct model files.
- [ ] Everything is wired in `main.go` through constructors.
