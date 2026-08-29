package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var ClientMigration = migration.Migration{
	Version: 4,
	Name:    "create_clients_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, schema.Table{
			Name:              "clients",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "name", Type: schema.TypeText, NotNull: true},
				{Name: "slug", Type: schema.TypeText, NotNull: true, Unique: true},
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'PENDING'"},
				{Name: "contact_email", Type: schema.TypeText},
				{Name: "contact_phone", Type: schema.TypeText},
				{Name: "address", Type: schema.TypeText},
				{Name: "created_by", Type: schema.TypeUUID, NotNull: true},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, "DROP TABLE IF EXISTS clients;")
	},
}

// UsersClientFKMigration adds the FK from users.client_id → clients(id).
// Must run after ClientMigration (v4) and AlterUsersAddRoleClientID (v3).
var UsersClientFKMigration = migration.Migration{
	Version: 5,
	Name:    "alter_users_add_client_id_fk",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, `
			ALTER TABLE users
				ADD CONSTRAINT fk_users_client_id
				FOREIGN KEY (client_id) REFERENCES clients(id);
		`)
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, `
			ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_client_id;
		`)
	},
}
