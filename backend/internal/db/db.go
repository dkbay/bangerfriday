package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func InitDB() (*sql.DB, error) {
	dbPath := "banger.db"
	
	_, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(dbPath)
			if err != nil {
				return nil, err
			}
			file.Close()
		}
	}

	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := createTables(database); err != nil {
		return nil, err
	}

	return database, nil
}

func createTables(db *sql.DB) error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		spotify_id TEXT UNIQUE,
		display_name TEXT,
		access_token TEXT,
		refresh_token TEXT,
		token_expiry INTEGER,
		is_admin INTEGER DEFAULT 0,
		created_at INTEGER
	);`

	themeSuggestionsTable := `
	CREATE TABLE IF NOT EXISTS theme_suggestions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		theme TEXT,
		status TEXT DEFAULT 'pending',
		created_at INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	dailyPlaylistsTable := `
	CREATE TABLE IF NOT EXISTS daily_playlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT UNIQUE,
		playlist_id TEXT,
		theme TEXT,
		created_at INTEGER
	);`

	archivesTable := `
	CREATE TABLE IF NOT EXISTS archives (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT,
		playlist_id TEXT,
		playlist_name TEXT,
		track_count INTEGER,
		created_at INTEGER
	);`

	_, err := db.Exec(usersTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(themeSuggestionsTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(dailyPlaylistsTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(archivesTable)
	if err != nil {
		return err
	}

	return nil
}