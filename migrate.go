// Copyright 2018 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package migrate

import (
	"github.com/gozix/di"
	"github.com/gozix/glue/v3"
	gzSQL "github.com/gozix/sql/v3"
	gzZap "github.com/gozix/zap/v3"

	"github.com/gozix/sql-migrate/v3/internal/command"
)

type (
	// Bundle implements the glue.Bundle interface.
	Bundle struct {
		path       string
		table      string
		schema     string
		dialect    string
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

// Bundle implements glue.Bundle interface.
var _ glue.Bundle = (*Bundle)(nil)

// Connection option.
func Connection(value string) Option {
	return optionFunc(func(b *Bundle) {
		b.connection = value
	})
}

// Dialect option.
func Dialect(value string) Option {
	return optionFunc(func(b *Bundle) {
		b.dialect = value
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
		connection: gzSQL.DEFAULT,
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
func (b *Bundle) Build(builder di.Builder) error {
	var tag = "cli.cmd.migrate.subcommand"

	return builder.Apply(
		di.Provide(
			command.NewMigrateConstructor(b.path, b.table, b.schema, b.dialect, b.connection),
			di.Constraint(0, di.WithTags(tag)),
			glue.AsCliCommand(),
		),
		di.Provide(command.NewMigrateDown, di.Tags{{
			Name: tag,
		}}),
		di.Provide(command.NewMigrateUp, di.Tags{{
			Name: tag,
		}}),
	)
}

// DependsOn implements the glue.DependsOn interface.
func (b *Bundle) DependsOn() []string {
	return []string{
		gzSQL.BundleName,
		gzZap.BundleName,
	}
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
