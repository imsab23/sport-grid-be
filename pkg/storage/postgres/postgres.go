package postgres

import (
	"context"
	"sport-grid-be/pkg/config"

	mig "sport-grid-be/pkg/storage/postgres/migration"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	dbimpl "github.com/imsab23/platform-be/infra/storage/postgres/dbimpl"
	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
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

	err = Migration(ctx, conn)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Migration(ctx context.Context, db db.DB) error {
	runner := migration.NewRunner(db)

	err := runner.Run(ctx, []migration.Migration{
		mig.UserMigation,
		mig.PlayerProfileMigration,
		mig.UserPasswordResetToken,
		mig.AlterPlayersDateOfBirth,
		mig.AlterUsersAddRoleClientID,
		mig.ClientsMigration,
		mig.AlterUsersClientIDFK,
		mig.AlterPlayersAddCredentials,
	})
	if err != nil {
		return err
	}

	return nil
}
