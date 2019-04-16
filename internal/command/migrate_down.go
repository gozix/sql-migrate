// Package command contains cli command definitions.
package command

import (
	"path/filepath"
	"strconv"

	"github.com/gozix/glue"
	sqlBundle "github.com/gozix/sql"
	zapBundle "github.com/gozix/zap"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sarulabs/di"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// DefMigrateDownName is command definition name.
const DefMigrateDownName = "cli.cmd.migrate_down"

// DefMigrateDown is command definition getter.
func DefMigrateDown() di.Def {
	return di.Def{
		Name: DefMigrateDownName,
		Build: func(ctn di.Container) (_ interface{}, err error) {
			return &cobra.Command{
				Use:   "down",
				Short: "Down migrations",
				Args:  cobra.ExactArgs(0),
				RunE: func(cmd *cobra.Command, args []string) (err error) {
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

					var max int64 = 1
					if len(args) > 0 {
						if max, err = strconv.ParseInt(args[0], 10, 64); err != nil {
							return err
						}
					}

					var glueRegistry glue.Registry
					if err = ctn.Fill(glue.DefRegistry, &glueRegistry); err != nil {
						return err
					}

					var path = cmd.Flag("path").Value.String()
					if !filepath.IsAbs(path) {
						var appPath string
						if err = glueRegistry.Fill("app.path", &appPath); err != nil {
							return err
						}

						path = filepath.Join(appPath, path)
					}

					var n int
					if n, err = migrate.ExecMax(
						db.Master(),
						driver,
						&migrate.FileMigrationSource{Dir: path},
						migrate.Down,
						int(max),
					); err != nil {
						return err
					}

					var logger *zapBundle.Logger
					if err = ctn.Fill(zapBundle.BundleName, &logger); err != nil {
						return err
					}

					logger.Info("Down migrations", zap.Int("count", n))

					return nil
				},
			}, nil
		},
	}
}
