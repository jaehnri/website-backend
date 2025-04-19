package spotify

// Expected response for https://accounts.spotify.com/api/token in the refresh token flow.
// See https://developer.spotify.com/documentation/web-api/tutorials/refreshing-tokens.
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
