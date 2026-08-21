package postgres

import (
	"context"
	"sport-grid-be/pkg/config"

	db "github.com/imsab23/platform-be/infra/storage/postgres"
	dbimpl "github.com/imsab23/platform-be/infra/storage/postgres/dbimpl"
	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
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

var allMigrations = []migration.Migration{
	{
		Version: 1,
		Name:    "create_users_table",
		Up: func(ctx context.Context, exec migration.ExecFunc) error {
			return exec(ctx, schema.Table{
				Name: "users",
				Columns: []schema.Column{
					{Name: "id", Type: schema.TypeUUID, PrimaryKey: true, Default: "gen_random_uuid()"},
					{Name: "first_name", Type: schema.TypeText, NotNull: true},
					{Name: "middle_name", Type: schema.TypeText, NotNull: false},
					{Name: "last_name", Type: schema.TypeText, NotNull: true},
					{Name: "email", Type: schema.TypeText, NotNull: true, Unique: true},
					{Name: "contact_number", Type: schema.TypeText, NotNull: false},
					{Name: "password", Type: schema.TypeText, NotNull: true},
					{Name: "created_by", Type: schema.TypeText, NotNull: true},
					{Name: "updated_by", Type: schema.TypeText, NotNull: true},
					{Name: "created_at", Type: schema.TypeText, NotNull: true},
					{Name: "updated_at", Type: schema.TypeText, NotNull: true},
				},
			}.DDL())
		},
		Down: func(ctx context.Context, exec migration.ExecFunc) error {
			return exec(ctx, "DROP TABLE IF EXISTS users;")
		},
	},
}

func Migration(ctx context.Context, db db.DB) error {
	runner := migration.NewRunner(db)

	err := runner.Run(ctx, allMigrations)
	if err != nil {
		return err
	}

	return nil
}
