package migration

import (
	"context"

	"github.com/imsab23/platform-be/infra/storage/postgres/migration"
	"github.com/imsab23/platform-be/infra/storage/postgres/schema"
)

var MatchMigration = migration.Migration{
	Version: 16,
	Name:    "create_matches_table",

	Up: func(ctx context.Context, exec migration.ExecFunc) error {
		// matches
		if err := exec(ctx, schema.Table{
			Name:              "matches",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "tournament_id", Type: schema.TypeUUID, NotNull: true, References: "tournaments(id)"},
				{Name: "division_id", Type: schema.TypeUUID, NotNull: true, References: "tournament_divisions(id)"},
				{Name: "bracket_id", Type: schema.TypeUUID, References: "brackets(id)"},
				{Name: "bracket_node_id", Type: schema.TypeUUID, References: "bracket_nodes(id)"},
				{Name: "round", Type: schema.TypeInteger, NotNull: true, Default: "1"},
				{Name: "match_number", Type: schema.TypeInteger, NotNull: true},
				{Name: "court_id", Type: schema.TypeUUID, References: "courts(id)"},
				{Name: "scheduled_at", Type: schema.TypeTimestamptz},
				// SCHEDULED | READY | IN_PROGRESS | COMPLETED | CANCELLED | POSTPONED
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'SCHEDULED'"},
				// NORMAL_WIN | WALKOVER | FORFEIT | DISQUALIFICATION | BYE
				{Name: "result_type", Type: schema.TypeText},
				{Name: "winner_registration_id", Type: schema.TypeUUID},
				{Name: "winner_team_id", Type: schema.TypeUUID},
				{Name: "notes", Type: schema.TypeText},
				{Name: "finalized_by", Type: schema.TypeUUID},
				{Name: "finalized_at", Type: schema.TypeTimestamptz},
			},
		}.DDL()); err != nil {
			return err
		}

		// match_participants
		if err := exec(ctx, schema.Table{
			Name: "match_participants",
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "match_id", Type: schema.TypeUUID, NotNull: true, References: "matches(id)"},
				{Name: "registration_id", Type: schema.TypeUUID},
				{Name: "team_id", Type: schema.TypeUUID},
				{Name: "side", Type: schema.TypeSmallInt, NotNull: true},
				{Name: "is_winner", Type: schema.TypeBoolean, NotNull: true, Default: "false"},
				{Name: "created_at", Type: schema.TypeTimestamptz, NotNull: true, Default: "NOW()"},
			},
		}.DDL()); err != nil {
			return err
		}

		// match_games (individual sets/games/periods within a match)
		return exec(ctx, schema.Table{
			Name:              "match_games",
			IncludeTimestamps: true,
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeUUID, PrimaryKey: true},
				{Name: "match_id", Type: schema.TypeUUID, NotNull: true, References: "matches(id)"},
				{Name: "game_number", Type: schema.TypeSmallInt, NotNull: true},
				{Name: "score_side1", Type: schema.TypeInteger, NotNull: true, Default: "0"},
				{Name: "score_side2", Type: schema.TypeInteger, NotNull: true, Default: "0"},
				{Name: "is_tiebreak", Type: schema.TypeBoolean, NotNull: true, Default: "false"},
				// PENDING | COMPLETED
				{Name: "status", Type: schema.TypeText, NotNull: true, Default: "'PENDING'"},
			},
		}.DDL())
	},

	Down: func(ctx context.Context, exec migration.ExecFunc) error {
		if err := exec(ctx, "DROP TABLE IF EXISTS match_games;"); err != nil {
			return err
		}
		if err := exec(ctx, "DROP TABLE IF EXISTS match_participants;"); err != nil {
			return err
		}
		return exec(ctx, "DROP TABLE IF EXISTS matches;")
	},
}
