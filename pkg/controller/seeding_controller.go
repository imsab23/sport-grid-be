package controller

import (
	"sport-grid-be/pkg/role"
	"sport-grid-be/pkg/seeding"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
)

func (s *Server) NewSeedingController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/divisions/{divisionId}/seedings", func(r router.Router) {
		r.Use(wrapNetHTTPMiddleware(authzmw.RequireAnyRole(
			string(role.SuperAdmin), string(role.ClientAdmin), string(role.TournamentStaff),
		)))
		r.GET("/", s.searchSeedingHandler)
		r.POST("/generate", s.generateSeedingHandler)
		r.POST("/assign", s.assignSeedHandler)
	})
}

func (s *Server) searchSeedingHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	result, err := s.Dependencies.SeedingSvc.GetByDivision(c.Context(), d.ID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) generateSeedingHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	seededBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd seeding.GenerateSeedingCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.DivisionID = d.ID
	cmd.SeededBy = seededBy

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.SeedingSvc.GenerateSeeding(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) assignSeedHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	seededBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd seeding.AssignSeedCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.DivisionID = d.ID
	cmd.SeededBy = seededBy

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.SeedingSvc.AssignSeed(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}
