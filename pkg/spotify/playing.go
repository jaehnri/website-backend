package spotify

// Expected response for https://api.spotify.com/v1/me/player/currently-playing.
// See https://developer.spotify.com/documentation/web-api/reference/get-the-users-currently-playing-track.
type CurrentPlayingResponse struct {
	IsPlaying  bool `json:"is_playing"`
	ProgressMs int  `json:"progress_ms"`
	Item       Item `json:"item"`
}

// Item represents the currently playing track.
type Item struct {
	DurationMs int      `json:"duration_ms"`
	SongName   string   `json:"name"`
	Artists    []Artist `json:"artists"`
}

// Artist of a given Item.
type Artist struct {
	Name string `json:"name"`
}

// Expected response for https://api.spotify.com/v1/me/player/recently-played.
// See https://developer.spotify.com/documentation/web-api/reference/get-recently-played.
type LastPlayedResponse struct {
	Items []PlayedItem `json:"items"`
}

// PlayedItem represents a recently played track.
type PlayedItem struct {
	Track Track `json:"track"`
}

// Track represents a song.
type Track struct {
	Name       string   `json:"name"`
	Artists    []Artist `json:"artists"`
	DurationMs int      `json:"duration_ms"`
}
