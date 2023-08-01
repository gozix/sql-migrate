// Copyright 2018 Sergey Novichkov. All rights reserved.
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gozix/di"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const migrationFileExtension = "sql"

// NewMigrateCreate is subcommand constructor.
func NewMigrateCreate(ctn di.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create named migration",
		Long: `Normally each migration is run within a transaction. 
	However some SQL commands (for example creating an index concurrently in PostgreSQL) 
	cannot be executed inside a transaction. In order to execute such a command in a migration, 
	the migration can be run using the notransaction option https://github.com/rubenv/sql-migrate`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ctn.Call(func(ctx context.Context, logger *zap.Logger) (err error) {
				var (
					name     = args[0]
					path     = cmd.Flag("path").Value.String()
					fileName string
				)
				if !filepath.IsAbs(path) {
					var appPath, ok = ctx.Value("app.path").(string)
					if !ok {
						return errors.New("app.path is undefined")
					}

					path = filepath.Join(appPath, path)
				}

				if fileName, err = create(path, name, migrationFileExtension); err != nil {
					return err
				}

				logger.Info("Created migration", zap.String("migration", fileName))

				return nil
			})
		},
	}
}

func create(migrationDir string, name string, extension string) (string, error) {
	var (
		fileName    string
		fileContent string
		file        *os.File
		err         error
	)

	fileName = filepath.Join(
		filepath.Clean(migrationDir),
		fmt.Sprintf("%s_%s.%s",
			// version
			strconv.FormatInt(time.Now().Unix(), 10),
			name,
			extension,
		),
	)

	fileContent = strings.Join([]string{
		"-- +migrate Up",
		"-- SQL in section 'Up' is executed when this migration is applied",
		"",
		"",
		"",
		"-- +migrate Down",
		"-- SQL section 'Down' is executed when this migration is rolled back",
		"",
	}, "\n")

	file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return fileName, err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.WriteString(fileContent)

	return fileName, err
}
