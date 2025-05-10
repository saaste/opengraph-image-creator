package server

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/saaste/opengraph-image-creator/internal/pkg/cache"
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

	// Calculate ETag value
	etagData := []byte(fmt.Sprintf("%s / %s / %s", title, site, date))
	eTag := fmt.Sprintf("%x", md5.Sum(etagData))

	w.Header().Set("Cache-Control", fmt.Sprintf("max-age: %d", int64(appConfig.MaxCache.Seconds())))
	w.Header().Set("ETag", eTag)

	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("ERROR: failed to execute the template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleOpenGraphPngRequest(w http.ResponseWriter, r *http.Request) {
	handleOpenGraphImageRequest(w, r, cache.ImageTypePng)
}

func handleOpenGraphJpegRequest(w http.ResponseWriter, r *http.Request) {
	handleOpenGraphImageRequest(w, r, cache.ImageTypeJpeg)
}

func handleOpenGraphImageRequest(w http.ResponseWriter, r *http.Request, imageType cache.ImageType) {
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

	secret := r.URL.Query().Get("secret")

	// Calculate ETag value
	etagData := []byte(fmt.Sprintf("%s / %s / %s", title, site, date))
	eTag := fmt.Sprintf("%x", md5.Sum(etagData))

	w.Header().Set("Cache-Control", fmt.Sprintf("max-age: %d", int64(appConfig.MaxCache.Seconds())))
	w.Header().Set("ETag", eTag)

	// Try to get image from the cache
	cachedImage, err := cache.TryGetImageFromCache(appConfig, eTag, imageType)
	if err != nil {
		log.Printf("Failed to get image from the cache: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var outputFileName string
	switch imageType {
	case cache.ImageTypePng:
		outputFileName = "opengraph.png"
		w.Header().Set("Content-Type", "image/png")
	case cache.ImageTypeJpeg:
		outputFileName = "opengraph.jpg"
		w.Header().Set("Content-Type", "image/jpeg")
	default:
		log.Printf("Invalid image type: %d\n", imageType)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if cachedImage != nil {
		http.ServeContent(w, r, outputFileName, cachedImage.ModTime, bytes.NewReader(cachedImage.Data))
		return
	}

	imageBytes, err := image.TakeScreenshot(fmt.Sprintf("http://localhost:8080/?title=%s&site=%s&date=%s&secret=%s", url.QueryEscape(title), url.QueryEscape(site), url.QueryEscape(date), url.QueryEscape(secret)))

	if err != nil {
		log.Printf("Failed to take screenshot: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	cachedImage, err = cache.SaveImageToCache(appConfig, eTag, imageBytes, imageType)
	if err != nil {
		log.Printf("Failed to save image to cache: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, outputFileName, time.Now(), bytes.NewReader(cachedImage.Data))
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
