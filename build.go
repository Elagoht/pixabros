//go:build ignore

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	// Phase 2: Render static HTML
	log.Println("Phase 2: Rendering HTML pages...")
	if err := renderAllPages(); err != nil {
		log.Fatalf("Render pages: %v", err)
	}

	log.Println("=== Build complete ===")
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
