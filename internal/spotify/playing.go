package spotify

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	spotifyapi "github.com/jaehnri/website-backend/pkg/spotify"
	_ "github.com/joho/godotenv/autoload"
)

const (
	CurrentlyPlayingEndpoint = "https://api.spotify.com/v1/me/player/currently-playing"
)

type SpotifyClient struct {
	authProvider *AuthProvider
	httpClient   http.Client
}

type CurrentSong struct {
	IsPlaying bool `json:"is_playing"`

	ProgressMs int `json:"progress_ms"`
	DurationMs int `json:"song_duration_ms"`

	Song   string `json:"song"`
	Artist string `json:"artist"`
}

func NewSpotifyClient() *SpotifyClient {
	return &SpotifyClient{
		authProvider: NewAuthProvider(),
		httpClient:   http.Client{},
	}
}

// HandleNowPlaying receives an HTTP request and returns the current playing song
// in CurrentSong format.
func (s *SpotifyClient) HandleNowPlaying(w http.ResponseWriter, r *http.Request) {
	playingSong, err := s.getCurrentPlayingSong()
	if err != nil {
		http.Error(w, "failed to fetch current playing song", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(playingSong)
	if err != nil {
		http.Error(w, "failed to encode json response", http.StatusInternalServerError)
		return
	}
}

func (s *SpotifyClient) getCurrentPlayingSong() (*CurrentSong, error) {
	req, err := s.buildCurrentPlayingSongRequest()
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	currentPlayingResponse, err := parseCurrentPlayingSongResponse(resp)
	if err != nil {
		return nil, err
	}

	return convertAPIResponseToBackendResponse(currentPlayingResponse), nil
}

func (s *SpotifyClient) buildCurrentPlayingSongRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", CurrentlyPlayingEndpoint, nil)
	if err != nil {
		log.Println("failed to create current playing song request:", err)
		return nil, err
	}

	accessToken, err := s.authProvider.GetAccessToken()
	if err != nil {
		log.Println("failed to fetch access token:", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	return req, err
}

func parseCurrentPlayingSongResponse(resp *http.Response) (*spotifyapi.CurrentPlayingResponse, error) {
	// TODO: When not listening anything, the current playing endpoint returns 204.
	// In this case, I should probably "/me/player/recently-played". This would
	// require the "user-read-recently-played" scope though.
	// Ignoring this use case for now.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read current playing song response body:", err)
		return nil, err
	}

	var currentPlayingResponse spotifyapi.CurrentPlayingResponse
	err = json.Unmarshal(bodyBytes, &currentPlayingResponse)
	if err != nil {
		log.Println("failed to unmarshal current playing song JSON:", err)
		return nil, err
	}

	return &currentPlayingResponse, nil
}

func convertAPIResponseToBackendResponse(apiResponse *spotifyapi.CurrentPlayingResponse) *CurrentSong {
	return &CurrentSong{
		IsPlaying: apiResponse.IsPlaying,

		ProgressMs: apiResponse.ProgressMs,

		DurationMs: apiResponse.Item.DurationMs,
		Song:       apiResponse.Item.SongName,

		// Too many artists make it ugly. The first one is enough.
		// Plus, feats often include other artists in the name anyway.
		Artist: apiResponse.Item.Artists[0].Name,
	}
}
