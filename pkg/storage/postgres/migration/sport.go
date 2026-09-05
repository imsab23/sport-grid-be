package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var SportMigration = migration.Migration{
	Version: 6,
	Name:    "create_sports_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name:              "sports",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "name", Type: schema.TypeText, NotNull: true},
				{Name: "slug", Type: schema.TypeText, NotNull: true, Unique: true},
				{Name: "description", Type: schema.TypeText},
				{Name: "is_active", Type: schema.TypeBoolean, NotNull: true, Default: "true"},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS sports;")
	},
}

var SportScoringConfigMigration = migration.Migration{
	Version: 7,
	Name:    "create_sport_scoring_configs_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name:              "sport_scoring_configs",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "sport_id", Type: schema.TypeUUID, NotNull: true, References: "sports(id)"},
				{Name: "best_of", Type: schema.TypeSmallInt, NotNull: true, Default: "3"},
				{Name: "target_score", Type: schema.TypeSmallInt, NotNull: true, Default: "11"},
				{Name: "win_by", Type: schema.TypeSmallInt, NotNull: true, Default: "2"},
				{Name: "max_score", Type: schema.TypeSmallInt},
				{Name: "sets_count", Type: schema.TypeSmallInt},
				{Name: "periods_count", Type: schema.TypeSmallInt},
				{Name: "tiebreaker_type", Type: schema.TypeText},
				{Name: "win_condition", Type: schema.TypeText},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS sport_scoring_configs;")
	},
}
