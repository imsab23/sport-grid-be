package controller

import (
	"sport-grid-be/pkg/role"
	"sport-grid-be/pkg/sport"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

func (s *Server) NewSportController(r router.Router) {
	r.Group("/sports", func(r router.Router) {
		r.GET("/", s.searchSportHandler)
		r.GET("/{id}", s.getSportHandler)

		r.Group("", func(r router.Router) {
			r.Use(wrapNetHTTPMiddleware(authzmw.RequireRole(string(role.SuperAdmin))))
			r.POST("/", s.createSportHandler)
			r.PUT("/{id}", s.updateSportHandler)
			r.POST("/{id}/activate", s.activateSportHandler)
			r.POST("/{id}/deactivate", s.deactivateSportHandler)
		})
	})
}

func (s *Server) createSportHandler(c *router.Ctx) error {
	var cmd sport.CreateSportCommand
	err := c.BindJson(&cmd)
	if err != nil {
		return err
	}

	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.SportSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) searchSportHandler(c *router.Ctx) error {
	var (
		query sport.SearchSportQuery
		m     meta.Meta
	)

	err := c.BindQuery(&query)
	if err != nil {
		return err
	}
	err = c.BindQuery(&m)
	if err != nil {
		return err
	}
	query.Meta = &m

	result, err := s.Dependencies.SportSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Sports, result.Meta)
	return nil
}

func (s *Server) getSportHandler(c *router.Ctx) error {
	id := c.Param("id")
	if !validate.UUID(id) {
		response.BadRequest(c.ResponseWriter(), "Invalid sport ID")
		return nil
	}

	result, err := s.Dependencies.SportSvc.GetByID(c.Context(), uuid.MustParse(id))
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) updateSportHandler(c *router.Ctx) error {
	id := c.Param("id")
	if !validate.UUID(id) {
		response.BadRequest(c.ResponseWriter(), "Invalid sport ID")
		return nil
	}

	var cmd sport.UpdateSportCommand
	err := c.BindJson(&cmd)
	if err != nil {
		return err
	}

	err = cmd.Validate()
	if err != nil {
		return err
	}
	cmd.ID = uuid.MustParse(id)

	result, err := s.Dependencies.SportSvc.Update(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) activateSportHandler(c *router.Ctx) error {
	id := c.Param("id")
	if !validate.UUID(id) {
		response.BadRequest(c.ResponseWriter(), "Invalid sport ID")
		return nil
	}

	err := s.Dependencies.SportSvc.Activate(c.Context(), uuid.MustParse(id))
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) deactivateSportHandler(c *router.Ctx) error {
	id := c.Param("id")
	if !validate.UUID(id) {
		response.BadRequest(c.ResponseWriter(), "Invalid sport ID")
		return nil
	}

	err := s.Dependencies.SportSvc.Deactivate(c.Context(), uuid.MustParse(id))
	if err != nil {
		return err
	}

	response.Success(c.ResponseWriter())
	return nil
}
