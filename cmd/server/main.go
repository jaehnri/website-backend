package main

import (
	"github.com/jaehnri/website-backend/internal/server"
)

func main() {
	s := server.NewServer(":8080")
	s.Run()
}
