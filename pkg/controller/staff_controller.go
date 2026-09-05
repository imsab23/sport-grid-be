package controller

import (
	"sport-grid-be/pkg/authz"
	"sport-grid-be/pkg/role"
	"sport-grid-be/pkg/user"

	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

func (s *Server) NewStaffController(r router.Router) {
	r.Group("/staff", func(r router.Router) {
		r.POST("/", s.createStaffHandler)
		r.GET("/", s.searchStaffHandler)
	})
}

// CreateStaffRequest intentionally excludes role/client_id — both are
// derived server-side from the authenticated Client Admin's own identity.
type CreateStaffRequest struct {
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	ContactNumber *string `json:"contact_number"`
	Password      string  `json:"password"`
}

func (s *Server) createStaffHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	clientID, err := authz.ClientID(id)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var req CreateStaffRequest
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
		Role:          role.TournamentStaff,
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

	response.SuccessWithMessage(c.ResponseWriter(), "Tournament staff created successfully")
	return nil
}

func (s *Server) searchStaffHandler(c *router.Ctx) error {
	id := identity.FromContext(c.Context())
	if id == nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	clientID, err := authz.ClientID(id)
	if err != nil {
		response.Forbidden(c.ResponseWriter())
		return nil
	}

	var (
		query user.SearchUserQuery
		m     meta.Meta
	)
	err = c.BindQuery(&query)
	if err != nil {
		return err
	}
	err = c.BindQuery(&m)
	if err != nil {
		return err
	}
	query.Meta = &m

	// Tenant isolation: staff search is always scoped to the caller's own client.
	staffRole := role.TournamentStaff
	query.ClientID = &clientID
	query.Role = &staffRole

	result, err := s.Dependencies.UserSvc.Search(c.Context(), &query)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(c.ResponseWriter(), result.Users, result.Meta)
	return nil
}
