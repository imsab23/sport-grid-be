package controller

import (
	"sport-grid-be/pkg/client"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

func (s *Server) NewClientController(r router.Router) {
	r.Group("/clients", func(r router.Router) {
		r.POST("/", s.createClientHandler)
		r.GET("/", s.searchClientHandler)
		r.GET("/{id}", s.getClientHandler)
	})
}

func (s *Server) createClientHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	createdBy, err := uuid.Parse(id.Subject)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var cmd client.CreateClientCommand
	err = c.BindJson(&cmd)
	if err != nil {
		return err
	}

	cmd.CreatedBy = createdBy
	err = cmd.Validate()
	if err != nil {
		return err
	}

	result, err := s.Dependencies.ClientSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}

func (s *Server) searchClientHandler(c *router.Ctx) error {
	var (
		query client.SearchClientQuery
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

	result, err := s.Dependencies.ClientSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Clients, result.Meta)
	return nil
}

func (s *Server) getClientHandler(c *router.Ctx) error {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c.ResponseWriter(), "Invalid client ID")
		return nil
	}

	result, err := s.Dependencies.ClientSvc.GetByID(c.Context(), clientID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(c.ResponseWriter(), result)
	return nil
}
