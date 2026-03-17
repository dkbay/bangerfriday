package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int64
	SpotifyID    string
	DisplayName  string
	AccessToken  string
	RefreshToken string
	TokenExpiry  int64
	IsAdmin      bool
	CreatedAt    int64
}

type ThemeSuggestion struct {
	ID        int64
	UserID    int64
	Theme     string
	Status    string
	CreatedAt int64
	UserName  string
}

type DailyPlaylist struct {
	ID         int64
	Date       string
	PlaylistID string
	Theme      string
	CreatedAt  int64
}

type Archive struct {
	ID           int64
	Date         string
	PlaylistID   string
	PlaylistName string
	TrackCount   int
	CreatedAt    int64
}

type TrackAddition struct {
	ID        int64
	VideoID   string
	Date      string
	AddedBy   string
	CreatedAt int64
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserBySpotifyID(spotifyID string) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, spotify_id, display_name, access_token, refresh_token, token_expiry, is_admin, created_at
		FROM users WHERE spotify_id = ?
	`, spotifyID).Scan(&user.ID, &user.SpotifyID, &user.DisplayName, &user.AccessToken, &user.RefreshToken, &user.TokenExpiry, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByID(id int64) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, spotify_id, display_name, access_token, refresh_token, token_expiry, is_admin, created_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.SpotifyID, &user.DisplayName, &user.AccessToken, &user.RefreshToken, &user.TokenExpiry, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) CreateUser(user *User) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO users (spotify_id, display_name, access_token, refresh_token, token_expiry, is_admin, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, user.SpotifyID, user.DisplayName, user.AccessToken, user.RefreshToken, user.TokenExpiry, user.IsAdmin, user.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) UpdateUser(user *User) error {
	_, err := s.db.Exec(`
		UPDATE users SET display_name = ?, access_token = ?, refresh_token = ?, token_expiry = ?, is_admin = ?
		WHERE id = ?
	`, user.DisplayName, user.AccessToken, user.RefreshToken, user.TokenExpiry, user.IsAdmin, user.ID)
	return err
}

func (s *Store) SetUserAdmin(userID int64, isAdmin bool) error {
	_, err := s.db.Exec(`UPDATE users SET is_admin = ? WHERE id = ?`, isAdmin, userID)
	return err
}

func (s *Store) GetDailyPlaylist(date string) (*DailyPlaylist, error) {
	var dp DailyPlaylist
	err := s.db.QueryRow(`
		SELECT id, date, playlist_id, theme, created_at
		FROM daily_playlists WHERE date = ?
	`, date).Scan(&dp.ID, &dp.Date, &dp.PlaylistID, &dp.Theme, &dp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &dp, nil
}

func (s *Store) CreateDailyPlaylist(dp *DailyPlaylist) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO daily_playlists (date, playlist_id, theme, created_at)
		VALUES (?, ?, ?, ?)
	`, dp.Date, dp.PlaylistID, dp.Theme, dp.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) UpdateDailyPlaylistTheme(date, theme string) error {
	_, err := s.db.Exec(`UPDATE daily_playlists SET theme = ? WHERE date = ?`, theme, date)
	return err
}

func (s *Store) UpsertDailyPlaylistTheme(date, theme string) error {
	_, err := s.db.Exec(`
		INSERT INTO daily_playlists (date, playlist_id, theme, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET theme = excluded.theme
	`, date, "", theme, time.Now().Unix())
	return err
}

func (s *Store) UpdateDailyPlaylistID(date, playlistID string) error {
	_, err := s.db.Exec(`UPDATE daily_playlists SET playlist_id = ? WHERE date = ?`, playlistID, date)
	return err
}

func (s *Store) AddThemeSuggestion(ts *ThemeSuggestion) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO theme_suggestions (user_id, theme, status, created_at)
		VALUES (?, ?, ?, ?)
	`, ts.UserID, ts.Theme, ts.Status, ts.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetThemeSuggestions(status string) ([]ThemeSuggestion, error) {
	rows, err := s.db.Query(`
		SELECT ts.id, ts.user_id, ts.theme, ts.status, ts.created_at, u.display_name
		FROM theme_suggestions ts
		JOIN users u ON ts.user_id = u.id
		WHERE ts.status = ?
		ORDER BY ts.created_at DESC
	`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []ThemeSuggestion
	for rows.Next() {
		var ts ThemeSuggestion
		if err := rows.Scan(&ts.ID, &ts.UserID, &ts.Theme, &ts.Status, &ts.CreatedAt, &ts.UserName); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, ts)
	}
	return suggestions, nil
}

func (s *Store) DeleteThemeSuggestion(id int64) error {
	_, err := s.db.Exec("DELETE FROM theme_suggestions WHERE id = ?", id)
	return err
}

func (s *Store) GetCurrentTheme() (string, error) {
	date := time.Now().Format("2006-01-02")
	dp, err := s.GetDailyPlaylist(date)
	if err != nil {
		return "", nil
	}
	return dp.Theme, nil
}

func (s *Store) CreateArchive(archive *Archive) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO archives (date, playlist_id, playlist_name, track_count, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, archive.Date, archive.PlaylistID, archive.PlaylistName, archive.TrackCount, archive.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetArchives() ([]Archive, error) {
	rows, err := s.db.Query(`
		SELECT id, date, playlist_id, playlist_name, track_count, created_at
		FROM archives
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var archives []Archive
	for rows.Next() {
		var a Archive
		if err := rows.Scan(&a.ID, &a.Date, &a.PlaylistID, &a.PlaylistName, &a.TrackCount, &a.CreatedAt); err != nil {
			return nil, err
		}
		archives = append(archives, a)
	}
	return archives, nil
}

func (s *Store) SetUserName(name string) error {
	_, err := s.db.Exec(`
		INSERT INTO user_names (name, created_at)
		VALUES (?, ?)
	`, name, time.Now().Unix())
	return err
}

func (s *Store) GetUserName() (string, error) {
	var name string
	err := s.db.QueryRow(`SELECT name FROM user_names ORDER BY id DESC LIMIT 1`).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (s *Store) AddTrackAddition(track *TrackAddition) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO track_additions (video_id, date, added_by, created_at)
		VALUES (?, ?, ?, ?)
	`, track.VideoID, track.Date, track.AddedBy, track.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) RemoveTrackAddition(videoID string) error {
	_, err := s.db.Exec(`DELETE FROM track_additions WHERE video_id = ?`, videoID)
	return err
}

func (s *Store) GetTrackAdditions(date string) ([]TrackAddition, error) {
	rows, err := s.db.Query(`
		SELECT id, video_id, date, added_by, created_at
		FROM track_additions
		WHERE date = ?
		ORDER BY created_at DESC
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var additions []TrackAddition
	for rows.Next() {
		var t TrackAddition
		if err := rows.Scan(&t.ID, &t.VideoID, &t.Date, &t.AddedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		additions = append(additions, t)
	}
	return additions, nil
}
