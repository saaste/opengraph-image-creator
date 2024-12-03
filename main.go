package main

import (
	"github.com/saaste/opengraph-image-creator/cmd/server"
)

type Data struct {
	Title string
	Site  string
	Date  string
}

func main() {
	server.StartServer(8080)
}
