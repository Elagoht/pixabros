//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Joker/jade"
	"github.com/yuin/goldmark"
)

// TemplateStore holds compiled templates.
type TemplateStore struct {
	pages map[string]*template.Template
}

// PageData is the top-level data passed to every page template.
type PageData struct {
	Title        string
	Description  string
	Studio       string
	Tagline      string
	CurrentPath  string
	Socials      Socials
	NavLinks     []NavLink
	SocialMailto string

	// Page-specific data
	Bros         []BrotherVM
	Games        []GameVM
	Game         *GameVM
	DevlogPosts  []DevlogPostVM
	Post         *DevlogPostVM
	WorkingOn    []WorkingOnVM
	PressKit     *PressKitVM
	ErrorMessage string
	Success      bool

	// Asset paths
	DistCSS  string
	DistJS   string
	LogoPath string
}

// ---- View-model types (all fields pre-computed, no template functions needed) ----

type BrotherVM struct {
	Name     string
	Roles    []string
	Bio      string
	Avatar   string
	Initials string
}

type GameVM struct {
	Slug        string
	Title       string
	Genre       string
	Description string
	ImageURL    string
	Year        int
	Links       GameLinks
}

type DevlogPostVM struct {
	Slug          string
	Title         string
	Date          time.Time
	FormattedDate string // pre-computed "January 2, 2006"
	RawDate       string
	Content       template.HTML
	Excerpt       string
	URL           string
}

type WorkingOnVM struct {
	Title         string
	Status        string
	Progress      int
	Description   string
	Image         string
	Tags          []string
	ProgressStyle string
}

type PressKitVM struct {
	Studio             string
	Founded            int
	Location           string
	Website            string
	PressContact       string
	PressContactMailto string
	Socials            Socials
	Description        string
	History            string
	Team               []PressKitTeamMember
	LogoPath           string
}

type NavLink struct {
	Label string
	Path  string
}

func navLinks() []NavLink {
	return []NavLink{
		{Label: "Home", Path: "/"},
		{Label: "Devlog", Path: "/devlog"},
		{Label: "What We're Working On", Path: "/working-on"},
		{Label: "Press Kit", Path: "/press-kit"},
		{Label: "Contact", Path: "/contact"},
	}
}

// ---- Template Compilation ----

func LoadTemplates() (*TemplateStore, error) {
	store := &TemplateStore{pages: make(map[string]*template.Template)}

	files, err := os.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("reading templates: %w", err)
	}

	funcMap := template.FuncMap{
		"join":  strings.Join,
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".pug") {
			continue
		}
		path := filepath.Join("templates", f.Name())
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		name := strings.TrimSuffix(f.Name(), ".pug")

		goSrc, err := jade.Parse(name, src)
		if err != nil {
			return nil, fmt.Errorf("compiling %s: %w", path, err)
		}

		tmpl, err := template.New(name).Funcs(funcMap).Parse(goSrc)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		store.pages[name] = tmpl
		log.Printf("Compiled template: %s", name)
	}
	return store, nil
}

func (ts *TemplateStore) Render(name string, data any) ([]byte, error) {
	tmpl, ok := ts.pages[name]
	if !ok {
		return nil, fmt.Errorf("template %q not found", name)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// ---- Devlog Parsing ----

type DevlogPost struct {
	Slug    string
	Title   string
	Date    time.Time
	RawDate string
	Content template.HTML
	Excerpt string
}

var md = goldmark.New()

func renderMarkdown(text string) template.HTML {
	var buf bytes.Buffer
	if err := md.Convert([]byte(text), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(text))
	}
	return template.HTML(buf.String())
}

func parseDevlogPosts(dir string) ([]DevlogPost, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var posts []DevlogPost
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			log.Printf("Warning: cannot read devlog entry %s: %v", entry.Name(), err)
			continue
		}
		post, err := parseDevlogFile(entry.Name(), raw)
		if err != nil {
			log.Printf("Warning: cannot parse devlog entry %s: %v", entry.Name(), err)
			continue
		}
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts, nil
}

func parseDevlogFile(filename string, raw []byte) (DevlogPost, error) {
	text := string(raw)
	post := DevlogPost{}

	if strings.HasPrefix(text, "---") {
		rest := strings.TrimPrefix(text, "---")
		if idx := strings.Index(rest, "---"); idx >= 0 {
			fm := rest[:idx]
			body := rest[idx+3:]
			parseFrontmatter(fm, &post)
			post.Content = renderMarkdown(body)
			post.Excerpt = makeExcerpt(body, 200)
		}
	} else {
		post.Content = renderMarkdown(text)
		post.Excerpt = makeExcerpt(text, 200)
		post.Title = strings.TrimSuffix(filename, ".md")
	}

	if post.Slug == "" {
		name := strings.TrimSuffix(filename, ".md")
		if len(name) > 11 && name[4] == '-' && name[7] == '-' && name[10] == '-' {
			post.Slug = name[11:]
			post.RawDate = name[:10]
		} else {
			post.Slug = name
		}
	}

	if post.Date.IsZero() && post.RawDate != "" {
		if t, err := time.Parse("2006-01-02", post.RawDate); err == nil {
			post.Date = t
		}
	}
	return post, nil
}

func parseFrontmatter(fm string, post *DevlogPost) {
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "title":
			post.Title = val
		case "slug":
			post.Slug = val
		case "date":
			post.RawDate = val
			if t, err := time.Parse("2006-01-02", val); err == nil {
				post.Date = t
			}
		}
	}
}

func makeExcerpt(raw string, maxLen int) string {
	// Render markdown to HTML so we can strip tags and get clean text
	var buf bytes.Buffer
	if err := md.Convert([]byte(raw), &buf); err != nil {
		raw = strings.TrimSpace(raw)
		if len(raw) <= maxLen {
			return raw
		}
		excerpt := raw[:maxLen]
		if idx := strings.LastIndex(excerpt, " "); idx > 0 {
			excerpt = excerpt[:idx]
		}
		return excerpt + "..."
	}

	// Strip HTML tags to get plain text
	html := buf.String()
	plain := stripTags(html)

	// Unescape HTML entities (leave them as-is for the excerpt, but trim)
	plain = strings.TrimSpace(plain)
	if len(plain) <= maxLen {
		return plain
	}
	excerpt := plain[:maxLen]
	if idx := strings.LastIndex(excerpt, " "); idx > 0 {
		excerpt = excerpt[:idx]
	}
	return excerpt + "..."
}

// stripTags removes HTML tags and common entities from a string.
func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			// add a space in case adjacent text would merge
			b.WriteByte(' ')
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	// Collapse whitespace
	out := strings.Join(strings.Fields(b.String()), " ")
	// Unescape common HTML entities
	out = strings.ReplaceAll(out, "&amp;", "&")
	out = strings.ReplaceAll(out, "&lt;", "<")
	out = strings.ReplaceAll(out, "&gt;", ">")
	out = strings.ReplaceAll(out, "&quot;", "\"")
	out = strings.ReplaceAll(out, "&#39;", "'")
	out = strings.ReplaceAll(out, "&rsquo;", "'")
	out = strings.ReplaceAll(out, "&mdash;", "—")
	return out
}

// ---- View Model Conversions ----

type assetResolver interface {
	path(key string) string
}

func toBrotherVM(bros []Brother, _ assetResolver) []BrotherVM {
	result := make([]BrotherVM, len(bros))
	for i, b := range bros {
		parts := strings.Fields(b.Name)
		init := strings.ToUpper(b.Name[:min(2, len(b.Name))])
		if len(parts) >= 2 {
			init = strings.ToUpper(string(parts[0][0]) + string(parts[1][0]))
		}
		result[i] = BrotherVM{
			Name:     b.Name,
			Roles:    b.Roles,
			Bio:      b.Bio,
			Avatar:   b.Avatar,
			Initials: init,
		}
	}
	return result
}

func toGameVM(games []Game, am assetResolver) []GameVM {
	result := make([]GameVM, len(games))
	for i, g := range games {
		result[i] = GameVM{
			Slug:        g.Slug,
			Title:       g.Title,
			Genre:       g.Genre,
			Description: g.Description,
			ImageURL:    am.path("public/gallery/" + g.Screenshot),
			Year:        g.Year,
			Links:       g.Links,
		}
	}
	return result
}

func toDevlogPostVM(posts []DevlogPost) []DevlogPostVM {
	result := make([]DevlogPostVM, len(posts))
	for i, p := range posts {
		result[i] = DevlogPostVM{
			Slug:          p.Slug,
			Title:         p.Title,
			Date:          p.Date,
			FormattedDate: p.Date.Format("January 2, 2006"),
			RawDate:       p.RawDate,
			Content:       p.Content,
			Excerpt:       p.Excerpt,
			URL:           "/devlog/" + p.Slug,
		}
	}
	return result
}

func toDevlogPostVMSingle(post *DevlogPost) *DevlogPostVM {
	if post == nil {
		return nil
	}
	return &DevlogPostVM{
		Slug:          post.Slug,
		Title:         post.Title,
		Date:          post.Date,
		FormattedDate: post.Date.Format("January 2, 2006"),
		RawDate:       post.RawDate,
		Content:       post.Content,
		Excerpt:       post.Excerpt,
		URL:           "/devlog/" + post.Slug,
	}
}

func toWorkingOnVM(items []WorkingOnItem) []WorkingOnVM {
	result := make([]WorkingOnVM, len(items))
	for i, w := range items {
		result[i] = WorkingOnVM{
			Title:         w.Title,
			Status:        w.Status,
			Progress:      w.Progress,
			Description:   w.Description,
			Image:         w.Image,
			Tags:          w.Tags,
			ProgressStyle: fmt.Sprintf("width: %d%%", w.Progress),
		}
	}
	return result
}

func toPressKitVM(pk *PressKit, am assetResolver) *PressKitVM {
	if pk == nil {
		return nil
	}
	return &PressKitVM{
		Studio:             pk.Studio,
		Founded:            pk.Founded,
		Location:           pk.Location,
		Website:            pk.Website,
		PressContact:       pk.PressContact,
		PressContactMailto: "mailto:" + pk.PressContact,
		Socials:            pk.Socials,
		Description:        pk.Description,
		History:            pk.History,
		Team:               pk.Team,
		LogoPath:           am.path("public/assets/pixabros.png"),
	}
}
