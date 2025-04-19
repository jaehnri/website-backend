package server

import(
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	httpAddress string
}

func NewServer(httpAddress string) *Server {
	return &Server{
		httpAddress: httpAddress,
	}
}

func (s *Server) startHTTPServer() {
	log.Println("starting sample HTTP server")

	http.HandleFunc("/spotify", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received an HTTP call!")
	})
	log.Fatal(http.ListenAndServe(s.httpAddress, nil))
}

func (s *Server) Run() {
	go s.startHTTPServer()

	sigint := make(chan os.Signal, 1)
	// interrupt signal sent from terminal
	signal.Notify(sigint, os.Interrupt)
	// sigterm signal sent from kubernetes
	signal.Notify(sigint, syscall.SIGTERM)

	<-sigint
	log.Fatal("SIGINT received, shutting down server")
}