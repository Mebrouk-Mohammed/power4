(function () {
  // Physique simple (px/s^2). Augmentez g pour accélérer la chute.
  // Valeur par défaut augmentée pour des animations plus rapides.
  const DEFAULT_G = 2000;

  function animateFreeFall(element, startYpx, targetYpx, opts = {}) { // fonction qui anime la gravité de chute libre (intégration analytique)
    const g = opts.g ?? DEFAULT_G; // constante de gravité
    const initialV = opts.initialV ?? 0;
    const maxDt = opts.maxDt ?? 0.016; // ~60fps

    element.style.willChange = 'transform';
    // position initiale (top) doit déjà être définie sur l'élément

    return new Promise(resolve => {
      const startT = performance.now();
      // s(t) = v0 * t + 0.5 * g * t^2
      function step(now) {
        let t = (now - startT) / 1000;
        // clamp t to avoid huge jumps in case of throttling
        if (!isFinite(t) || t < 0) t = 0;

        // déplacement le long de la direction (targetY - startY)
        const dir = targetYpx >= startYpx ? 1 : -1;
        const s = initialV * t + 0.5 * g * t * t;
        const y = startYpx + dir * s;

        element.style.transform = `translate3d(0, ${y - startYpx}px, 0)`;

        const arrived = (dir === 1 && y >= targetYpx) || (dir === -1 && y <= targetYpx);
        if (arrived) {
          element.style.transform = `translate3d(0, ${targetYpx - startYpx}px, 0)`;
          resolve();
          return;
        }

        requestAnimationFrame(step);
      }

      requestAnimationFrame(step);
    });
  }

  // cree un jeton clone correspondant au visuel de la cellule
  function spawnTokenClone(cell) {
    // si la cellule contient un élément 'token' (variante néon), on l'utilise comme source visuelle
    const srcToken = cell.querySelector && (cell.querySelector('.token') || cell.querySelector('.token-p1') || cell.querySelector('.token-p2') || cell.querySelector('[class*="token-p1"]') || cell.querySelector('[class*="token-p2"]'));
    const sourceElement = srcToken || cell;
    const rect = sourceElement.getBoundingClientRect();
    const clone = document.createElement('div');
    clone.className = 'token-clone';
    clone.style.position = 'fixed';
    clone.style.left = rect.left + 'px';
  // apparait au dessus de la fenetre pour que le jeton tombe du haut

  // si la gravité est inversée, on fait apparaitre le clone en dessous de l'écran
  const inverted = (function(){ try { const b = document.querySelector('.gravity-btn[href*="inverted=true"]'); return !!(b && b.classList && b.classList.contains('active')); } catch(e){ return false; } })();
  const startY = inverted ? (window.innerHeight + Math.max(80, rect.height * 1.5)) : -80;
    clone.dataset._startY = startY;
    clone.style.top = startY + 'px';
    clone.style.width = rect.width + 'px';
    clone.style.height = rect.height + 'px';
  clone.style.zIndex = 9999;
  // réduire l'ombre pour soulager le rendu GPU pendant l'animation
  clone.style.boxShadow = '0 2px 6px rgba(0,0,0,0.25)';
    // copier le visuel depuis la source (token) si présente, sinon depuis la cellule
    const bg = getComputedStyle(sourceElement).backgroundImage || getComputedStyle(sourceElement).backgroundColor || getComputedStyle(cell).backgroundImage || getComputedStyle(cell).backgroundColor;
    clone.style.background = bg;
    clone.style.backgroundSize = 'contain';
    clone.style.backgroundRepeat = 'no-repeat';
    clone.style.backgroundPosition = 'center';
    document.body.appendChild(clone);
    return { clone, startY, targetTop: rect.top };
  }

  // quand la cellule est remplie le jeton ne tombe pas
  function onCellFilled(cell) {
    try {
  const { clone, startY, targetTop } = spawnTokenClone(cell);
  // accélère la chute pour l'animation de placement automatique
  animateFreeFall(clone, startY, targetTop, { g: 1500, restitution: 0.06 })
        .then(() => {
          clone.remove();
        })
        .catch(() => { clone.remove(); });
    } catch (e) {
      console.error('physics spawn error', e);
    }
  }

  // detection de la cellule lorsqu'elle remplie
    function installObserver() {
    // accepter plusieurs variantes de templates : board-inner, board-inner-neon, board-inner-neon-medium, ...
    const board = document.querySelector('.board-inner, .board-inner-neon, .board-inner-neon-medium, .board-inner-neon-large') || document.querySelector('[class*="board-inner"]');
    if (!board) return;
    const mo = new MutationObserver(muts => {
      for (const m of muts) {
        // Cas 1 : changement de classe sur une cellule (ancienne logique)
        if (m.type === 'attributes' && m.attributeName === 'class' && m.target.classList) {
          const target = m.target;
          if ((target.classList.contains('p1') || target.classList.contains('p2')) && !target.dataset._animated) {
            target.dataset._animated = '1';
            onCellFilled(target);
          }
        }

        // Cas 2 : nœuds ajoutés (ex : insertion d'un <div class="token..."> dans les variantes neon)
        if (m.type === 'childList' && m.addedNodes.length) {
          m.addedNodes.forEach(node => {
            if (node.nodeType !== 1) return;
            const el = node;

            // recueillir les éléments qui ressemblent à des tokens
            const possibleTokens = [];
            if (el.matches && (el.matches('.token') || el.matches('.token-p1') || el.matches('.token-p2') || el.matches('[class*="token-p1"]') || el.matches('[class*="token-p2"]'))) {
              possibleTokens.push(el);
            }
            if (el.querySelectorAll) {
              el.querySelectorAll('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]').forEach(t => possibleTokens.push(t));
            }

            possibleTokens.forEach(tokenEl => {
              // trouver la cellule/bouton cible (parent le plus proche)
              const targetCell = tokenEl.closest('button, .cell, .hole-neon, .hole-neon-medium, .hole-neon-large') || tokenEl.parentElement;
              if (targetCell && !targetCell.dataset._animated) {
                targetCell.dataset._animated = '1';
                onCellFilled(targetCell);
              }
            });
          });
        }
      }
    });

    mo.observe(board, { attributes: true, subtree: true, attributeFilter: ['class'], childList: true });
    console.log('physics observer installed');
  }

  // récupère le joueur courant depuis le message de statut (texte affiché)
  function getCurrentPlayerFromStatus() {
    try {
      const status = document.querySelector('.status');
      if (!status) return 1;
      const txt = status.textContent || '';
      const m = txt.match(/joueur\s*(\d+)/i) || txt.match(/player\s*(\d+)/i);
      if (m && m[1]) return Number(m[1]);
    } catch (e) { /* silent */ }
    return 1;
  }

  // détecte si la gravité est inversée 
  function isInvertedGravity() {
    try {
      const b = document.querySelector('.gravity-btn[href*="inverted=true"]');
      return !!(b && b.classList && b.classList.contains('active'));
    } catch (e) { return false; }
  }

  // Trouve la cellule (élément bouton) cible la plus basse pour une colonne donnée
  function findTargetCellForColumn(col) {
    const gridSelectors = ['.game-grid', '.game-grid-neon', '.game-grid-neon-medium', '.game-grid-neon-large', '.game-grid-neon'];
    const inverted = isInvertedGravity();
    for (const sel of gridSelectors) {
      const grid = document.querySelector(sel);
      if (!grid) continue;
      const colButtons = Array.from(grid.querySelectorAll(`button[name="col"][value="${col}"]`));
      if (!colButtons.length) continue;
      if (inverted) {
        // gravité normale : parcourir du haut vers le bas (ordre DOM naturel)
        for (let i = 0; i < colButtons.length; i++) {
          const btn = colButtons[i];
          const hasTokenChild = btn.querySelector && btn.querySelector('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]');
          const hasClassToken = btn.classList && (btn.classList.contains('p1') || btn.classList.contains('p2'));
          if (!hasTokenChild && !hasClassToken) return btn;
        }
      } else {
        // gravité inversé: parcourir de bas en haut
        for (let i = colButtons.length - 1; i >= 0; i--) {
          const btn = colButtons[i];
          const hasTokenChild = btn.querySelector && btn.querySelector('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]');
          const hasClassToken = btn.classList && (btn.classList.contains('p1') || btn.classList.contains('p2'));
          if (!hasTokenChild && !hasClassToken) return btn;
        }
      }
    }
    // fallback: chercher dans tout le document
    const all = Array.from(document.querySelectorAll(`button[name="col"][value="${col}"]`));
    if (inverted) {
      for (let i = 0; i < all.length; i++) {
        const btn = all[i];
        const hasTokenChild = btn.querySelector && btn.querySelector('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]');
        const hasClassToken = btn.classList && (btn.classList.contains('p1') || btn.classList.contains('p2'));
        if (!hasTokenChild && !hasClassToken) return btn;
      }
    } else {
      for (let i = all.length - 1; i >= 0; i--) {
        const btn = all[i];
        const hasTokenChild = btn.querySelector && btn.querySelector('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]');
        const hasClassToken = btn.classList && (btn.classList.contains('p1') || btn.classList.contains('p2'));
        if (!hasTokenChild && !hasClassToken) return btn;
      }
    }
    return null;
  }

  // Crée un clone visuel pour le joueur donné et l'anime vers targetEl.
  // Renvoie une Promise résolue quand l'animation est terminée.
  function spawnFallingCloneForPlayer(player, targetEl, opts = {}) {
    const rect = targetEl.getBoundingClientRect();
    const clone = document.createElement('div');
    clone.className = 'token-clone';
    clone.style.position = 'fixed';
    clone.style.width = rect.width + 'px';
    clone.style.height = rect.height + 'px';
    clone.style.left = rect.left + 'px';
  // si la gravité est inversée et qu'aucun startY n'est fourni, commencer sous l'écran
  const startY = opts.startY ?? (isInvertedGravity() ? (window.innerHeight + Math.max(80, rect.height * 1.5)) : -Math.max(80, rect.height * 1.5));
    clone.style.top = startY + 'px';
    clone.style.zIndex = 9999;
    const img = player === 1 ? "url('/static/img/orange_token.png')" : "url('/static/img/purple_token.png')";
    clone.style.backgroundImage = img;
    clone.style.backgroundSize = 'contain';
    clone.style.backgroundRepeat = 'no-repeat';
    clone.style.backgroundPosition = 'center';
    document.body.appendChild(clone);
    return animateFreeFall(clone, startY, rect.top, { g: opts.g ?? 700, restitution: opts.restitution ?? 0.06 }).then(() => { clone.remove(); });
  }

  // Intercepte les clics sur les boutons de contrôle (contrôles de colonne)
  function handleControlButtonClick(e) {
    // bouton de contrôle (flèche de la colonne)
    const btn = e.currentTarget;
    if (!btn || btn.disabled) return;
    e.preventDefault();
    const col = btn.value;
    if (col == null) { btn.form && btn.form.submit(); return; }
    const target = findTargetCellForColumn(col);
    if (!target) {
      // colonne pleine : petit feedback visuel
      btn.animate([{ transform: 'translateY(-3px)' }, { transform: 'translateY(0)' }], { duration: 200 });
      return;
    }
  const player = getCurrentPlayerFromStatus();
  // jouer l'animation puis soumettre le formulaire
  // accélération augmentée pour une chute plus rapide
  spawnFallingCloneForPlayer(player, target, { g: 1500, restitution: 0.06 }).then(() => {
      // soumettre le formulaire original pour effectuer le coup côté serveur
      if (btn.form) {
        if (typeof btn.form.requestSubmit === 'function') {
          btn.form.requestSubmit(btn);
        } else {
          const hidden = document.createElement('input');
          hidden.type = 'hidden';
          hidden.name = btn.name || 'col';
          hidden.value = btn.value;
          hidden.dataset._physics = '1';
          btn.form.appendChild(hidden);
          btn.form.submit();
          setTimeout(() => { try { hidden.remove(); } catch (e) {} }, 1000);
        }
      } else {
        // fallback : créer un formulaire et le soumettre
        const f = document.createElement('form');
        f.method = 'post';
        f.action = '/play';
        const inp = document.createElement('input'); inp.type = 'hidden'; inp.name = 'col'; inp.value = col; f.appendChild(inp); document.body.appendChild(f); f.submit();
      }
    }).catch(() => {
      if (btn.form) {
        try {
          if (typeof btn.form.requestSubmit === 'function') btn.form.requestSubmit(btn);
          else btn.form.submit();
        } catch (e) { btn.form.submit(); }
      }
    });
  }

  // Attache les listeners aux boutons de contrôle
  function attachControlListeners() {
    // trouver les boutons de contrôle (dans les conteneurs de contrôles)
    const allControlBtns = Array.from(document.querySelectorAll('button[name="col"]')).filter(b => !!b.closest('.controls, .neon-controls, .neon-controls-medium, .neon-controls-large, .sizes'));
    allControlBtns.forEach(b => {
      // éviter d'attacher deux fois
      if (b.dataset._physicsBound) return;
      b.addEventListener('click', handleControlButtonClick);
      b.dataset._physicsBound = '1';
    });
  }

  // Précharger les images de jetons pour éviter des décodages pendant l'animation,
  // puis installer observateur et listeners.
  document.addEventListener('DOMContentLoaded', () => {
    try {
      ['/static/img/orange_token.png', '/static/img/purple_token.png'].forEach(src => { const i = new Image(); i.src = src; });
    } catch (e) { /* ignore */ }
    installObserver();
    attachControlListeners();
  });
})();
