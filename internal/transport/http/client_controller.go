package controller

import (
	"sport-grid-be/internal/client"
	"sport-grid-be/internal/role/permission"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

var (
	reqClientCreate   = authzmw.RequirePermission(string(permission.CreateClient))
	reqClientView     = authzmw.RequirePermission(string(permission.ViewClient))
	reqClientUpdate   = authzmw.RequirePermission(string(permission.UpdateClient))
	reqClientDelete   = authzmw.RequirePermission(string(permission.DeleteClient))
	reqClientActivate = authzmw.RequirePermission(string(permission.ActivateClient))
	reqClientSuspend  = authzmw.RequirePermission(string(permission.SuspendClient))
)

func (server *Server) NewClientController(routerGroup router.Router) {
	routerGroup.Group("/clients", func(clientRouter router.Router) {

		clientRouter.POST("/", server.createClientHandler, reqClientCreate)
		clientRouter.GET("/", server.searchClientHandler, reqClientView)
		clientRouter.GET("/{id}", server.getClientHandler, reqClientView)
		clientRouter.POST("/{id}/activate", server.activateClientHandler, reqClientActivate)
		clientRouter.POST("/{id}/suspend", server.suspendClientHandler, reqClientSuspend)
	})
}

func (server *Server) createClientHandler(ctx *router.Ctx) error {
	identityRecord := identity.FromContext(ctx.Context())
	if identityRecord == nil {
		return apperror.ErrForbidden
	}

	createdBy, err := uuid.Parse(identityRecord.Subject)
	if err != nil {
		return apperror.ErrForbidden
	}

	var command client.CreateClientCommand
	err = ctx.BindJson(&command)
	if err != nil {
		return err
	}

	command.CreatedBy = createdBy
	err = command.Validate()
	if err != nil {
		return err
	}

	createdClient, err := server.Dependencies.ClientSvc.Create(ctx.Context(), &command)
	if err != nil {
		return err
	}

	response.SuccessWithResult(ctx.ResponseWriter(), createdClient)
	return nil
}

func (server *Server) searchClientHandler(ctx *router.Ctx) error {
	var searchQuery client.SearchClientQuery
	var searchMeta meta.Meta

	err := ctx.BindQuery(&searchQuery)
	if err != nil {
		return err
	}

	err = ctx.BindQuery(&searchMeta)
	if err != nil {
		return err
	}
	searchQuery.Meta = &searchMeta

	searchResult, err := server.Dependencies.ClientSvc.Search(ctx.Context(), &searchQuery)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(ctx.ResponseWriter(), searchResult.Clients, searchResult.Meta)
	return nil
}

func (server *Server) getClientHandler(ctx *router.Ctx) error {
	clientID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return apperror.ErrBadRequest
	}

	clientRecord, err := server.Dependencies.ClientSvc.GetByID(ctx.Context(), clientID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(ctx.ResponseWriter(), clientRecord)
	return nil
}

func (server *Server) activateClientHandler(ctx *router.Ctx) error {
	clientID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = server.Dependencies.ClientSvc.Activate(ctx.Context(), clientID)
	if err != nil {
		return err
	}

	clientRecord, err := server.Dependencies.ClientSvc.GetByID(ctx.Context(), clientID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(ctx.ResponseWriter(), clientRecord)
	return nil
}

func (server *Server) suspendClientHandler(ctx *router.Ctx) error {
	clientID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = server.Dependencies.ClientSvc.Suspend(ctx.Context(), clientID)
	if err != nil {
		return err
	}

	clientRecord, err := server.Dependencies.ClientSvc.GetByID(ctx.Context(), clientID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(ctx.ResponseWriter(), clientRecord)
	return nil
}
