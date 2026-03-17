# Banger Friday 🎵

A shared playlist app for Friday bangers with YouTube integration.

## Features

- **YouTube API** - Search and add tracks from YouTube
- **Daily Playlist** - Automatically creates a playlist for each day
- **Theme System** - Users suggest themes, admin picks one
- **Auto-Archive** - End-of-day archive creates permanent playlist
- **Admin Panel** - Pick random/manual theme, archive playlist

## Prerequisites

- Go 1.26+
- Node.js 20+
- Google Cloud project with YouTube Data API access

## Setup

### 1. YouTube App Setup

1. Create OAuth credentials in Google Cloud Console
2. Enable YouTube Data API v3
3. Add redirect URI: `http://localhost:8080/auth/callback`
4. Get Client ID, Client Secret, and a Refresh Token

### 2. Local Development

```bash
# Copy environment file
cp .env.example .env
# Edit .env with your YouTube credentials

# Backend
cd backend
go mod tidy
go run cmd/main.go

# Frontend (in another terminal)
cd frontend
npm install
npm run dev
```

### 3. Docker Deployment

```bash
# Build and run
docker-compose up --build
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| YOUTUBE_CLIENT_ID | From Google Cloud OAuth credentials | required |
| YOUTUBE_CLIENT_SECRET | From Google Cloud OAuth credentials | required |
| YOUTUBE_REFRESH_TOKEN | OAuth refresh token for playlist operations | required |
| YOUTUBE_REDIRECT_URL | OAuth callback URL | http://localhost:8080/auth/callback |
| ADMIN_USERS | Comma-separated admin names | lbk,mby |
| ADMIN_PASSWORD | Password to make users admin | changeme |

## Making Someone Admin

1. User logs in normally
2. Call the API (or use a tool like curl):
```bash
curl -X POST http://localhost:8080/auth/make-admin \
  -H "Content-Type: application/json" \
  -d '{"password": "changeme", "user_id": 1}'
```
(Replace `user_id` with the actual user ID from the database)

## Usage Flow

1. **Theme**: Admin clicks a suggested theme or enters one manually
2. **Add Bangers**: Users search and add songs throughout the day
3. **Play**: The playlist plays on a shared speaker/TV
4. **Archive**: At end of day, admin clicks "Archive Playlist" to save it permanently

## Project Structure

```text
banger-friday/
├── backend/
│   ├── cmd/main.go         # Entry point
│   ├── internal/
│   │   ├── handlers/       # HTTP handlers
│   │   ├── models/         # Database models
│   │   ├── youtube/        # YouTube API service
│   │   └── db/             # Database setup
│   └── go.mod
├── frontend/
│   ├── src/App.vue         # Main Vue component
│   ├── index.html
│   └── package.json
├── docker-compose.yml
└── .env.example
```
