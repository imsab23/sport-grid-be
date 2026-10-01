package controller

import (
	"sport-grid-be/pkg/court"
	"sport-grid-be/pkg/role"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

func (s *Server) NewCourtController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/courts", func(r router.Router) {
		r.GET("/", s.searchCourtHandler)
		r.GET("/{id}", s.getCourtHandler)

		r.Group("", func(r router.Router) {
			r.Use(wrapNetHTTPMiddleware(authzmw.RequireAnyRole(
				string(role.SuperAdmin), string(role.ClientAdmin), string(role.TournamentStaff),
			)))
			r.POST("/", s.createCourtHandler)
			r.PUT("/{id}", s.updateCourtHandler)
			r.POST("/{id}/status", s.setCourtStatusHandler)
		})
	})
}

// getOwnedCourt fetches a court and verifies it belongs to the tournament in the path.
func (s *Server) getOwnedCourt(c *router.Ctx, tournamentID uuid.UUID) (*court.Court, error) {
	cid := c.Param("id")
	if !validate.UUID(cid) {
		response.BadRequest(c.ResponseWriter(), "Invalid court ID")
		return nil, errHandledResponse
	}

	ct, err := s.Dependencies.CourtSvc.GetByID(c.Context(), uuid.MustParse(cid))
	if err != nil {
		return nil, err
	}
	if ct.TournamentID != tournamentID {
		return nil, court.ErrCourtNotFound
	}

	return ct, nil
}

func (s *Server) createCourtHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd court.CreateCourtCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.TournamentID = tUUID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.CourtSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) searchCourtHandler(c *router.Ctx) error {
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
		query court.SearchCourtQuery
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
	query.TournamentID = tUUID

	result, err := s.Dependencies.CourtSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Courts, result.Meta)
	return nil
}

func (s *Server) getCourtHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	ct, err := s.getOwnedCourt(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	response.SuccessWithResult(c.ResponseWriter(), ct)
	return nil
}

func (s *Server) updateCourtHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	ct, err := s.getOwnedCourt(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd court.UpdateCourtCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ID = ct.ID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.CourtSvc.Update(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) setCourtStatusHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	ct, err := s.getOwnedCourt(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd court.SetCourtStatusCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ID = ct.ID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	err = s.Dependencies.CourtSvc.SetStatus(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}
