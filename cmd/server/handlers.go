package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/saaste/opengraph-image-creator/internal/pkg/config"
	"github.com/saaste/opengraph-image-creator/internal/pkg/image"
)

type TemplateData struct {
	Title string
	Site  string
	Date  string
}

func handleRootRequest(w http.ResponseWriter, r *http.Request) {
	appConfig, err := config.Load()
	if err != nil {
		log.Printf("ERROR: failed to load app config: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if appConfig.Secret != "" && r.URL.Query().Get("secret") != appConfig.Secret {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles("template.html")
	if err != nil {
		log.Printf("ERROR: failed to parse template files: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Title Not Set"
	}
	for _, char := range appConfig.LineBreakChars {
		title = strings.Replace(title, char, fmt.Sprintf("%s<br/>", strings.TrimSpace(char)), 1)
	}

	site := r.URL.Query().Get("site")
	if site == "" {
		site = appConfig.Site
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format(appConfig.DateFormat)
	}

	data := &TemplateData{
		Title: title,
		Site:  site,
		Date:  date,
	}

	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("ERROR: failed to execute the template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleOpenGraphRequest(w http.ResponseWriter, r *http.Request) {
	appConfig, err := config.Load()
	if err != nil {
		log.Printf("ERROR: failed to load app config: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if appConfig.Secret != "" && r.URL.Query().Get("secret") != appConfig.Secret {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Title Not Set"
	}

	site := r.URL.Query().Get("site")
	if site == "" {
		site = appConfig.Site
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format(appConfig.DateFormat)
	}

	imageBytes, err := image.TakeScreenshot(fmt.Sprintf("http://localhost:8080/?title=%s&site=%s&date=%s", url.QueryEscape(title), url.QueryEscape(site), url.QueryEscape(date)))
	if err != nil {
		log.Printf("Failed to take screenshot: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "opengraph.png", time.Now(), bytes.NewReader(imageBytes))
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
