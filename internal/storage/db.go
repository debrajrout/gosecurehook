package storage

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		log.Fatalf("failed to open sqlite db: %v", err)
	}

	createTables(db)
	return db
}

func createTables(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			body TEXT,
			headers TEXT,
			received_at TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS dlq (
			id TEXT PRIMARY KEY,
			body TEXT,
			headers TEXT,
			received_at TEXT
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatalf("failed to create table: %v", err)
		}
	}
}
