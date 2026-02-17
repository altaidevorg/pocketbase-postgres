//go:build !no_default_driver

package core

import (
	"os"
	"strings"

	"github.com/pocketbase/dbx"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func DefaultDBConnect(dbPath string) (*dbx.DB, error) {
	if val := os.Getenv("PB_DB_CONNECT"); val != "" {
		if strings.HasPrefix(val, "postgres://") {
			return dbx.Open("postgres", val)
		}
	}

	if strings.HasPrefix(dbPath, "postgres://") {
		return dbx.Open("postgres", dbPath)
	}

	// Note: the busy_timeout pragma must be first because
	// the connection needs to be set to block on busy before WAL mode
	// is set in case it hasn't been already set by another connection.
	pragmas := "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-32000)"

	db, err := dbx.Open("sqlite", dbPath+pragmas)
	if err != nil {
		return nil, err
	}

	return db, nil
}
