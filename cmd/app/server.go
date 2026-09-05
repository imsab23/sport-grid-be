package app

import (
	"context"

	"sport-grid-be/pkg/auth"
	"sport-grid-be/pkg/client"
	"sport-grid-be/pkg/config"
	"sport-grid-be/pkg/controller"
	"sport-grid-be/pkg/player"
	"sport-grid-be/pkg/storage/postgres"
	"sport-grid-be/pkg/user"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	logger "github.com/imsab23/platform-be/observability/logging"
	"github.com/imsab23/platform-be/pkg/lifecycle"
)

type Server struct {
	runner *lifecycle.Runner
	db     db.DB
}

func NewServer(ctx context.Context) (*Server, error) {
	log, err := logger.NewLogger("app-server")
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	database, err := postgres.InitDB(ctx, cfg)
	if err != nil {
		return nil, err
	}

	userSvc, err := user.NewService(database)
	if err != nil {
		return nil, err
	}

	playerSvc, err := player.NewService(database)
	if err != nil {
		return nil, err
	}

	authSvc, err := auth.NewService(userSvc, playerSvc, database, &cfg.JWT)
	if err != nil {
		return nil, err
	}

	clientSvc, err := client.NewService(database)
	if err != nil {
		return nil, err
	}

	restServer, err := controller.NewServer(&controller.Dependencies{
		UserSvc:   userSvc,
		AuthSvc:   authSvc,
		PlayerSvc: playerSvc,
		ClientSvc: clientSvc,
	}, cfg)
	if err != nil {
		return nil, err
	}

	runner := lifecycle.New(lifecycle.WithLogger(log))
	runner.Add(restServer.Run)

	return &Server{
		runner: runner,
		db:     database,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	defer s.close()

	err := s.runner.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) close() {
	s.db.Close()
}
