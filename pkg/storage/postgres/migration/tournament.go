package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var TournamentMigration = migration.Migration{
	Version: 8,
	Name:    "create_tournaments_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name:              "tournaments",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "client_id", Type: schema.TypeUUID, NotNull: true, References: "clients(id)"},
				{Name: "sport_id", Type: schema.TypeUUID, NotNull: true, References: "sports(id)"},
				{Name: "name", Type: schema.TypeText, NotNull: true},
				{Name: "description", Type: schema.TypeText},
				{Name: "venue", Type: schema.TypeText},
				{Name: "address", Type: schema.TypeText},
				{Name: "start_date", Type: schema.TypeTimestamptz, NotNull: true},
				{Name: "end_date", Type: schema.TypeTimestamptz, NotNull: true},
				{Name: "registration_open_at", Type: schema.TypeTimestamptz},
				{Name: "registration_close_at", Type: schema.TypeTimestamptz},
				{Name: "registration_fee", Type: schema.TypeNumeric, NotNull: true, Default: "0"},
				{Name: "max_participants", Type: schema.TypeInteger},
				{Name: "contact_name", Type: schema.TypeText},
				{Name: "contact_email", Type: schema.TypeText},
				{Name: "contact_phone", Type: schema.TypeText},
				{Name: "rules", Type: schema.TypeText},
				{Name: "terms", Type: schema.TypeText},
				{Name: "payment_instructions", Type: schema.TypeText},
				{Name: "format", Type: schema.TypeText, NotNull: true},
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'DRAFT'"},
				{Name: "created_by", Type: schema.TypeUUID, NotNull: true},
				{Name: "cancelled_reason", Type: schema.TypeText},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS tournaments;")
	},
}

var TournamentDivisionMigration = migration.Migration{
	Version: 9,
	Name:    "create_tournament_divisions_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name:              "tournament_divisions",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "tournament_id", Type: schema.TypeUUID, NotNull: true, References: "tournaments(id)"},
				{Name: "name", Type: schema.TypeText, NotNull: true},
				{Name: "gender", Type: schema.TypeText},
				{Name: "min_age", Type: schema.TypeSmallInt},
				{Name: "max_age", Type: schema.TypeSmallInt},
				{Name: "skill_level", Type: schema.TypeText},
				{Name: "participant_type", Type: schema.TypeText, NotNull: true, Default: "'INDIVIDUAL'"},
				{Name: "capacity", Type: schema.TypeInteger},
				{Name: "registration_fee", Type: schema.TypeNumeric},
				{Name: "format", Type: schema.TypeText},
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'ACTIVE'"},
				// Scoring overrides per division (if null, use sport default)
				{Name: "best_of", Type: schema.TypeSmallInt},
				{Name: "target_score", Type: schema.TypeSmallInt},
				{Name: "win_by", Type: schema.TypeSmallInt},
				{Name: "max_score", Type: schema.TypeSmallInt},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS tournament_divisions;")
	},
}
