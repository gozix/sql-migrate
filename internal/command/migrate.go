// Copyright 2018 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package command

import (
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
)

// NewMigrateConstructor is command constructor getter.
func NewMigrateConstructor(
	path, table, schema, dialect, connection string,
) func(subcommands []*cobra.Command) *cobra.Command {
	migrate.SetTable(table)
	migrate.SetSchema(schema)

	return func(subcommands []*cobra.Command) *cobra.Command {
		var cmd = &cobra.Command{
			Use:   "sql-migrate [command]",
			Short: "Database migrations",
		}

		cmd.AddCommand(subcommands...)

		cmd.PersistentFlags().String("path", path, "Path to migrations")
		cmd.PersistentFlags().String("dialect", dialect, "Dialect name")
		cmd.PersistentFlags().String("connection", connection, "Connection name")

		return cmd
	}
}
