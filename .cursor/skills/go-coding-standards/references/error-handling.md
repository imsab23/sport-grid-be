# Error Handling

Read when handling, wrapping, returning, logging, or mapping errors.

## Contents
1. The standard form
2. Reusing and shadowing `err`
3. Wrapping
4. Handle each error once
5. Sentinel and typed errors
6. Layer responsibilities
7. Never ignore errors
8. Panics
9. Goroutines and multiple errors

## 1. The standard form

Assign, then check, on separate lines. This lets you set a breakpoint on the assignment and inspect the returned values before the check runs.

```go
user, err := s.userService.GetByID(ctx, userID)
if err != nil {
	return fmt.Errorf("get user %s: %w", userID, err)
}
```

For functions that return only an error:

```go
err := s.registrationRepository.Save(ctx, registration)
if err != nil {
	return fmt.Errorf("save registration %s: %w", registration.ID, err)
}
```

Not allowed:

```go
if err := s.registrationRepository.Save(ctx, registration); err != nil { ... }
if user, err := s.userService.GetByID(ctx, userID); err != nil { ... }
switch err := doSomething(); err { ... }
```

The same reasoning applies to `ok` checks: prefer a separate statement.

```go
registration, found := registrationByID[registrationID]
if !found {
	return ErrRegistrationNotFound
}
```

## 2. Reusing and shadowing `err`

- The first assignment in a function uses `:=`. Later assignments of only `err` use `=`.
- `:=` with at least one new variable on the left is fine and reuses the existing `err` in the same scope.
- Do not declare `err :=` again inside a nested block that is meant to update an outer `err`: it shadows it silently. Assign with `=`, or restructure so the nested block returns.

```go
err := validateRegistrationRequest(request)
if err != nil {
	return fmt.Errorf("validate registration request: %w", err)
}

err = s.registrationRepository.Save(ctx, registration)
if err != nil {
	return fmt.Errorf("save registration %s: %w", registration.ID, err)
}
```

## 3. Wrapping

- Wrap with `fmt.Errorf("<verb> <object>: %w", err)`. Lowercase, no trailing punctuation.
- Describe the operation that failed, with identifiers that help debugging: `"save registration %s: %w"`.
- Do not write `"failed to ..."` or `"error ..."` prefixes. The chain would read "failed to a: failed to b: failed to c".
- Do not put secrets or personal data in messages (tokens, payment reference numbers if sensitive in your policy, minors' details).
- Use `%w` so callers can use `errors.Is` and `errors.As`. Use `%v` only when deliberately hiding the cause across a boundary.
- Do not wrap the same error twice at the same level or add context that the callee already added.

## 4. Handle each error once

Either return the error (wrapped) or handle it (log, retry, convert, recover). Not both. Log at the boundary where the error is finally handled: the HTTP handler, background job runner, or `main`. Lower layers wrap and return.

```go
// Wrong: logs and returns, producing duplicate log lines
err := s.registrationRepository.Save(ctx, registration)
if err != nil {
	s.logger.ErrorContext(ctx, "save registration", slog.Any("error", err))
	return err
}
```

## 5. Sentinel and typed errors

- Define business errors as sentinels: `var ErrRegistrationClosed = errors.New("registration closed")`. Name them `ErrXxx`.
- Use a typed error when callers need data: `type CapacityReachedError struct { DivisionID string }` with the method `Error() string`. Name the type `XxxError`.
- Compare with `errors.Is` and `errors.As`, never `==` on wrapped errors and never by matching error strings.
- Keep the list of business errors small and meaningful. They map to user-facing outcomes (for example capacity reached, registration closed, payment already reviewed, invalid status transition).

## 6. Layer responsibilities

| Layer | Does |
|---|---|
| Repository | Translates driver errors into domain errors (for example `sql.ErrNoRows` becomes `ErrRegistrationNotFound`); wraps everything else with the operation and identifiers. Does not log. |
| Service / domain | Enforces business rules, returns domain errors, wraps repository errors with context. Does not log and return. |
| Handler / job runner / `main` | Maps domain errors to responses or retry decisions; logs unexpected errors once; never leaks internal error text to clients. |

Do not let `database/sql`, driver, or HTTP-client errors cross into the handler layer unwrapped from their meaning: callers should reason about domain errors, not drivers.

## 7. Never ignore errors

- Do not discard with `_` unless ignoring is genuinely correct (for example best-effort cleanup). In that case add a short comment saying why, since this is an important comment.
- `defer` of `Close()` on readers: ignoring the error is acceptable only when nothing can be done about it. For writers, files you write, and transactions, handle the error.
- `rows.Err()` must be checked after iterating rows.

## 8. Panics

- No `panic` in business or library code. Return errors.
- `panic` is only for programmer errors that should crash at startup (for example an invalid regex constant or missing required wiring), usually in `main` or a clearly named `Must...` helper used only at startup and in tests.
- Recover only at boundaries (HTTP server middleware, goroutine entry points) to log and fail the request or job safely.

## 9. Goroutines and multiple errors

- Use `errgroup` so the first error cancels siblings and `Wait` returns it.
- Combine independent errors with `errors.Join` when all must be reported.
- A goroutine that cannot return an error must report it through a channel or the group, never drop it.
