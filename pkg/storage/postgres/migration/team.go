package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var TeamMigration = migration.Migration{
	Version: 12,
	Name:    "create_teams_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		if err := exec(ctx, schema.Table{
			Name:              "teams",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "tournament_id", Type: schema.TypeUUID, NotNull: true, References: "tournaments(id)"},
				{Name: "division_id", Type: schema.TypeUUID, NotNull: true, References: "tournament_divisions(id)"},
				{Name: "registration_id", Type: schema.TypeUUID, References: "tournament_registrations(id)"},
				{Name: "name", Type: schema.TypeText, NotNull: true},
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'ACTIVE'"},
			},
		}.DDL()); err != nil {
			return err
		}
		return exec(ctx, schema.Table{
			Name: "team_members",
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "team_id", Type: schema.TypeUUID, NotNull: true, References: "teams(id)"},
				{Name: "player_id", Type: schema.TypeUUID, NotNull: true, References: "players(id)"},
				{Name: "role", Type: schema.TypeText, NotNull: true, Default: "'MEMBER'"},
				{Name: "created_at", Type: schema.TypeTimestamptz, NotNull: true, Default: "NOW()"},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		if err := exec(ctx, "DROP TABLE IF EXISTS team_members;"); err != nil {
			return err
		}
		return exec(ctx, "DROP TABLE IF EXISTS teams;")
	},
}
