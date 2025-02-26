// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package ddl

const (
	// CreateSecretTable represents a query to
	// create the secrets table for Vela.
	CreateSecretTable = `
CREATE TABLE
IF NOT EXISTS
secrets (
	id            SERIAL PRIMARY KEY,
	type          VARCHAR(100),
	org           VARCHAR(250),
	repo          VARCHAR(250),
	team          VARCHAR(250),
	name          VARCHAR(250),
	value         BYTEA,
	images        VARCHAR(1000),
	events        VARCHAR(1000),
	allow_command BOOLEAN,
	created_at    INTEGER,
	created_by    VARCHAR(250),
	updated_at    INTEGER,
	updated_by    VARCHAR(250),
	UNIQUE(type, org, repo, name),
	UNIQUE(type, org, team, name)
);
`

	// CreateSecretTypeOrgRepo represents a query to create an
	// index on the secrets table for the type, org and repo columns.
	//
	// nolint: gosec // ignore false positive
	CreateSecretTypeOrgRepo = `
CREATE INDEX
IF NOT EXISTS
secrets_type_org_repo
ON secrets (type, org, repo);
`

	// CreateSecretTypeOrgTeam represents a query to create an
	// index on the secrets table for the type, org and team columns.
	//
	// nolint: gosec // ignore false positive
	CreateSecretTypeOrgTeam = `
CREATE INDEX
IF NOT EXISTS
secrets_type_org_team
ON secrets (type, org, team);
`

	// CreateSecretTypeOrg represents a query to create an
	// index on the secrets table for the type, and org columns.
	//
	// nolint: gosec // ignore false positive
	CreateSecretTypeOrg = `
CREATE INDEX
IF NOT EXISTS
secrets_type_org
ON secrets (type, org);
`
)
