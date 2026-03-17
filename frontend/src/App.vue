<template>
  <div class="app">
    <header class="header">
      <h1>Banger Friday</h1>
      <div class="name-section" v-if="userName">
        <span>{{ userName }}</span>
        <button @click="showNameInput = true" class="btn-small">Change</button>
      </div>
    </header>

    <div v-if="notification" class="notification" :class="notification.type">
      {{ notification.message }}
    </div>

    <div v-if="!userName && !showNameInput" class="name-prompt">
      <h2>Welcome to Banger Friday</h2>
      <p>Enter your name to start adding bangers</p>
      <div class="name-input-row">
        <input 
          v-model="tempName" 
          @keyup.enter="setName"
          placeholder="Your name..."
          class="name-input"
        >
        <button @click="setName" class="btn-primary">Continue</button>
      </div>
    </div>

    <div v-if="showNameInput && userName" class="name-input-section">
      <input 
        v-model="tempName" 
        @keyup.enter="setName"
        placeholder="Your name..."
        class="name-input"
      >
      <button @click="setName" class="btn-primary">Save</button>
      <button @click="showNameInput = false" class="btn-secondary">Cancel</button>
    </div>

    <main class="main" v-if="userName">
      <!-- Admin Panel - visible only for LBK/MBY -->
      <section v-if="isAdmin" class="admin-section">
        <div class="admin-header">
          <h3>Admin</h3>
        </div>
        
        <div class="admin-content">
          <div v-if="themeSuggestions.length" class="suggestions-list">
            <div 
              v-for="s in themeSuggestions" 
              :key="s.ID" 
              class="suggestion-item"
            >
              <span class="suggestion-text" @click="pickTheme(s.Theme)">{{ s.Theme }}</span>
              <button class="btn-remove-suggestion" @click.stop="removeSuggestion(s.ID)">x</button>
            </div>
          </div>
          <div v-else class="no-suggestions">
            No theme suggestions
          </div>
          
          <div class="admin-row">
            <input v-model="manualTheme" placeholder="Or enter theme manually..." class="theme-input">
            <button @click.prevent="pickTheme(manualTheme)" class="btn-primary">Set Theme</button>
          </div>
          
          <button @click="archivePlaylist" class="btn-danger">End of Day - Archive</button>
        </div>
      </section>

      <!-- Theme Section -->
      <section class="theme-section">
        <div v-if="theme" class="current-theme">
          <span class="theme-label">Theme</span>
          <span class="theme-value">{{ theme }}</span>
        </div>
        <div v-else class="no-theme">
          <span class="no-theme-text">No theme</span>
        </div>
        
        <div class="theme-suggestion">
          <input 
            v-model="newTheme" 
            @keyup.enter.prevent="submitTheme"
            placeholder="Suggest theme..."
            class="theme-input"
          >
          <button @click.prevent="submitTheme" class="btn-primary">Submit</button>
        </div>
      </section>

      <!-- Player Section -->
      <section class="player-section" v-if="tracks.length > 0">
        <div class="video-container">
          <div v-if="currentIndex === -1" class="video-placeholder">
            <span>Select a track to play</span>
          </div>
          <div id="youtube-player"></div>
        </div>
        <div class="player-controls">
          <button @click="togglePlay" class="btn-control">
            {{ isPlaying ? 'Pause' : 'Play' }}
          </button>
          <button @click="skipNext" class="btn-control">Next</button>
          <button @click="toggleLoop" class="btn-control" :class="{ active: loopEnabled }">Loop</button>
          <button @click="toggleShuffle" class="btn-control" :class="{ active: shuffleEnabled }">Shuffle</button>
          <span class="now-playing" v-if="currentTrack">
            {{ currentTrack.name }}
          </span>
        </div>
      </section>

      <!-- Search Section -->
      <section class="search-section">
        <h3>Add a Banger</h3>
        <div v-if="!theme" class="waiting-message">
          Waiting for theme to be picked...
        </div>
        <div v-else>
          <div class="search-box">
            <input 
              v-model="searchQuery" 
              @input="debounceSearch"
              placeholder="Search for a video..."
              class="search-input"
            >
          </div>
          
          <div v-if="searchResults.length" class="search-results">
            <div 
              v-for="track in searchResults" 
              :key="track.id" 
              class="track-item"
              @click.prevent="addTrack(track.id)"
            >
              <img :src="track.album_art" :alt="track.name" class="track-art">
              <div class="track-info">
                <span class="track-name">{{ track.name }}</span>
                <span class="track-artist">{{ track.artist }}</span>
              </div>
              <button class="btn-add" @click.prevent.stop="addTrack(track.id)">+</button>
            </div>
          </div>
        </div>
      </section>

      <!-- Playlist Section -->
      <section class="playlist-section">
        <h3>Today's Bangers ({{ tracks.length }})</h3>
        <div v-if="tracks.length" class="track-list">
          <div 
            v-for="(track, index) in tracks" 
            :key="track.id" 
            class="track-item"
            :class="{ active: currentIndex === index }"
            @click="playTrack(index)"
          >
            <img :src="track.album_art" :alt="track.name" class="track-art">
            <div class="track-info">
              <span class="track-name">{{ track.name }}</span>
              <span v-if="track.added_by" class="added-by">{{ track.added_by }}</span>
            </div>
            <button v-if="isAdmin" class="btn-remove" @click.stop="removeTrack(track.id)">x</button>
          </div>
        </div>
        <p v-else class="no-tracks">No tracks added yet</p>
      </section>

      <!-- Archive Section -->
      <section class="archive-section">
        <h3>Previous Bangers</h3>
        <div v-if="archives.length" class="archive-list">
          <a 
            v-for="archive in archives" 
            :key="archive.id" 
            class="archive-item"
            :href="`https://youtube.com/playlist?list=${archive.playlist_id}`"
            target="_blank"
          >
            <span class="archive-date">{{ archive.date }}</span>
            <span class="archive-name">{{ archive.playlist_name }}</span>
            <span class="archive-count">{{ archive.track_count }} tracks</span>
          </a>
        </div>
        <p v-else class="no-archives">No archives yet</p>
      </section>
    </main>

  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      userName: '',
      tempName: '',
      showNameInput: false,
      theme: '',
      newTheme: '',
      manualTheme: '',
      searchQuery: '',
      searchResults: [],
      tracks: [],
      archives: [],
      themeSuggestions: [],
      searchTimeout: null,
      player: null,
      playerReady: false,
      currentIndex: -1,
      isPlaying: false,
      playlistIds: [],
      notification: null,
      refreshInterval: null,
      loopEnabled: localStorage.getItem('bangerfriday_loop') === 'true',
      shuffleEnabled: localStorage.getItem('bangerfriday_shuffle') === 'true',
      shuffleOrder: [],
      waitingForNewTracks: false,
      userIsAdmin: false,
      youtubeInitAttempts: 0,
      maxYouTubeInitRetries: 50
    }
  },
  computed: {
    isAdmin() {
      return this.userIsAdmin
    },
    currentTrack() {
      if (this.currentIndex >= 0 && this.currentIndex < this.tracks.length) {
        return this.tracks[this.currentIndex]
      }
      return null
    }
  },
  async mounted() {
    window.onYouTubeIframeAPIReady = () => {
      this.youtubeInitAttempts = 0
      this.initYouTubePlayer()
    }

    this.userName = localStorage.getItem('bangerfriday_name') || ''
    if (this.userName) {
      await this.apiCall('/api/user/set-name', {
        method: 'POST',
        body: JSON.stringify({ name: this.userName })
      })
      await this.loadData()
    }
    this.initYouTubePlayer()
    
    this.refreshInterval = setInterval(() => {
      if (this.userName) {
        this.loadDataSilent()
      }
    }, 5000)
  },
  beforeUnmount() {
    if (this.player) {
      this.player.destroy()
    }
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval)
    }
    if (window.onYouTubeIframeAPIReady) {
      window.onYouTubeIframeAPIReady = null
    }
  },
  watch: {
    tracks: {
      handler(newTracks, oldTracks) {
        if (newTracks.length > (oldTracks ? oldTracks.length : 0)) {
          if (this.player && this.player.addToPlaylist) {
            try {
              const newIds = newTracks.map(t => t.id)
              for (let i = (oldTracks ? oldTracks.length : 0); i < newTracks.length; i++) {
                this.player.addToPlaylist(newIds[i])
              }
            } catch (e) {
              console.warn('Could not add to playlist:', e)
            }
          }

          if (this.waitingForNewTracks && newTracks.length > 0) {
            const newTrackIndex = newTracks.length - 1
            this.currentIndex = newTrackIndex
            this.waitingForNewTracks = false
            if (this.shuffleEnabled) {
              this.generateShuffleOrder()
            }
            if (this.player && this.playerReady && this.player.loadVideoById) {
              this.player.loadVideoById({ videoId: newTracks[newTrackIndex].id })
            }
          }
        }
      },
      deep: true
    }
  },
  methods: {
    initYouTubePlayer() {
      if (window.YT && window.YT.Player) {
        this.youtubeInitAttempts = 0
        this.player = new window.YT.Player('youtube-player', {
          height: '360',
          width: '100%',
          playerVars: {
            autoplay: 0,
            controls: 1,
            modestbranding: 1,
            rel: 0,
            listType: 'playlist',
            list: ''
          },
          events: {
            onReady: () => {
              this.playerReady = true
            },
            onStateChange: (event) => {
              if (event.data === window.YT.PlayerState.ENDED) {
                this.playNext()
              } else if (event.data === window.YT.PlayerState.PLAYING) {
                this.isPlaying = true
              } else if (event.data === window.YT.PlayerState.PAUSED) {
                this.isPlaying = false
              }
            }
          }
        })
      } else if (this.youtubeInitAttempts < this.maxYouTubeInitRetries) {
        this.youtubeInitAttempts += 1
        setTimeout(() => this.initYouTubePlayer(), 100)
      } else {
        console.error('YouTube API failed to load after retries')
      }
    },
    playTrack(index) {
      this.currentIndex = index
      if (this.player && this.playerReady) {
        const videoId = this.tracks[index].id
        if (this.player.loadVideoById) {
          this.player.loadVideoById({ videoId: videoId })
        }
      }
    },
    playNext() {
      if (this.tracks.length === 0) return

      let nextIndex = -1

      if (this.shuffleEnabled && this.shuffleOrder.length > 0) {
        const currentPos = this.shuffleOrder.indexOf(this.currentIndex)
        if (currentPos >= 0 && currentPos < this.shuffleOrder.length - 1) {
          nextIndex = this.shuffleOrder[currentPos + 1]
        } else {
          nextIndex = this.shuffleOrder[0]
        }
      } else {
        if (this.currentIndex < this.tracks.length - 1) {
          nextIndex = this.currentIndex + 1
        } else if (this.loopEnabled) {
          nextIndex = 0
        }
      }

      if (nextIndex === -1) {
        this.waitingForNewTracks = true
        this.isPlaying = false
        return
      }

      this.currentIndex = nextIndex
      this.waitingForNewTracks = false
      if (this.player && this.playerReady && this.player.loadVideoById) {
        this.player.loadVideoById({ videoId: this.tracks[this.currentIndex].id })
      }
    },
    skipNext() {
      this.playNext()
    },
    togglePlay() {
      if (this.player && this.playerReady) {
        if (this.isPlaying) {
          this.player.pauseVideo()
        } else {
          if (this.currentIndex === -1 && this.tracks.length > 0) {
            this.currentIndex = 0
            this.waitingForNewTracks = false
            this.player.loadVideoById({ videoId: this.tracks[0].id })
          } else if (this.currentIndex >= 0) {
            this.player.playVideo()
          }
        }
      }
    },
    toggleLoop() {
      this.loopEnabled = !this.loopEnabled
      localStorage.setItem('bangerfriday_loop', this.loopEnabled)
    },
    toggleShuffle() {
      this.shuffleEnabled = !this.shuffleEnabled
      localStorage.setItem('bangerfriday_shuffle', this.shuffleEnabled)
      if (this.shuffleEnabled) {
        this.generateShuffleOrder()
      } else {
        this.shuffleOrder = []
      }
    },
    generateShuffleOrder() {
      const indices = Array.from({ length: this.tracks.length }, (_, i) => i)
      for (let i = indices.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [indices[i], indices[j]] = [indices[j], indices[i]]
      }
      this.shuffleOrder = indices
    },
    async apiCall(endpoint, options = {}) {
      try {
        const response = await fetch(endpoint, {
          ...options,
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            ...options.headers
          }
        })

        const contentType = response.headers.get('content-type') || ''
        const body = contentType.includes('application/json')
          ? await response.json().catch(() => ({}))
          : {}

        if (!response.ok) {
          return {
            error: body.error || `${response.status} ${response.statusText}`.trim()
          }
        }

        return body
      } catch (e) {
        return { error: `network error: ${e.message}` }
      }
    },
    showNotification(message, type = 'info') {
      this.notification = { message, type }
      setTimeout(() => {
        this.notification = null
      }, 3000)
    },
    async loadDataSilent() {
      try {
        const [themeData, playlistData] = await Promise.all([
          this.apiCall('/api/theme'),
          this.apiCall('/api/playlist/today')
        ])
        
        const newTracks = playlistData.tracks || []
        const oldIds = this.tracks.map(t => t.id)
        const newIds = newTracks.map(t => t.id)
        const hasChanges = JSON.stringify(oldIds) !== JSON.stringify(newIds)
        
        this.theme = themeData.theme || ''
        
        if (hasChanges) {
          this.tracks = newTracks
        }
      } catch (e) {
        // Silent fail
      }
    },
    async setName() {
      const name = this.tempName.trim()
      if (!name) return
      
      this.userName = name
      localStorage.setItem('bangerfriday_name', name)
      this.showNameInput = false
      this.tempName = ''
      
      await this.apiCall('/api/user/set-name', {
        method: 'POST',
        body: JSON.stringify({ name })
      })
      await this.loadData()
    },
    async loadData() {
      try {
        const [themeData, playlistData, archiveData, adminData] = await Promise.all([
          this.apiCall('/api/theme'),
          this.apiCall('/api/playlist/today'),
          this.apiCall('/api/archive'),
          this.apiCall('/api/admin')
        ])

        this.theme = themeData.theme || ''
        this.tracks = playlistData.tracks || []
        this.archives = archiveData.archives || []
        this.userIsAdmin = Boolean(adminData && adminData.is_admin)
        
        if (this.isAdmin) {
          const suggestionsData = await this.apiCall('/api/theme/suggestions')
          this.themeSuggestions = suggestionsData.suggestions || []
        }
        
        if (this.tracks.length > 0 && this.currentIndex === -1) {
          this.currentIndex = -1
        }
      } catch (e) {
        console.warn('Failed to load data:', e)
      }
    },
    async submitTheme() {
      if (!this.newTheme.trim()) return
      const result = await this.apiCall('/api/theme', {
        method: 'POST',
        body: JSON.stringify({ theme: this.newTheme })
      })
      if (result.error) {
        this.showNotification(result.error, 'error')
      } else {
        this.newTheme = ''
        this.showNotification('Theme suggested', 'success')
      }
    },
    async pickTheme(theme) {
      if (!theme) return
      const result = await this.apiCall('/api/theme/pick', {
        method: 'POST',
        body: JSON.stringify({ theme })
      })

      if (result.error) {
        this.showNotification(result.error, 'error')
        return
      }

      this.theme = result.theme || theme
      this.manualTheme = ''
      const suggestionsData = await this.apiCall('/api/theme/suggestions')
      this.themeSuggestions = suggestionsData.suggestions || []
      await this.loadData()
    },
    async archivePlaylist() {
      if (!confirm('Archive today playlist?')) return
      await this.apiCall('/api/playlist/archive', { method: 'POST' })
      await this.loadData()
      this.showNotification('Playlist archived', 'success')
    },
    debounceSearch() {
      clearTimeout(this.searchTimeout)
      this.searchTimeout = setTimeout(() => this.search(), 300)
    },
    async search() {
      if (!this.searchQuery.trim()) {
        this.searchResults = []
        return
      }
      const data = await this.apiCall(`/api/search?q=${encodeURIComponent(this.searchQuery)}`)
      this.searchResults = data.tracks || []
    },
    async addTrack(trackId, event) {
      if (event) {
        event.preventDefault()
        event.stopPropagation()
      }
      try {
        const trackToAdd = this.searchResults.find(t => t.id === trackId)
        
        const result = await this.apiCall('/api/playlist/add', {
          method: 'POST',
          body: JSON.stringify({ track_id: trackId })
        })
        
        if (result.error) {
          this.showNotification(result.error, 'error')
        } else {
          this.showNotification('Track added', 'success')
          
          if (trackToAdd) {
            this.tracks.push({
              id: trackToAdd.id,
              name: trackToAdd.name,
              artist: trackToAdd.artist,
              album_art: trackToAdd.album_art,
              uri: trackToAdd.uri,
              added_by: this.userName
            })
            
            if (this.player && this.player.addToPlaylist) {
              try {
                this.player.addToPlaylist(trackId)
              } catch (e) {
                console.warn('Could not add to player:', e)
              }
            }
          }
        }
        
        this.searchResults = []
        this.searchQuery = ''
      } catch (e) {
        this.showNotification('Failed to add track', 'error')
      }
    },
    async removeTrack(trackId) {
      try {
        const result = await this.apiCall('/api/playlist/remove', {
          method: 'POST',
          body: JSON.stringify({ track_id: trackId })
        })
        
        if (result.error) {
          this.showNotification(result.error, 'error')
        } else {
          this.showNotification('Track removed', 'success')
          this.tracks = this.tracks.filter(t => t.id !== trackId)
        }
      } catch (e) {
        this.showNotification('Failed to remove track', 'error')
      }
    },
    async removeSuggestion(id) {
      try {
        const result = await this.apiCall('/api/theme/suggestion/delete', {
          method: 'POST',
          body: JSON.stringify({ id })
        })
        
        if (result.error) {
          this.showNotification(result.error, 'error')
        } else {
          this.themeSuggestions = this.themeSuggestions.filter(s => s.ID !== id)
        }
      } catch (e) {
        this.showNotification('Failed to remove suggestion', 'error')
      }
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #121212;
  min-height: 100vh;
  color: #fff;
}

.app {
  max-width: 800px;
  margin: 0 auto;
  padding: 24px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #2d2d2d;
  margin-bottom: 24px;
}

.header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

.name-section {
  display: flex;
  gap: 12px;
  align-items: center;
  color: #9ca3af;
}

.btn-small {
  padding: 6px 12px;
  border: 1px solid #2d2d2d;
  border-radius: 6px;
  cursor: pointer;
  background: transparent;
  color: #9ca3af;
  font-size: 12px;
}

.btn-small:hover {
  background: #1a1a1a;
}

.notification {
  position: fixed;
  top: 20px;
  right: 20px;
  padding: 12px 20px;
  font-weight: 500;
  z-index: 1001;
  font-size: 14px;
  border-radius: 6px;
}

.notification.success {
  background: #22c55e;
  color: #fff;
}

.notification.error {
  background: #ff4444;
  color: #fff;
}

.name-prompt, .name-input-section {
  padding: 48px;
  border: 1px solid #2d2d2d;
  text-align: center;
  border-radius: 6px;
  background: #1a1a1a;
}

.name-prompt h2, .name-input-section h2 {
  margin-bottom: 8px;
  font-weight: 500;
}

.name-prompt p {
  color: #9ca3af;
  margin-bottom: 24px;
}

.name-input-row {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.name-input {
  padding: 12px 16px;
  border: 1px solid #2d2d2d;
  border-radius: 6px;
  font-size: 14px;
  background: #1a1a1a;
  color: #fff;
  width: 200px;
}

.main {
  margin-top: 24px;
}

.admin-section {
  border: 1px solid #2d2d2d;
  margin-bottom: 24px;
  border-radius: 6px;
  background: #1a1a1a;
}

.admin-header {
  padding: 12px 16px;
  border-bottom: 1px solid #2d2d2d;
  background: #222;
}

.admin-header h3 {
  font-size: 14px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #9ca3af;
}

.admin-content {
  padding: 16px;
}

.admin-row {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.suggestions-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
  max-height: 200px;
  overflow-y: auto;
}

.suggestion-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #222;
  border: 1px solid #2d2d2d;
  font-size: 14px;
  border-radius: 6px;
}

.suggestion-text {
  cursor: pointer;
  color: #e5e5e5;
}

.suggestion-text:hover {
  color: #22c55e;
}

.btn-remove-suggestion {
  background: transparent;
  border: none;
  color: #666;
  cursor: pointer;
  font-size: 14px;
  padding: 0 4px;
}

.btn-remove-suggestion:hover {
  color: #fff;
}

.suggestion-item:hover {
  background: #252525;
}

.no-suggestions {
  color: #666;
  font-size: 13px;
  padding: 8px 0;
}

.btn-danger {
  padding: 10px 20px;
  border: none;
  background: #22c55e;
  color: #fff;
  font-weight: 500;
  cursor: pointer;
  font-size: 14px;
}

.btn-danger:hover {
  background: #16a34a;
}

.theme-section {
  margin-bottom: 24px;
}

.current-theme {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
}

.theme-label {
  color: #9ca3af;
  font-size: 13px;
  text-transform: lowercase;
}

.theme-value {
  font-size: 32px;
  font-weight: 600;
  letter-spacing: -0.5px;
  color: #22c55e;
}

.no-theme {
  margin-bottom: 16px;
}

.no-theme-text {
  color: #9ca3af;
  font-size: 13px;
}

.theme-suggestion {
  display: flex;
  gap: 8px;
}

.theme-input {
  flex: 1;
  padding: 8px 0;
  border: none;
  border-bottom: 1px solid #2d2d2d;
  font-size: 14px;
  background: transparent;
  color: #fff;
}

.theme-input::placeholder {
  color: #666;
}

.theme-input:focus {
  outline: none;
  border-bottom-color: #22c55e;
}

.btn-primary {
  padding: 10px 20px;
  border: none;
  background: #22c55e;
  color: #fff;
  font-weight: 500;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary:hover {
  background: #16a34a;
}

.btn-secondary {
  padding: 10px 20px;
  border: 1px solid #2d2d2d;
  background: transparent;
  color: #9ca3af;
  cursor: pointer;
  font-size: 14px;
  border-radius: 6px;
}

.player-section {
  margin-bottom: 24px;
  border: 1px solid #2d2d2d;
  border-radius: 6px;
  background: #1a1a1a;
}

.video-container {
  width: 100%;
  aspect-ratio: 16/9;
  background: #0a0a0a;
  position: relative;
}

.video-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  font-size: 14px;
  z-index: 10;
  background: #0a0a0a;
}

#youtube-player {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.player-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-top: 1px solid #2d2d2d;
}

.btn-control {
  padding: 8px 16px;
  border: 1px solid #22c55e;
  background: transparent;
  color: #22c55e;
  cursor: pointer;
  font-size: 13px;
  border-radius: 6px;
}

.btn-control:hover {
  background: #22c55e;
  color: #fff;
}

.btn-control.active {
  background: #22c55e;
  color: #fff;
}

.now-playing {
  color: #9ca3af;
  font-size: 13px;
}

.search-section {
  margin-bottom: 24px;
}

.search-section h3 {
  font-size: 14px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #9ca3af;
  margin-bottom: 12px;
}

.waiting-message {
  color: #9ca3af;
  padding: 16px;
  border: 1px solid #2d2d2d;
  text-align: center;
  border-radius: 6px;
  background: #1a1a1a;
}

.search-box {
  margin-bottom: 12px;
}

.search-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #2d2d2d;
  font-size: 14px;
  background: #1a1a1a;
  color: #fff;
  border-radius: 6px;
}

.search-input::placeholder {
  color: #666;
}

.search-results, .track-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.track-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #1a1a1a;
  cursor: pointer;
  border-radius: 6px;
}

.track-item:hover {
  background: #222;
}

.track-item.active {
  background: #222;
}

.track-art {
  width: 100px;
  height: 56px;
  object-fit: cover;
  background: #333;
}

.track-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.track-name {
  font-weight: 500;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.track-artist {
  color: #666;
  font-size: 13px;
}

.added-by {
  color: #666;
  font-size: 12px;
}

.btn-add {
  width: 32px;
  height: 32px;
  border: 1px solid #2d2d2d;
  background: #22c55e;
  color: #fff;
  cursor: pointer;
  font-size: 18px;
  flex-shrink: 0;
  border-radius: 6px;
}

.btn-add:hover {
  background: #16a34a;
}

.btn-remove {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: #666;
  cursor: pointer;
  font-size: 16px;
  flex-shrink: 0;
}

.btn-remove:hover {
  color: #fff;
}

.no-tracks, .no-archives {
  color: #9ca3af;
  padding: 24px;
  text-align: center;
  border: 1px solid #2d2d2d;
  border-radius: 6px;
  background: #1a1a1a;
}

.archive-section {
  margin-top: 24px;
}

.archive-section h3 {
  font-size: 14px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #9ca3af;
  margin-bottom: 12px;
}

.archive-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.archive-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 16px;
  background: #1a1a1a;
  text-decoration: none;
  color: #fff;
  border-radius: 6px;
}

.archive-item:hover {
  background: #222;
}

.archive-date {
  color: #fff;
  font-weight: 500;
}

.archive-name {
  flex: 1;
  margin-left: 16px;
  color: #9ca3af;
}

.archive-count {
  color: #9ca3af;
}
</style>
