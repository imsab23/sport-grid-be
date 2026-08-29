package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var SeedingMigration = migration.Migration{
	Version: 14,
	Name:    "create_seedings_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		if err := exec(ctx, schema.Table{
			Name: "seedings",
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "division_id", Type: schema.TypeUUID, NotNull: true, References: "tournament_divisions(id)"},
				{Name: "registration_id", Type: schema.TypeUUID, NotNull: true, References: "tournament_registrations(id)"},
				{Name: "seed_number", Type: schema.TypeInteger, NotNull: true},
				// MANUAL | RANDOM | RANKING | REGISTRATION_ORDER
				{Name: "method", Type: schema.TypeText, NotNull: true, Default: "'MANUAL'"},
				{Name: "seeded_by", Type: schema.TypeUUID, NotNull: true},
				{Name: "seeded_at", Type: schema.TypeTimestamptz, NotNull: true},
				{Name: "created_at", Type: schema.TypeTimestamptz, NotNull: true, Default: "NOW()"},
			},
		}.DDL()); err != nil {
			return err
		}
		if err := exec(ctx, `
			ALTER TABLE seedings
				ADD CONSTRAINT uq_seeding_registration UNIQUE (division_id, registration_id);
		`); err != nil {
			return err
		}
		return exec(ctx, `
			ALTER TABLE seedings
				ADD CONSTRAINT uq_seeding_seed_number UNIQUE (division_id, seed_number);
		`)
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS seedings;")
	},
}

var BracketMigration = migration.Migration{
	Version: 15,
	Name:    "create_brackets_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		if err := exec(ctx, schema.Table{
			Name:              "brackets",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "tournament_id", Type: schema.TypeUUID, NotNull: true, References: "tournaments(id)"},
				{Name: "division_id", Type: schema.TypeUUID, NotNull: true, References: "tournament_divisions(id)"},
				{Name: "format", Type: schema.TypeText, NotNull: true},
				// ACTIVE | RESET
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'ACTIVE'"},
				{Name: "generated_by", Type: schema.TypeUUID, NotNull: true},
				{Name: "generated_at", Type: schema.TypeTimestamptz, NotNull: true},
				{Name: "reset_reason", Type: schema.TypeText},
			},
		}.DDL()); err != nil {
			return err
		}

		// bracket_nodes uses a UUID for next_node_id without FK (self-referential managed at app level)
		return exec(ctx, schema.Table{
			Name:              "bracket_nodes",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "bracket_id", Type: schema.TypeUUID, NotNull: true, References: "brackets(id)"},
				{Name: "round", Type: schema.TypeInteger, NotNull: true},
				{Name: "position", Type: schema.TypeInteger, NotNull: true},
				{Name: "registration_id", Type: schema.TypeUUID},
				{Name: "team_id", Type: schema.TypeUUID},
				{Name: "is_bye", Type: schema.TypeBoolean, NotNull: true, Default: "false"},
				// UUID of the bracket_node this node's winner advances to (no FK to avoid circular dep)
				{Name: "next_node_id", Type: schema.TypeUUID},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		if err := exec(ctx, "DROP TABLE IF EXISTS bracket_nodes;"); err != nil {
			return err
		}
		return exec(ctx, "DROP TABLE IF EXISTS brackets;")
	},
}
