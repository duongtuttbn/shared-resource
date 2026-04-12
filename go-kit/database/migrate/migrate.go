package migrate

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	migrateV4 "github.com/golang-migrate/migrate/v4"

	// import posgres.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// import file.
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/spf13/cobra"
)

const versionTimeFormat = "20060102150405"

// Command return common migrate command for application.
func Command(sourceURL string, databaseURL string) *cobra.Command {
	return CommandWithConfigResolver(func() (string, string) {
		return sourceURL, databaseURL
	})
}

type configResolver func() (sourceURL string, databaseURL string)

func CommandWithConfigResolver(configResolverFn configResolver) *cobra.Command {
	// Migration should always run on development mode
	logger := logrus.New()

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "database migration command",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "lift migration up to date",
		Run: func(_ *cobra.Command, _ []string) {
			m, err := migrateV4.New(configResolverFn())
			if err != nil {
				logger.WithError(err).Fatal("Error create migration")
			}

			logger.Info("migration up")
			if err := m.Up(); err != nil && !errors.Is(err, migrateV4.ErrNoChange) {
				logger.Fatal(err.Error())
			}
		},
	}, &cobra.Command{
		Use:   "down",
		Short: "step down migration by N(int)",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			m, err := migrateV4.New(configResolverFn())
			if err != nil {
				logger.WithError(err).Fatal("Error create migration")
			}

			down, err := strconv.Atoi(args[0])
			if err != nil {
				logger.WithError(err).Fatal("rev should be a number")
			}

			logger.WithField("down", -down).Info("migration down")
			if err := m.Steps(-down); err != nil {
				logger.Fatal(err.Error())
			}
		},
	}, &cobra.Command{
		Use:   "force",
		Short: "Enforce dirty migration with verion (int)",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			m, err := migrateV4.New(configResolverFn())
			if err != nil {
				logger.WithError(err).Fatal("Error create migration")
			}

			ver, err := strconv.Atoi(args[0])
			if err != nil {
				logger.WithError(err).Fatal("rev should be a number")
			}

			logger.WithField("ver", ver).Info("force")

			if err := m.Force(ver); err != nil {
				logger.Fatal(err.Error())
			}
		},
	}, &cobra.Command{
		Use:  "create",
		Args: cobra.MinimumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			sourceURL, _ := configResolverFn()
			folder := strings.ReplaceAll(sourceURL, "file://", "")
			now := time.Now()
			ver := now.Format(versionTimeFormat)
			name := strings.Join(args, "-")

			up := fmt.Sprintf("%s/%s_%s.up.sql", folder, ver, name)
			down := fmt.Sprintf("%s/%s_%s.down.sql", folder, ver, name)

			logger.WithField("name", name).Info("create migration")
			logger.WithField("up", up).Info("up script")
			logger.WithField("down", up).Info("down script")

			if err := os.WriteFile(up, []byte{}, 0o644); err != nil {
				logger.WithError(err).Fatal("Create migration up error")
			}
			if err := os.WriteFile(down, []byte{}, 0o644); err != nil {
				logger.WithError(err).Fatal("Create migration down error")
			}
		},
	})
	return cmd
}
