package controller

import (
	"sport-grid-be/internal/player"
	"sport-grid-be/internal/role"
	"sport-grid-be/internal/role/permission"
	"time"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

var requirePlayerView = authzmw.RequirePermission(permission.ViewPlayer)

func (s *Server) NewAdminPlayerController(r router.Router) {
	r.Group("/players", func(r router.Router) {
		r.GET("/", s.searchPlayerHandler, requirePlayerView)
		r.GET("/{id}", s.getPlayerByID, requirePlayerView)
	})
}

func (s *Server) NewPlayerController(r router.Router) {
	r.Group("/player-info", func(r router.Router) {
		r.GET("/me", s.getPlayer)
		r.PUT("/me", s.updateMeHandler)
	})
}

func (s *Server) searchPlayerHandler(c *router.Ctx) error {
	var (
		query player.SearchPlayerQuery
		meta  meta.Meta
	)

	err := c.BindQuery(&query)
	if err != nil {
		return err
	}
	err = c.BindQuery(&meta)
	if err != nil {
		return err
	}
	query.Meta = &meta

	result, err := s.Dependencies.PlayerSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Players, result.Meta)
	return nil
}

func (s *Server) getPlayerByID(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	idUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c.ResponseWriter(), "Invalid player ID")
		return nil
	}

	result, err := s.Dependencies.PlayerSvc.GetByID(c.Context(), idUUID)
	if err != nil {
		return err
	}

	// Super Admin can view any player; others can only view their own profile.
	if !authz.HasRole(id, string(role.SuperAdmin)) && result.ID.String() != id.Subject {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) getPlayer(c *router.Ctx) error {
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

	result, err := s.Dependencies.PlayerSvc.GetByID(c.Context(), playerID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) updateMeHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		return nil
	}

	playerID, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd player.UpdatePlayerCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}

	cmd.ID = playerID
	cmd.UpdatedAt = time.Now().UTC()
	cmd.UpdatedBy = playerID
	prof, err := s.Dependencies.PlayerSvc.Update(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), prof)
	return nil
}
