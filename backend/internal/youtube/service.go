package youtube

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"banger-friday/internal/models"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Service struct {
	config       *oauth2.Config
	refreshToken string
	store        *models.Store
	httpClient   *http.Client
}

type Video struct {
	ID        string
	Title     string
	Channel   string
	Thumbnail string
	URL       string
}

type PlaylistItem struct {
	Video   Video
	AddedBy string
}

func NewService(refreshToken string, store *models.Store) *Service {
	clientID := os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")
	redirectURL := os.Getenv("YOUTUBE_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/callback"
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{youtube.YoutubeScope},
		Endpoint:     google.Endpoint,
	}

	return &Service{
		config:       config,
		refreshToken: refreshToken,
		store:        store,
	}
}

func (s *Service) getClient(ctx context.Context) (*http.Client, error) {
	token := &oauth2.Token{
		RefreshToken: s.refreshToken,
		TokenType:    "Bearer",
	}

	ts := s.config.TokenSource(ctx, token)
	newToken, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %v", err)
	}

	if newToken.RefreshToken != "" && newToken.RefreshToken != s.refreshToken {
		s.refreshToken = newToken.RefreshToken
	}

	return s.config.Client(ctx, newToken), nil
}

func (s *Service) SearchVideos(query string) ([]Video, error) {
	ctx := context.Background()

	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("error creating YouTube service: %v", err)
	}

	searchQuery := query + " music"
	call := service.Search.List([]string{"snippet"}).
		Q(searchQuery).
		Type("video").
		VideoCategoryId("10").
		MaxResults(10)

	response, err := call.Do()
	if err != nil {
		call2 := service.Search.List([]string{"snippet"}).
			Q(searchQuery).
			Type("video").
			MaxResults(10)
		response, err = call2.Do()
		if err != nil {
			return nil, fmt.Errorf("error searching YouTube: %v", err)
		}
	}

	var videos []Video
	for _, item := range response.Items {
		title := item.Snippet.Title
		channel := item.Snippet.ChannelTitle

		lowerTitle := strings.ToLower(title)
		lowerChannel := strings.ToLower(channel)
		if strings.Contains(lowerTitle, "channel") ||
			strings.Contains(lowerTitle, "playlist") ||
			strings.Contains(lowerTitle, "movie") ||
			(strings.Contains(lowerTitle, "trailer") && !strings.Contains(lowerTitle, "music")) ||
			strings.Contains(lowerChannel, "channel") {
			continue
		}

		thumbnails := item.Snippet.Thumbnails
		thumbnail := ""
		if thumbnails.Maxres != nil {
			thumbnail = thumbnails.Maxres.Url
		} else if thumbnails.High != nil {
			thumbnail = thumbnails.High.Url
		} else if thumbnails.Medium != nil {
			thumbnail = thumbnails.Medium.Url
		}

		videos = append(videos, Video{
			ID:        item.Id.VideoId,
			Title:     title,
			Channel:   channel,
			Thumbnail: thumbnail,
			URL:       "https://youtube.com/watch?v=" + item.Id.VideoId,
		})
	}

	return videos, nil
}

func (s *Service) CreateOrGetPlaylist(ctx context.Context, date, theme string) (string, error) {
	hasDailyPlaylistRecord := false
	existingTheme := ""

	dp, err := s.store.GetDailyPlaylist(date)
	if err == nil {
		hasDailyPlaylistRecord = true
		existingTheme = dp.Theme
		if dp.PlaylistID != "" {
			return dp.PlaylistID, nil
		}
		if theme == "" {
			theme = existingTheme
		}
	} else if err != sql.ErrNoRows {
		return "", err
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return "", err
	}

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("error creating YouTube service: %v", err)
	}

	playlistName := fmt.Sprintf("Bangers %s", date)
	if theme != "" {
		playlistName = fmt.Sprintf("Bangers %s - %s", date, theme)
	}

	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       playlistName,
			Description: "Banger Friday playlist",
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "public",
		},
	}

	resp, err := service.Playlists.Insert([]string{"snippet", "status"}, playlist).Do()
	if err != nil {
		return "", fmt.Errorf("error creating playlist: %v", err)
	}

	if hasDailyPlaylistRecord {
		err = s.store.UpdateDailyPlaylistID(date, resp.Id)
		if err != nil {
			return "", err
		}

		if theme != "" && existingTheme != theme {
			err = s.store.UpdateDailyPlaylistTheme(date, theme)
			if err != nil {
				return "", err
			}
		}
	} else {
		newDP := &models.DailyPlaylist{
			Date:       date,
			PlaylistID: resp.Id,
			Theme:      theme,
			CreatedAt:  time.Now().Unix(),
		}
		_, err = s.store.CreateDailyPlaylist(newDP)
		if err != nil {
			return "", err
		}
	}

	return resp.Id, nil
}

func (s *Service) AddVideoToPlaylist(ctx context.Context, playlistID, videoID, addedBy string) error {
	client, err := s.getClient(ctx)
	if err != nil {
		return err
	}

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error creating YouTube service: %v", err)
	}

	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			Title:       "Added by " + addedBy,
			Description: "Added via Banger Friday",
			PlaylistId:  playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}

	_, err = service.PlaylistItems.Insert([]string{"snippet", "contentDetails"}, playlistItem).Do()
	if err != nil {
		return fmt.Errorf("error adding video to playlist: %v", err)
	}

	date := time.Now().Format("2006-01-02")
	_, err = s.store.AddTrackAddition(&models.TrackAddition{
		VideoID:   videoID,
		Date:      date,
		AddedBy:   addedBy,
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		log.Printf("warning: failed to track addition video=%s added_by=%s date=%s err=%v", videoID, addedBy, date, err)
	}

	return nil
}

func (s *Service) RemoveVideoFromPlaylist(ctx context.Context, playlistID, videoID string) error {
	client, err := s.getClient(ctx)
	if err != nil {
		return err
	}

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error creating YouTube service: %v", err)
	}

	var playlistItemID string
	pageToken := ""
	for {
		call := service.PlaylistItems.List([]string{"id", "snippet"}).
			PlaylistId(playlistID).
			MaxResults(50)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		response, err := call.Do()
		if err != nil {
			return fmt.Errorf("error getting playlist items: %v", err)
		}

		for _, item := range response.Items {
			if item.Snippet.ResourceId.VideoId == videoID {
				playlistItemID = item.Id
				break
			}
		}

		if playlistItemID != "" || response.NextPageToken == "" {
			break
		}

		pageToken = response.NextPageToken
	}

	if playlistItemID == "" {
		return fmt.Errorf("video not found in playlist")
	}

	err = service.PlaylistItems.Delete(playlistItemID).Do()
	if err != nil {
		return fmt.Errorf("error removing video from playlist: %v", err)
	}

	if err := s.store.RemoveTrackAddition(videoID); err != nil {
		log.Printf("warning: failed to remove track addition video=%s err=%v", videoID, err)
	}

	return nil
}

func (s *Service) GetPlaylistVideos(ctx context.Context, playlistID string) ([]PlaylistItem, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("error creating YouTube service: %v", err)
	}

	call := service.PlaylistItems.List([]string{"snippet"}).
		PlaylistId(playlistID).
		MaxResults(50)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error getting playlist items: %v", err)
	}

	date := time.Now().Format("2006-01-02")
	additions, _ := s.store.GetTrackAdditions(date)

	var items []PlaylistItem
	for _, item := range response.Items {
		thumbnails := item.Snippet.Thumbnails
		thumbnail := ""
		if thumbnails.Maxres != nil {
			thumbnail = thumbnails.Maxres.Url
		} else if thumbnails.High != nil {
			thumbnail = thumbnails.High.Url
		} else if thumbnails.Medium != nil {
			thumbnail = thumbnails.Medium.Url
		}

		addedBy := "Unknown"
		for _, a := range additions {
			if a.VideoID == item.Snippet.ResourceId.VideoId {
				addedBy = a.AddedBy
				break
			}
		}

		items = append(items, PlaylistItem{
			Video: Video{
				ID:        item.Snippet.ResourceId.VideoId,
				Title:     item.Snippet.Title,
				Channel:   item.Snippet.ChannelTitle,
				Thumbnail: thumbnail,
				URL:       "https://youtube.com/watch?v=" + item.Snippet.ResourceId.VideoId,
			},
			AddedBy: addedBy,
		})
	}

	return items, nil
}

func (s *Service) CreateArchivePlaylist(ctx context.Context, date, theme string, trackCount int) (string, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return "", err
	}

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("error creating YouTube service: %v", err)
	}

	playlistName := fmt.Sprintf("Bangers %s", date)
	if theme != "" {
		playlistName = fmt.Sprintf("Bangers %s - %s", date, theme)
	}

	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       playlistName,
			Description: fmt.Sprintf("Archived Banger Friday - %d tracks", trackCount),
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "public",
		},
	}

	resp, err := service.Playlists.Insert([]string{"snippet", "status"}, playlist).Do()
	if err != nil {
		return "", fmt.Errorf("error creating archive playlist: %v", err)
	}

	archive := &models.Archive{
		Date:         date,
		PlaylistID:   resp.Id,
		PlaylistName: playlistName,
		TrackCount:   trackCount,
		CreatedAt:    time.Now().Unix(),
	}
	_, err = s.store.CreateArchive(archive)
	if err != nil {
		return "", err
	}

	return resp.Id, nil
}

func (s *Service) KeepAuthAlive(ctx context.Context) error {
	client, err := s.getClient(ctx)
	if err != nil {
		return err
	}

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error creating YouTube service: %v", err)
	}

	_, err = service.Playlists.List([]string{"id"}).Mine(true).MaxResults(1).Do()
	if err != nil {
		return fmt.Errorf("error running keepalive check: %v", err)
	}

	return nil
}

func (s *Service) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state)
}

func (s *Service) ExchangeToken(ctx context.Context, code string) (string, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	return token.RefreshToken, nil
}
