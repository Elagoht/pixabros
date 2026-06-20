<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/Elagoht/pixabros/main/public/assets/pixabros.png">
    <img src="https://raw.githubusercontent.com/Elagoht/pixabros/main/public/assets/pixabros.png" width="200" alt="PixaBros">
  </picture>
</p>

<p align="center">
  <a href="https://pixabros.com"><img src="https://img.shields.io/badge/website-pixabros.com-ff1a6c?style=flat-square" alt="Website"></a>
  <a href="https://github.com/Elagoht/pixabros/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-GPLv3-00f0ff?style=flat-square" alt="License MIT"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go 1.26"></a>
  <a href="https://elagoht.itch.io"><img src="https://img.shields.io/badge/itch.io-elagoht-fa5c5c?style=flat-square&logo=itchdotio&logoColor=white" alt="itch.io"></a>
</p>

Two brothers making games together. This is our studio website — a retro arcade-themed static site built with Go and Pug templates. Everything from code and art to music and game design, done in-house.

## Quick start

```bash
make run
```

Opens at `http://localhost:8080` (set `PORT=3000` to change).

## Structure

```
├── main.go              # HTTP server, router, rate limiter
├── build.go             # Asset hashing, game zip extraction
├── content.go           # Content types & JSON loader
├── templates.go         # Template compilation (Pug → Go), view models
├── render.go            # Static HTML page renderer (build time)
├── embedded.go          # Embed.FS for dist/ assets
├── install.sh           # Systemd service installer for Linux
├── Makefile             # Build, run, bundle-linux targets
│
├── content/             # JSON content files
│   ├── site.json        # Studio name, tagline, socials, CTA
│   ├── bros.json        # Brother bios, roles, avatars
│   ├── games.json       # 14 games with descriptions, links
│   └── awards.json      # Game jam awards
│
├── templates/           # Pug (Jade) templates
│   ├── _partials/       # Head, nav, footer, overlays
│   ├── index.pug        # Home — hero, bros, games grid, modal
│   ├── play.pug         # Arcade TV + NES console + cartridge shelf
│   ├── devlog.pug       # Devlog list
│   ├── devlog-post.pug  # Single devlog article
│   ├── awards.pug       # Awards showcase
│   ├── contact.pug      # Contact form
│   └── 404.pug          # GAME OVER screen
│
├── devlog/              # Markdown devlog posts (frontmatter + body)
├── static/              # CSS and JS (hashed at build)
├── public/              # Static assets (hashed at build)
│   ├── assets/          # Logo, bro photos
│   ├── awards/          # Award certificate images
│   ├── gallery/         # Game screenshots
│   └── browser-games/   # Playable game zips (extracted to dist/embeds/)
│
└── dist/                # Build output (gitignored)
    ├── pages/           # Rendered static HTML
    ├── embeds/          # Extracted browser games (not embedded in binary)
    ├── css/, js/        # Hashed immutable files
    └── manifest.json    # Original → hashed path mapping
```

## Build pipeline

```
make assets          # Hash assets → compile Pug → render HTML → extract game zips
make build           # assets + compile Go binary
make run             # build + start server
make bundle-linux    # assets + static Linux binary + tar.gz
```

The binary embeds pages, CSS, JS, gallery, and awards. Large game embeds stay on disk under `dist/embeds/`.

## Deployment

```bash
make bundle-linux
scp pixabros-linux.tar.gz furkan@your-server:
ssh furkan@your-server
mkdir pixabros && tar -xzf pixabros-linux.tar.gz -C pixabros && cd pixabros
sudo PORT=80 ./install.sh
```

This creates a `pixabros` system user, installs to `/opt/pixabros`, writes a systemd service, and starts it.

```bash
systemctl status pixabros
journalctl -u pixabros -f
```

## Content editing

Edit JSON files in `content/`, write devlog posts as Markdown in `devlog/`, then rebuild:

```bash
make assets
```

On the production server, send SIGHUP to reload pages from disk without restarting:

```bash
sudo systemctl kill -s SIGHUP pixabros
```

## Tech

- **Go** 1.26 — HTTP server, embed, content hash
- **Pug** via [Joker/jade](https://github.com/Joker/jade) — templates
- **Goldmark** — Markdown rendering for devlog
- **Press Start 2P + IBM Plex Mono** — typography
- **Zero JavaScript dependencies** — vanilla JS for all interactions

## License

[GPLv3](LICENSE) — free software, forever.
