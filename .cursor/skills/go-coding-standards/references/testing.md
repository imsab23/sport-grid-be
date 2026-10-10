# Testing

Read when writing or reviewing Go tests.

## Rules

- Test behavior through the public API of the package; do not test private internals unless there is no other seam.
- Use table-driven tests when the same logic is checked with different inputs. Each case has a descriptive `name`. The loop variable is `testCase`.
- Cover the error paths and boundaries: not found, invalid transition, capacity reached, duplicate, unauthorized, timeout or cancellation, rollback.
- Run with `go test -race ./...`.
- No `time.Sleep` for synchronization; use channels, `sync.WaitGroup`, or an injected clock.
- Inject the clock and ID generators so results are deterministic.
- Prefer small hand-written fakes over generated mocks when the interface is small; they make tests easier to read and debug. Use the mocking approach already in the repository.
- Tests that need a database use a real test database or the repository's established test harness, and each test cleans up after itself.
- Mark independent tests and subtests with `t.Parallel()` only when they share no state.
- Test helpers call `t.Helper()`.
- Check error identity with `errors.Is`, not by comparing message strings.
- Keep assertions specific: assert the status that changed, the error returned, and any side effect (saved record, emitted event).

## Naming

`Test<Unit>_<Scenario>_<ExpectedOutcome>`, with the allowed short parameter name `t *testing.T`.

```go
func TestHasAvailableSlot(t *testing.T) {
	testCases := []struct {
		name                string
		capacity            int
		occupiedSlotCount   int
		expectedHasSlot     bool
	}{
		{name: "division with free slots has an available slot", capacity: 16, occupiedSlotCount: 15, expectedHasSlot: true},
		{name: "full division has no available slot", capacity: 16, occupiedSlotCount: 16, expectedHasSlot: false},
		{name: "zero capacity has no available slot", capacity: 0, occupiedSlotCount: 0, expectedHasSlot: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			division := Division{Capacity: testCase.capacity, OccupiedSlotCount: testCase.occupiedSlotCount}

			actualHasSlot := division.HasAvailableSlot()

			if actualHasSlot != testCase.expectedHasSlot {
				t.Errorf("HasAvailableSlot() = %t, want %t", actualHasSlot, testCase.expectedHasSlot)
			}
		})
	}
}
```

Use the same assert/require library the repository already uses; do not introduce a second one.

## What to test for this domain

Where it applies, add tests for: invalid state transitions are rejected; a user from organization A cannot read or change organization B's data; two concurrent registrations cannot exceed capacity; the same request submitted twice has one effect; cancelling or postponing leaves no registration or payment in an inconsistent state. The business rules themselves are in `sport-grid-business-flow`.
