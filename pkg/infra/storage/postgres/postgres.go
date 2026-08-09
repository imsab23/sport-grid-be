package postgres

import (
	"context"
	"sport-grid-be/pkg/config"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	dbimpl "github.com/imsab23/platform-be/infra/storage/postgres/dbimpl"
)

func InitDB(ctx context.Context, cfg *config.Config) (db.DB, error) {
	conn, err := dbimpl.NewPostgresDB(ctx, db.DatabaseConfig{
		Host:     cfg.PostgresDB.Host,
		Port:     cfg.PostgresDB.Port,
		Database: cfg.PostgresDB.Database,
		Username: cfg.PostgresDB.Username,
		Password: cfg.PostgresDB.Password,
		SSLMode:  cfg.PostgresDB.SSLMode,
	})
	if err != nil {
		return nil, err
	}

	return conn, nil
}
