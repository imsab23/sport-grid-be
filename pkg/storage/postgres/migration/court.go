package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var CourtMigration = migration.Migration{
	Version: 13,
	Name:    "create_courts_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name:              "courts",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "tournament_id", Type: schema.TypeUUID, NotNull: true, References: "tournaments(id)"},
				{Name: "name", Type: schema.TypeText, NotNull: true},
				// AVAILABLE | OCCUPIED | MAINTENANCE | CLOSED
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'AVAILABLE'"},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS courts;")
	},
}
