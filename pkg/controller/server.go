package controller

import (
	"context"
	"sport-grid-be/pkg/config"

	logger "github.com/imsab23/platform-be/observability/logging"
	"github.com/imsab23/platform-be/pkg/http/response"
	"github.com/imsab23/platform-be/pkg/http/router"
	chirtr "github.com/imsab23/platform-be/pkg/http/router/chi"
	"github.com/imsab23/platform-be/pkg/http/server"
	"github.com/imsab23/platform-be/pkg/middleware/secheaders"
)

type Server struct {
	router       router.Router
	Dependencies *Dependencies
	cfg          *config.Config
	log          logger.Logger
}

type Dependencies struct {
}

func NewServer(deps *Dependencies, cfg *config.Config) (*Server, error) {
	log, _ := logger.NewLogger("restserver")

	r := chirtr.New(chirtr.Options{
		Logger: log,
		SecHeadersConfig: secheaders.Config{
			HSTSMaxAge:            31536000,
			HSTSIncludeSubDomains: true,
		},
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

}

func healthHandler(c *router.Ctx) error {
	response.Success(c.ResponseWriter(), "Hello")
	return nil
}

func (s *Server) Run(ctx context.Context) error {
	s.registerRoutes()

	srv, err := server.New(server.DefaultConfig("localhost:9000"), s.router)
	if err != nil {
		return err
	}

	s.log.Info("Starting server on port 9000")

	err = srv.ListenAndServe(ctx)
	if err != nil {
		return err
	}

	return nil
}
