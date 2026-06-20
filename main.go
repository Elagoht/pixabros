package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ---- HTML Page Store ----

type htmlPage struct {
	body []byte
	etag string
}

// htmlStore loads pages from dist/pages/ into memory.
type htmlStore struct {
	mu    sync.RWMutex
	pages map[string]htmlPage // path → page
}

func newHTMLStore() *htmlStore { return &htmlStore{pages: make(map[string]htmlPage)} }

func (s *htmlStore) load(dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newPages := make(map[string]htmlPage)

	walk := func(base string) error {
		return filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(info.Name(), ".html") {
				return nil
			}

			body, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			hash := fmt.Sprintf(`"%x"`, sha256.Sum256(body))

			// Derive URL path from the file path
			rel, _ := filepath.Rel(dir, path)
			urlPath := "/" + strings.TrimSuffix(rel, ".html")

			// Special case: index.html → /
			if urlPath == "/index" {
				urlPath = "/"
			}
			// devlog/<slug> → /devlog/<slug>
			if strings.HasPrefix(urlPath, "/devlog/") {
				// already correct
			}

			newPages[urlPath] = htmlPage{body: body, etag: hash}
			log.Printf("  loaded: %s → %d bytes", urlPath, len(body))
			return nil
		})
	}

	if err := walk(dir); err != nil {
		return err
	}

	s.pages = newPages
	log.Printf("Loaded %d HTML pages into memory", len(newPages))
	return nil
}

func (s *htmlStore) get(path string) (htmlPage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pages[path]
	return p, ok
}

func (s *htmlStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pages)
}

var pages *htmlStore

// ---- Rate Limiter ----

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]int
}

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{buckets: make(map[string]int)}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) allow(key string, max int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.buckets[key] >= max {
		return false
	}
	rl.buckets[key]++
	return true
}

func (rl *rateLimiter) cleanup() {
	for range time.Tick(1 * time.Minute) {
		rl.mu.Lock()
		rl.buckets = make(map[string]int)
		rl.mu.Unlock()
	}
}

// ---- HTML Handler ----

func serveHTML(w http.ResponseWriter, r *http.Request, status int, page htmlPage) {
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == page.etag || strings.Contains(match, page.etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("ETag", page.etag)
	w.WriteHeader(status)
	_, _ = w.Write(page.body)
}

func handlePage(name string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, ok := pages.get(name)
		if !ok {
			serve404(w, r)
			return
		}
		serveHTML(w, r, status, page)
	}
}

func handleDevlogPage(w http.ResponseWriter, r *http.Request) {
	// /devlog → devlog index
	if r.URL.Path == "/devlog" {
		page, ok := pages.get("/devlog")
		if !ok {
			serve404(w, r)
			return
		}
		serveHTML(w, r, http.StatusOK, page)
		return
	}
	// /devlog/<slug> → individual post
	page, ok := pages.get(r.URL.Path)
	if !ok {
		serve404(w, r)
		return
	}
	serveHTML(w, r, http.StatusOK, page)
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handleContactPost(w, r)
		return
	}
	page, ok := pages.get("/contact")
	if !ok {
		serve404(w, r)
		return
	}
	serveHTML(w, r, http.StatusOK, page)
}

func handleContactPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Failed to parse form."})
		return
	}

	msg := map[string]string{
		"name":    strings.TrimSpace(r.FormValue("name")),
		"email":   strings.TrimSpace(r.FormValue("email")),
		"subject": strings.TrimSpace(r.FormValue("subject")),
		"message": strings.TrimSpace(r.FormValue("message")),
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	if msg["name"] == "" || msg["email"] == "" || msg["message"] == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Name, email, and message are required."})
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "Internal error."})
		return
	}

	if err := os.MkdirAll("messages", 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "Server error."})
		return
	}

	f, err := os.OpenFile("messages/messages.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "Failed to store message."})
		return
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "Failed to store message."})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func serve404(w http.ResponseWriter, r *http.Request) {
	page, ok := pages.get("/404")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 - Not Found"))
		return
	}
	serveHTML(w, r, http.StatusNotFound, page)
}

// ---- Static Assets ----


// ---- Router ----

func setupRouter(rl *rateLimiter) http.Handler {
	mux := http.NewServeMux()

	// HTML pages from memory
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			p, ok := pages.get("/")
			if !ok {
				serve404(w, r)
				return
			}
			serveHTML(w, r, http.StatusOK, p)
			return
		}
		serve404(w, r)
	})
	mux.HandleFunc("/devlog", handleDevlogPage)
	mux.HandleFunc("/devlog/", handleDevlogPage)
	mux.HandleFunc("/play", handlePage("/play", http.StatusOK))
	mux.HandleFunc("/awards", handlePage("/awards", http.StatusOK))

	// Contact with rate limiting on POST
	mux.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = fwd
			}
			if !rl.allow(ip, 5) {
				writeJSON(w, http.StatusTooManyRequests, map[string]any{"ok": false, "error": "Too many messages. Please wait a minute."})
				return
			}
		}
		handleContact(w, r)
	})

	// Hashed immutable assets + browser game embeds
	mux.HandleFunc("/dist/", func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Clean(filepath.Join("dist", strings.TrimPrefix(r.URL.Path, "/dist/")))
		if strings.Contains(r.URL.Path, "..") || !strings.HasPrefix(p, "dist"+string(filepath.Separator)) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		data, err := os.ReadFile(p)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		// Detect content type from extension
		ct := "application/octet-stream"
		switch filepath.Ext(p) {
		case ".html":
			ct = "text/html; charset=utf-8"
		case ".js":
			ct = "application/javascript"
		case ".css":
			ct = "text/css; charset=utf-8"
		case ".wasm":
			ct = "application/wasm"
		case ".png":
			ct = "image/png"
		case ".jpg", ".jpeg", ".JPG", ".JPEG":
			ct = "image/jpeg"
		case ".webp":
			ct = "image/webp"
		case ".pck":
			ct = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ct)
		w.Write(data)
	})

	// Public assets (unhashed fallback)
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	return mux
}

// ---- Main ----

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("Starting PixaBros (static mode)...")

	// Load HTML pages into memory
	pages = newHTMLStore()
	if err := pages.load("dist/pages"); err != nil {
		log.Fatalf("Failed to load HTML pages: %v", err)
	}

	rl := newRateLimiter()
	router := setupRouter(rl)

	// SIGHUP reloads HTML pages + content from disk
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)
	go func() {
		for range sigCh {
			log.Println("SIGHUP: reloading HTML pages from disk...")
			if err := pages.load("dist/pages"); err != nil {
				log.Printf("Error reloading pages: %v", err)
			}
			log.Printf("Reload complete. %d pages in memory.", pages.count())
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on :%s (%d pages in memory)", port, pages.count())
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
