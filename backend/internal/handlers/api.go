package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"banger-friday/internal/models"
	internalyoutube "banger-friday/internal/youtube"

	"github.com/gin-gonic/gin"
)

type Config struct {
	ServerPort    string
	AdminUsersCSV string
}

type APIHandler struct {
	youtubeService *internalyoutube.Service
	store          *models.Store
	config         Config
	adminUsers     map[string]struct{}
}

func NewAPIHandler(youtubeService *internalyoutube.Service, store *models.Store, config Config) *APIHandler {
	adminUsers := make(map[string]struct{})
	for _, user := range strings.Split(config.AdminUsersCSV, ",") {
		normalized := strings.TrimSpace(strings.ToLower(user))
		if normalized != "" {
			adminUsers[normalized] = struct{}{}
		}
	}

	return &APIHandler{
		youtubeService: youtubeService,
		store:          store,
		config:         config,
		adminUsers:     adminUsers,
	}
}

func (h *APIHandler) requesterName(c *gin.Context) string {
	userName, _ := c.Cookie("user_name")
	return strings.TrimSpace(userName)
}

func (h *APIHandler) isRequesterAdmin(c *gin.Context) bool {
	userName := strings.ToLower(h.requesterName(c))
	if userName == "" {
		return false
	}
	_, ok := h.adminUsers[userName]
	return ok
}

func (h *APIHandler) GetTheme(c *gin.Context) {
	theme, err := h.store.GetCurrentTheme()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"theme": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"theme": theme})
}

func (h *APIHandler) SubmitTheme(c *gin.Context) {
	userName := h.requesterName(c)
	if userName == "" {
		userName = "Anonymous"
	}

	var request struct {
		Theme string `json:"theme"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	existing, err := h.store.GetThemeSuggestions("pending")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check suggestions"})
		return
	}
	for _, s := range existing {
		if s.Theme == request.Theme {
			c.JSON(http.StatusBadRequest, gin.H{"error": "suggestion already exists"})
			return
		}
	}

	suggestion := &models.ThemeSuggestion{
		UserID:    0,
		UserName:  userName,
		Theme:     request.Theme,
		Status:    "pending",
		CreatedAt: time.Now().Unix(),
	}

	_, err = h.store.AddThemeSuggestion(suggestion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add suggestion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "theme suggestion added"})
}

func (h *APIHandler) PickTheme(c *gin.Context) {
	if !h.isRequesterAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}

	var request struct {
		Theme string `json:"theme"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	date := time.Now().Format("2006-01-02")
	err := h.store.UpsertDailyPlaylistTheme(date, request.Theme)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update theme"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"theme": request.Theme})
}

func (h *APIHandler) GetThemeSuggestions(c *gin.Context) {
	suggestions, err := h.store.GetThemeSuggestions("pending")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get suggestions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

func (h *APIHandler) DeleteThemeSuggestion(c *gin.Context) {
	var request struct {
		ID int64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	suggestion, err := h.store.GetThemeSuggestionByID(request.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "suggestion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load suggestion"})
		return
	}

	requester := h.requesterName(c)
	if !h.isRequesterAdmin(c) && !strings.EqualFold(requester, suggestion.UserName) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not allowed to delete this suggestion"})
		return
	}

	err = h.store.DeleteThemeSuggestion(request.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete suggestion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "suggestion deleted"})
}

func (h *APIHandler) GetTodayPlaylist(c *gin.Context) {
	date := time.Now().Format("2006-01-02")
	dp, err := h.store.GetDailyPlaylist(date)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"tracks": []interface{}{}})
		return
	}

	videos, err := h.youtubeService.GetPlaylistVideos(c.Request.Context(), dp.PlaylistID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"playlist_id": dp.PlaylistID,
			"theme":       dp.Theme,
			"tracks":      []interface{}{},
		})
		return
	}

	trackList := make([]gin.H, 0, len(videos))
	for _, v := range videos {
		trackList = append(trackList, gin.H{
			"id":        v.Video.ID,
			"name":      v.Video.Title,
			"artist":    v.Video.Channel,
			"album_art": v.Video.Thumbnail,
			"uri":       v.Video.URL,
			"added_by":  v.AddedBy,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"playlist_id": dp.PlaylistID,
		"theme":       dp.Theme,
		"tracks":      trackList,
	})
}

func (h *APIHandler) AddToPlaylist(c *gin.Context) {
	theme, _ := h.store.GetCurrentTheme()
	if theme == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "waiting for theme to be picked"})
		return
	}

	userName, _ := c.Cookie("user_name")
	if userName == "" {
		userName = "Anonymous"
	}

	var request struct {
		TrackID string `json:"track_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	date := time.Now().Format("2006-01-02")
	dp, err := h.store.GetDailyPlaylist(date)

	playlistID := ""
	if err != nil || dp.PlaylistID == "" {
		themeToUse := theme
		if err == nil && dp.Theme != "" {
			themeToUse = dp.Theme
		}

		playlistID, err = h.youtubeService.CreateOrGetPlaylist(c.Request.Context(), date, themeToUse)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create playlist"})
			return
		}

		if err := h.store.UpdateDailyPlaylistID(date, playlistID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist playlist"})
			return
		}
	} else {
		playlistID = dp.PlaylistID
	}

	err = h.youtubeService.AddVideoToPlaylist(c.Request.Context(), playlistID, request.TrackID, userName)
	if err != nil {
		log.Printf("error adding track %s to playlist %s: %v", request.TrackID, playlistID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add track"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "track added"})
}

func (h *APIHandler) SearchTracks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing query"})
		return
	}

	videos, err := h.youtubeService.SearchVideos(query)
	if err != nil {
		log.Printf("search error for query %q: %v", query, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search"})
		return
	}

	results := make([]gin.H, 0, len(videos))
	for _, video := range videos {
		results = append(results, gin.H{
			"id":        video.ID,
			"name":      video.Title,
			"artist":    video.Channel,
			"album_art": video.Thumbnail,
			"uri":       video.URL,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tracks": results})
}

func (h *APIHandler) RemoveFromPlaylist(c *gin.Context) {
	date := time.Now().Format("2006-01-02")
	dp, err := h.store.GetDailyPlaylist(date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no playlist for today"})
		return
	}

	var request struct {
		TrackID string `json:"track_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	requester := h.requesterName(c)
	if requester == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "name is required"})
		return
	}

	if !h.isRequesterAdmin(c) {
		addition, err := h.store.GetTrackAdditionByVideoID(date, request.TrackID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusForbidden, gin.H{"error": "only admins can remove this track"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate track ownership"})
			return
		}

		if !strings.EqualFold(addition.AddedBy, requester) {
			c.JSON(http.StatusForbidden, gin.H{"error": "only admins or track owner can remove"})
			return
		}
	}

	err = h.youtubeService.RemoveVideoFromPlaylist(c.Request.Context(), dp.PlaylistID, request.TrackID)
	if err != nil {
		log.Printf("error removing track %s from playlist %s: %v", request.TrackID, dp.PlaylistID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove track"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "track removed"})
}

func (h *APIHandler) ArchivePlaylist(c *gin.Context) {
	if !h.isRequesterAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}

	date := time.Now().Format("2006-01-02")
	dp, err := h.store.GetDailyPlaylist(date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no playlist for today"})
		return
	}

	playlistID, err := h.youtubeService.CreateArchivePlaylist(c.Request.Context(), date, dp.Theme, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create archive"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "playlist archived", "playlist_id": playlistID})
}

func (h *APIHandler) GetArchives(c *gin.Context) {
	archives, err := h.store.GetArchives()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get archives"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"archives": archives})
}

func (h *APIHandler) IsAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"is_admin": h.isRequesterAdmin(c)})
}

func (h *APIHandler) SetUserName(c *gin.Context) {
	var request struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if request.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
		return
	}

	err := h.store.SetUserName(request.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save name"})
		return
	}

	c.SetCookie("user_name", request.Name, 86400*30, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "name saved"})
}

func (h *APIHandler) GetUserName(c *gin.Context) {
	userName, _ := c.Cookie("user_name")
	if userName == "" {
		name, err := h.store.GetUserName()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"name": ""})
			return
		}
		userName = name
		c.SetCookie("user_name", userName, 86400*30, "/", "", false, true)
	}
	c.JSON(http.StatusOK, gin.H{"name": userName})
}
