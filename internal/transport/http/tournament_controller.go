package controller

import (
	"sport-grid-be/internal/role/permission"
	"sport-grid-be/internal/tournament"
	"sport-grid-be/internal/transport"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	requireTournamentCreate            = authzmw.RequirePermission(permission.CreateTournament)
	requireTournamentUpdate            = authzmw.RequirePermission(permission.UpdateTournament)
	requireTournamentView              = authzmw.RequirePermission(permission.ViewTournament)
	requireTournamentPublish           = authzmw.RequirePermission(permission.PublishTournament)
	requireTournamentOpenRegistration  = authzmw.RequirePermission(permission.OpenTournamentRegistration)
	requireTournamentCloseRegistration = authzmw.RequirePermission(permission.CloseTournamentRegistration)
	requireTournamentStart             = authzmw.RequirePermission(permission.StartTournament)
	requireTournamentComplete          = authzmw.RequirePermission(permission.CompleteTournament)
	requireTournamentCancel            = authzmw.RequirePermission(permission.CancelTournament)
)

func (s *Server) NewTournamentController(r router.Router) {
	r.Group("/tournaments", func(r router.Router) {
		r.GET("/", s.searchTournamentHandler, requireTournamentView)
		r.GET("/{id}", s.getTournamentHandler, requireTournamentView)
		r.POST("/", s.createTournamentHandler, requireTournamentCreate)
		r.PUT("/{id}", s.updateTournamentHandler, requireTournamentUpdate)
		r.POST("/{id}/publish", s.publishTournamentHandler, requireTournamentPublish)
		r.POST("/{id}/open-registration", s.openRegistrationTournamentHandler, requireTournamentOpenRegistration)
		r.POST("/{id}/close-registration", s.closeRegistrationTournamentHandler, requireTournamentCloseRegistration)
		r.POST("/{id}/start", s.startTournamentHandler, requireTournamentStart)
		r.POST("/{id}/complete", s.completeTournamentHandler, requireTournamentComplete)
		r.POST("/{id}/cancel", s.cancelTournamentHandler, requireTournamentCancel)
	})
}

func (s *Server) createTournamentHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd tournament.CreateTournamentCommand
	err := c.BindJson(&cmd)
	if err != nil {
		return err
	}

	// Tenant scope always derives from identity, never from the request body.
	if !transport.IsSuperAdmin(id) {
		clientID, err := transport.ClientID(id)
		if err != nil {
			return err
		}
		cmd.ClientID = clientID
	}

	createdBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}
	cmd.CreatedBy = createdBy

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.TournamentSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) searchTournamentHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var (
		query tournament.SearchTournamentQuery
		m     meta.Meta
	)

	err := c.BindQuery(&query)
	if err != nil {
		return err
	}
	err = c.BindQuery(&m)
	if err != nil {
		return err
	}
	query.Meta = &m

	if !transport.IsSuperAdmin(id) {
		clientID, err := transport.ClientID(id)
		if err != nil {
			return err
		}
		query.ClientID = &clientID
	}

	result, err := s.Dependencies.TournamentSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Tournaments, result.Meta)
	return nil
}

func (s *Server) getTournamentHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tid := c.Param("id")
	if !validate.CheckUUIDv7Format(tid) {
		response.BadRequest(c.ResponseWriter(), "Invalid tournament ID")
		return nil
	}

	result, err := s.Dependencies.TournamentSvc.GetByID(c.Context(), uuid.MustParse(tid))
	if err != nil {
		return err
	}

	if err := transport.RequireClientOwnership(id, result.ClientID); err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) updateTournamentHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tid := c.Param("id")
	if !validate.CheckUUIDv7Format(tid) {
		response.BadRequest(c.ResponseWriter(), "Invalid tournament ID")
		return nil
	}
	tUUID := uuid.MustParse(tid)

	existing, err := s.Dependencies.TournamentSvc.GetByID(c.Context(), tUUID)
	if err != nil {
		return err
	}
	if err := transport.RequireClientOwnership(id, existing.ClientID); err != nil {
		return err
	}

	var cmd tournament.UpdateTournamentCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ID = tUUID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.TournamentSvc.Update(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

// tournamentTransitionHandler fetches the tournament, enforces tenant ownership,
// then applies the given state transition.
func (s *Server) tournamentTransitionHandler(c *router.Ctx, apply func(id uuid.UUID) error) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tid := c.Param("id")
	if !validate.CheckUUIDv7Format(tid) {
		response.BadRequest(c.ResponseWriter(), "Invalid tournament ID")
		return nil
	}
	tUUID := uuid.MustParse(tid)

	existing, err := s.Dependencies.TournamentSvc.GetByID(c.Context(), tUUID)
	if err != nil {
		return err
	}
	if err := transport.RequireClientOwnership(id, existing.ClientID); err != nil {
		return err
	}

	err = apply(tUUID)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) publishTournamentHandler(c *router.Ctx) error {
	return s.tournamentTransitionHandler(c, func(id uuid.UUID) error {
		return s.Dependencies.TournamentSvc.Publish(c.Context(), id)
	})
}

func (s *Server) openRegistrationTournamentHandler(c *router.Ctx) error {
	return s.tournamentTransitionHandler(c, func(id uuid.UUID) error {
		return s.Dependencies.TournamentSvc.OpenRegistration(c.Context(), id)
	})
}

func (s *Server) closeRegistrationTournamentHandler(c *router.Ctx) error {
	return s.tournamentTransitionHandler(c, func(id uuid.UUID) error {
		return s.Dependencies.TournamentSvc.CloseRegistration(c.Context(), id)
	})
}

func (s *Server) startTournamentHandler(c *router.Ctx) error {
	return s.tournamentTransitionHandler(c, func(id uuid.UUID) error {
		return s.Dependencies.TournamentSvc.Start(c.Context(), id)
	})
}

func (s *Server) completeTournamentHandler(c *router.Ctx) error {
	return s.tournamentTransitionHandler(c, func(id uuid.UUID) error {
		return s.Dependencies.TournamentSvc.Complete(c.Context(), id)
	})
}

func (s *Server) cancelTournamentHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tid := c.Param("id")
	if !validate.CheckUUIDv7Format(tid) {
		response.BadRequest(c.ResponseWriter(), "Invalid tournament ID")
		return nil
	}
	tUUID := uuid.MustParse(tid)

	existing, err := s.Dependencies.TournamentSvc.GetByID(c.Context(), tUUID)
	if err != nil {
		return err
	}
	if err := transport.RequireClientOwnership(id, existing.ClientID); err != nil {
		return err
	}

	var cmd tournament.CancelTournamentCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ID = tUUID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	err = s.Dependencies.TournamentSvc.Cancel(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}
