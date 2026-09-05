package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var RegistrationMigration = migration.Migration{
	Version: 10,
	Name:    "create_tournament_registrations_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		if err := exec(ctx, schema.Table{
			Name:              "tournament_registrations",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "tournament_id", Type: schema.TypeUUID, NotNull: true, References: "tournaments(id)"},
				{Name: "division_id", Type: schema.TypeUUID, NotNull: true, References: "tournament_divisions(id)"},
				{Name: "player_id", Type: schema.TypeUUID, NotNull: true, References: "players(id)"},
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'PENDING'"},
				{Name: "payment_status", Type: schema.TypeText, NotNull: true, Default: "'NOT_SUBMITTED'"},
				{Name: "notes", Type: schema.TypeText},
				{Name: "registered_at", Type: schema.TypeTimestamptz, NotNull: true},
				{Name: "approved_at", Type: schema.TypeTimestamptz},
				{Name: "approved_by", Type: schema.TypeUUID},
			},
		}.DDL()); err != nil {
			return err
		}
		// One active registration per player per division.
		return exec(ctx, `
			ALTER TABLE tournament_registrations
				ADD CONSTRAINT uq_registration_player_division
				UNIQUE (division_id, player_id);
		`)
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS tournament_registrations;")
	},
}

var PaymentSubmissionMigration = migration.Migration{
	Version: 11,
	Name:    "create_payment_submissions_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name:              "payment_submissions",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "registration_id", Type: schema.TypeUUID, NotNull: true, References: "tournament_registrations(id)"},
				{Name: "amount_expected", Type: schema.TypeNumeric, NotNull: true},
				{Name: "amount_submitted", Type: schema.TypeNumeric, NotNull: true},
				{Name: "payment_method", Type: schema.TypeText, NotNull: true},
				{Name: "reference_number", Type: schema.TypeText},
				{Name: "proof_reference", Type: schema.TypeText},
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'SUBMITTED'"},
				// EXACT | UNDERPAID | OVERPAID
				{Name: "amount_status", Type: schema.TypeText, NotNull: true, Default: "'EXACT'"},
				{Name: "submitted_at", Type: schema.TypeTimestamptz, NotNull: true},
				{Name: "verified_at", Type: schema.TypeTimestamptz},
				{Name: "verified_by", Type: schema.TypeUUID},
				{Name: "rejection_reason", Type: schema.TypeText},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS payment_submissions;")
	},
}
