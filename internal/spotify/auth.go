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
	"sync"
	"time"

	spotifyapi "github.com/jaehnri/website-backend/pkg/spotify"
)

const (
	// TokenExpiryBuffer is used to refresh tokens a bit before they expire,
	// avoiding issues if there's a slight delay in the refresh process or
	// the next API call.
	TokenExpiryBuffer = 30 * time.Second

	ClientIDEnv     = "CLIENT_ID_ENV"
	ClientSecretEnv = "CLIENT_SECRET_ENV"
	RefreshTokenEnv = "REFRESH_TOKEN_ENV"

	TokenEndpoint = "https://accounts.spotify.com/api/token"
)

var (
	once         sync.Once
	authProvider *AuthProvider
)

// AuthProvider is a thread-safe module that manages access tokens to the Spotify API.
type AuthProvider struct {
	httpClient http.Client

	clientID     string
	clientSecret string

	// refreshToken is an "infinite-lived" token used to generate new accessTokens.
	refreshToken string

	// accessToken is used in all Spotify API requests.
	accessToken string

	// lock protects accessToken.
	lock sync.RWMutex

	// expiresAt holds the accessToken expiration time.
	expiresAt time.Time
}

func NewAuthProvider() *AuthProvider {
	if authProvider == nil {
		once.Do(func() {
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

			authProvider = &AuthProvider{
				clientID:     clientID,
				clientSecret: clientSecret,
				refreshToken: refreshToken,
				lock:         sync.RWMutex{},
				httpClient:   http.Client{},
			}
		})
	}

	return authProvider
}

// GetAccessToken grants valid access tokens to the Spotify API.
// This method is thread-safe.
func (a *AuthProvider) GetAccessToken() (string, error) {
	a.lock.RLock()
	if a.isTokenFresh() {
		defer a.lock.RUnlock()
		return a.accessToken, nil
	}
	a.lock.RUnlock()

	// Token is refreshed if:
	// 1. Access token was never set.
	// 2. Token has expired.
	err := a.refreshAccessToken()
	if err != nil {
		log.Println("failed to refresh access token: ", err)

		// Note that in case of failure, we return the old token.
		// Token services commonly extend token expirations when there are outages.
		return a.accessToken, err
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.accessToken, nil
}

// refreshAccessToken uses the "infinite-lived" refresh token to get a new access token.
// This method is thread-safe
func (a *AuthProvider) refreshAccessToken() error {
	a.lock.Lock()
	defer a.lock.Unlock()

	// Repeat this check as the token might have already been refreshed by the
	// time RW lock is acquired.
	if a.isTokenFresh() {
		return nil
	}

	req, err := a.buildRefreshTokenRequest()
	if err != nil {
		return err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		log.Println("failed to do refresh token request")
		return err
	}
	defer resp.Body.Close()

	refreshTokenResponse, err := parseRefreshTokenResponse(resp)
	if err != nil {
		return err
	}

	a.accessToken = refreshTokenResponse.AccessToken
	a.expiresAt = time.Now().Add(time.Duration(refreshTokenResponse.ExpiresIn) * time.Second)
	return nil
}

// Checks if the access token is still fresh
func (a *AuthProvider) isTokenFresh() bool {
	return time.Now().Add(TokenExpiryBuffer).Before(a.expiresAt)
}

func (a *AuthProvider) buildRefreshTokenRequest() (*http.Request, error) {
	header := []byte(a.clientID + ":" + a.clientSecret)
	encodedAuthorizationHeader := base64.RawStdEncoding.EncodeToString(header)

	// Create form data
	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", a.refreshToken)
	encodedFormData := formData.Encode()

	// Create HTTP request
	req, err := http.NewRequest("POST", TokenEndpoint, strings.NewReader(encodedFormData))
	if err != nil {
		log.Println("failed to create refresh token post request:", err)
		return nil, err
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+encodedAuthorizationHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
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
