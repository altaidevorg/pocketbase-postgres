package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.SystemMigrations.Add(&core.Migration{
		Up: func(txApp core.App) error {
			var query string
			if core.GetDBDriverName(txApp) == "postgres" {
				query = `
					CREATE TABLE IF NOT EXISTS {{_logs}} (
						[[id]]      TEXT PRIMARY KEY DEFAULT ('r'||substring(md5(random()::text) from 1 for 14)) NOT NULL,
						[[level]]   INTEGER DEFAULT 0 NOT NULL,
						[[message]] TEXT DEFAULT '' NOT NULL,
						[[data]]    JSON DEFAULT '{}' NOT NULL,
						[[created]] TEXT DEFAULT (to_char(statement_timestamp() at time zone 'UTC', 'YYYY-MM-DD" "HH24:MI:SS.MS"Z"')) NOT NULL
					);

					CREATE INDEX IF NOT EXISTS idx_logs_level on {{_logs}} ([[level]]);
					CREATE INDEX IF NOT EXISTS idx_logs_message on {{_logs}} ([[message]]);
					CREATE INDEX IF NOT EXISTS idx_logs_created_hour on {{_logs}} ((substring([[created]] from 1 for 13) || ':00:00'));
				`
			} else {
				query = `
					CREATE TABLE IF NOT EXISTS {{_logs}} (
						[[id]]      TEXT PRIMARY KEY DEFAULT ('r'||lower(hex(randomblob(7)))) NOT NULL,
						[[level]]   INTEGER DEFAULT 0 NOT NULL,
						[[message]] TEXT DEFAULT "" NOT NULL,
						[[data]]    JSON DEFAULT "{}" NOT NULL,
						[[created]] TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL
					);

					CREATE INDEX IF NOT EXISTS idx_logs_level on {{_logs}} ([[level]]);
					CREATE INDEX IF NOT EXISTS idx_logs_message on {{_logs}} ([[message]]);
					CREATE INDEX IF NOT EXISTS idx_logs_created_hour on {{_logs}} (strftime('%Y-%m-%d %H:00:00', [[created]]));
				`
			}

			_, execErr := txApp.AuxDB().NewQuery(query).Execute()

			return execErr
		},
		Down: func(txApp core.App) error {
			_, err := txApp.AuxDB().DropTable("_logs").Execute()
			return err
		},
		ReapplyCondition: func(txApp core.App, runner *core.MigrationsRunner, fileName string) (bool, error) {
			// reapply only if the _logs table doesn't exist
			exists := txApp.AuxHasTable("_logs")
			return !exists, nil
		},
	})
}
