// Copyright (c) 2021 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package postgres

import (
	"database/sql"
	"time"

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
	}

	client struct {
		config   *config
		Postgres *gorm.DB
	}
)

// New returns a Database implementation that integrates with a Postgres instance.
//
// nolint: golint // ignore returning unexported client
func New(opts ...ClientOpt) (*client, error) {
	// create new Postgres client
	c := new(client)

	// create new fields
	c.config = new(config)
	c.Postgres = new(gorm.DB)

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

	return c, nil
}

// NewTest returns a Database implementation that integrates with a fake Postgres instance.
//
// This function is intended for running tests only.
//
// nolint: golint // ignore returning unexported client
func NewTest(_sql *sql.DB) (*client, error) {
	// create new Postgres client
	c := new(client)

	// create new fields
	c.config = &config{
		CompressionLevel: 3,
		EncryptionKey:    "A1B2C3D4E5G6H7I8J9K0LMNOPQRSTUVW",
	}
	c.Postgres = new(gorm.DB)

	// create the new Postgres database client
	//
	// https://pkg.go.dev/gorm.io/gorm#Open
	_database, err := gorm.Open(postgres.New(postgres.Config{
		Conn: _sql,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	c.Postgres = _database

	return c, nil
}