package controller

import (
	"sport-grid-be/pkg/division"
	"sport-grid-be/pkg/role"
	"sport-grid-be/pkg/team"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

// ownedDivision resolves and authorizes the parent tournament, then verifies the division
// in the path belongs to it.
func (s *Server) ownedDivision(c *router.Ctx, identityID *identity.Identity) (*division.Division, error) {
	tUUID, err := s.ownedTournament(c, identityID)
	if err != nil {
		return nil, err
	}

	did := c.Param("divisionId")
	if !validate.UUID(did) {
		response.BadRequest(c.ResponseWriter(), "Invalid division ID")
		return nil, errHandledResponse
	}

	d, err := s.Dependencies.DivisionSvc.GetByID(c.Context(), uuid.MustParse(did))
	if err != nil {
		return nil, err
	}
	if d.TournamentID != tUUID {
		return nil, division.ErrDivisionNotFound
	}

	return d, nil
}

// getOwnedTeam fetches a team and verifies it belongs to the division in the path.
func (s *Server) getOwnedTeam(c *router.Ctx, divisionID uuid.UUID) (*team.Team, error) {
	tid := c.Param("id")
	if !validate.UUID(tid) {
		response.BadRequest(c.ResponseWriter(), "Invalid team ID")
		return nil, errHandledResponse
	}

	t, err := s.Dependencies.TeamSvc.GetByID(c.Context(), uuid.MustParse(tid))
	if err != nil {
		return nil, err
	}
	if t.DivisionID != divisionID {
		return nil, team.ErrTeamNotFound
	}

	return t, nil
}

// NewTeamController registers staff-facing team management routes (reads + roster admin).
func (s *Server) NewTeamController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/divisions/{divisionId}/teams", func(r router.Router) {
		r.GET("/", s.searchTeamHandler)
		r.GET("/{id}", s.getTeamHandler)
		r.GET("/{id}/members", s.listTeamMembersHandler)

		r.Group("", func(r router.Router) {
			r.Use(wrapNetHTTPMiddleware(authzmw.RequireAnyRole(
				string(role.SuperAdmin), string(role.ClientAdmin), string(role.TournamentStaff),
			)))
			r.POST("/{id}/disband", s.disbandTeamHandler)
			r.POST("/{id}/members", s.addTeamMemberHandler)
			r.POST("/{id}/members/{memberId}/remove", s.removeTeamMemberHandler)
		})
	})
}

// NewPlayerTeamController registers player-facing, self-service team creation.
func (s *Server) NewPlayerTeamController(r router.Router) {
	r.Group("/tournaments/{tournamentId}/divisions/{divisionId}/teams", func(r router.Router) {
		r.POST("/", s.createTeamHandler)
	})
	r.Group("/teams", func(r router.Router) {
		r.POST("/{id}/leave", s.leaveTeamHandler)
	})
}

func (s *Server) createTeamHandler(c *router.Ctx) error {
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

	var cmd team.CreateTeamCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.TournamentID = uuid.MustParse(tid)
	cmd.DivisionID = uuid.MustParse(did)
	cmd.CreatorPlayerID = playerID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.TeamSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) leaveTeamHandler(c *router.Ctx) error {
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

	tid := c.Param("id")
	if !validate.UUID(tid) {
		response.BadRequest(c.ResponseWriter(), "Invalid team ID")
		return nil
	}
	teamID := uuid.MustParse(tid)

	members, err := s.Dependencies.TeamSvc.ListMembers(c.Context(), teamID)
	if err != nil {
		return err
	}

	var memberID uuid.UUID
	found := false
	for _, m := range members {
		if m.PlayerID == playerID {
			memberID = m.ID
			found = true
			break
		}
	}
	if !found {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	err = s.Dependencies.TeamSvc.RemoveMember(c.Context(), teamID, memberID)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) searchTeamHandler(c *router.Ctx) error {
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
		query team.SearchTeamQuery
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
	query.DivisionID = d.ID

	result, err := s.Dependencies.TeamSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Teams, result.Meta)
	return nil
}

func (s *Server) getTeamHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	t, err := s.getOwnedTeam(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	response.SuccessWithResult(c.ResponseWriter(), t)
	return nil
}

func (s *Server) listTeamMembersHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	t, err := s.getOwnedTeam(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	members, err := s.Dependencies.TeamSvc.ListMembers(c.Context(), t.ID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), members)
	return nil
}

func (s *Server) disbandTeamHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	t, err := s.getOwnedTeam(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	err = s.Dependencies.TeamSvc.Disband(c.Context(), t.ID)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) addTeamMemberHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	t, err := s.getOwnedTeam(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	var cmd team.AddTeamMemberCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.TeamID = t.ID

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.TeamSvc.AddMember(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) removeTeamMemberHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	d, err := s.ownedDivision(c, id)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	t, err := s.getOwnedTeam(c, d.ID)
	if err != nil {
		return handleOwnedTournamentErr(c, err)
	}

	mid := c.Param("memberId")
	if !validate.UUID(mid) {
		response.BadRequest(c.ResponseWriter(), "Invalid member ID")
		return nil
	}

	err = s.Dependencies.TeamSvc.RemoveMember(c.Context(), t.ID, uuid.MustParse(mid))
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}
