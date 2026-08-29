package controller

import (
	"sport-grid-be/pkg/player"
	"sport-grid-be/pkg/user"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	"github.com/imsab23/platform-be/pkg/security/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
)

func (s *Server) NewPlayerController(r router.Router) {
	// Admin routes
	r.Group("/players", func(r router.Router) {
		r.GET("/{id}", s.getPlayerByID)
	})

	// Player routes
	r.Group("/player-info", func(r router.Router) {
		r.GET("/me", s.getPlayer)
		r.PUT("/me", s.updateMeHandler)
	})
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

	result, err := s.Dependencies.PlayerSvc.GetByPlayerID(c.Context(), idUUID)
	if err != nil {
		return err
	}

	// Super Admin can view any player; others can only view their own profile.
	if !authz.HasRole(id, user.RoleSuperAdmin) && result.UserID.String() != id.Subject {
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

	userID, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	result, err := s.Dependencies.PlayerSvc.GetByID(c.Context(), userID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) updateMeHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	userID, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd player.UpdatePlayerCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}

	cmd.UserID = userID
	prof, err := s.Dependencies.PlayerSvc.Update(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), prof)
	return nil
}
