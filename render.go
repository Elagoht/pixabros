//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func filterPlayable(games []GameVM) []GameVM {
	var out []GameVM
	for _, g := range games {
		if g.Playable {
			out = append(out, g)
		}
	}
	return out
}

// ---- Render all static HTML pages during build ----

type assetMap map[string]string

func loadManifest(path string) (assetMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := assetMap{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (m assetMap) path(key string) string {
	if hashed, ok := m[key]; ok {
		return "/dist/" + hashed
	}
	return "/" + key
}

func renderAllPages() error {
	// Load manifest for asset paths
	am, err := loadManifest("dist/manifest.json")
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	outputDir := "dist/pages"
	if err := os.MkdirAll(filepath.Join(outputDir, "devlog"), 0755); err != nil {
		return err
	}

	// Load content
	if err := LoadContent(); err != nil {
		return fmt.Errorf("loading content: %w", err)
	}
	c := GetContent()

	// Compile templates
	ts, err := LoadTemplates()
	if err != nil {
		return fmt.Errorf("compiling templates: %w", err)
	}

	// Pre-compute common fields
	distCSS := am.path("static/css/site.css")
	distJS := am.path("static/js/site.js")
	logoPath := am.path("public/assets/pixabros.png")
	socialMailto := "mailto:" + c.Site.Socials.Email

	makePD := func(title, desc, curPath string) PageData {
		return PageData{
			Title:        title,
			Description:  desc,
			Studio:       c.Site.Studio,
			Tagline:      c.Site.Tagline,
			CurrentPath:  curPath,
			Socials:      c.Site.Socials,
			SocialMailto: socialMailto,
			NavLinks:     navLinks(),
			DistCSS:      distCSS,
			DistJS:       distJS,
			LogoPath:     logoPath,
		}
	}

	write := func(name string, pd PageData) error {
		html, err := ts.Render(name, pd)
		if err != nil {
			return err
		}
		fullPath := filepath.Join(outputDir, name+".html")
		if err := os.WriteFile(fullPath, html, 0644); err != nil {
			return err
		}
		log.Printf("  page: %s.html (%d bytes)", name, len(html))
		return nil
	}

	// ---- Index ----
	pd := makePD("PixaBros — Brothers make games", c.Site.Description, "/")
	pd.Bros = toBrotherVM(c.Bros, am)
	pd.Games = toGameVM(c.Games, am)
	if err := write("index", pd); err != nil {
		return fmt.Errorf("index: %w", err)
	}

	// ---- Play ----
	pdPlay := makePD("Play — PixaBros", "Play PixaBros games right in your browser", "/play")
	allGames := toGameVM(c.Games, am)
	pdPlay.Games = allGames
	pdPlay.PlayableGames = filterPlayable(allGames)
	if err := write("play", pdPlay); err != nil {
		return fmt.Errorf("play: %w", err)
	}

	// ---- Devlog index ----
	posts, err := parseDevlogPosts("devlog")
	if err != nil {
		log.Printf("Warning: parsing devlog: %v", err)
	}
	pdDevlog := makePD("Devlog — PixaBros", "Development log for PixaBros games", "/devlog")
	pdDevlog.DevlogPosts = toDevlogPostVM(posts)
	if err := write("devlog", pdDevlog); err != nil {
		return fmt.Errorf("devlog: %w", err)
	}

	// ---- Devlog posts ----
	for _, post := range posts {
		pdPost := makePD(post.Title+" — PixaBros Devlog", post.Excerpt, "/devlog/"+post.Slug)
		pdPost.Post = toDevlogPostVMSingle(&post)
		pdPost.DevlogPosts = toDevlogPostVM(posts)
		html, err := ts.Render("devlog-post", pdPost)
		if err != nil {
			return fmt.Errorf("devlog/%s: %w", post.Slug, err)
		}
		p := filepath.Join(outputDir, "devlog", post.Slug+".html")
		if err := os.WriteFile(p, html, 0644); err != nil {
			return err
		}
		log.Printf("  page: devlog/%s.html (%d bytes)", post.Slug, len(html))
	}

	// ---- Working On ----
	pdWip := makePD("What We're Working On — PixaBros", "Current projects in development at PixaBros", "/working-on")
	pdWip.WorkingOn = toWorkingOnVM(c.WorkingOn)
	if err := write("working-on", pdWip); err != nil {
		return fmt.Errorf("working-on: %w", err)
	}

	// ---- Contact ----
	pdContact := makePD("Contact — PixaBros", "Get in touch with PixaBros", "/contact")
	if err := write("contact", pdContact); err != nil {
		return fmt.Errorf("contact: %w", err)
	}

	// ---- Press Kit ----
	pdPress := makePD("Press Kit — PixaBros", "Press resources for PixaBros", "/press-kit")
	pdPress.PressKit = toPressKitVM(&c.PressKit, am)
	pdPress.Games = toGameVM(c.Games, am)
	if err := write("press-kit", pdPress); err != nil {
		return fmt.Errorf("press-kit: %w", err)
	}

	// ---- 404 ----
	pd404 := makePD("404 — PixaBros", "Page not found", "")
	if err := write("404", pd404); err != nil {
		return fmt.Errorf("404: %w", err)
	}

	log.Printf("Rendered all pages to dist/pages/")
	return nil
}
