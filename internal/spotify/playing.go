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

type Spotify struct {
	clientID     string
	clientSecret string

	accessToken  string
	refreshToken string

	httpClient http.Client
}

func NewSpotify() *Spotify {
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

	return &Spotify{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		httpClient:   http.Client{},
	}
}

func HandleCurrentSongRequest(w http.ResponseWriter, r *http.Request) {
	s := NewSpotify()

	err := s.refreshAccessToken()
	if err != nil {
		return
	}

	err = s.getCurrentPlayingSong()
	if err != nil {
		return
	}
}

func (s *Spotify) refreshAccessToken() error {
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

func (s *Spotify) getCurrentPlayingSong() error {
	req, err := s.buildCurrentPlayingSongRequest()
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	currentPlayingResponse, err := parseCurrentPlayingSongResponse(resp)
	if err != nil {
		return err
	}

	log.Println("Song playing:", currentPlayingResponse.Item.SongName)
	return nil
}

func (s *Spotify) buildRefreshTokenRequest() (*http.Request, error) {
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

func (s *Spotify) buildCurrentPlayingSongRequest() (*http.Request, error) {
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
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic("failed to read current playing song response body:", err)
		return nil, err
	}

	var currentPlayingResponse spotifyapi.CurrentPlayingResponse
	err = json.Unmarshal(bodyBytes, &currentPlayingResponse)
	if err != nil {
		log.Panic("failed to unmarshal current playing song JSON:", err)
		return nil, err
	}

	return &currentPlayingResponse, nil
}
