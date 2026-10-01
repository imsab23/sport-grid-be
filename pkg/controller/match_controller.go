package controller

import (
	"sport-grid-be/pkg/match"
	"sport-grid-be/pkg/role"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

func (s *Server) NewMatchController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/divisions/{divisionId}/matches", func(r router.Router) {
		r.Use(wrapNetHTTPMiddleware(authzmw.RequireAnyRole(
			string(role.SuperAdmin), string(role.ClientAdmin), string(role.TournamentStaff),
		)))
		r.GET("/", s.searchMatchHandler)
		r.GET("/{id}", s.getMatchHandler)
		r.GET("/{id}/games", s.getMatchGamesHandler)
		r.POST("/generate", s.generateMatchesHandler)
		r.POST("/{id}/schedule", s.scheduleMatchHandler)
		r.POST("/{id}/start", s.startMatchHandler)
		r.POST("/{id}/games", s.recordMatchGameHandler)
		r.POST("/{id}/finalize", s.finalizeMatchHandler)
		r.POST("/{id}/cancel", s.cancelMatchHandler)
	})
}

// getOwnedMatch fetches a match and verifies it belongs to the division in the path.
func (s *Server) getOwnedMatch(c *router.Ctx, divisionID uuid.UUID) (*match.Match, error) {
	mid := c.Param("id")
	if !validate.UUID(mid) {
		response.BadRequest(c.ResponseWriter(), "Invalid match ID")
		return nil, errHandledResponse
	}

	m, err := s.Dependencies.MatchSvc.GetByID(c.Context(), uuid.MustParse(mid))
	if err != nil {
		return nil, err
	}
	if m.DivisionID != divisionID {
		return nil, match.ErrMatchNotFound
	}

	return m, nil
}

func (s *Server) generateMatchesHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var body struct {
		BracketID uuid.UUID `json:"bracket_id"`
	}
	err = c.BindJson(&body)
	if err != nil {
		return err
	}

	b, err := s.Dependencies.BracketSvc.GetByID(c.Context(), body.BracketID)
	if err != nil {
		return err
	}
	if b.DivisionID != d.ID {
		return match.ErrMatchNotFound
	}

	result, err := s.Dependencies.MatchSvc.GenerateMatches(c.Context(), b.ID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) searchMatchHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var (
		query match.SearchMatchQuery
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
	query.TournamentID = d.TournamentID
	query.DivisionID = &d.ID

	result, err := s.Dependencies.MatchSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Matches, result.Meta)
	return nil
}

func (s *Server) getMatchHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	m, err := s.getOwnedMatch(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	response.SuccessWithResult(c.ResponseWriter(), m)
	return nil
}

func (s *Server) getMatchGamesHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	m, err := s.getOwnedMatch(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	games, err := s.Dependencies.MatchSvc.GetGames(c.Context(), m.ID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), games)
	return nil
}

func (s *Server) scheduleMatchHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	m, err := s.getOwnedMatch(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd match.ScheduleMatchCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.MatchID = m.ID

	result, err := s.Dependencies.MatchSvc.Schedule(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) startMatchHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	m, err := s.getOwnedMatch(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	err = s.Dependencies.MatchSvc.Start(c.Context(), m.ID)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) recordMatchGameHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	m, err := s.getOwnedMatch(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd match.RecordGameCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.MatchID = m.ID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.MatchSvc.RecordGame(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) finalizeMatchHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	m, err := s.getOwnedMatch(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	finalizedBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd match.FinalizeMatchCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.MatchID = m.ID
	cmd.FinalizedBy = finalizedBy

	err = cmd.Validate()
	if err != nil {
		return err
	}

	err = s.Dependencies.MatchSvc.Finalize(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) cancelMatchHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	m, err := s.getOwnedMatch(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd match.CancelMatchCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.MatchID = m.ID

	err = s.Dependencies.MatchSvc.Cancel(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}
