// Package command contains cli command definitions.
package command

import (
	"github.com/gozix/glue/v2"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sarulabs/di/v2"
	"github.com/spf13/cobra"
)

// DefMigrateName is command definition name.
const DefMigrateName = "cli.cmd.migrate"

// DefMigrate is command definition getter.
func DefMigrate(path, table, schema, dialect, connection string) di.Def {
	migrate.SetTable(table)
	migrate.SetSchema(schema)

	return di.Def{
		Name: DefMigrateName,
		Tags: []di.Tag{{
			Name: glue.TagCliCommand,
		}},
		Build: func(ctn di.Container) (_ interface{}, err error) {
			var cmdUp *cobra.Command
			if err = ctn.Fill(DefMigrateUpName, &cmdUp); err != nil {
				return nil, err
			}

			var cmdDown *cobra.Command
			if err = ctn.Fill(DefMigrateDownName, &cmdDown); err != nil {
				return nil, err
			}

			var cmd = &cobra.Command{
				Use:   "sql-migrate [command]",
				Short: "Database migrations",
			}

			cmd.AddCommand(cmdUp)
			cmd.AddCommand(cmdDown)

			cmd.PersistentFlags().String("path", path, "Path to migrations")
			cmd.PersistentFlags().String("dialect", dialect, "Dialect name")
			cmd.PersistentFlags().String("connection", connection, "Connection name")

			return cmd, nil
		},
	}
}
