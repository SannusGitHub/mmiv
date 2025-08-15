package controller

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func OpenSQL() {
	var err error
	db, err = sql.Open("sqlite3", "./mmiv.db")
	if err != nil {
		log.Fatal(err)
	}

	WriteToSQL(`
		CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY,
		username TEXT NOT NULL,
		postcontent TEXT NOT NULL,
		imagepath TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		pinned INTEGER NOT NULL DEFAULT 0,
		locked INTEGER NOT NULL DEFAULT 0,
		isanonymous INTEGER NOT NULL DEFAULT 0,
	)`)

	WriteToSQL(`
		CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		rank INTEGER NOT NULL
	)`)

	WriteToSQL(`
		CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY,
		parentpostid INTEGER NOT NULL,
		username TEXT NOT NULL,
		postcontent TEXT NOT NULL,
		imagepath TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		isanonymous INTEGER NOT NULL DEFAULT 0
	)`)

	WriteToSQL(`
		CREATE TABLE IF NOT EXISTS global_ids (
		id INTEGER PRIMARY KEY AUTOINCREMENT
	)`)

	if err = db.Ping(); err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
}

func WriteToSQL(execString string, args ...interface{}) error {
	_, err := db.Exec(execString, args...)
	if err != nil {
		return err
	}
	return nil
}

func QueryFromSQL(execString string, args ...interface{}) string {
	var result string

	err := db.QueryRow(execString, args...).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return "not found"
		}

		log.Fatal(err)
	}

	return result
}

func CloseSQL() {
	db.Close()
}
