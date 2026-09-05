package controller

import (
	"sport-grid-be/pkg/auth"

	"github.com/imsab23/platform-be/pkg/http/clientip"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
)

func (s *Server) NewAuthController(r router.Router) {
	r.Group("/auth", func(r router.Router) {
		r.POST("/register", s.registerHandler)
		r.POST("/login", s.loginHandler)
		r.POST("/refresh", s.refreshHandler)
	})
}

func (s *Server) registerHandler(c *router.Ctx) error {
	var cmd auth.RegisterUserCommand

	err := c.BindJson(&cmd)
	if err != nil {
		return err
	}

	err = s.Dependencies.AuthSvc.Register(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithMessage(c.ResponseWriter(), "Registration successful")
	return nil
}

func (s *Server) loginHandler(c *router.Ctx) error {
	var cmd auth.LoginUserCommand

	err := c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ClientIP = clientip.FromXForwardedFor(c.Request())

	result, err := s.Dependencies.AuthSvc.LoginUser(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) refreshHandler(c *router.Ctx) error {
	var cmd auth.RefreshTokenCommand

	err := c.BindJson(&cmd)
	if err != nil {
		return err
	}
	cmd.ClientIP = clientip.FromXForwardedFor(c.Request())

	result, err := s.Dependencies.AuthSvc.RefreshToken(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}
