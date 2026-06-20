/* ============================================================
   PixaBros — Gamified Interactions
   ============================================================ */

(function () {
  'use strict';

  // ---- DOM Ready ----
  function ready(fn) {
    if (document.readyState !== 'loading') { fn(); return; }
    document.addEventListener('DOMContentLoaded', fn);
  }

  ready(function () {
    initMobileNav();
    initActiveNav();
    initParallax();
    initParticles();
    initCursorTrail();
    initGameCards();
    initGameModal();
    initTypewriter();
    initContactForm();
    initKonamiCode();
    initCRTToggle();
    initAchievementTracking();
  });

  // ---- Mobile Navigation ----
  function initMobileNav() {
    var toggle = document.getElementById('nav-toggle');
    var links = document.querySelector('.nav-links');
    if (!toggle || !links) return;
    toggle.addEventListener('click', function () {
      links.classList.toggle('open');
    });
  }

  // ---- Active Nav Highlighting ----
  function initActiveNav() {
    var path = window.location.pathname;
    var links = document.querySelectorAll('.nav-links a');
    links.forEach(function (link) {
      var href = link.getAttribute('href');
      if (href === path || (href !== '/' && path.indexOf(href) === 0)) {
        link.classList.add('active');
      }
    });
  }

  // ---- Parallax Scrolling ----
  function initParallax() {
    var ticking = false;
    window.addEventListener('scroll', function () {
      if (!ticking) {
        requestAnimationFrame(function () {
          document.documentElement.style.setProperty(
            '--scroll-y', String(window.scrollY)
          );
          ticking = false;
        });
        ticking = true;
      }
    }, { passive: true });
  }

  // ---- Floating Particles ----
  function initParticles() {
    var container = document.getElementById('particles');
    if (!container) return;

    var count = 30;
    var fragment = document.createDocumentFragment();
    for (var i = 0; i < count; i++) {
      var particle = document.createElement('div');
      particle.className = 'particle';
      particle.style.left = Math.random() * 100 + '%';
      particle.style.animationDuration = (8 + Math.random() * 12) + 's';
      particle.style.animationDelay = Math.random() * 10 + 's';
      particle.style.width = (4 + Math.random() * 8) + 'px';
      particle.style.height = particle.style.width;
      fragment.appendChild(particle);
    }
    container.appendChild(fragment);
  }

  // ---- Cursor Trail ----
  function initCursorTrail() {
    // Don't enable on touch devices
    if ('ontouchstart' in window) return;

    var trail = [];
    var maxTrail = 12;
    var lastMove = Date.now();
    var active = false;

    // Pre-create dots
    for (var i = 0; i < maxTrail; i++) {
      var dot = document.createElement('div');
      dot.className = 'cursor-trail-dot';
      dot.style.display = 'none';
      document.body.appendChild(dot);
      trail.push({ el: dot, x: 0, y: 0 });
    }

    var trailIdx = 0;

    document.addEventListener('mousemove', function (e) {
      if (!active) {
        active = true;
        updateTrail();
      }
      lastMove = Date.now();
      var t = trail[trailIdx % maxTrail];
      t.x = e.clientX;
      t.y = e.clientY;
      t.el.style.display = 'block';
      t.el.style.left = e.clientX + 'px';
      t.el.style.top = e.clientY + 'px';
      t.el.style.animation = 'none';
      // Force reflow
      void t.el.offsetWidth;
      t.el.style.animation = 'trail-fade 0.6s ease-out forwards';
      trailIdx++;
    });

    function updateTrail() {
      if (Date.now() - lastMove > 500) {
        active = false;
        for (var i = 0; i < maxTrail; i++) {
          trail[i].el.style.display = 'none';
        }
        return;
      }
      requestAnimationFrame(updateTrail);
    }
  }

  // ---- Game Cards ----
  function initGameCards() {
    // Cards are functional via CSS hover. Data attributes store slug for modal.
  }

  // ---- Game Detail Modal ----
  function initGameModal() {
    var modal = document.getElementById('game-modal');
    if (!modal) return;

    var closeBtn = modal.querySelector('.modal-close');
    var backdrop = modal.querySelector('.modal-backdrop');
    var gamesData = [];

    // Try to extract game data from the cards (rendered server-side)
    // We'll use a minimal inline data approach: store game info in data attributes
    var cards = document.querySelectorAll('.game-card');
    cards.forEach(function (card) {
      card.addEventListener('click', function () {
        var title = card.querySelector('.game-title').textContent;
        var genre = card.querySelector('.game-genre').textContent;
        var screenshot = card.querySelector('.game-thumb').src;
        var desc = '';
        openModal(title, genre, screenshot, desc);
      });
    });

    closeBtn.addEventListener('click', closeModal);
    backdrop.addEventListener('click', closeModal);

    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') closeModal();
    });

    function openModal(title, genre, screenshot, description) {
      document.getElementById('modal-title').textContent = title;
      document.getElementById('modal-genre').textContent = genre;
      document.getElementById('modal-screenshot').src = screenshot;
      document.getElementById('modal-screenshot').alt = title;
      document.getElementById('modal-description').textContent = description;

      // Look up links from the games data
      var linksDiv = document.getElementById('modal-links');
      linksDiv.innerHTML = '';
      var game = findGameData(title);
      if (game) {
        document.getElementById('modal-description').textContent = game.description || description;
        if (game.links && game.links.itchio) {
          var a = document.createElement('a');
          a.href = game.links.itchio;
          a.className = 'pixel-btn-small';
          a.textContent = 'Play on itch.io';
          a.target = '_blank';
          a.rel = 'noopener';
          linksDiv.appendChild(a);
        }
        if (game.links && game.links.steam) {
          var b = document.createElement('a');
          b.href = game.links.steam;
          b.className = 'pixel-btn-small';
          b.textContent = 'Wishlist on Steam';
          b.target = '_blank';
          b.rel = 'noopener';
          linksDiv.appendChild(b);
        }
      }

      modal.classList.add('open');
      modal.setAttribute('aria-hidden', 'false');
      document.body.style.overflow = 'hidden';

      // Track for achievements
      trackGameView(title);
    }

    function closeModal() {
      modal.classList.remove('open');
      modal.setAttribute('aria-hidden', 'true');
      document.body.style.overflow = '';
    }

    function findGameData(title) {
      // Try to find the game in the page's embedded data
      var scripts = document.querySelectorAll('script[type="application/json"].game-data');
      for (var i = 0; i < scripts.length; i++) {
        try {
          var g = JSON.parse(scripts[i].textContent);
          if (g.title === title) return g;
        } catch (e) { /* skip */ }
      }
      return null;
    }
  }

  // ---- Typewriter Effect ----
  function initTypewriter() {
    var el = document.getElementById('typewriter');
    if (!el) return;

    var text = el.getAttribute('data-text') || el.textContent;
    el.textContent = '';
    el.classList.add('typing');

    var i = 0;
    var speed = 60;
    function type() {
      if (i < text.length) {
        el.textContent += text.charAt(i);
        i++;
        setTimeout(type, speed + Math.random() * 30);
      } else {
        el.classList.remove('typing');
        // Keep cursor blinking via class for a moment
        setTimeout(function () {
          el.classList.add('typing');
        }, 100);
      }
    }
    setTimeout(type, 400);
  }

  // ---- Contact Form (AJAX) ----
  function initContactForm() {
    var form = document.getElementById('contact-form');
    if (!form) return;
    var submitBtn = document.getElementById('contact-submit');
    var errorDiv = document.getElementById('contact-error');
    var successDiv = document.getElementById('contact-success');

    form.addEventListener('submit', function (e) {
      e.preventDefault();
      var params = 'name=' + encodeURIComponent(document.getElementById('name').value)
        + '&email=' + encodeURIComponent(document.getElementById('email').value)
        + '&subject=' + encodeURIComponent((document.getElementById('subject') || {}).value || '')
        + '&message=' + encodeURIComponent(document.getElementById('message').value);

      if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Sending...';
      }
      if (errorDiv) errorDiv.style.display = 'none';

      fetch('/contact', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: params
      })
      .then(function (r) { return r.json(); })
      .then(function (data) {
        if (data.ok) {
          if (form) form.style.display = 'none';
          if (successDiv) successDiv.style.display = 'block';
        } else {
          if (errorDiv) {
            errorDiv.textContent = data.error || 'Something went wrong.';
            errorDiv.style.display = 'block';
          }
          if (submitBtn) {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Send Message';
          }
        }
      })
      .catch(function () {
        if (errorDiv) {
          errorDiv.textContent = 'Network error. Please try again.';
          errorDiv.style.display = 'block';
        }
        if (submitBtn) {
          submitBtn.disabled = false;
          submitBtn.textContent = 'Send Message';
        }
      });
    });
  }

  // ---- Konami Code ----
  function initKonamiCode() {
    var konami = [38, 38, 40, 40, 37, 39, 37, 39, 66, 65];
    var pos = 0;

    document.addEventListener('keydown', function (e) {
      if (e.keyCode === konami[pos]) {
        pos++;
        if (pos === konami.length) {
          activateKonami();
          pos = 0;
        }
      } else {
        pos = (e.keyCode === 38) ? 1 : 0;
      }
    });
  }

  function activateKonami() {
    var overlay = document.getElementById('konami-overlay');
    if (!overlay) return;

    // Pixel explosion effect
    spawnPixelExplosion();

    overlay.classList.add('open');
    overlay.setAttribute('aria-hidden', 'false');

    var closeBtn = document.getElementById('konami-close');
    closeBtn.addEventListener('click', function () {
      overlay.classList.remove('open');
      overlay.setAttribute('aria-hidden', 'true');
    });

    // Auto-close on backdrop click
    overlay.addEventListener('click', function (e) {
      if (e.target === overlay) {
        overlay.classList.remove('open');
        overlay.setAttribute('aria-hidden', 'true');
      }
    });

    // Track achievement
    showToast('🏆 Achievement Unlocked: Konami Code Discoverer!');
  }

  function spawnPixelExplosion() {
    var count = 80;
    var colors = ['#e94560', '#00ff88', '#ffdd00', '#ff6b81', '#0f3460'];
    for (var i = 0; i < count; i++) {
      var pixel = document.createElement('div');
      pixel.style.cssText =
        'position:fixed;z-index:6000;pointer-events:none;' +
        'width:' + (4 + Math.random() * 12) + 'px;' +
        'height:' + (4 + Math.random() * 12) + 'px;' +
        'background:' + colors[Math.floor(Math.random() * colors.length)] + ';' +
        'left:50%;top:50%;' +
        'image-rendering:pixelated;' +
        'transition:all ' + (0.5 + Math.random() * 1) + 's cubic-bezier(0,1,0.3,1);';
      document.body.appendChild(pixel);

      requestAnimationFrame(function () {
        pixel.style.transform =
          'translate(' + ((Math.random() - 0.5) * 600) + 'px,' +
          ((Math.random() - 0.5) * 600) + 'px) rotate(' + (Math.random() * 360) + 'deg)';
        pixel.style.opacity = '0';
      });

      setTimeout(function () {
        pixel.remove();
      }, 2000);
    }
  }

  // ---- CRT Toggle ----
  function initCRTToggle() {
    var toggle = document.getElementById('crt-toggle');
    var overlay = document.getElementById('crt-overlay');
    if (!toggle || !overlay) return;

    function setState(on) {
      if (on) {
        overlay.classList.add('on');
        toggle.classList.add('active');
        toggle.textContent = 'TV CRT ON';
      } else {
        overlay.classList.remove('on');
        toggle.classList.remove('active');
        toggle.textContent = 'TV CRT OFF';
      }
    }

    var saved = localStorage.getItem('crt-mode') === 'on';
    setState(saved);

    toggle.addEventListener('click', function () {
      var next = !overlay.classList.contains('on');
      localStorage.setItem('crt-mode', next ? 'on' : 'off');
      setState(next);
    });
  }

  // ---- Achievement Tracking ----
  function initAchievementTracking() {
    var viewed = JSON.parse(sessionStorage.getItem('pixabros-viewed-games') || '[]');
    var devlogRead = parseInt(sessionStorage.getItem('pixabros-devlog-read') || '0', 10);
    var pagesVisited = parseInt(sessionStorage.getItem('pixabros-pages-visited') || '0', 10);

    // Track page visit
    pagesVisited++;
    sessionStorage.setItem('pixabros-pages-visited', String(pagesVisited));

    // Page visit achievements
    if (pagesVisited === 3) {
      showToast('🗺️ Achievement: Explorer — You visited 3 pages!');
    }
    if (pagesVisited === 5) {
      showToast('🧭 Achievement: Cartographer — You visited all pages!');
    }
  }

  function trackGameView(title) {
    var viewed = JSON.parse(sessionStorage.getItem('pixabros-viewed-games') || '[]');
    if (viewed.indexOf(title) === -1) {
      viewed.push(title);
      sessionStorage.setItem('pixabros-viewed-games', JSON.stringify(viewed));

      if (viewed.length === 3) {
        showToast('👀 Achievement: Window Shopper — You viewed 3 games!');
      }
      if (viewed.length === 7) {
        showToast('🎮 Achievement: Game Curator — You viewed 7 games!');
      }
      if (viewed.length === 14) {
        showToast('👑 Achievement: Completionist — You viewed every game!');
      }
    }
  }

  // Check devlog read achievement (called from devlog pages)
  var currentPath = window.location.pathname;
  if (currentPath.indexOf('/devlog/') === 0 && currentPath !== '/devlog') {
    var read = parseInt(sessionStorage.getItem('pixabros-devlog-read') || '0', 10) + 1;
    sessionStorage.setItem('pixabros-devlog-read', String(read));
    if (read === 1) {
      showToast('📖 Achievement: Reader — You read your first devlog post!');
    }
  }

  // ---- Toast System ----
  function showToast(message) {
    var container = document.getElementById('toast-container');
    if (!container) return;

    var toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    container.appendChild(toast);

    // Auto-remove after animation
    setTimeout(function () {
      if (toast.parentNode) toast.parentNode.removeChild(toast);
    }, 4000);
  }
})();
