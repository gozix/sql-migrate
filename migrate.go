package migrate

import (
	"path/filepath"
	"strconv"

	"github.com/gozix/glue"
	"github.com/gozix/sql"
	zapBundle "github.com/gozix/zap"
	"github.com/rubenv/sql-migrate"
	"github.com/sarulabs/di"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type (
	// Bundle implements the glue.Bundle interface.
	Bundle struct {
		path       string
		table      string
		schema     string
		connection string
	}

	// Option interface.
	Option interface {
		apply(b *Bundle)
	}

	// optionFunc wraps a func so it satisfies the Option interface.
	optionFunc func(b *Bundle)
)

// BundleName is default definition name.
const BundleName = "sql-migrate"

// Connection option.
func Connection(value string) Option {
	return optionFunc(func(b *Bundle) {
		b.connection = value
	})
}

// Path option.
func Path(value string) Option {
	return optionFunc(func(b *Bundle) {
		b.path = value
	})
}

// Table option.
func Table(value string) Option {
	return optionFunc(func(b *Bundle) {
		b.table = value
	})
}

// Schema option.
func Schema(value string) Option {
	return optionFunc(func(b *Bundle) {
		b.schema = value
	})
}

// NewBundle create bundle instance.
func NewBundle(options ...Option) (b *Bundle) {
	b = &Bundle{
		path:       "migrations",
		table:      "migration",
		connection: sql.DEFAULT,
	}

	for _, option := range options {
		option.apply(b)
	}

	return b
}

// Name implements the glue.Bundle interface.
func (b *Bundle) Name() string {
	return BundleName
}

// Build implements the glue.Bundle interface.
func (b *Bundle) Build(builder *di.Builder) error {
	return builder.Add(di.Def{
		Name: BundleName,
		Tags: []di.Tag{{
			Name: glue.TagCliCommand,
		}},
		Build: func(ctn di.Container) (_ interface{}, err error) {
			// dependencies
			var logger *zapBundle.Logger
			if err = ctn.Fill(zapBundle.BundleName, &logger); err != nil {
				return nil, err
			}

			var sqlRegistry *sql.Registry
			if err = ctn.Fill(sql.BundleName, &sqlRegistry); err != nil {
				return nil, err
			}

			var glueRegistry glue.Registry
			if err = ctn.Fill(glue.DefRegistry, &glueRegistry); err != nil {
				return nil, err
			}

			// set globals
			migrate.SetTable(b.table)
			migrate.SetSchema(b.schema)

			// build command tree
			var cmd = &cobra.Command{
				Use:   "sql-migrate [command]",
				Short: "Database migrations",
			}

			cmd.AddCommand(
				&cobra.Command{
					Use:   "up",
					Short: "Execute a migration to latest available version",
					Args:  cobra.ExactArgs(0),
					RunE:  b.upCmdHandler(logger, sqlRegistry, glueRegistry),
				},
				&cobra.Command{
					Use:   "down <max>",
					Short: "Rollback a migration to <max> available version",
					Args:  cobra.MaximumNArgs(1),
					RunE:  b.downCmdHandler(logger, sqlRegistry, glueRegistry),
				},
			)

			cmd.PersistentFlags().String("path", b.path, "Path to migrations")
			cmd.PersistentFlags().String("connection", b.connection, "Connection name")

			return cmd, nil
		},
	})
}

// DependsOn implements the glue.DependsOn interface.
func (b *Bundle) DependsOn() []string {
	return []string{sql.BundleName, zapBundle.BundleName}
}

// upCmdHandler is migrate up command handler.
func (b *Bundle) upCmdHandler(
	logger *zapBundle.Logger,
	sqlRegistry *sql.Registry,
	glueRegistry glue.Registry,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) (err error) {
		var (
			conn = cmd.Flag("connection").Value.String()
			db   *sql.DB
		)

		if db, err = sqlRegistry.ConnectionWithName(conn); err != nil {
			return err
		}

		var driver string
		if driver, err = sqlRegistry.DriverWithName(conn); err != nil {
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
		if n, err = migrate.Exec(
			db.Master(),
			driver,
			&migrate.FileMigrationSource{Dir: path},
			migrate.Up,
		); err != nil {
			return err
		}

		logger.Info("Applied migrations", zap.Int("count", n))

		return nil
	}
}

// DownCmdHandler is migrate down command handler.
func (b *Bundle) downCmdHandler(
	logger *zapBundle.Logger,
	sqlRegistry *sql.Registry,
	glueRegistry glue.Registry,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		var (
			conn = cmd.Flag("connection").Value.String()
			db   *sql.DB
		)

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

		logger.Info("Down migrations", zap.Int("count", n))

		return nil
	}
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
