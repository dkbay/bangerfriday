package db

import (
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	dbPath := "banger.db"

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
		user_name TEXT,
		theme TEXT,
		status TEXT DEFAULT 'pending',
		created_at INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`

	userNamesTable := `
	CREATE TABLE IF NOT EXISTS user_names (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		created_at INTEGER
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

	trackAdditionsTable := `
	CREATE TABLE IF NOT EXISTS track_additions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		video_id TEXT,
		date TEXT,
		added_by TEXT,
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

	_, err = db.Exec(userNamesTable)
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

	_, err = db.Exec(trackAdditionsTable)
	if err != nil {
		return err
	}

	_, err = db.Exec("ALTER TABLE theme_suggestions ADD COLUMN user_name TEXT")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_spotify_id ON users(spotify_id)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_theme_suggestions_user_id ON theme_suggestions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_theme_suggestions_status_created_at ON theme_suggestions(status, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_track_additions_date ON track_additions(date)",
	}

	for _, indexStmt := range indexes {
		if _, err := db.Exec(indexStmt); err != nil {
			return err
		}
	}

	return nil
}
