package spotify

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	spotifyapi "github.com/jaehnri/website-backend/pkg/spotify"
	_ "github.com/joho/godotenv/autoload"
)

const (
	ClientIDEnv     = "CLIENT_ID_ENV"
	ClientSecretEnv = "CLIENT_SECRET_ENV"
	RefreshTokenEnv = "REFRESH_TOKEN_ENV"

	TokenEndpoint            = "https://accounts.spotify.com/api/token"
	CurrentlyPlayingEndpoint = "https://api.spotify.com/v1/me/player/currently-playing"
)

type SpotifyClient struct {
	clientID     string
	clientSecret string

	accessToken  string
	refreshToken string

	httpClient http.Client
}

type CurrentSong struct {
	IsPlaying bool `json:"is_playing"`

	ProgressMs int `json:"progress_ms"`
	DurationMs int `json:"song_duration_ms"`

	Song   string `json:"song"`
	Artist string `json:"artist"`
}

func NewSpotifyClient() *SpotifyClient {
	clientID, exists := os.LookupEnv(ClientIDEnv)
	if !exists {
		log.Panic("couldn't retrieve client ID")
	}

	clientSecret, exists := os.LookupEnv(ClientSecretEnv)
	if !exists {
		log.Panic("couldn't retrieve client secret")
	}

	refreshToken, exists := os.LookupEnv(RefreshTokenEnv)
	if !exists {
		log.Panic("couldn't retrieve refresh token")
	}

	return &SpotifyClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		httpClient:   http.Client{},
	}
}

// HandleNowPlaying receives an HTTP request and returns the current playing song
// in CurrentSong format.
func (s *SpotifyClient) HandleNowPlaying(w http.ResponseWriter, r *http.Request) {
	// TODO: Instead of fetching a new access token everytime, I should check if the
	// current access token is still valid.
	err := s.refreshAccessToken()
	if err != nil {
		http.Error(w, "failed to refresh access token", http.StatusInternalServerError)
		return
	}

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

func (s *SpotifyClient) refreshAccessToken() error {
	req, err := s.buildRefreshTokenRequest()
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	refreshTokenResponse, err := parseRefreshTokenResponse(resp)
	if err != nil {
		return err
	}

	s.accessToken = refreshTokenResponse.AccessToken
	return nil
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

func (s *SpotifyClient) buildRefreshTokenRequest() (*http.Request, error) {
	header := []byte(s.clientID + ":" + s.clientSecret)
	encodedAuthorizationHeader := base64.RawStdEncoding.EncodeToString(header)

	// Create form data
	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", s.refreshToken)
	encodedFormData := formData.Encode()

	// Create HTTP request
	req, err := http.NewRequest("POST", TokenEndpoint, strings.NewReader(encodedFormData))
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+encodedAuthorizationHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func (s *SpotifyClient) buildCurrentPlayingSongRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", CurrentlyPlayingEndpoint, nil)
	if err != nil {
		log.Println("failed to create current playing song request:", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	return req, err
}

func parseRefreshTokenResponse(resp *http.Response) (*spotifyapi.RefreshTokenResponse, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read refresh token response body")
		return nil, err
	}

	var refreshTokenResponse spotifyapi.RefreshTokenResponse
	err = json.Unmarshal(bodyBytes, &refreshTokenResponse)
	if err != nil {
		log.Println("failed to unmarshal access token JSON:", err)
		return nil, err
	}

	return &refreshTokenResponse, nil
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

		// TODO: Perhaps do some joining logic here IDK
		Artist: apiResponse.Item.Artists[0].Name,
	}
}
