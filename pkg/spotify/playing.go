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
