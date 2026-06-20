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

// Embed everything under dist/ except embeds/ and browser-games/ (they're large game files served from disk).
//
//go:embed dist/css/*
//go:embed dist/js/*
//go:embed dist/assets/*
//go:embed dist/gallery/*
//go:embed dist/awards/*
//go:embed dist/pages/*
//go:embed dist/manifest.json
var distFS embed.FS

// loadPages loads HTML pages, preferring embedded dist/ over disk.
func (s *htmlStore) loadPages() error {
	if err := s.loadFromFS(distFS, "dist/pages"); err == nil && s.count() > 0 {
		log.Println("Pages loaded from embedded dist/")
		return nil
	}
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
