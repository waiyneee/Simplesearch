package storage

import (
	"database/sql"
	// "log"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// type documentdb struct{
// 	id int
// 	url string
// 	title string
// 	body string
// 	createdAt string
// }

func OpenDbInstance(path string) (*sql.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("db path is empty")
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir failed: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// defer db.Close() we will not close it until used
	db.SetMaxOpenConns(1)

	// Speed + safety balance
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA temp_store = MEMORY;",
		"PRAGMA foreign_keys = ON;",
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("pragma failed (%s): %w", p, err)
		}
	}

	return db, nil
}
