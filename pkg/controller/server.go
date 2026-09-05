package controller

import (
	"context"
	"sport-grid-be/pkg/auth"
	"sport-grid-be/pkg/client"
	"sport-grid-be/pkg/config"

	"sport-grid-be/pkg/player"
	"sport-grid-be/pkg/role"
	"sport-grid-be/pkg/user"

	logger "github.com/imsab23/platform-be/observability/logging"
	"github.com/imsab23/platform-be/pkg/http/clientip"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	chirtr "github.com/imsab23/platform-be/pkg/http/router/chi"
	"github.com/imsab23/platform-be/pkg/http/server"
	authmw "github.com/imsab23/platform-be/pkg/middleware/auth"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
)

type Server struct {
	router       router.Router
	Dependencies *Dependencies
	cfg          *config.Config
	log          logger.Logger
}

type Dependencies struct {
	UserSvc   user.Service
	AuthSvc   auth.Service
	PlayerSvc player.Service
	ClientSvc client.Service
}

func NewServer(deps *Dependencies, cfg *config.Config) (*Server, error) {
	log, _ := logger.NewLogger("restserver")

	r := chirtr.New(chirtr.Options{
		Logger:               log,
		EnableRequestLogging: true,
	})

	return &Server{
		router:       r,
		Dependencies: deps,
		cfg:          cfg,
		log:          log,
	}, nil
}

func (s *Server) registerRoutes() {
	r := s.router

	r.GET("/", healthHandler)
	s.NewAuthController(r)

	// CMS routes
	r.Group("/api/v1", func(api router.Router) {
		// Public routes.
		s.NewAuthController(api)

		// Protected routes.
		api.Group("", func(protected router.Router) {
			protected.Use(
				wrapNetHTTPMiddleware(
					authmw.New(
						s.Dependencies.AuthSvc.VerifyToken(),
						authmw.WithClientIPFunc(clientip.FromXForwardedFor),
					),
				),
			)

			// Super Admin only — user and client management.
			protected.Group("", func(superAdmin router.Router) {
				superAdmin.Use(wrapNetHTTPMiddleware(authzmw.RequireRole(string(role.SuperAdmin))))
				s.NewUserController(superAdmin)
				s.NewClientController(superAdmin)
			})

			// Client Admin only — tournament staff management.
			protected.Group("", func(clientAdmin router.Router) {
				clientAdmin.Use(wrapNetHTTPMiddleware(authzmw.RequireRole(string(role.ClientAdmin))))
				s.NewStaffController(clientAdmin)
			})

			s.NewPlayerController(protected)
		})
	})
}

func healthHandler(c *router.Ctx) error {
	response.Success(c.ResponseWriter())
	return nil
}

func (s *Server) Run(ctx context.Context) error {
	s.registerRoutes()

	srv, err := server.New(server.DefaultConfig(s.cfg.Server.Addr()), s.router)
	if err != nil {
		return err
	}

	s.log.Info("Starting server on " + s.cfg.Server.Addr())

	err = srv.ListenAndServe(ctx)
	if err != nil {
		return err
	}

	return nil
}
