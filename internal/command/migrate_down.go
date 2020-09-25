// Package command contains cli command definitions.
package command

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"

	"github.com/gozix/glue/v2"
	sqlBundle "github.com/gozix/sql/v2"
	zapBundle "github.com/gozix/zap/v2"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sarulabs/di/v2"
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
						dialect = cmd.Flag("dialect").Value.String()
						conn    = cmd.Flag("connection").Value.String()
						db      *sqlBundle.DB
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

					var max int64 = 1
					if len(args) > 0 {
						if max, err = strconv.ParseInt(args[0], 10, 64); err != nil {
							return err
						}
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

					if dialect == "" {
						dialect = driver
					}

					var n int
					if n, err = migrate.ExecMax(
						db.Master(),
						dialect,
						&migrate.FileMigrationSource{Dir: path},
						migrate.Down,
						int(max),
					); err != nil {
						return err
					}

					var logger *zap.Logger
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
