package controller

import (
	"context"
	"sport-grid-be/internal/auth"
	"sport-grid-be/internal/client"
	"sport-grid-be/internal/config"
	"sport-grid-be/internal/role"

	"sport-grid-be/internal/bracket"
	"sport-grid-be/internal/court"
	"sport-grid-be/internal/division"
	"sport-grid-be/internal/match"
	"sport-grid-be/internal/player"
	"sport-grid-be/internal/registration"
	"sport-grid-be/internal/seeding"
	"sport-grid-be/internal/sport"
	"sport-grid-be/internal/team"
	"sport-grid-be/internal/tournament"
	"sport-grid-be/internal/user"

	logger "github.com/imsab23/platform-be/observability/logging"
	"github.com/imsab23/platform-be/pkg/http/clientip"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	chirtr "github.com/imsab23/platform-be/pkg/http/router/chi"
	"github.com/imsab23/platform-be/pkg/http/server"
	authmw "github.com/imsab23/platform-be/pkg/middleware/auth"
	authzmw "github.com/imsab23/platform-be/pkg/middleware/authz"
)

var (
	RequireUser   = authzmw.RequireUserType(string(auth.User))
	RequirePlayer = authzmw.RequireUserType(string(auth.Player))

	RequireSuperAdmin                 = authzmw.RequireRole(string(role.SuperAdmin))
	RequireClientAdmin                = authzmw.RequireRole(string(role.ClientAdmin))
	RequireRegistrationManager        = authzmw.RequireRole(string(role.RegistrationManager))
	RequirePaymentReviewer            = authzmw.RequireRole(string(role.PaymentReviewer))
	RequireContentAnnouncementManager = authzmw.RequireRole(string(role.ContentAnnouncementManager))
	RequireScorer                     = authzmw.RequireRole(string(role.Scorer))
	RequireReferee                    = authzmw.RequireRole(string(role.Referee))
	RequireTournamentStaff            = authzmw.RequireRole(string(role.TournamentStaff))
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
				chirtr.FromHTTPMiddleware(
					authmw.New(
						s.Dependencies.AuthSvc.VerifyToken(),
						authmw.WithClientIPFunc(clientip.FromXForwardedFor),
					),
				),
			)

			s.NewTeamController(protected)

			// user type only routes
			protected.Group("", func(reqUser router.Router) {
				reqUser.Use(RequireUser)
				s.NewUserController(reqUser)
				s.NewClientController(reqUser)
				s.NewSportController(reqUser)
				s.NewTournamentController(reqUser)
				s.NewDivisionController(reqUser)
				s.NewRegistrationController(reqUser)
				s.NewCourtController(reqUser)
				s.NewSeedingController(reqUser)
				s.NewBracketController(reqUser)
				s.NewMatchController(reqUser)
				s.NewAdminPlayerController(reqUser)
			})

			// player type only routes
			protected.Group("", func(reqPlayer router.Router) {
				reqPlayer.Use(RequirePlayer)
				s.NewPlayerController(reqPlayer)
				s.NewPlayerRegistrationController(reqPlayer)
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
