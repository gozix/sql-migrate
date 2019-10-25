// Package command contains cli command definitions.
package command

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/gozix/glue/v2"
	sqlBundle "github.com/gozix/sql/v2"
	zapBundle "github.com/gozix/zap/v2"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sarulabs/di/v2"
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

					var registry *sqlBundle.Registry
					if err = ctn.Fill(sqlBundle.BundleName, &registry); err != nil {
						return err
					}

					if db, err = registry.ConnectionWithName(conn); err != nil {
						return err
					}

					var driver string
					if driver, err = registry.DriverWithName(conn); err != nil {
						return err
					}

					var ctx context.Context
					if err = ctn.Fill(glue.DefContext, &ctx); err != nil {
						return err
					}

					var path = cmd.Flag("path").Value.String()
					if !filepath.IsAbs(path) {
						var appPath, ok = ctx.Value("app.path").(string)
						if !ok {
							return errors.New("app.path is undefined")
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

					var logger *zap.Logger
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
