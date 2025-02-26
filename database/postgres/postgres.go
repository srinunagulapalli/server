// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package postgres

import (
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-vela/server/database/postgres/ddl"
	"github.com/go-vela/types/constants"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type (
	config struct {
		// specifies the address to use for the Postgres client
		Address string
		// specifies the level of compression to use for the Postgres client
		CompressionLevel int
		// specifies the connection duration to use for the Postgres client
		ConnectionLife time.Duration
		// specifies the maximum idle connections for the Postgres client
		ConnectionIdle int
		// specifies the maximum open connections for the Postgres client
		ConnectionOpen int
		// specifies the encryption key to use for the Postgres client
		EncryptionKey string
		// specifies to skip creating tables and indexes for the Postgres client
		SkipCreation bool
	}

	client struct {
		config   *config
		Postgres *gorm.DB
		// https://pkg.go.dev/github.com/sirupsen/logrus#Entry
		Logger *logrus.Entry
	}
)

// New returns a Database implementation that integrates with a Postgres instance.
//
// nolint: revive // ignore returning unexported client
func New(opts ...ClientOpt) (*client, error) {
	// create new Postgres client
	c := new(client)

	// create new fields
	c.config = new(config)
	c.Postgres = new(gorm.DB)

	// create new logger for the client
	//
	// https://pkg.go.dev/github.com/sirupsen/logrus?tab=doc#StandardLogger
	logger := logrus.StandardLogger()

	// create new logger for the client
	//
	// https://pkg.go.dev/github.com/sirupsen/logrus?tab=doc#NewEntry
	c.Logger = logrus.NewEntry(logger).WithField("database", c.Driver())

	// apply all provided configuration options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	// create the new Postgres database client
	//
	// https://pkg.go.dev/gorm.io/gorm#Open
	_postgres, err := gorm.Open(postgres.Open(c.config.Address), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// set the Postgres database client in the Postgres client
	c.Postgres = _postgres

	// setup database with proper configuration
	err = setupDatabase(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// NewTest returns a Database implementation that integrates with a fake Postgres instance.
//
// This function is intended for running tests only.
//
// nolint: revive // ignore returning unexported client
func NewTest() (*client, sqlmock.Sqlmock, error) {
	// create new Postgres client
	c := new(client)

	// create new fields
	c.config = &config{
		CompressionLevel: 3,
		ConnectionLife:   30 * time.Minute,
		ConnectionIdle:   2,
		ConnectionOpen:   0,
		EncryptionKey:    "A1B2C3D4E5G6H7I8J9K0LMNOPQRSTUVW",
		SkipCreation:     false,
	}
	c.Postgres = new(gorm.DB)

	// create new logger for the client
	//
	// https://pkg.go.dev/github.com/sirupsen/logrus?tab=doc#StandardLogger
	logger := logrus.StandardLogger()

	// create new logger for the client
	//
	// https://pkg.go.dev/github.com/sirupsen/logrus?tab=doc#NewEntry
	c.Logger = logrus.NewEntry(logger)

	// create the new mock sql database
	//
	// https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock#New
	_sql, _mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, err
	}

	// create the new mock Postgres database client
	//
	// https://pkg.go.dev/gorm.io/gorm#Open
	c.Postgres, err = gorm.Open(
		postgres.New(postgres.Config{Conn: _sql}),
		&gorm.Config{SkipDefaultTransaction: true},
	)
	if err != nil {
		return nil, nil, err
	}

	return c, _mock, nil
}

// setupDatabase is a helper function to setup
// the database with the proper configuration.
func setupDatabase(c *client) error {
	// capture database/sql database from gorm database
	//
	// https://pkg.go.dev/gorm.io/gorm#DB.DB
	_sql, err := c.Postgres.DB()
	if err != nil {
		return err
	}

	// set the maximum amount of time a connection may be reused
	//
	// https://golang.org/pkg/database/sql/#DB.SetConnMaxLifetime
	_sql.SetConnMaxLifetime(c.config.ConnectionLife)

	// set the maximum number of connections in the idle connection pool
	//
	// https://golang.org/pkg/database/sql/#DB.SetMaxIdleConns
	_sql.SetMaxIdleConns(c.config.ConnectionIdle)

	// set the maximum number of open connections to the database
	//
	// https://golang.org/pkg/database/sql/#DB.SetMaxOpenConns
	_sql.SetMaxOpenConns(c.config.ConnectionOpen)

	// verify connection to the database
	err = c.Ping()
	if err != nil {
		return err
	}

	// check if we should skip creating database objects
	if c.config.SkipCreation {
		c.Logger.Warning("skipping creation of data tables and indexes in the postgres database")

		return nil
	}

	// create the tables in the database
	err = createTables(c)
	if err != nil {
		return err
	}

	// create the indexes in the database
	err = createIndexes(c)
	if err != nil {
		return err
	}

	return nil
}

// createTables is a helper function to setup
// the database with the necessary tables.
func createTables(c *client) error {
	c.Logger.Trace("creating data tables in the postgres database")

	// create the builds table
	err := c.Postgres.Exec(ddl.CreateBuildTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableBuild, err)
	}

	// create the hooks table
	err = c.Postgres.Exec(ddl.CreateHookTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableHook, err)
	}

	// create the logs table
	err = c.Postgres.Exec(ddl.CreateLogTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableLog, err)
	}

	// create the repos table
	err = c.Postgres.Exec(ddl.CreateRepoTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableRepo, err)
	}

	// create the secrets table
	err = c.Postgres.Exec(ddl.CreateSecretTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableSecret, err)
	}

	// create the services table
	err = c.Postgres.Exec(ddl.CreateServiceTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableService, err)
	}

	// create the steps table
	err = c.Postgres.Exec(ddl.CreateStepTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableStep, err)
	}

	// create the users table
	err = c.Postgres.Exec(ddl.CreateUserTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableUser, err)
	}

	// create the workers table
	err = c.Postgres.Exec(ddl.CreateWorkerTable).Error
	if err != nil {
		return fmt.Errorf("unable to create %s table: %w", constants.TableWorker, err)
	}

	return nil
}

// createIndexes is a helper function to setup
// the database with the necessary indexes.
func createIndexes(c *client) error {
	c.Logger.Trace("creating data indexes in the postgres database")

	// create the builds_repo_id index for the builds table
	err := c.Postgres.Exec(ddl.CreateBuildRepoIDIndex).Error
	if err != nil {
		return fmt.Errorf("unable to create builds_repo_id index for the %s table: %w", constants.TableBuild, err)
	}

	// create the builds_status index for the builds table
	err = c.Postgres.Exec(ddl.CreateBuildStatusIndex).Error
	if err != nil {
		return fmt.Errorf("unable to create builds_status index for the %s table: %w", constants.TableBuild, err)
	}

	// create the builds_created index for the builds table
	err = c.Postgres.Exec(ddl.CreateBuildCreatedIndex).Error
	if err != nil {
		return fmt.Errorf("unable to create builds_created index for the %s table: %w", constants.TableBuild, err)
	}

	// create the hooks_repo_id index for the hooks table
	err = c.Postgres.Exec(ddl.CreateHookRepoIDIndex).Error
	if err != nil {
		return fmt.Errorf("unable to create hooks_repo_id index for the %s table: %w", constants.TableHook, err)
	}

	// create the logs_build_id index for the logs table
	err = c.Postgres.Exec(ddl.CreateLogBuildIDIndex).Error
	if err != nil {
		return fmt.Errorf("unable to create logs_build_id index for the %s table: %w", constants.TableLog, err)
	}

	// create the repos_org_name index for the repos table
	err = c.Postgres.Exec(ddl.CreateRepoOrgNameIndex).Error
	if err != nil {
		return fmt.Errorf("unable to create repos_org_name index for the %s table: %w", constants.TableRepo, err)
	}

	// create the secrets_type_org_repo index for the secrets table
	err = c.Postgres.Exec(ddl.CreateSecretTypeOrgRepo).Error
	if err != nil {
		return fmt.Errorf("unable to create secrets_type_org_repo index for the %s table: %w", constants.TableSecret, err)
	}

	// create the secrets_type_org_team index for the secrets table
	err = c.Postgres.Exec(ddl.CreateSecretTypeOrgTeam).Error
	if err != nil {
		return fmt.Errorf("unable to create secrets_type_org_team index for the %s table: %w", constants.TableSecret, err)
	}

	// create the secrets_type_org index for the secrets table
	err = c.Postgres.Exec(ddl.CreateSecretTypeOrg).Error
	if err != nil {
		return fmt.Errorf("unable to create secrets_type_org index for the %s table: %w", constants.TableSecret, err)
	}

	// create the users_refresh index for the users table
	err = c.Postgres.Exec(ddl.CreateUserRefreshIndex).Error
	if err != nil {
		return fmt.Errorf("unable to create users_refresh index for the %s table: %w", constants.TableUser, err)
	}

	// create the workers_hostname_address index for the workers table
	err = c.Postgres.Exec(ddl.CreateWorkerHostnameAddressIndex).Error
	if err != nil {
		return fmt.Errorf("unable to create workers_hostname_address index for the %s table: %w", constants.TableWorker, err)
	}

	return nil
}
