package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/go-chi/chi/v5"
)

type Data struct {
	Title string
	Site  string
	Date  string
}

func main() {
	r := chi.NewRouter()
	r.Get("/", handleRequest)
	r.Get("/opengraph.jpg", handleOpenGraphRequest)
	handleStaticFiles(r, "/static", http.Dir("ui/static"))

	err := http.ListenAndServe(fmt.Sprintf(":%d", 8080), r)
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("INFO: Server closed")
	} else if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("ui/template.html")
	if err != nil {
		log.Printf("ERROR: failed to parse template files: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Title Not Set"
	}

	site := r.URL.Query().Get("site")
	if site == "" {
		site = "saaste.net"
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("02.01.2006")
	}

	data := &Data{
		Title: title,
		Site:  site,
		Date:  date,
	}

	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("ERROR: failed to execute the template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleStaticFiles(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		log.Fatal("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func handleOpenGraphRequest(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Title Not Set"
	}

	site := r.URL.Query().Get("site")
	if site == "" {
		site = "saaste.net"
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("02.01.2006")
	}

	imageBytes, err := takeScreenshot(fmt.Sprintf("http://localhost:8080/?title=%s&site=%s&date=%s", url.QueryEscape(title), url.QueryEscape(site), url.QueryEscape(date)))
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	w.Header().Set("Content-type", "image/jpeg")
	w.Write(imageBytes)
}

func takeScreenshot(url string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Creating timeout for 15 seconds
	ctx, cancel = context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	var buf []byte

	err := chromedp.Run(ctx,
		emulation.SetUserAgentOverride("OpenGraphImageCreator"),
		chromedp.Navigate(url),
		chromedp.Screenshot(".image", &buf, chromedp.NodeVisible),
	)

	if err != nil {
		return buf, err
	}

	return buf, nil
}
