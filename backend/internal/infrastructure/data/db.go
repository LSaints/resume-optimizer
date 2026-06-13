package data

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func NewConnection() *sql.DB {
	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	createTables(db)

	return db
}

func createTables(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS resumes (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		original_name TEXT NOT NULL,
		raw_text TEXT NOT NULL,
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
