// Copyright 2018 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package command

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/gozix/di"
	gzSQL "github.com/gozix/sql/v3"

	"github.com/iqoption/nap"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewMigrateUp is subcommand constructor.
func NewMigrateUp(ctn di.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Execute a migration to latest available version",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			return ctn.Call(func(ctx context.Context, logger *zap.Logger, registry *gzSQL.Registry) error {
				var (
					dialect = cmd.Flag("dialect").Value.String()
					conn    = cmd.Flag("connection").Value.String()
					db      *nap.DB
				)

				if db, err = registry.ConnectionWithName(conn); err != nil {
					return err
				}

				var driver string
				if driver, err = registry.DriverWithName(conn); err != nil {
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
				if n, err = migrate.Exec(
					db.Master(),
					dialect,
					&migrate.FileMigrationSource{Dir: path},
					migrate.Up,
				); err != nil {
					return err
				}

				logger.Info("Applied migrations", zap.Int("count", n))

				return nil
			})
		},
	}
}
