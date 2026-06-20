//go:build ignore

package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	log.SetFlags(log.Ltime)
	log.Println("=== PixaBros Build ===")

	// Phase 1: Hash assets
	log.Println("Phase 1: Hashing assets...")
	am := make(assetMap)

	if err := os.MkdirAll("dist", 0755); err != nil {
		log.Fatalf("Failed to create dist/: %v", err)
	}
	if err := hashDir("static", "dist", am); err != nil {
		log.Fatalf("Hashing static/: %v", err)
	}
	if err := hashDir("public", "dist", am); err != nil {
		log.Fatalf("Hashing public/: %v", err)
	}

	data, err := json.MarshalIndent(am, "", "  ")
	if err != nil {
		log.Fatalf("Marshal manifest: %v", err)
	}
	if err := os.WriteFile("dist/manifest.json", data, 0644); err != nil {
		log.Fatalf("Write manifest: %v", err)
	}
	log.Printf("  Hashed %d assets → dist/manifest.json", len(am))

	// Phase 2: Extract browser games
	log.Println("Phase 2: Extracting browser games...")
	if err := extractBrowserGames(); err != nil {
		log.Fatalf("Extract browser games: %v", err)
	}

	// Phase 3: Render static HTML
	log.Println("Phase 2: Rendering HTML pages...")
	if err := renderAllPages(); err != nil {
		log.Fatalf("Render pages: %v", err)
	}

	log.Println("=== Build complete ===")
}

func extractBrowserGames() error {
	srcDir := "public/browser-games"
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("  No browser games found.")
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}
		slug := strings.TrimSuffix(entry.Name(), ".zip")
		dest := filepath.Join("dist", "embeds", slug)
		if err := os.MkdirAll(dest, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dest, err)
		}

		zipPath := filepath.Join(srcDir, entry.Name())
		r, err := zip.OpenReader(zipPath)
		if err != nil {
			return fmt.Errorf("opening %s: %w", zipPath, err)
		}
		defer r.Close()

		for _, f := range r.File {
			// Skip __MACOSX junk and directories
			if strings.HasPrefix(f.Name, "__MACOSX") || strings.HasSuffix(f.Name, "/") {
				continue
			}
			// Flatten: strip any leading directories
			name := filepath.Base(f.Name)
			if name == "" || strings.HasPrefix(name, ".") {
				continue
			}

			outPath := filepath.Join(dest, name)
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("reading %s from zip: %w", f.Name, err)
			}
			out, err := os.Create(outPath)
			if err != nil {
				rc.Close()
				return fmt.Errorf("creating %s: %w", outPath, err)
			}
			_, err = io.Copy(out, rc)
			rc.Close()
			out.Close()
			if err != nil {
				return fmt.Errorf("writing %s: %w", outPath, err)
			}
		}
		log.Printf("  extracted: %s → dist/embeds/%s/ (%d files)", entry.Name(), slug, len(r.File))
	}
	return nil
}

func hashDir(srcRoot, dstRoot string, am assetMap) error {
	return filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])[:12]

		ext := filepath.Ext(info.Name())
		base := strings.TrimSuffix(info.Name(), ext)
		hashedName := fmt.Sprintf("%s.%s%s", base, hashStr, ext)

		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		dstRel := filepath.Join(filepath.Dir(rel), hashedName)
		dstPath := filepath.Join(dstRoot, dstRel)

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", dstPath, err)
		}

		key := filepath.ToSlash(filepath.Join(srcRoot, rel))
		val := filepath.ToSlash(dstRel)
		am[key] = val

		log.Printf("  asset: %s → %s", key, val)
		return nil
	})
}
