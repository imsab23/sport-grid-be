package controller

import (
	"sport-grid-be/pkg/role"
	"sport-grid-be/pkg/user"

	"github.com/google/uuid"

	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	"github.com/imsab23/platform-be/pkg/security/identity"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

func (s *Server) NewUserController(r router.Router) {
	r.Group("/users", func(r router.Router) {
		r.POST("/", s.createUserHandler)
		r.GET("/", s.searchUserHandler)
		r.GET("/{id}", s.getUserHandler)
	})
}

func (s *Server) createUserHandler(c *router.Ctx) error {
	signedInUser := identity.FromContext(c.Context())
	var cmd user.CreateUserCommand

	err := c.BindJson(&cmd)
	if err != nil {
		return apperror.ErrBadRequest
	}

	// Super Admin can create any user; Client Admin can only create Tournament Staff users.
	switch signedInUser.Roles[0] {
	case string(role.SuperAdmin):
	case string(role.ClientAdmin):
		if cmd.Role != role.TournamentStaff {
			return apperror.ErrForbidden
		}
	default:
		return apperror.ErrForbidden
	}

	err = cmd.Validate()
	if err != nil {
		return err
	}

	_, err = s.Dependencies.UserSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithMessage(c.ResponseWriter(), "User created successfully")
	return nil
}

func (s *Server) searchUserHandler(c *router.Ctx) error {
	var (
		query user.SearchUserQuery
		meta  meta.Meta
	)

	err := c.BindQuery(&query)
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = c.BindQuery(&meta)
	if err != nil {
		return apperror.ErrBadRequest
	}

	query.Meta = &meta

	result, err := s.Dependencies.UserSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Users, result.Meta)
	return nil
}

func (s *Server) getUserHandler(c *router.Ctx) error {
	id := c.Param("id")
	isUUID := validate.UUID(id)
	if !isUUID {
		return apperror.ErrBadRequest
	}

	u, err := s.Dependencies.UserSvc.GetByID(c.Context(), uuid.MustParse(id))
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), u)
	return nil
}

func (s *Server) updateUserPasswordHandler(c *router.Ctx) error {
	signedInUser := identity.FromContext(c.Context())
	var cmd user.UpdateUserPasswordCommand

	err := c.BindJson(&cmd)
	if err != nil {
		return apperror.ErrBadRequest
	}

	cmd.ID = uuid.MustParse(signedInUser.UserID)

	err = cmd.Validate()
	if err != nil {
		return err
	}

	err = s.Dependencies.UserSvc.UpdatePassword(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithMessage(c.ResponseWriter(), "Password updated successfully")
	return nil
}
