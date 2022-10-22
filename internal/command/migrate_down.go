// Copyright 2018 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package command

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"

	"github.com/gozix/di"
	gzSQL "github.com/gozix/sql/v3"

	"github.com/iqoption/nap"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewMigrateDown is subcommand constructor.
func NewMigrateDown(ctn di.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Down migrations",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctn.Call(func(ctx context.Context, logger *zap.Logger, registry *gzSQL.Registry) (err error) {
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

				var max int64 = 1
				if len(args) > 0 {
					if max, err = strconv.ParseInt(args[0], 10, 64); err != nil {
						return err
					}
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

				logger.Info("Down migrations", zap.Int("count", n))

				return nil
			})
		},
	}
}
