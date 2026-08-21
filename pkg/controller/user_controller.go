package controller

import (
	"sport-grid-be/pkg/user"

	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
)

func (s *Server) NewUserController(r router.Router) {
	r.Group("/users", func(r router.Router) {
		r.POST("/", s.createUser)
		r.GET("/{id}", s.getUserHandler)
	})
}

func (s *Server) createUser(c *router.Ctx) error {
	var cmd user.User

	err := c.Bind(&cmd)
	if err != nil {
		return err
	}

	err = s.Dependencies.UserSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter(), "User created successfully")
	return nil
}

func (s *Server) getUserHandler(c *router.Ctx) error {
	id := c.Param("id")
	user, err := s.Dependencies.UserSvc.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), "", user)
	return nil
}
