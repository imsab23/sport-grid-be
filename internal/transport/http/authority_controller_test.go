package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"sport-grid-be/internal/client"
	"sport-grid-be/internal/role"
	"sport-grid-be/internal/role/permission"
	"sport-grid-be/internal/user"

	"github.com/google/uuid"
	chirtr "github.com/imsab23/platform-be/pkg/http/router/chi"
	"github.com/imsab23/platform-be/pkg/security/identity"
	"github.com/imsab23/platform-be/pkg/util/meta"
)

const (
	testAuthorityUserID   = "0199b88b-cd21-7f3a-bf6c-158dbfae85e7"
	testAuthorityClientID = "0199b88b-cd21-7f3a-bf6c-158dbfae85e8"
	testAuthorityTargetID = "0199b88b-cd21-7f3a-bf6c-158dbfae85e9"
)

type authorityClientService struct{}

func (service *authorityClientService) Create(context.Context, *client.CreateClientCommand) (*client.Client, error) {
	return &client.Client{ID: uuid.MustParse(testAuthorityClientID), Name: "Test Client"}, nil
}

func (service *authorityClientService) GetByID(context.Context, uuid.UUID) (*client.Client, error) {
	return &client.Client{ID: uuid.MustParse(testAuthorityClientID), Name: "Test Client"}, nil
}

func (service *authorityClientService) Search(context.Context, *client.SearchClientQuery) (*client.SearchClientResult, error) {
	return &client.SearchClientResult{Clients: []*client.Client{}, Meta: &meta.Meta{}}, nil
}

func (service *authorityClientService) Suspend(context.Context, uuid.UUID) error  { return nil }
func (service *authorityClientService) Activate(context.Context, uuid.UUID) error { return nil }

type authorityUserService struct{}

func (service *authorityUserService) Search(context.Context, *user.SearchUserQuery) (*user.SearchUserResult, error) {
	return &user.SearchUserResult{Users: []*user.User{}, Meta: &meta.Meta{}}, nil
}

func (service *authorityUserService) Create(context.Context, *user.CreateUserCommand) (*user.User, error) {
	return &user.User{ID: uuid.MustParse(testAuthorityTargetID)}, nil
}

func (service *authorityUserService) GetByID(context.Context, uuid.UUID) (*user.User, error) {
	return &user.User{ID: uuid.MustParse(testAuthorityTargetID)}, nil
}

func (service *authorityUserService) GetByIDForActor(_ context.Context, actor role.Actor, userID uuid.UUID) (*user.User, error) {
	return &user.User{ID: userID, ClientID: actor.ClientID}, nil
}

func (service *authorityUserService) GetByEmail(context.Context, string) (*user.User, error) {
	return nil, nil
}

func (service *authorityUserService) UpdateLastLogin(context.Context, uuid.UUID) error { return nil }
func (service *authorityUserService) UpdatePassword(context.Context, *user.UpdateUserPasswordCommand) error {
	return nil
}
func (service *authorityUserService) ValidatePassword(context.Context, string, string) error {
	return nil
}

func requestWithRole(method string, path string, actorRole role.Role) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	return request.WithContext(identity.WithIdentity(request.Context(), &identity.Identity{
		Subject:     testAuthorityUserID,
		UserID:      testAuthorityUserID,
		UserType:    "user",
		ClientID:    testAuthorityClientID,
		Roles:       []string{string(actorRole)},
		Permissions: permission.ForRole(actorRole),
	}))
}

func TestClientControllerRejectsClientAdmin(t *testing.T) {
	router := chirtr.New(chirtr.Options{})
	server := &Server{Dependencies: &Dependencies{ClientSvc: &authorityClientService{}}}
	server.NewClientController(router)

	testCases := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/clients/"},
		{method: http.MethodGet, path: "/clients/" + testAuthorityClientID},
		{method: http.MethodPost, path: "/clients/" + testAuthorityClientID + "/suspend"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.method+" "+testCase.path, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, requestWithRole(testCase.method, testCase.path, role.ClientAdmin))
			if responseRecorder.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusForbidden)
			}
		})
	}
}

func TestClientControllerAllowsSuperAdminRead(t *testing.T) {
	router := chirtr.New(chirtr.Options{})
	server := &Server{Dependencies: &Dependencies{ClientSvc: &authorityClientService{}}}
	server.NewClientController(router)

	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, requestWithRole(http.MethodGet, "/clients/"+testAuthorityClientID, role.SuperAdmin))
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusOK)
	}
}

func TestUserControllerRejectsOperationalStaff(t *testing.T) {
	router := chirtr.New(chirtr.Options{})
	server := &Server{Dependencies: &Dependencies{UserSvc: &authorityUserService{}}}
	server.NewUserController(router)

	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, requestWithRole(http.MethodGet, "/users/", role.Scorer))
	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusForbidden)
	}
}

func TestUserControllerAllowsClientAdminReadInOwnScope(t *testing.T) {
	router := chirtr.New(chirtr.Options{})
	server := &Server{Dependencies: &Dependencies{UserSvc: &authorityUserService{}}}
	server.NewUserController(router)

	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, requestWithRole(http.MethodGet, "/users/"+testAuthorityTargetID, role.ClientAdmin))
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", responseRecorder.Code, http.StatusOK)
	}
}
