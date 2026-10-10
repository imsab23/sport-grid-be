# Examples

Read when you want concrete before/after code. Business names are from Sport Grid, but the patterns apply anywhere. Assumes Go 1.22 or newer.

## Contents
1. Service method: before and after
2. Domain method
3. Repository method
4. HTTP handler
5. Fan-out with errgroup

## 1. Service method: before and after

Before (inline errors, vague names, noise comments, logs and returns):

```go
// WithdrawReg withdraws
func (s *Svc) Wd(ctx context.Context, o, i string) error {
	// get the registration
	if r, err := s.repo.Get(ctx, o, i); err != nil {
		log.Println(err)
		return err
	} else {
		// withdraw it
		if err := r.Wd(time.Now()); err != nil {
			return err
		}
		// save it
		if err := s.repo.Save(ctx, r); err != nil {
			return err
		}
	}
	return nil
}
```

After:

```go
func (s *RegistrationService) WithdrawRegistration(ctx context.Context, organizationID string, registrationID string) error {
	registration, err := s.registrationRepository.GetByID(ctx, organizationID, registrationID)
	if err != nil {
		return fmt.Errorf("get registration %s: %w", registrationID, err)
	}

	err = registration.Withdraw(s.clock.Now())
	if err != nil {
		return fmt.Errorf("withdraw registration %s: %w", registrationID, err)
	}

	err = s.registrationRepository.Save(ctx, organizationID, registration)
	if err != nil {
		return fmt.Errorf("save registration %s: %w", registrationID, err)
	}

	return nil
}
```

What changed: assign-then-check, descriptive names, wrapped errors with identifiers, injected clock, no logging in the service, no comments restating code, tenant ID passed explicitly.

## 2. Domain method

```go
func (r *Registration) Withdraw(withdrawnAt time.Time) error {
	if r.Status.IsTerminal() {
		return fmt.Errorf("withdraw from status %s: %w", r.Status, ErrInvalidStatusTransition)
	}

	r.Status = RegistrationStatusWithdrawn
	r.WithdrawnAt = withdrawnAt

	return nil
}
```

The transition rule lives in the domain type, so every caller gets the same validation.

## 3. Repository method

```go
func (r *RegistrationRepository) GetByID(ctx context.Context, organizationID string, registrationID string) (Registration, error) {
	queryContext, cancel := context.WithTimeout(ctx, databaseQueryTimeout)
	defer cancel()

	row := r.database.QueryRowContext(queryContext, selectRegistrationByIDQuery, organizationID, registrationID)

	var registration Registration
	err := row.Scan(&registration.ID, &registration.DivisionID, &registration.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return Registration{}, ErrRegistrationNotFound
	}
	if err != nil {
		return Registration{}, fmt.Errorf("scan registration %s: %w", registrationID, err)
	}

	return registration, nil
}
```

The query is parameterized and filtered by organization ID; the driver's not-found error becomes a domain error; every other error is wrapped with context.

## 4. HTTP handler

```go
func (h *RegistrationHandler) WithdrawRegistration(responseWriter http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	organizationID := organizationIDFromContext(ctx)
	registrationID := request.PathValue("registrationID")

	err := h.registrationService.WithdrawRegistration(ctx, organizationID, registrationID)
	if errors.Is(err, ErrRegistrationNotFound) {
		writeErrorResponse(responseWriter, http.StatusNotFound, "registration not found")
		return
	}
	if errors.Is(err, ErrInvalidStatusTransition) {
		writeErrorResponse(responseWriter, http.StatusConflict, "registration cannot be withdrawn in its current status")
		return
	}
	if err != nil {
		h.logger.ErrorContext(ctx, "withdraw registration",
			slog.String("organizationID", organizationID),
			slog.String("registrationID", registrationID),
			slog.Any("error", err),
		)
		writeErrorResponse(responseWriter, http.StatusInternalServerError, "internal error")
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}
```

The handler is the boundary: it maps domain errors to responses, logs unexpected errors once, and never exposes internal error text. The organization ID comes from the authenticated context, not from a request field.

## 5. Fan-out with errgroup

```go
func (s *DivisionSummaryService) LoadDivisionSummary(ctx context.Context, organizationID string, divisionID string) (DivisionSummary, error) {
	var confirmedRegistrationCount int
	var waitlistedRegistrationCount int

	group, groupContext := errgroup.WithContext(ctx)

	group.Go(func() error {
		var err error
		confirmedRegistrationCount, err = s.registrationRepository.CountByStatus(groupContext, organizationID, divisionID, RegistrationStatusConfirmed)
		if err != nil {
			return fmt.Errorf("count confirmed registrations: %w", err)
		}

		return nil
	})

	group.Go(func() error {
		var err error
		waitlistedRegistrationCount, err = s.registrationRepository.CountByStatus(groupContext, organizationID, divisionID, RegistrationStatusWaitlisted)
		if err != nil {
			return fmt.Errorf("count waitlisted registrations: %w", err)
		}

		return nil
	})

	err := group.Wait()
	if err != nil {
		return DivisionSummary{}, fmt.Errorf("load division summary %s: %w", divisionID, err)
	}

	return DivisionSummary{
		ConfirmedRegistrationCount:  confirmedRegistrationCount,
		WaitlistedRegistrationCount: waitlistedRegistrationCount,
	}, nil
}
```

The group has an owner (`Wait`), the first error cancels the sibling call, and each goroutine writes to its own variable, so there is no data race.
