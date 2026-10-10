# Naming

Read when naming anything. The rule: a reader should know what a name refers to or does without looking at its definition.

## Contents
1. Variables
2. Allowed short names and banned names
3. Functions and methods
4. Types, interfaces, constants
5. Packages and files
6. Errors
7. Tests
8. Acronyms
9. Quick examples

## 1. Variables

Name the thing for what it is in the business, not for its type or position.

| Instead of | Write |
|---|---|
| `t, err := s.userSvc.GetByID(ctx, id)` | `user, err := s.userService.GetByID(ctx, userID)` |
| `t, err := s.tournamentService.GetByID(...)` | `tournament, err := s.tournamentService.GetByID(...)` |
| `res`, `ret`, `out`, `result` | `registrationSummary`, `approvedPayment`, `availableSlotCount` |
| `data`, `info`, `obj`, `item`, `val` | the actual noun: `paymentRecord`, `divisionConfiguration` |
| `n`, `cnt` | `registrationCount`, `waitlistedCount` |
| `err2`, `err3` | reuse `err` with `=` after the first assignment (see `error-handling.md`) |
| `list` | `activeRegistrations` |
| `m` (map) | `registrationByID`, `refundStatusByPaymentID` |

Naming patterns:

- Collections are plural: `registrations`, `divisions`. Range variables are the singular: `for _, registration := range registrations`.
- Maps read as "value by key": `divisionByID`, `countByStatus`.
- Booleans start with `is`, `has`, `can`, `should`: `isRegistrationOpen`, `hasAvailableSlot`, `canApprovePayment`.
- Counts end in `Count`; durations and timeouts say so: `paymentDeadlineDuration`, `requestTimeout`; times end in `At`: `submittedAt`, `expiresAt`.
- IDs say whose: `organizationID`, `registrationID`, `reviewerUserID`. Never a bare `id` when more than one ID is in scope.
- Do not encode the type in the name (`userStruct`, `nameString`).

## 2. Allowed short names and banned names

**Allowed:** `ctx` (context), `err`, `ok`/`found` (comma-ok), `id` (only when it is unambiguous), `db`, `i` and `j` for trivial index-only loops of a few lines, method receivers (see below), and `t *testing.T` / `b *testing.B` as test parameters. Type parameters may be `T`, `K`, `V` only for trivially generic helpers; otherwise use a descriptive name such as `Entity` or `Key`.

**Receivers:** short, consistent per type, derived from the type name (`s` for `RegistrationService`, `h` for `RegistrationHandler`). Do not mix names for one type.

**Banned as variable or field names:** `t` (except `*testing.T`), `tmp`, `temp`, `res`, `resp` (write `response` or what it is), `ret`, `out`, `data`, `val`, `value`, `obj`, `item`, `info`, `x`, `y`, `a`, `b`, `foo`, `bar`, `thing`, `stuff`, and abbreviations `Svc`, `Mgr`, `Cfg`, `Repo`, `Impl`, `Ctrl`, `Util`.

**Banned as package or type names:** `util`, `utils`, `common`, `helpers`, `base`, `misc`, `shared`, `manager`, `processor`.

## 3. Functions and methods

A function name is a verb plus an object plus any qualifier needed to know exactly what it is for. If you need a comment to explain what the function does, the name is wrong.

| Vague | Specific |
|---|---|
| `Process` | `ConfirmRegistrationAfterPaymentApproval` |
| `Handle` | `RejectPaymentWithReason` |
| `CheckCapacity` | `HasAvailableSlot` |
| `GetData` | `ListActiveRegistrationsByDivision` |
| `Run`, `Do`, `Execute` | `ExpireUnpaidRegistrations`, `PromoteNextWaitlistedPlayer` |
| `Validate` | `ValidateRegistrationAgainstEligibilityRules` |
| `Update` | `ChangeTournamentStatusToPostponed` |

Verb conventions:

- `Get...` returns the thing or a not-found error. `GetByID` is fine on a type whose name already says what it returns (`userService.GetByID`); add a qualifier when there are several lookups (`GetByEmail`).
- `List...` returns a slice (empty, never nil, unless the repository standard says otherwise). `Count...` returns a number.
- `Create...`, `Save...`, `Update...`, `Delete...` write data. Prefer a domain verb when one exists (`Confirm`, `Withdraw`, `Cancel`, `Postpone`).
- `Is...`, `Has...`, `Can...` return `bool` and have no side effects.
- `Validate...` returns an error. `Calculate...` or `Compute...` is pure.
- `New...` constructs. `Must...` panics and is only for startup and tests.
- Avoid `And` chains: a name with more than one `And` usually means two functions.
- Keep names to about five words; if longer, the function is probably doing too much.

## 4. Types, interfaces, constants

- Types are nouns from the business domain: `Registration`, `PaymentVerification`, `DivisionCapacity`.
- Avoid stutter with the package name: `registration.Service`, not `registration.RegistrationService`.
- Interfaces describe behavior and are defined where they are used, kept small. No `I` prefix and no `Interface` suffix. Single-method interfaces use the `-er` form (`PaymentVerifier`).
- Enums are typed constants with the type as prefix, not ALL_CAPS: `RegistrationStatusConfirmed`, `PaymentStatusUnderReview`.
- Unexported by default; export only what other packages need.

## 5. Packages and files

- Packages: short, lowercase, singular, no underscores, named for what they provide (`registration`, `payment`, `tournament`).
- Files: lowercase `snake_case.go` named for the main thing inside (`registration_service.go`, `payment_verification_repository.go`). Test files end in `_test.go`.
- One concept per file; do not create `utils.go` or `helpers.go`.

## 6. Errors

- Sentinels: `ErrRegistrationClosed`, `ErrCapacityReached`. Typed errors: `CapacityReachedError`.
- Error message strings are lowercase with no trailing punctuation.

## 7. Tests

- Test function: `Test<Unit>_<Scenario>_<ExpectedOutcome>`, for example `TestConfirmRegistration_PaymentApproved_SetsStatusConfirmed`.
- Table cases have a descriptive `name` field and the loop variable is `testCase`, not `tc` or `tt`.
- Fakes and test data are named for their meaning: `registrationWithExpiredPaymentDeadline`, `fullDivision`.

## 8. Acronyms

Acronyms keep one case: `userID`, `ID`, `URL`, `HTTPClient`, `JSONPayload`, `SQLQuery`, `APIKey`. Not `userId`, `Url`, `HttpClient`.

## 9. Quick examples

```go
// Hard to read
func (s *Svc) Proc(ctx context.Context, i string) error {
	r, err := s.repo.Get(ctx, i)
	...
}

// Easy to read
func (s *RegistrationService) ConfirmRegistrationAfterPaymentApproval(ctx context.Context, registrationID string) error {
	registration, err := s.registrationRepository.GetByID(ctx, registrationID)
	...
}
```
