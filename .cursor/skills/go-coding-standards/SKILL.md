---
name: go-coding-standards
description: Go coding standards for this project, covering error handling style (a separate `err :=` statement followed by `if err != nil`, never the inline `if err := ...; err != nil` form), descriptive naming (no single-letter or vague variable names, function names that state their purpose), comment policy (only when important), readability, production best practices (context, timeouts, logging, concurrency, transactions, configuration, security), testing, and a review checklist. Use this skill whenever writing, editing, refactoring, reviewing, or generating Go code or Go tests, even for small changes, and whenever answering questions about Go style, naming, error handling, or code structure.
---

# Go Coding Standards

How we write Go here. Business rules live in `sport-grid-business-flow`; folder layout lives in the project-structure skill. This skill is about the code itself.

**Priorities, in order:** 1. easy to debug, 2. easy to read, 3. safe in production, 4. idiomatic Go (when it does not conflict with 1 to 3).

Some rules below deliberately override common Go habits (inline `if err :=`, short names). Follow these rules, not the habit.

Assume the Go version in `go.mod` (these standards expect Go 1.22 or newer). Check the repository's existing code before adding a new pattern; match what is already there when it does not conflict with these rules.

## The rules that matter most

### 1. Error handling: assign first, then check

Never use the inline form. Assign the error on its own line, then check it on the next line, so you can put a breakpoint on the assignment and inspect values.

```go
// Wrong
if err := s.registrationRepository.Save(ctx, registration); err != nil {
	return fmt.Errorf("save registration: %w", err)
}

// Right
err := s.registrationRepository.Save(ctx, registration)
if err != nil {
	return fmt.Errorf("save registration %s: %w", registration.ID, err)
}
```

The same applies to `switch` and `if` initializers that call functions: use a separate statement. Details, wrapping, sentinel errors, and layering rules are in `references/error-handling.md`.

### 2. Names say exactly what the thing is

```go
// Wrong
t, err := s.userSvc.GetByID(ctx, id)
res, err := s.repo.Do(ctx, x)

// Right
user, err := s.userService.GetByID(ctx, userID)
registrationCount, err := s.registrationRepository.CountActiveByDivision(ctx, divisionID)
```

- No single-letter or vague names (`t`, `tmp`, `res`, `data`, `val`, `obj`, `item`, `info`). Allowed short names: `ctx`, `err`, `ok`, `id`, `db`, trivial index `i`/`j`, method receivers, and `t *testing.T`.
- No abbreviated fields or types (`Svc`, `Mgr`, `Cfg`, `Repo`): write `userService`, `configuration`, `registrationRepository`.
- Function names are a verb plus the object plus any qualifier needed to know what it is for: `ConfirmRegistrationAfterPaymentApproval`, not `Process` or `Handle`.

Full conventions and a banned-names list are in `references/naming.md`.

### 3. Comments: only when important

Code and names explain *what*. Add a comment only when a reader would otherwise miss something important that the code cannot say:

- why something non-obvious is done (a business rule, constraint, ordering requirement);
- a concurrency, locking, or transaction reason;
- a workaround, with a ticket or issue reference;
- a security, performance, or data-safety warning;
- the contract of an exported API when its name and signature do not make it clear (side effects, returned errors, units, thread safety).

Never write comments that restate the code, section banners, or commented-out code (delete it; version control keeps history). `TODO` is allowed only with a ticket reference. Machine directives (`//go:build`, `//go:generate`, `//nolint:name // reason`) are fine. The `exported` doc-comment lint rule is turned off in `assets/golangci.yml` for this reason.

### 4. Small, flat, readable functions

- Guard clauses and early returns; no `else` after a `return`.
- Aim for functions under about 40 lines and nesting no deeper than 3. If a function needs a comment to separate sections, split it into named functions.
- At most 4 parameters; beyond that, use a named parameter struct.
- No naked returns. Avoid named results unless they document two same-typed returns.
- No boolean flag parameters that change behavior; use two functions or a named type.
- One level of abstraction per function.
- Prefer concrete types and simple code. Use generics, reflection, and `any` only when they remove real duplication.

### 5. Production defaults

Every change must be safe to run in production. The short list (details in `references/production-practices.md`):

- `ctx context.Context` is the first parameter of anything that does I/O; always propagate it.
- Every external call (database, HTTP, queue) has a timeout.
- Structured logging with `log/slog`; handle each error once (return or log, not both); never log secrets or personal data.
- No package-level mutable state, no `init()` side effects; inject dependencies through constructors.
- Every goroutine has an owner and a way to stop; use `errgroup`.
- Money as integer minor units plus currency; time in UTC; inject a clock instead of calling `time.Now()` in business logic.
- SQL is parameterized and tenant-scoped; transactions are explicit and always rolled back on failure.
- Configuration is validated at startup and the program fails fast.

### 6. Tests

Table-driven tests with descriptive case names, error paths covered, run with `-race`. See `references/testing.md`.

## Review checklist

Before finishing any Go change, confirm:

- [ ] No `if err := ...; err != nil` anywhere in the diff.
- [ ] No vague or single-letter names (outside the allowed list); function names state their purpose.
- [ ] No comment that restates the code; important comments explain *why*.
- [ ] Errors are wrapped with context using `%w`, handled once, and never ignored silently.
- [ ] `ctx` is passed through and external calls have timeouts.
- [ ] No new globals, `init()` side effects, or goroutines without an owner.
- [ ] Queries are parameterized and scoped to the tenant where the data is tenant-owned.
- [ ] Tests cover the new behavior including error paths; `go test -race ./...` passes.
- [ ] `gofumpt`, `goimports`, and `golangci-lint run` pass with `assets/golangci.yml`.

## References

| Read when | File |
|---|---|
| Handling, wrapping, returning, or mapping errors | `references/error-handling.md` |
| Naming anything (variables, functions, types, packages, files, tests) | `references/naming.md` |
| Context, logging, concurrency, database access, config, security, HTTP | `references/production-practices.md` |
| Writing or reviewing tests | `references/testing.md` |
| Want concrete before/after code (service, repository, handler, transaction) | `references/examples.md` |
| Setting up linting that enforces these rules | `assets/golangci.yml` |

## Tooling enforces what it can

Prefer tools over prose. Formatting and most of these rules are checked by `gofumpt`, `goimports`, and `golangci-lint` (config in `assets/golangci.yml`, including the linters that reject inline error handling and short names). Run them before declaring work done. If a rule here conflicts with a linter in the repository's own configuration, report the conflict instead of silently picking one.
