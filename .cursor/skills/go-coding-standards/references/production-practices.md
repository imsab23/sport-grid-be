# Production Practices

Read when touching I/O, concurrency, configuration, persistence, logging, security, or service startup. For tenant, authorization, concurrency, and state-machine rules specific to Sport Grid, see the business flow skill and the matching engineering skills.

## Contents
1. Context and timeouts
2. Dependencies and construction
3. Logging
4. Concurrency
5. Database access and transactions
6. Money and time
7. Configuration
8. Security
9. HTTP services
10. API and struct design
11. Dependencies and tooling

## 1. Context and timeouts

- `ctx context.Context` is the first parameter of every function that does I/O or may block. Never store a context in a struct. Never pass `nil`.
- Derive request-scoped values (organization ID, request ID, actor) from `ctx`; do not smuggle them through globals.
- Every call to an external system has a deadline: database queries, HTTP clients, queues. Use `context.WithTimeout` and always `defer cancel()`.

```go
queryContext, cancel := context.WithTimeout(ctx, databaseQueryTimeout)
defer cancel()
```

- Respect cancellation: long loops check `ctx.Err()`; blocking selects include `<-ctx.Done()`.

## 2. Dependencies and construction

- Build objects with constructors: `NewRegistrationService(registrationRepository, paymentRepository, clock, logger)`. Dependencies are explicit parameters, not fetched from globals.
- Accept interfaces, return concrete types. Define each interface where it is consumed, with only the methods that consumer uses.
- No package-level mutable state and no `init()` with side effects. Package-level `var` is for sentinel errors and compiled constants (regex, templates).
- Constructors validate required dependencies and return an error (or panic only when called from startup wiring) instead of leaving nil fields.

## 3. Logging

- Use `log/slog` with structured fields; pass the logger as a dependency.
- Use `slog.Any("error", err)` and `...Context(ctx, ...)` variants so request-scoped fields attach.
- Levels: `Error` for unexpected failures that need attention, `Warn` for degraded but handled situations, `Info` for significant lifecycle events, `Debug` for diagnostics.
- Log once, at the boundary (see `error-handling.md`).
- Never log secrets, tokens, full payment references, or personal data (especially minors' data). Log IDs instead.
- Include organization ID and request or job ID in boundary logs so a problem can be traced to a tenant and a request.

## 4. Concurrency

- Every goroutine has an owner who can stop it and wait for it. No fire-and-forget goroutines in request paths.
- Use `golang.org/x/sync/errgroup` for fan-out work; derive the group context from the request context.
- Protect shared state with a mutex next to the data it guards, or confine it to one goroutine. Document lock ordering when more than one lock exists.
- Prefer passing data over sharing memory; channels have a clear sender who closes them.
- Do not spawn unbounded goroutines; limit with `errgroup.SetLimit` or a worker pool.
- Run tests with `-race`.
- Concurrency at the **database level** (capacity, waitlist promotion, payment review, status changes) is solved with transactions, constraints, and atomic updates, not in-process mutexes; the service may run several instances.

## 5. Database access and transactions

- Parameterized queries only; never build SQL by concatenating input.
- Every query on tenant-owned data filters by organization ID. Passing the organization ID explicitly to repository methods makes the scope visible.
- Keep transactions short and explicit. Begin, defer a rollback that tolerates "already committed", commit, and handle the commit error.

```go
transaction, err := s.database.BeginTx(ctx, nil)
if err != nil {
	return fmt.Errorf("begin transaction: %w", err)
}
defer func() {
	rollbackErr := transaction.Rollback()
	if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
		s.logger.ErrorContext(ctx, "roll back transaction", slog.Any("error", rollbackErr))
	}
}()

// ... work using transaction ...

err = transaction.Commit()
if err != nil {
	return fmt.Errorf("commit transaction: %w", err)
}
```

- Always `defer rows.Close()` and check `rows.Err()` after iteration.
- Use unique constraints for duplicate protection and idempotency; rely on them, not on a read-then-insert check alone.
- No unbounded queries on request paths: paginate and cap results.

## 6. Money and time

- Money is an integer count of minor units plus a currency code. Never `float32` or `float64`.
- Store and compare times in UTC as `time.Time`; convert to the venue's time zone only for display and for deadline rules that are defined in local time.
- Use `time.Duration` for durations, not raw integers.
- Business logic does not call `time.Now()` directly; inject a `Clock` interface so tests can control time.

## 7. Configuration

- Read configuration from the environment (or the repository's existing mechanism) into one typed struct at startup.
- Validate it at startup and fail fast with a clear message listing what is missing.
- No secrets in the repository, logs, or error messages.
- No machine-specific paths.

## 8. Security

- Validate and normalize input at the boundary; domain types enforce their own invariants in constructors.
- Enforce authorization and tenant scope on the server using identity from the verified token or session, never from request fields alone.
- Compare secrets with `crypto/subtle.ConstantTimeCompare`. Use `crypto/rand` for tokens, never `math/rand`.
- Do not expose internal error text, stack traces, or SQL errors to clients.
- Run `gosec` (via golangci-lint) and `govulncheck` in CI.

## 9. HTTP services

- Handlers are thin: decode and validate, call a service, map the result or domain error to a response.
- Configure `http.Server` timeouts (at least `ReadHeaderTimeout`, plus read, write, and idle timeouts); never use `http.ListenAndServe` bare in production.
- Shut down gracefully on SIGTERM: stop accepting requests, wait for in-flight ones with a deadline, then close resources.
- Provide liveness and readiness endpoints.
- Return a consistent error shape and the right status codes; map domain errors explicitly.

## 10. API and struct design

- Make the zero value useful where practical; avoid constructors that only set defaults.
- Prefer small structs and small interfaces. Pass named parameter structs when there are more than four inputs.
- Use named types for IDs and statuses to prevent mixing them up (`type RegistrationID string`) where the repository does so, or at least distinct parameter names.
- Do not return a nil slice or map where an empty one is expected by callers; do not return `nil, nil` for "not found" (return an error).
- Make invalid states hard to represent: status fields are typed, transitions go through methods that validate them.

## 11. Dependencies and tooling

- Keep dependencies few; prefer the standard library when adequate. Run `go mod tidy`.
- CI runs: `gofumpt`, `goimports`, `go vet`, `golangci-lint run`, `go test -race ./...`, and `govulncheck ./...`.
- Pin the Go version in `go.mod` and use the same version in CI and local development.
