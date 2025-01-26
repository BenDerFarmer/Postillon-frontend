package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Post struct {
	ID      string
	Heading string
	Body    []string
	Images  map[uint8]Image
	Extra   map[uint8]string
}

type HomePost struct {
	ID      string
	Heading string
	Image   string
}

type Image struct {
	Space       uint8
	Source      string
	Description string
}

//go:embed public/*
var embeddedFiles embed.FS

func main() {

	http.HandleFunc("/", Handler)
	http.HandleFunc("/post/{id...}", PostHandler)
	http.HandleFunc("/img/", ProxyHandler)

	fmt.Println("Listening on :6060")
	http.ListenAndServe(":6060", nil)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	p := "public" + r.URL.Path

	query := ""

	if r.URL.Query().Has("q") {
		query = r.URL.Query().Get("q")
	}

	if p == "public/" {
		posts := fetchNewPosts(query)
		Home(posts).Render(context.Background(), w)
		return
	}

	file, err := embeddedFiles.Open(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set MIME type based on file extension
	ext := filepath.Ext(p)
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	http.ServeContent(w, r, filepath.Base(p), time.Now(), strings.NewReader(string(content)))
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimRight(r.PathValue("id"), "/")

	post := fetch("https://www.der-postillon.com/" + id + ".html")

	postComponent(post).Render(context.Background(), w)
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	targetURL := "https://blogger.googleusercontent.com" + r.URL.Path

	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Failed to copy image", http.StatusInternalServerError)
		return
	}
}
