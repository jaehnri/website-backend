package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jaehnri/website-backend/internal/ideas"
	"github.com/jaehnri/website-backend/internal/spotify"
)

type Server struct {
	httpAddress   string
	spotifyClient *spotify.SpotifyClient
	ideasClient   *ideas.IdeasClient
}

func NewServer(httpAddress string) *Server {
	return &Server{
		httpAddress:   httpAddress,
		spotifyClient: spotify.NewSpotifyClient(),
		ideasClient:   ideas.NewIdeasClient(),
	}
}

func (s *Server) startHTTPServer() {
	log.Println("starting HTTP server")

	http.HandleFunc("/now-playing", s.spotifyClient.HandleNowPlaying)
	http.HandleFunc("/ideas", s.ideasClient.HandleGetIdeas)

	log.Fatal(http.ListenAndServe(s.httpAddress, nil))
}

func (s *Server) Run() {
	go s.startHTTPServer()

	sigint := make(chan os.Signal, 1)
	// interrupt signal sent from terminal
	signal.Notify(sigint, os.Interrupt)
	// sigterm signal sent from kubernetes (if I ever deploy this on kubernetes lol)
	signal.Notify(sigint, syscall.SIGTERM)

	<-sigint
	log.Fatal("SIGINT received, shutting down server")
}
