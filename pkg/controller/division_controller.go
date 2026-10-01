package controller

import (
	"errors"

	"sport-grid-be/pkg/authz"
	"sport-grid-be/pkg/division"
	"sport-grid-be/pkg/role"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

// errHandledResponse signals the response was already written (e.g. BadRequest); the caller should return nil.
var errHandledResponse = errors.New("response already written")

func handleOwnedTournamentErr(_ *router.Ctx, err error) error {
	if errors.Is(err, errHandledResponse) {
		return nil
	}
	return err
}

func (s *Server) NewDivisionController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/divisions", func(r router.Router) {
		r.GET("/", s.searchDivisionHandler)
		r.GET("/{id}", s.getDivisionHandler)

		r.Group("", func(r router.Router) {
			r.Use(wrapNetHTTPMiddleware(authzmw.RequireAnyRole(
				string(role.SuperAdmin), string(role.ClientAdmin), string(role.TournamentStaff),
			)))
			r.POST("/", s.createDivisionHandler)
			r.PUT("/{id}", s.updateDivisionHandler)
			r.POST("/{id}/activate", s.activateDivisionHandler)
			r.POST("/{id}/deactivate", s.deactivateDivisionHandler)
		})
	})
}

// ownedTournament resolves and authorizes the parent tournament from the URL path.
func (s *Server) ownedTournament(c *router.Ctx, identityID *identity.Identity) (uuid.UUID, error) {
	tid := c.Param("tournamentId")
	if !validate.UUID(tid) {
		response.BadRequest(c.ResponseWriter(), "Invalid tournament ID")
		return uuid.Nil, errHandledResponse
	}
	tUUID := uuid.MustParse(tid)

	t, err := s.Dependencies.TournamentSvc.GetByID(c.Context(), tUUID)
	if err != nil {
		return uuid.Nil, err
	}
	if err := authz.RequireClientOwnership(identityID, t.ClientID); err != nil {
		return uuid.Nil, err
	}

	return tUUID, nil
}

func (s *Server) createDivisionHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd division.CreateDivisionCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.TournamentID = tUUID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.DivisionSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) searchDivisionHandler(c *router.Ctx) error {
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
		query division.SearchDivisionQuery
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

	result, err := s.Dependencies.DivisionSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Divisions, result.Meta)
	return nil
}

// getOwnedDivision fetches a division and verifies it belongs to the tournament in the path.
func (s *Server) getOwnedDivision(c *router.Ctx, tournamentID uuid.UUID) (*division.Division, error) {
	did := c.Param("id")
	if !validate.UUID(did) {
		response.BadRequest(c.ResponseWriter(), "Invalid division ID")
		return nil, errHandledResponse
	}

	d, err := s.Dependencies.DivisionSvc.GetByID(c.Context(), uuid.MustParse(did))
	if err != nil {
		return nil, err
	}
	if d.TournamentID != tournamentID {
		return nil, division.ErrDivisionNotFound
	}

	return d, nil
}

func (s *Server) getDivisionHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	d, err := s.getOwnedDivision(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	response.SuccessWithResult(c.ResponseWriter(), d)
	return nil
}

func (s *Server) updateDivisionHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	d, err := s.getOwnedDivision(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd division.UpdateDivisionCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ID = d.ID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.DivisionSvc.Update(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) activateDivisionHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	d, err := s.getOwnedDivision(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	err = s.Dependencies.DivisionSvc.Activate(c.Context(), d.ID)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) deactivateDivisionHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	tUUID, err := s.ownedTournament(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	d, err := s.getOwnedDivision(c, tUUID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	err = s.Dependencies.DivisionSvc.Deactivate(c.Context(), d.ID)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}
