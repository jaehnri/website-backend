# website-backend
Back-end for joaohenri.io.

## Running locally

### Spotify API
A valid Spotify integration is required to run this API. 
See https://developer.spotify.com/dashboard.

This specific integration requires access to:
- Web API
- Web Playback SDK

### Refresh token

Follow https://developer.spotify.com/documentation/web-api/tutorials/code-flow, i.e., manually claim an authorization code and use it to redeem the refresh token.

### Executing

Set these 3 environment variables:
- CLIENT_ID_ENV
- CLIENT_SECRET_ENV
- REFRESH_TOKEN_ENV

The first two are available at https://developer.spotify.com/dashboard. The third was picked up in the last step.

Then, simply:
```bash
make run
```
