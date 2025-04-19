package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jaehnri/website-backend/internal/spotify"
)

type Server struct {
	httpAddress string

	// TODO: The server should be composed of a spotify module.
}

func NewServer(httpAddress string) *Server {
	return &Server{
		httpAddress: httpAddress,
	}
}

func (s *Server) startHTTPServer() {
	log.Println("starting sample HTTP server")

	http.HandleFunc("/spotify", spotify.HandleCurrentSongRequest)

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
