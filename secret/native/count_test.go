// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package native

import (
	"testing"

	"github.com/go-vela/server/database/sqlite"
	"github.com/go-vela/types/library"
)

func TestNative_Count(t *testing.T) {
	// setup types
	sec := new(library.Secret)
	sec.SetID(1)
	sec.SetOrg("foo")
	sec.SetRepo("bar")
	sec.SetName("baz")
	sec.SetValue("foob")
	sec.SetType("repo")
	sec.SetImages([]string{"foo", "bar"})
	sec.SetEvents([]string{"foo", "bar"})

	want := 1

	// setup database
	db, _ := sqlite.NewTest()

	defer func() {
		db.Sqlite.Exec("delete from secrets;")
		_sql, _ := db.Sqlite.DB()
		_sql.Close()
	}()

	_ = db.CreateSecret(sec)

	// run test
	s, err := New(
		WithDatabase(db),
	)
	if err != nil {
		t.Errorf("New returned err: %v", err)
	}

	got, err := s.Count("repo", "foo", "bar", []string{})
	if err != nil {
		t.Errorf("Count returned err: %v", err)
	}

	if got != int64(want) {
		t.Errorf("Count is %v, want %v", got, want)
	}
}

func TestNative_Count_Invalid(t *testing.T) {
	// setup database
	db, _ := sqlite.NewTest()
	_sql, _ := db.Sqlite.DB()
	_sql.Close()

	// run test
	s, err := New(
		WithDatabase(db),
	)
	if err != nil {
		t.Errorf("New returned err: %v", err)
	}

	got, err := s.Count("repo", "foo", "bar", []string{})
	if err == nil {
		t.Errorf("Count should have returned err")
	}

	if got != 0 {
		t.Errorf("Count is %v, want 0", got)
	}
}
