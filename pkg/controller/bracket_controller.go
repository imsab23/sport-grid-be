package controller

import (
	"sport-grid-be/pkg/bracket"
	"sport-grid-be/pkg/role"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

func (s *Server) NewBracketController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/divisions/{divisionId}/bracket", func(r router.Router) {
		r.Use(wrapNetHTTPMiddleware(authzmw.RequireAnyRole(
			string(role.SuperAdmin), string(role.ClientAdmin), string(role.TournamentStaff),
		)))
		r.GET("/{id}", s.getBracketHandler)
		r.POST("/generate", s.generateBracketHandler)
		r.POST("/{id}/reset", s.resetBracketHandler)
	})
}

// bracketResult bundles the bracket with its nodes for a single response.
type bracketResult struct {
	*bracket.Bracket
	Nodes []*bracket.BracketNode `json:"nodes"`
}

func (s *Server) generateBracketHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	generatedBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	result, err := s.Dependencies.BracketSvc.Generate(c.Context(), &bracket.GenerateBracketCommand{
		DivisionID:  d.ID,
		GeneratedBy: generatedBy,
	})
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) getBracketHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	bid := c.Param("id")
	if !validate.UUID(bid) {
		response.BadRequest(c.ResponseWriter(), "Invalid bracket ID")
		return nil
	}

	b, err := s.Dependencies.BracketSvc.GetByID(c.Context(), uuid.MustParse(bid))
	if err != nil {
		return err
	}
	if b.DivisionID != d.ID {
		return bracket.ErrBracketNotFound
	}

	nodes, err := s.Dependencies.BracketSvc.GetNodes(c.Context(), b.ID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), &bracketResult{Bracket: b, Nodes: nodes})
	return nil
}

func (s *Server) resetBracketHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	bid := c.Param("id")
	if !validate.UUID(bid) {
		response.BadRequest(c.ResponseWriter(), "Invalid bracket ID")
		return nil
	}

	existing, err := s.Dependencies.BracketSvc.GetByID(c.Context(), uuid.MustParse(bid))
	if err != nil {
		return err
	}
	if existing.DivisionID != d.ID {
		return bracket.ErrBracketNotFound
	}

	generatedBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd bracket.ResetBracketCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ID = existing.ID
	cmd.GeneratedBy = generatedBy

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.BracketSvc.Reset(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}
