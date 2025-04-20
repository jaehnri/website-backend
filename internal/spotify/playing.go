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
	LastPlayedSongEndpoint   = "https://api.spotify.com/v1/me/player/recently-played"
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

	// TODO: Think about this later!
	w.Header().Set("Access-Control-Allow-Origin", "*")
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

	// Current playing endpoint returns 204 when no track is on.
	// In this case, we can get the last played song instead.
	if resp.StatusCode == 204 {
		return s.getLastPlayedSong()
	}

	currentPlayingResponse, err := parseCurrentPlayingSongResponse(resp)
	if err != nil {
		return nil, err
	}

	return convertCurrentPlayingResponse(currentPlayingResponse), nil
}

func (s *SpotifyClient) getLastPlayedSong() (*CurrentSong, error) {
	req, err := s.buildLastPlayedSongRequest()
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	lastPlayedResponse, err := parseLastPlayedSongResponse(resp)
	if err != nil {
		return nil, err
	}

	return convertLastPlayedToResponse(lastPlayedResponse), nil
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
	return req, nil
}

func (s *SpotifyClient) buildLastPlayedSongRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", LastPlayedSongEndpoint, nil)
	if err != nil {
		log.Println("failed to create current playing song request:", err)
		return nil, err
	}

	accessToken, err := s.authProvider.GetAccessToken()
	if err != nil {
		log.Println("failed to fetch access token:", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("limit", "1")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+accessToken)
	return req, nil
}

func parseCurrentPlayingSongResponse(resp *http.Response) (*spotifyapi.CurrentPlayingResponse, error) {
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

func parseLastPlayedSongResponse(resp *http.Response) (*spotifyapi.LastPlayedResponse, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read last played song response body:", err)
		return nil, err
	}

	var lastPlayedResponse spotifyapi.LastPlayedResponse
	err = json.Unmarshal(bodyBytes, &lastPlayedResponse)
	if err != nil {
		log.Println("failed to unmarshal current playing song JSON:", err)
		return nil, err
	}

	return &lastPlayedResponse, nil
}

func convertCurrentPlayingResponse(apiResponse *spotifyapi.CurrentPlayingResponse) *CurrentSong {
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

func convertLastPlayedToResponse(apiResponse *spotifyapi.LastPlayedResponse) *CurrentSong {
	return &CurrentSong{
		IsPlaying: false,

		ProgressMs: apiResponse.Items[0].Track.DurationMs,

		DurationMs: apiResponse.Items[0].Track.DurationMs,
		Song:       apiResponse.Items[0].Track.Name,

		// Too many artists make it ugly. The first one is enough.
		// Plus, feats often include other artists in the name anyway.
		Artist: apiResponse.Items[0].Track.Artists[0].Name,
	}
}
