package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var PlayerProfileMigration = migration.Migration{
	Version: 1,
	Name:    "create_players_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name: "players",

			Columns: []schema.Column{
				{
					Name:       "id",
					Type:       schema.TypeUUID,
					PrimaryKey: true,
				},
				{
					Name:    "first_name",
					Type:    schema.TypeText,
					NotNull: true,
				},
				{
					Name: "middle_name",
					Type: schema.TypeText,
				},
				{
					Name:    "last_name",
					Type:    schema.TypeText,
					NotNull: true,
				},
				{
					Name: "date_of_birth",
					Type: schema.TypeText,
				},
				{
					Name: "gender",
					Type: schema.TypeText,
				},
				{
					Name: "phone",
					Type: schema.TypeText,
				},
				{
					Name: "profile_image_url",
					Type: schema.TypeText,
				},
				{
					Name: "address",
					Type: schema.TypeText,
				},
				{
					Name: "emergency_contact_name",
					Type: schema.TypeText,
				},
				{
					Name: "emergency_contact_phone",
					Type: schema.TypeText,
				},
				{
					Name:    "created_at",
					Type:    schema.TypeTimestamptz,
					NotNull: true,
				},
				{
					Name:    "updated_at",
					Type:    schema.TypeTimestamptz,
					NotNull: true,
				},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS players;")
	},
}

// AlterPlayersDateOfBirth corrects the date_of_birth column from TEXT to TIMESTAMPTZ.
// NULLIF guards against empty strings that would fail the cast.
var AlterPlayersDateOfBirth = migration.Migration{
	Version: 2,
	Name:    "alter_players_date_of_birth_type",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx,
			"ALTER TABLE players ALTER COLUMN date_of_birth TYPE TIMESTAMPTZ USING NULLIF(date_of_birth, '')::TIMESTAMPTZ;",
		)
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx,
			"ALTER TABLE players ALTER COLUMN date_of_birth TYPE TEXT USING date_of_birth::TEXT;",
		)
	},
}

// AlterPlayersAddCredentials makes players a standalone, self-authenticating
// entity — players have no relation to the users table.
var AlterPlayersAddCredentials = migration.Migration{
	Version: 6,
	Name:    "alter_players_add_credentials",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, `
			ALTER TABLE players
				ADD COLUMN IF NOT EXISTS email TEXT NOT NULL,
				ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL,
				ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;
			ALTER TABLE players ADD CONSTRAINT uq_players_email UNIQUE (email);
		`)
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, `
			ALTER TABLE players
				DROP COLUMN IF EXISTS email,
				DROP COLUMN IF EXISTS password_hash,
				DROP COLUMN IF EXISTS last_login_at;
		`)
	},
}
