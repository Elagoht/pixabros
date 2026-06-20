//go:build ignore

package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// ---- Site Config ----

type Socials struct {
	Twitter string `json:"twitter"`
	Itchio  string `json:"itchio"`
	Discord string `json:"discord"`
	Github  string `json:"github"`
	Email   string `json:"email"`
}

type CTA struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type SiteConfig struct {
	Studio      string  `json:"studio"`
	Tagline     string  `json:"tagline"`
	Description string  `json:"description"`
	Founded     int     `json:"founded"`
	Location    string  `json:"location"`
	Socials     Socials `json:"socials"`
	CTA         CTA     `json:"cta"`
}

// ---- Brothers ----

type Brother struct {
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	Bio    string   `json:"bio"`
	Avatar string   `json:"avatar"`
}

// ---- Games ----

type GameLinks struct {
	Itchio string `json:"itchio"`
	Steam  string `json:"steam"`
}

type Game struct {
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Genre       string    `json:"genre"`
	Description string    `json:"description"`
	Screenshot  string    `json:"screenshot"`
	Links       GameLinks `json:"links"`
	Year        int       `json:"year"`
}

// ---- Press Kit ----

type PressKitTeamMember struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

type PressKitLogos struct {
	Default string `json:"default"`
}

type PressKit struct {
	Studio        string              `json:"studio"`
	Founded       int                 `json:"founded"`
	Location      string              `json:"location"`
	Website       string              `json:"website"`
	PressContact  string              `json:"pressContact"`
	Socials       Socials             `json:"socials"`
	Description   string              `json:"description"`
	History       string              `json:"history"`
	Team          []PressKitTeamMember `json:"team"`
	Logos         PressKitLogos       `json:"logos"`
	FeaturedGames []string            `json:"featuredGames"`
}

// ---- Working On ----

type WorkingOnItem struct {
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	Progress    int      `json:"progress"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	Tags        []string `json:"tags"`
}

// ---- Content Store ----

type Content struct {
	Site      SiteConfig
	Bros      []Brother
	Games     []Game
	PressKit  PressKit
	WorkingOn []WorkingOnItem
}

var content Content
var contentMu sync.RWMutex

func loadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func LoadContent() error {
	contentMu.Lock()
	defer contentMu.Unlock()

	if err := loadJSON("content/site.json", &content.Site); err != nil {
		return err
	}
	if err := loadJSON("content/bros.json", &content.Bros); err != nil {
		return err
	}
	if err := loadJSON("content/games.json", &content.Games); err != nil {
		return err
	}
	if err := loadJSON("content/press-kit.json", &content.PressKit); err != nil {
		return err
	}
	if err := loadJSON("content/working-on.json", &content.WorkingOn); err != nil {
		return err
	}

	log.Printf("Loaded: %d brothers, %d games, %d working-on items",
		len(content.Bros), len(content.Games), len(content.WorkingOn))
	return nil
}

func GetContent() Content {
	contentMu.RLock()
	defer contentMu.RUnlock()
	return content
}
