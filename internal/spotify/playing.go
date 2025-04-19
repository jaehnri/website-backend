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
	}
}

func (s *Spotify) refreshAccessToken() {
	header := []byte(s.clientID + ":" + s.clientSecret)
	encodedAuthorizationHeader := base64.RawStdEncoding.EncodeToString(header)

	// Create the form data
	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", s.refreshToken)
	encodedFormData := formData.Encode()

	// Create the HTTP request
	req, err := http.NewRequest("POST", TokenEndpoint, strings.NewReader(encodedFormData))
	if err != nil {
		log.Panic("Error creating request:", err)
		return
	}

	// Set the headers
	req.Header.Set("Authorization", "Basic "+encodedAuthorizationHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create an HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Print the response status
	log.Println("Status:", resp.Status)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic("Error reading response body:", err)
		return
	}

	var accessTokenResponse spotifyapi.RefreshTokenResponse
	err = json.Unmarshal(bodyBytes, &accessTokenResponse)
	if err != nil {
		log.Panic("Error unmarshaling JSON:", err)
		return
	}

	s.accessToken = accessTokenResponse.AccessToken
	log.Println("Access Token:", s.accessToken)
}

func (s *Spotify) getCurrentPlaying() {
	// Create the HTTP request
	req, err := http.NewRequest("GET", CurrentlyPlayingEndpoint, nil)
	if err != nil {
		log.Panic("Error creating request:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	// Create an HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Print the response status
	log.Println("Status:", resp.Status)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic("Error reading response body:", err)
		return
	}

	var currentPlayingResponse spotifyapi.CurrentPlayingResponse
	err = json.Unmarshal(bodyBytes, &currentPlayingResponse)
	if err != nil {
		log.Panic("Error unmarshaling JSON:", err)
		return
	}

	log.Println("Song playing:", currentPlayingResponse.Item.SongName)

}

func LogSpotify(w http.ResponseWriter, r *http.Request) {
	log.Println("received an HTTP call!")

	s := NewSpotify()
	s.refreshAccessToken()
	s.getCurrentPlaying()
}
