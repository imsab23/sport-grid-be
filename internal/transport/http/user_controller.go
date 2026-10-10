package controller

import (
	"sport-grid-be/internal/role/permission"
	"sport-grid-be/internal/transport"
	"sport-grid-be/internal/user"

	"github.com/google/uuid"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
	"github.com/imsab23/platform-be/pkg/security/identity"
	apperror "github.com/imsab23/platform-be/pkg/util/error"
	"github.com/imsab23/platform-be/pkg/util/meta"
	"github.com/imsab23/platform-be/pkg/util/validate"
)

var (
	reqUserCreate = authzmw.RequirePermission(string(permission.CreateUser))
	reqUserView   = authzmw.RequirePermission(string(permission.ViewUser))
)

func (server *Server) NewUserController(routerGroup router.Router) {
	routerGroup.Group("/users", func(userRouter router.Router) {
		userRouter.POST("/", server.createUserHandler, reqUserCreate)
		userRouter.GET("/", server.searchUserHandler, reqUserView)
		userRouter.GET("/{id}", server.getUserHandler, reqUserView)
	})
}

func (server *Server) createUserHandler(ctx *router.Ctx) error {
	actor, err := transport.ActorFromIdentity(identity.FromContext(ctx.Context()))
	if err != nil {
		return err
	}

	var command user.CreateUserCommand
	err = ctx.BindJson(&command)
	if err != nil {
		return apperror.ErrBadRequest
	}
	command.Actor = actor

	createdUser, err := server.Dependencies.UserSvc.Create(ctx.Context(), &command)
	if err != nil {
		return err
	}

	response.SuccessWithResult(ctx.ResponseWriter(), createdUser)
	return nil
}

func (server *Server) searchUserHandler(ctx *router.Ctx) error {
	actor, err := transport.ActorFromIdentity(identity.FromContext(ctx.Context()))
	if err != nil {
		return err
	}

	var searchQuery user.SearchUserQuery
	var searchMeta meta.Meta

	err = ctx.BindQuery(&searchQuery)
	if err != nil {
		return apperror.ErrBadRequest
	}

	err = ctx.BindQuery(&searchMeta)
	if err != nil {
		return apperror.ErrBadRequest
	}

	searchQuery.Meta = &searchMeta
	searchQuery.Actor = actor

	searchResult, err := server.Dependencies.UserSvc.Search(ctx.Context(), &searchQuery)
	if err != nil {
		return err
	}

	response.SuccessWithMeta(ctx.ResponseWriter(), searchResult.Users, searchResult.Meta)
	return nil
}

func (server *Server) getUserHandler(ctx *router.Ctx) error {
	actor, err := transport.ActorFromIdentity(identity.FromContext(ctx.Context()))
	if err != nil {
		return err
	}

	userIDText := ctx.Param("id")
	if !validate.CheckUUIDv7Format(userIDText) {
		return apperror.ErrBadRequest
	}

	userID, err := uuid.Parse(userIDText)
	if err != nil {
		return apperror.ErrBadRequest
	}

	userRecord, err := server.Dependencies.UserSvc.GetByIDForActor(ctx.Context(), actor, userID)
	if err != nil {
		return err
	}

	response.SuccessWithResult(ctx.ResponseWriter(), userRecord)
	return nil
}

func (server *Server) updateUserPasswordHandler(ctx *router.Ctx) error {
	signedInUser := identity.FromContext(ctx.Context())
	if signedInUser == nil {
		return apperror.ErrForbidden
	}

	var command user.UpdateUserPasswordCommand
	err := ctx.BindJson(&command)
	if err != nil {
		return apperror.ErrBadRequest
	}

	userID, err := uuid.Parse(signedInUser.UserID)
	if err != nil {
		return apperror.ErrForbidden
	}
	command.ID = userID

	err = command.Validate()
	if err != nil {
		return err
	}

	err = server.Dependencies.UserSvc.UpdatePassword(ctx.Context(), &command)
	if err != nil {
		return err
	}

	response.SuccessWithMessage(ctx.ResponseWriter(), "Password updated successfully")
	return nil
}
