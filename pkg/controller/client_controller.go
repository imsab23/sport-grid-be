package controller

import (
	"sport-grid-be/pkg/client"
	"sport-grid-be/pkg/role"
	"sport-grid-be/pkg/user"

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
		r.PATCH("/{id}/suspend", s.suspendClientHandler)
		r.PATCH("/{id}/activate", s.activateClientHandler)
		r.POST("/{id}/admins", s.createClientAdminHandler)
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

func (s *Server) suspendClientHandler(c *router.Ctx) error {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c.ResponseWriter(), "Invalid client ID")
		return nil
	}

	err = s.Dependencies.ClientSvc.Suspend(c.Context(), clientID)
	if err != nil {
		return err
	}

	response.SuccessWithMessage(c.ResponseWriter(), "Client suspended")
	return nil
}

func (s *Server) activateClientHandler(c *router.Ctx) error {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c.ResponseWriter(), "Invalid client ID")
		return nil
	}

	err = s.Dependencies.ClientSvc.Activate(c.Context(), clientID)
	if err != nil {
		return err
	}

	response.SuccessWithMessage(c.ResponseWriter(), "Client activated")
	return nil
}

// CreateClientAdminRequest intentionally excludes role/client_id — both are
// derived server-side to prevent mass assignment and privilege escalation.
type CreateClientAdminRequest struct {
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	ContactNumber *string `json:"contact_number"`
	Password      string  `json:"password"`
}

func (s *Server) createClientAdminHandler(c *router.Ctx) error {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c.ResponseWriter(), "Invalid client ID")
		return nil
	}

	// Ensure the client exists before attaching an admin to it.
	_, err = s.Dependencies.ClientSvc.GetByID(c.Context(), clientID)
	if err != nil {
		return err
	}

	var req CreateClientAdminRequest
	err = c.BindJson(&req)
	if err != nil {
		return err
	}

	cmd := user.CreateUserCommand{
		FirstName:     req.FirstName,
		MiddleName:    req.MiddleName,
		LastName:      req.LastName,
		Email:         req.Email,
		ContactNumber: req.ContactNumber,
		Password:      req.Password,
		Role:          role.ClientAdmin,
		ClientID:      &clientID,
	}
	err = cmd.Validate()
	if err != nil {
		return err
	}

	_, err = s.Dependencies.UserSvc.Create(c.Context(), &cmd)
	if err != nil {
		return err
	}

	response.SuccessWithMessage(c.ResponseWriter(), "Client admin created successfully")
	return nil
}
