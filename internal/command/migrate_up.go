// Package command contains cli command definitions.
package command

import (
	"path/filepath"

	"github.com/gozix/glue"
	sqlBundle "github.com/gozix/sql"
	zapBundle "github.com/gozix/zap"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sarulabs/di"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// DefMigrateUpName is command definition name.
const DefMigrateUpName = "cli.cmd.migrate_up"

// DefMigrateUp is command definition getter.
func DefMigrateUp() di.Def {
	return di.Def{
		Name: DefMigrateUpName,
		Build: func(ctn di.Container) (_ interface{}, err error) {
			return &cobra.Command{
				Use:   "up",
				Short: "Execute a migration to latest available version",
				Args:  cobra.ExactArgs(0),
				RunE: func(cmd *cobra.Command, _ []string) (err error) {
					var (
						conn = cmd.Flag("connection").Value.String()
						db   *sqlBundle.DB
					)

					var sqlRegistry *sqlBundle.Registry
					if err = ctn.Fill(sqlBundle.BundleName, &sqlRegistry); err != nil {
						return err
					}

					if db, err = sqlRegistry.ConnectionWithName(conn); err != nil {
						return err
					}

					var driver string
					if driver, err = sqlRegistry.DriverWithName(conn); err != nil {
						return err
					}

					var glueRegistry glue.Registry
					if err = ctn.Fill(glue.DefRegistry, &glueRegistry); err != nil {
						return err
					}

					var path, _ = cmd.Flags().GetString("path")
					if !filepath.IsAbs(path) {
						var appPath string
						if err = glueRegistry.Fill("app.path", &appPath); err != nil {
							return err
						}

						path = filepath.Join(appPath, path)
					}

					var n int
					if n, err = migrate.Exec(
						db.Master(),
						driver,
						&migrate.FileMigrationSource{Dir: path},
						migrate.Up,
					); err != nil {
						return err
					}

					var logger *zapBundle.Logger
					if err = ctn.Fill(zapBundle.BundleName, &logger); err != nil {
						return err
					}

					logger.Info("Applied migrations", zap.Int("count", n))

					return nil
				},
			}, nil
		},
	}
}
