package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"sport-grid-be/internal/auth"

	chirtr "github.com/imsab23/platform-be/pkg/http/router/chi"
	"github.com/imsab23/platform-be/pkg/security/identity"
)

func TestNewTeamControllerRegistersTeamRoutesOnce(t *testing.T) {
	r := chirtr.New(chirtr.Options{})
	s := &Server{}
	s.NewTeamController(r)

	base := "/tournaments/0199b88b-cd21-7f3a-bf6c-158dbfae85e7/divisions/0199b88b-cd21-7f3a-bf6c-158dbfae85e8/teams"
	teamID := "0199b88b-cd21-7f3a-bf6c-158dbfae85e9"
	memberID := "0199b88b-cd21-7f3a-bf6c-158dbfae85ea"
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, base + "/"},
		{http.MethodGet, base + "/" + teamID},
		{http.MethodGet, base + "/" + teamID + "/members"},
		{http.MethodPost, base + "/"},
		{http.MethodPost, base + "/" + teamID + "/disband"},
		{http.MethodPost, base + "/" + teamID + "/members"},
		{http.MethodPost, base + "/" + teamID + "/members/" + memberID + "/remove"},
		{http.MethodPost, "/teams/" + teamID + "/leave"},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected protected route to return 403, got %d", rec.Code)
			}
		})
	}
}

func TestNewTeamControllerKeepsUserTypeBoundaries(t *testing.T) {
	r := chirtr.New(chirtr.Options{})
	s := &Server{}
	s.NewTeamController(r)

	base := "/tournaments/0199b88b-cd21-7f3a-bf6c-158dbfae85e7/divisions/0199b88b-cd21-7f3a-bf6c-158dbfae85e8/teams"
	teamID := "0199b88b-cd21-7f3a-bf6c-158dbfae85e9"
	tests := []struct {
		name     string
		userType auth.UserType
		method   string
		path     string
	}{
		{"player cannot read staff route", auth.Player, http.MethodGet, base + "/"},
		{"staff cannot create through player route", auth.User, http.MethodPost, base + "/"},
		{"staff cannot leave through player route", auth.User, http.MethodPost, "/teams/" + teamID + "/leave"},
		{"player cannot disband through staff route", auth.Player, http.MethodPost, base + "/" + teamID + "/disband"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			req = req.WithContext(identity.WithIdentity(req.Context(), &identity.Identity{
				Subject:  "0199b88b-cd21-7f3a-bf6c-158dbfae85eb",
				UserType: string(test.userType),
			}))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected wrong user type to return 403, got %d", rec.Code)
			}
		})
	}
}
