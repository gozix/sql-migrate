// Package migrate provide sql-migrate bundle.
package migrate

import (
	sqlBundle "github.com/gozix/sql"
	zapBundle "github.com/gozix/zap"
	"github.com/sarulabs/di"

	"github.com/gozix/sql-migrate/internal/command"
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
		connection: sqlBundle.DEFAULT,
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
	return builder.Add(
		command.DefMigrate(b.path, b.table, b.schema, b.connection),
		command.DefMigrateDown(),
		command.DefMigrateUp(),
	)
}

// DependsOn implements the glue.DependsOn interface.
func (b *Bundle) DependsOn() []string {
	return []string{
		sqlBundle.BundleName,
		zapBundle.BundleName,
	}
}

// apply implements Option.
func (f optionFunc) apply(bundle *Bundle) {
	f(bundle)
}
