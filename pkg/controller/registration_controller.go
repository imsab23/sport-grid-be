package controller

import (
	"sport-grid-be/pkg/registration"
	"sport-grid-be/pkg/role"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

// NewPlayerRegistrationController registers player-facing, self-service registration routes.
func (s *Server) NewPlayerRegistrationController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/divisions/{divisionId}", func(r router.Router) {
		r.POST("/register", s.createRegistrationHandler)
	})

	r.Group("/registrations", func(r router.Router) {
		r.GET("/", s.searchOwnRegistrationHandler)
		r.POST("/{id}/cancel", s.cancelOwnRegistrationHandler)
		r.POST("/{id}/payment-submissions", s.createPaymentSubmissionHandler)
	})
}

// NewRegistrationController registers staff-facing registration management routes,
// gated entirely by role since registration data includes other players' PII and payment info.
func (s *Server) NewRegistrationController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/registrations", func(r router.Router) {
		r.Use(wrapNetHTTPMiddleware(authzmw.RequireAnyRole(
			string(role.SuperAdmin), string(role.ClientAdmin), string(role.TournamentStaff),
		)))
		r.GET("/", s.searchTournamentRegistrationsHandler)
		r.GET("/{id}", s.getTournamentRegistrationHandler)
		r.POST("/{id}/approve", s.approveRegistrationHandler)
		r.POST("/{id}/cancel", s.cancelTournamentRegistrationHandler)
		r.POST("/{id}/payment-submissions/{subId}/verify", s.verifyPaymentSubmissionHandler)
		r.POST("/{id}/payment-submissions/{subId}/reject", s.rejectPaymentSubmissionHandler)
	})
}

// getOwnedRegistration fetches a registration and verifies it belongs to the tournament in the path.
func (s *Server) getOwnedRegistration(c *router.Ctx, tournamentID uuid.UUID) (*registration.Registration, error) {
	rid := c.Param("id")
	if !validate.UUID(rid) {
		response.BadRequest(c.ResponseWriter(), "Invalid registration ID")
		return nil, errHandledResponse
	}

	reg, err := s.Dependencies.RegistrationSvc.GetByID(c.Context(), uuid.MustParse(rid))
	if err != nil {
		return nil, err
	}
	if reg.TournamentID != tournamentID {
		return nil, registration.ErrRegistrationNotFound
	}

	return reg, nil
}

// ── Player-facing handlers ──────────────────────────────────────────────────

func (s *Server) createRegistrationHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tid := c.Param("tournamentId")
	did := c.Param("divisionId")
	if !validate.UUID(tid) || !validate.UUID(did) {
		response.BadRequest(c.ResponseWriter(), "Invalid tournament or division ID")
		return nil
	}

	playerID, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd registration.CreateRegistrationCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.TournamentID = uuid.MustParse(tid)
	cmd.DivisionID = uuid.MustParse(did)
	cmd.PlayerID = playerID

	result, err := s.Dependencies.RegistrationSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) searchOwnRegistrationHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	playerID, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var (
		query registration.SearchRegistrationQuery
		m     meta.Meta
	)
	err = c.BindQuery(&query)
	if err != nil {
		return err
	}
	err = c.BindQuery(&m)
	if err != nil {
		return err
	}
	query.Meta = &m
	query.PlayerID = &playerID

	result, err := s.Dependencies.RegistrationSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Registrations, result.Meta)
	return nil
}

func (s *Server) cancelOwnRegistrationHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	playerID, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	rid := c.Param("id")
	if !validate.UUID(rid) {
		response.BadRequest(c.ResponseWriter(), "Invalid registration ID")
		return nil
	}

	reg, err := s.Dependencies.RegistrationSvc.GetByID(c.Context(), uuid.MustParse(rid))
	if err != nil {
		return err
	}
	if reg.PlayerID != playerID {
		response.Forbidden(c.ResponseWriter())
		return nil
	}
	// Players may only withdraw before staff/payment confirmation; after that, staff must cancel.
	if reg.Status != registration.StatusPending {
		return registration.ErrInvalidRegistrationOp
	}

	err = s.Dependencies.RegistrationSvc.Cancel(c.Context(), reg.ID)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) createPaymentSubmissionHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	playerID, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	rid := c.Param("id")
	if !validate.UUID(rid) {
		response.BadRequest(c.ResponseWriter(), "Invalid registration ID")
		return nil
	}

	reg, err := s.Dependencies.RegistrationSvc.GetByID(c.Context(), uuid.MustParse(rid))
	if err != nil {
		return err
	}
	if reg.PlayerID != playerID {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd registration.CreatePaymentSubmissionCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.RegistrationID = reg.ID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.RegistrationSvc.CreatePaymentSubmission(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

// ── Staff-facing handlers ───────────────────────────────────────────────────

func (s *Server) searchTournamentRegistrationsHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var (
		query registration.SearchRegistrationQuery
		m     meta.Meta
	)
	err = c.BindQuery(&query)
	if err != nil {
		return err
	}
	err = c.BindQuery(&m)
	if err != nil {
		return err
	}
	query.Meta = &m
	query.TournamentID = &tUUID

	result, err := s.Dependencies.RegistrationSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Registrations, result.Meta)
	return nil
}

func (s *Server) getTournamentRegistrationHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	reg, err := s.getOwnedRegistration(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	response.SuccessWithResult(c.ResponseWriter(), reg)
	return nil
}

func (s *Server) approveRegistrationHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	reg, err := s.getOwnedRegistration(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	approvedBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	err = s.Dependencies.RegistrationSvc.Approve(c.Context(), reg.ID, approvedBy)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) cancelTournamentRegistrationHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	reg, err := s.getOwnedRegistration(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	err = s.Dependencies.RegistrationSvc.Cancel(c.Context(), reg.ID)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) verifyPaymentSubmissionHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	reg, err := s.getOwnedRegistration(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	subID := c.Param("subId")
	if !validate.UUID(subID) {
		response.BadRequest(c.ResponseWriter(), "Invalid payment submission ID")
		return nil
	}

	ps, err := s.Dependencies.RegistrationSvc.GetPaymentSubmissionByID(c.Context(), uuid.MustParse(subID))
	if err != nil {
		return err
	}
	if ps.RegistrationID != reg.ID {
		return registration.ErrPaymentNotFound
	}

	verifiedBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	err = s.Dependencies.RegistrationSvc.VerifyPaymentSubmission(c.Context(), ps.ID, verifiedBy)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) rejectPaymentSubmissionHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	reg, err := s.getOwnedRegistration(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	subID := c.Param("subId")
	if !validate.UUID(subID) {
		response.BadRequest(c.ResponseWriter(), "Invalid payment submission ID")
		return nil
	}

	ps, err := s.Dependencies.RegistrationSvc.GetPaymentSubmissionByID(c.Context(), uuid.MustParse(subID))
	if err != nil {
		return err
	}
	if ps.RegistrationID != reg.ID {
		return registration.ErrPaymentNotFound
	}

	var cmd registration.RejectPaymentSubmissionCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ID = ps.ID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	err = s.Dependencies.RegistrationSvc.RejectPaymentSubmission(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}
