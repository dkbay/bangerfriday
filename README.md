# Banger Friday 🎵

A shared playlist app for Friday bangers with Spotify integration.

## Features

- **Spotify OAuth** - Users login with Spotify to add tracks
- **Daily Playlist** - Automatically creates a playlist for each day
- **Theme System** - Users suggest themes, admin picks one
- **Auto-Archive** - End-of-day archive creates permanent playlist
- **Admin Panel** - Pick random/manual theme, archive playlist

## Prerequisites

- Go 1.21+
- Node.js 20+
- Spotify Developer Account

## Setup

### 1. Spotify App Setup

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create a new app
3. Add redirect URI: `http://localhost:8080/auth/callback`
4. Get Client ID and Client Secret

### 2. Local Development

```bash
# Copy environment file
cp .env.example .env
# Edit .env with your Spotify credentials

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
| SPOTIFY_CLIENT_ID | From Spotify Dashboard | required |
| SPOTIFY_CLIENT_SECRET | From Spotify Dashboard | required |
| REDIRECT_URI | OAuth callback URL | http://localhost:8080/auth/callback |
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

```
banger-friday/
├── backend/
│   ├── cmd/main.go         # Entry point
│   ├── internal/
│   │   ├── handlers/       # HTTP handlers
│   │   ├── models/         # Database models
│   │   ├── spotify/        # Spotify API service
│   │   └── db/             # Database setup
│   └── go.mod
├── frontend/
│   ├── src/App.vue         # Main Vue component
│   ├── index.html
│   └── package.json
├── docker-compose.yml
└── .env.example
```