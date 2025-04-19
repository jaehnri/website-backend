package main

import (
	"github.com/jaehnri/website-backend/internal/server"
)

const redirectURI = "https://www.joaohenri.io/"


func main() {
	s := server.NewServer(":8080")
	s.Run()
}
