package controller

import (
	"sport-grid-be/pkg/auth"

	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
)

func (s *Server) NewAuthController(r router.Router) {
	r.Group("/auth", func(r router.Router) {
		r.POST("/login", s.loginHandler)
	})
}

func (s *Server) loginHandler(c *router.Ctx) error {
	var cmd auth.LoginUserCommand

	err := c.Bind(&cmd)
	if err != nil {
		return err
	}

	result, err := s.Dependencies.AuthSvc.LoginUser(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), "Logged in successfully", result)
	return nil
}
