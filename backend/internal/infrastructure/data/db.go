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

	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		raw_description TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS resumes_optimized (
		id TEXT PRIMARY KEY,
		resume_id TEXT NOT NULL,
		job_id TEXT NOT NULL,
		system_prompt TEXT NOT NULL,
		user_prompt TEXT NOT NULL,
		raw_text TEXT NOT NULL,
		typst_content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (resume_id) REFERENCES resumes(id),
		FOREIGN KEY (job_id) REFERENCES jobs(id)
	);

	CREATE TABLE IF NOT EXISTS ats_evaluations (
		id TEXT PRIMARY KEY,
		resume_id TEXT NOT NULL,
		job_id TEXT NOT NULL,
		score REAL NOT NULL,
		summary TEXT NOT NULL,
		details TEXT NOT NULL,
		raw_response TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (resume_id) REFERENCES resumes(id),
		FOREIGN KEY (job_id) REFERENCES jobs(id)
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
