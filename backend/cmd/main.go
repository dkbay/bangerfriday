package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"banger-friday/internal/db"
	"banger-friday/internal/handlers"
	"banger-friday/internal/models"
	"banger-friday/internal/youtube"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Config struct {
	YouTubeClientID     string
	YouTubeClientSecret string
	YouTubeRefreshToken string
	ServerPort          string
	AdminUsers          string
}

func loadConfig() Config {
	return Config{
		YouTubeClientID:     getEnv("YOUTUBE_CLIENT_ID", ""),
		YouTubeClientSecret: getEnv("YOUTUBE_CLIENT_SECRET", ""),
		YouTubeRefreshToken: getEnv("YOUTUBE_REFRESH_TOKEN", ""),
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		AdminUsers:          getEnv("ADMIN_USERS", "lbk,mby"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func startYouTubeTokenKeepAlive(youtubeService *youtube.Service) {
	runKeepAlive := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := youtubeService.KeepAuthAlive(ctx); err != nil {
			log.Printf("YouTube token keepalive failed: %v", err)
			return
		}

		log.Printf("YouTube token keepalive succeeded")
	}

	runKeepAlive()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			runKeepAlive()
		}
	}()
}

func main() {
	config := loadConfig()

	if config.YouTubeClientID == "" || config.YouTubeClientSecret == "" || config.YouTubeRefreshToken == "" {
		log.Fatal("YOUTUBE_CLIENT_ID, YOUTUBE_CLIENT_SECRET, and YOUTUBE_REFRESH_TOKEN must be set")
	}

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	store := models.NewStore(database)
	youtubeService := youtube.NewService(config.YouTubeRefreshToken, store)
	startYouTubeTokenKeepAlive(youtubeService)
	apiHandler := handlers.NewAPIHandler(youtubeService, store, handlers.Config{ServerPort: config.ServerPort, AdminUsersCSV: config.AdminUsers})

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://bangerfriday.dk", "http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := router.Group("/api")
	{
		api.GET("/theme", apiHandler.GetTheme)
		api.POST("/theme", apiHandler.SubmitTheme)
		api.POST("/theme/pick", apiHandler.PickTheme)
		api.GET("/theme/suggestions", apiHandler.GetThemeSuggestions)
		api.POST("/theme/suggestion/delete", apiHandler.DeleteThemeSuggestion)
		api.GET("/playlist/today", apiHandler.GetTodayPlaylist)
		api.POST("/playlist/add", apiHandler.AddToPlaylist)
		api.POST("/playlist/remove", apiHandler.RemoveFromPlaylist)
		api.POST("/playlist/archive", apiHandler.ArchivePlaylist)
		api.GET("/archive", apiHandler.GetArchives)
		api.GET("/search", apiHandler.SearchTracks)
		api.POST("/user/set-name", apiHandler.SetUserName)
		api.GET("/user/name", apiHandler.GetUserName)
		api.GET("/admin", apiHandler.IsAdmin)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: router,
	}

	log.Printf("Server starting on port %s", config.ServerPort)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
