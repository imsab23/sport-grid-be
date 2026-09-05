package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var UserMigation = migration.Migration{
	Version: 1,
	Name:    "create_users_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name: "users",

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
					Name:    "email",
					Type:    schema.TypeText,
					NotNull: true,
					Unique:  true,
				},
				{
					Name: "contact_number",
					Type: schema.TypeText,
				},
				{
					Name:    "password_hash",
					Type:    schema.TypeText,
					NotNull: true,
				},
				{
					Name:    "status",
					Type:    schema.TypeSmallInt,
					NotNull: true,
					Default: "1",
				},
				{
					Name: "email_verified_at",
					Type: schema.TypeTimestamptz,
				},
				{
					Name: "last_login_at",
					Type: schema.TypeTimestamptz,
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
		return exec(ctx, "DROP TABLE IF EXISTS users;")
	},
}

var UserPasswordResetToken = migration.Migration{
	Version: 1,
	Name:    "create_password_reset_tokens_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name: "password_reset_tokens",

			Columns: []schema.Column{
				{
					Name:       "id",
					Type:       schema.TypeUUID,
					PrimaryKey: true,
				},
				{
					Name:       "user_id",
					Type:       schema.TypeUUID,
					NotNull:    true,
					Unique:     true,
					References: "users(id)",
				},
				{
					Name:    "token_hash",
					Type:    schema.TypeText,
					NotNull: true,
					Unique:  true,
				},
				{
					Name:    "expires_at",
					Type:    schema.TypeTimestamptz,
					NotNull: true,
				},
				{
					Name: "used_at",
					Type: schema.TypeTimestamptz,
				},
				{
					Name:    "created_at",
					Type:    schema.TypeTimestamptz,
					NotNull: true,
					Default: "NOW()",
				},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS password_reset_tokens;")
	},
}
