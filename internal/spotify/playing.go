package spotify

import (
	"log"
	"os"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
)

const (
	ClientIDEnv     = "CLIENT_ID_ENV"
	ClientSecretEnv = "CLIENT_SECRET_ENV"
)

type Spotify struct {
	clientID     string
	clientSecret string
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

	return &Spotify{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func LogSpotify(w http.ResponseWriter, r *http.Request) {
	log.Println("received an HTTP call!")
	
	s := NewSpotify()
	log.Println(s.clientID)

}