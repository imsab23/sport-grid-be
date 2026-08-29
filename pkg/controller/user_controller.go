package controller

import (
	"sport-grid-be/pkg/user"

	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

func (s *Server) NewUserController(r router.Router) {
	r.Group("/users", func(r router.Router) {
		r.POST("/", s.createUserHandler)
		r.GET("/", s.searchUserHandler)
		r.GET("/{id}", s.getUserHandler)
	})

}

func (s *Server) createUserHandler(c *router.Ctx) error {
	var cmd user.CreateUserCommand

	err := c.BindJson(&cmd)
	if err != nil {
		return err
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
		return err
	}

	err = c.BindQuery(&meta)
	if err != nil {
		return err
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

	u, err := s.Dependencies.UserSvc.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), u)
	return nil
}
