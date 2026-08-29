package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
)

// AlterUsersAddRoleClientID adds role (NOT NULL, default PLAYER) and
// client_id (nullable UUID, FK deferred to Phase 1B when clients table exists).
var AlterUsersAddRoleClientID = migration.Migration{
	Version: 3,
	Name:    "alter_users_add_role_client_id",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, `
			ALTER TABLE users
				ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'PLAYER',
				ADD COLUMN IF NOT EXISTS client_id UUID;
		`)
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		return exec(ctx, `
			ALTER TABLE users
				DROP COLUMN IF EXISTS role,
				DROP COLUMN IF EXISTS client_id;
		`)
	},
}
