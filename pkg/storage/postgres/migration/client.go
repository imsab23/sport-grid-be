package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var ClientsMigration = migration.Migration{
	Version: 4,
	Name:    "create_clients_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name: "clients",
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "name", Type: schema.TypeText, NotNull: true},
				{Name: "slug", Type: schema.TypeText, NotNull: true, Unique: true},
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'PENDING'"},
				{Name: "contact_name", Type: schema.TypeText},
				{Name: "contact_email", Type: schema.TypeText},
				{Name: "contact_phone", Type: schema.TypeText},
				{Name: "address", Type: schema.TypeText},
				{Name: "created_by", Type: schema.TypeUUID, NotNull: true, References: "users(id)"},
				{Name: "created_at", Type: schema.TypeTimestamptz, NotNull: true},
				{Name: "updated_at", Type: schema.TypeTimestamptz, NotNull: true},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS clients;")
	},
}

// AlterUsersClientIDFK adds the deferred FK now that clients exists.
var AlterUsersClientIDFK = migration.Migration{
	Version: 5,
	Name:    "alter_users_client_id_fk",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.ForeignKey{
			Name:             "fk_users_client_id",
			Table:            "users",
			Column:           "client_id",
			ReferencesTable:  "clients",
			ReferencesColumn: "id",
			OnDelete:         "RESTRICT",
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_client_id;")
	},
}
