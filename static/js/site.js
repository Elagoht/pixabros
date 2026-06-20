/* ============================================================
   PixaBros — Arcade After Hours
   ============================================================ */

(function () {
  'use strict';

  function ready(fn) {
    if (document.readyState !== 'loading') { fn(); return; }
    document.addEventListener('DOMContentLoaded', fn);
  }

  ready(function () {
    initMobileNav();
    initActiveNav();
    initTypewriter();
    initGameModal();
    initContactForm();
    initConsole();
    initKonamiCode();
    initCRTToggle();
  });

  /* ---- Mobile Navigation ---- */
  function initMobileNav() {
    var toggle = document.getElementById('nav-toggle');
    var links = document.querySelector('.nav-links');
    if (!toggle || !links) return;
    toggle.addEventListener('click', function () {
      links.classList.toggle('open');
    });
  }

  /* ---- Active Nav Highlighting ---- */
  function initActiveNav() {
    var path = window.location.pathname;
    var links = document.querySelectorAll('.nav-links a');
    for (var i = 0; i < links.length; i++) {
      var href = links[i].getAttribute('href');
      if (href === path || (href !== '/' && path.indexOf(href) === 0)) {
        links[i].classList.add('active');
      }
    }
  }

  /* ---- Typewriter Effect ---- */
  function initTypewriter() {
    var el = document.getElementById('typewriter');
    if (!el) return;

    var text = el.getAttribute('data-text') || el.textContent || '';
    el.textContent = '';
    el.classList.add('typing');

    var i = 0;
    var speed = 55;

    function type() {
      if (i < text.length) {
        el.textContent += text.charAt(i);
        i++;
        setTimeout(type, speed + Math.random() * 30);
      } else {
        el.classList.remove('typing');
        setTimeout(function () {
          el.classList.add('typing');
        }, 150);
      }
    }

    setTimeout(type, 400);
  }

  /* ---- Game Detail Modal ---- */
  function initGameModal() {
    var modal = document.getElementById('game-modal');
    if (!modal) return;

    var closeBtn = modal.querySelector('.game-modal-close');
    var backdrop = modal.querySelector('.game-modal-backdrop');

    var cards = document.querySelectorAll('.game-card');
    for (var i = 0; i < cards.length; i++) {
      cards[i].addEventListener('click', function () {
        var title = this.getAttribute('data-title') || '';
        var genre = this.getAttribute('data-genre') || '';
        var desc = this.getAttribute('data-desc') || '';
        var image = this.getAttribute('data-image') || '';
        var year = this.getAttribute('data-year') || '';
        var itch = this.getAttribute('data-itch') || '';

        setText(modal, '.game-modal-title', title);
        setText(modal, '.game-modal-genre', genre);
        setText(modal, '.game-modal-year', year);
        setText(modal, '.game-modal-desc', desc);

        var img = modal.querySelector('.game-modal-image');
        if (img) { img.src = image; img.alt = title; }

        var links = modal.querySelector('.game-modal-links');
        if (links) {
          links.innerHTML = '';
          if (itch) {
            var a = document.createElement('a');
            a.href = itch;
            a.className = 'arcade-btn-small';
            a.textContent = 'Play on itch.io';
            a.target = '_blank';
            a.rel = 'noopener';
            links.appendChild(a);
          }
        }

        modal.classList.add('open');
        modal.setAttribute('aria-hidden', 'false');
        document.body.style.overflow = 'hidden';
      });
    }

    function closeModal() {
      modal.classList.remove('open');
      modal.setAttribute('aria-hidden', 'true');
      document.body.style.overflow = '';
    }

    if (closeBtn) closeBtn.addEventListener('click', closeModal);
    if (backdrop) backdrop.addEventListener('click', closeModal);

    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && modal.classList.contains('open')) closeModal();
    });
  }

  function setText(parent, selector, text) {
    var el = parent.querySelector(selector);
    if (el) el.textContent = text;
  }

  /* ---- Contact Form (AJAX) ---- */
  function initContactForm() {
    var form = document.getElementById('contact-form');
    if (!form) return;
    var submitBtn = document.getElementById('contact-submit');
    var errorDiv = document.getElementById('contact-error');
    var successDiv = document.getElementById('contact-success');

    form.addEventListener('submit', function (e) {
      e.preventDefault();
      var nameEl = document.getElementById('name');
      var emailEl = document.getElementById('email');
      var subjectEl = document.getElementById('subject');
      var messageEl = document.getElementById('message');

      var params = 'name=' + encodeURIComponent(nameEl ? nameEl.value : '')
        + '&email=' + encodeURIComponent(emailEl ? emailEl.value : '')
        + '&subject=' + encodeURIComponent((subjectEl || {}).value || '')
        + '&message=' + encodeURIComponent(messageEl ? messageEl.value : '');

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

  /* ---- NES Console + Cartridge System ---- */
  function initConsole() {
    var shelf = document.getElementById('cartridge-shelf');
    if (!shelf) return;

    var iframe = document.getElementById('tv-iframe');
    var placeholder = document.getElementById('tv-placeholder');
    var fullscreenBtn = document.getElementById('tv-fullscreen');
    var led = document.getElementById('nes-led');
    var tvLed = document.getElementById('tv-on-led');
    var slotLabel = document.getElementById('slot-label');
    var infoPanel = document.getElementById('game-info-panel');
    var cartridges = shelf.querySelectorAll('.nes-cartridge');
    var selected = null;

    iframe.src = 'about:blank';

    function selectCartridge(cart) {
      if (selected === cart) return;
      var embedUrl = cart.getAttribute('data-embed');

      if (selected) {
        selected.classList.remove('cart-selected');
      }

      selected = cart;
      cart.classList.add('cart-selected');

      if (placeholder) placeholder.classList.add('hidden');
      iframe.src = embedUrl;
      iframe.classList.add('active');

      if (led) led.classList.add('on');
      if (tvLed) tvLed.classList.add('on');

      // Update slot label
      if (slotLabel) {
        slotLabel.textContent = cart.getAttribute('data-title') || 'INSERT CARTRIDGE';
      }

      // Update sidebar
      updateInfoPanel(cart);
    }

    function updateInfoPanel(cart) {
      if (!infoPanel) return;

      var title = cart.getAttribute('data-title') || '';
      var genre = cart.getAttribute('data-genre') || '';
      var desc = cart.getAttribute('data-desc') || '';
      var image = cart.getAttribute('data-image') || '';
      var year = cart.getAttribute('data-year') || '';
      var itch = cart.getAttribute('data-itch') || '';

      infoPanel.classList.add('active');

      var titleEl = infoPanel.querySelector('.game-info-title');
      var genreEl = infoPanel.querySelector('.game-info-genre');
      var yearEl = infoPanel.querySelector('.game-info-year');
      var descEl = infoPanel.querySelector('.game-info-desc');
      var thumbEl = infoPanel.querySelector('.game-info-thumb');
      var linksEl = infoPanel.querySelector('.game-info-links');

      if (titleEl) titleEl.textContent = title;
      if (genreEl) genreEl.textContent = genre;
      if (yearEl) yearEl.textContent = year;
      if (descEl) descEl.textContent = desc;
      if (thumbEl) {
        thumbEl.src = image;
        thumbEl.alt = title;
      }

      // Build links
      if (linksEl) {
        linksEl.innerHTML = '';
        if (itch) {
          var itchBtn = document.createElement('a');
          itchBtn.href = itch;
          itchBtn.className = 'arcade-btn-small';
          itchBtn.textContent = 'itch.io';
          itchBtn.target = '_blank';
          itchBtn.rel = 'noopener';
          linksEl.appendChild(itchBtn);
        }
      }
    }

    function resetConsole() {
      if (selected) {
        selected.classList.remove('cart-selected');
        selected = null;
      }

      iframe.src = 'about:blank';
      iframe.classList.remove('active');

      if (placeholder) placeholder.classList.remove('hidden');
      if (led) led.classList.remove('on');
      if (tvLed) tvLed.classList.remove('on');
      if (slotLabel) slotLabel.textContent = 'INSERT CARTRIDGE';

      // Reset sidebar
      if (infoPanel) infoPanel.classList.remove('active');
    }

    for (var i = 0; i < cartridges.length; i++) {
      cartridges[i].addEventListener('click', function () {
        selectCartridge(this);
      });
    }

    var resetBtn = document.querySelector('.arcade-reset-btn');
    if (resetBtn) {
      resetBtn.addEventListener('click', function () {
        resetConsole();
      });
    }

    if (fullscreenBtn) {
      fullscreenBtn.addEventListener('click', function () {
        var tvScreen = document.querySelector('.arcade-tv-screen');
        if (!tvScreen) return;
        if (document.fullscreenElement) {
          document.exitFullscreen();
        } else {
          var req = tvScreen.requestFullscreen || tvScreen.webkitRequestFullscreen;
          if (req) req.call(tvScreen);
        }
      });
    }
  }

  /* ---- Konami Code ---- */
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

    spawnPixelBurst();

    overlay.classList.add('open');
    overlay.setAttribute('aria-hidden', 'false');

    var closeBtn = document.getElementById('konami-close');
    if (closeBtn) {
      closeBtn.addEventListener('click', function () {
        overlay.classList.remove('open');
        overlay.setAttribute('aria-hidden', 'true');
      });
    }

    overlay.addEventListener('click', function (e) {
      if (e.target === overlay) {
        overlay.classList.remove('open');
        overlay.setAttribute('aria-hidden', 'true');
      }
    });
  }

  function spawnPixelBurst() {
    var count = 60;
    var colors = ['#ff1a6c', '#00f0ff', '#ffb700', '#f5f0e8'];
    for (var i = 0; i < count; i++) {
      var pixel = document.createElement('div');
      pixel.style.cssText =
        'position:fixed;z-index:6000;pointer-events:none;' +
        'width:' + (3 + Math.random() * 10) + 'px;' +
        'height:' + (3 + Math.random() * 10) + 'px;' +
        'background:' + colors[Math.floor(Math.random() * colors.length)] + ';' +
        'left:50%;top:50%;' +
        'image-rendering:pixelated;' +
        'transition:all ' + (0.4 + Math.random() * 0.8) + 's cubic-bezier(0,1,0.3,1);';
      document.body.appendChild(pixel);

      requestAnimationFrame(function () {
        pixel.style.transform =
          'translate(' + ((Math.random() - 0.5) * 500) + 'px,' +
          ((Math.random() - 0.5) * 500) + 'px) rotate(' + (Math.random() * 360) + 'deg)';
        pixel.style.opacity = '0';
      });

      setTimeout(function () {
        if (pixel.parentNode) pixel.parentNode.removeChild(pixel);
      }, 2000);
    }
  }

  /* ---- CRT Toggle ---- */
  function initCRTToggle() {
    var toggle = document.getElementById('crt-toggle');
    var overlay = document.getElementById('crt-overlay');
    if (!toggle || !overlay) return;

    function setState(on) {
      if (on) {
        overlay.classList.add('on');
        toggle.classList.add('active');
        toggle.textContent = 'CRT';
      } else {
        overlay.classList.remove('on');
        toggle.classList.remove('active');
        toggle.textContent = 'CRT';
      }
    }

    var saved = localStorage.getItem('pixabros-crt') === 'on';
    setState(saved);

    toggle.addEventListener('click', function () {
      var next = !overlay.classList.contains('on');
      localStorage.setItem('pixabros-crt', next ? 'on' : 'off');
      setState(next);
    });
  }
})();
