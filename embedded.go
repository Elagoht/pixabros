package main

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//go:embed dist/*
var distFS embed.FS

// loadPagesFromDist loads HTML pages, preferring embedded dist/ over disk.
func (s *htmlStore) loadPages() error {
	// Try embedded first
	if err := s.loadFromFS(distFS, "dist/pages"); err == nil && s.count() > 0 {
		log.Println("Pages loaded from embedded dist/")
		return nil
	}

	// Fallback: disk
	if err := s.load("dist/pages"); err != nil {
		return err
	}
	log.Println("Pages loaded from disk dist/")
	return nil
}

// loadFromFS loads HTML pages from an fs.FS subtree.
func (s *htmlStore) loadFromFS(fsys fs.FS, dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newPages := make(map[string]htmlPage)

	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".html") {
			return nil
		}

		body, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		// Derive URL path
		rel, _ := filepath.Rel(dir, path)
		urlPath := "/" + strings.TrimSuffix(rel, ".html")
		if urlPath == "/index" {
			urlPath = "/"
		}

		hash := fmt.Sprintf(`"%x"`, sha256.Sum256(body))
		newPages[urlPath] = htmlPage{body: body, etag: hash}
		return nil
	})
	if err != nil {
		return err
	}

	if len(newPages) == 0 {
		return os.ErrNotExist
	}

	s.pages = newPages
	log.Printf("Loaded %d HTML pages from embedded FS", len(newPages))
	return nil
}
