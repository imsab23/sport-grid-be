package controller

import (
	"context"
	"sport-grid-be/pkg/auth"
	"sport-grid-be/pkg/client"
	"sport-grid-be/pkg/config"

	"sport-grid-be/pkg/bracket"
	"sport-grid-be/pkg/court"
	"sport-grid-be/pkg/division"
	"sport-grid-be/pkg/match"
	"sport-grid-be/pkg/player"
	"sport-grid-be/pkg/registration"
	"sport-grid-be/pkg/seeding"
	"sport-grid-be/pkg/sport"
	"sport-grid-be/pkg/team"
	"sport-grid-be/pkg/tournament"
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
	UserSvc         user.Service
	AuthSvc         auth.Service
	PlayerSvc       player.Service
	ClientSvc       client.Service
	SportSvc        sport.Service
	TournamentSvc   tournament.Service
	DivisionSvc     division.Service
	RegistrationSvc registration.Service
	TeamSvc         team.Service
	CourtSvc        court.Service
	SeedingSvc      seeding.Service
	BracketSvc      bracket.Service
	MatchSvc        match.Service
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

			protected.Group("", func(reqUser router.Router) {
				reqUser.Use(wrapNetHTTPMiddleware(authzmw.RequireUserType(string(auth.User))))
				s.NewUserController(reqUser)
				s.NewClientController(reqUser)
				s.NewSportController(reqUser)
				s.NewTournamentController(reqUser)
				s.NewDivisionController(reqUser)
				s.NewRegistrationController(reqUser)
				s.NewTeamController(reqUser)
				s.NewCourtController(reqUser)
				s.NewSeedingController(reqUser)
				s.NewBracketController(reqUser)
				s.NewMatchController(reqUser)
			})

			protected.Group("", func(reqPlayer router.Router) {
				reqPlayer.Use(wrapNetHTTPMiddleware(authzmw.RequireUserType(string(auth.Player))))
				s.NewPlayerController(reqPlayer)
				s.NewPlayerRegistrationController(reqPlayer)
				s.NewPlayerTeamController(reqPlayer)
			})

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
