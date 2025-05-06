package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func StartServer(port int) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", handleRootRequest)
	r.Get("/opengraph.png", handleOpenGraphPngRequest)
	r.Get("/opengraph.jpg", handleOpenGraphJpegRequest)

	handleStaticFiles(r, "/static", http.Dir("static"))

	log.Printf("Server started: http://127.0.0.1:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("Server closed")
		return nil
	}

	return err
}
