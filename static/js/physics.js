(function () { // Fonction immédiatement invoquée pour isoler le code dans une portée locale
  // Constante de gravité par défaut (en px/s^2) pour la chute des jetons
  const DEFAULT_G = 2000;

  // Fonction qui anime une chute libre entre une position verticale de départ et une cible
  function animateFreeFall(element, startYpx, targetYpx, opts = {}) {
    const g = opts.g ?? DEFAULT_G;       // Gravité utilisée pour l'animation
    const initialV = opts.initialV ?? 0; // Vitesse initiale (en px/s)
    const maxDt = opts.maxDt ?? 0.016;   // Pas de temps maximum (~60fps), gardé au cas où

    element.style.willChange = 'transform'; // Indique au navigateur que "transform" va souvent changer

    // Retourne une promesse résolue en fin d'animation
    return new Promise(resolve => {
      const startT = performance.now(); // Temps de départ de l'animation

      // Fonction appelée à chaque frame par requestAnimationFrame
      function step(now) {
        let t = (now - startT) / 1000; // Temps écoulé en secondes
        if (!isFinite(t) || t < 0) t = 0; // Sécurise la valeur de t

        // Direction de déplacement selon la position de cible
        const dir = targetYpx >= startYpx ? 1 : -1;
        // Formule de déplacement: s(t) = v0 * t + 0.5 * g * t^2
        const s = initialV * t + 0.5 * g * t * t;
        const y = startYpx + dir * s; // Position verticale actuelle

        // Applique la translation relative sur l'élément
        element.style.transform = `translate3d(0, ${y - startYpx}px, 0)`;

        // Vérifie si on a atteint ou dépassé la position cible
        const arrived = (dir === 1 && y >= targetYpx) || (dir === -1 && y <= targetYpx);
        if (arrived) {
          // Force la position finale exacte
          element.style.transform = `translate3d(0, ${targetYpx - startYpx}px, 0)`;
          resolve(); // Termine la promesse
          return;
        }

        // Continue l'animation à la frame suivante
        requestAnimationFrame(step);
      }

      // Lance la première frame
      requestAnimationFrame(step);
    });
  }

  // Crée un clone visuel du jeton à partir d'une cellule cible
  function spawnTokenClone(cell) {
    // Cherche un jeton visuel dans la cellule (versions néon ou variantes)
    const srcToken = cell.querySelector && (
      cell.querySelector('.token') ||
      cell.querySelector('.token-p1') ||
      cell.querySelector('.token-p2') ||
      cell.querySelector('[class*="token-p1"]') ||
      cell.querySelector('[class*="token-p2"]')
    );

    const sourceElement = srcToken || cell;     // Élément source pour récupérer le style
    const rect = sourceElement.getBoundingClientRect(); // Position et taille absolue
    const clone = document.createElement('div'); // Création du clone
    clone.className = 'token-clone';            // Classe pour style éventuel
    clone.style.position = 'fixed';             // Position fixe par rapport à la fenêtre
    clone.style.left = rect.left + 'px';        // Position horizontale identique à la source

    // Détermine si la gravité est inversée à partir du bouton d'interface
    const inverted = (function () {
      try {
        const b = document.querySelector('.gravity-btn[href*="inverted=true"]');
        return !!(b && b.classList && b.classList.contains('active'));
      } catch (e) {
        return false;
      }
    })();

    // Position de départ verticale:
    // - si gravité normale: au-dessus de l'écran
    // - si gravité inversée: en dessous de l'écran
    const startY = inverted
      ? (window.innerHeight + Math.max(80, rect.height * 1.5))
      : -80;

    clone.dataset._startY = startY.toString(); // Stocke la position de départ en data-attribute
    clone.style.top = startY + 'px';           // Applique la position verticale de départ
    clone.style.width = rect.width + 'px';     // Largeur identique à la cellule
    clone.style.height = rect.height + 'px';   // Hauteur identique à la cellule
    clone.style.zIndex = 9999;                 // S'assure qu'il est au-dessus du reste

    // Réduit l'ombre pour limiter le coût GPU durant l'animation
    clone.style.boxShadow = '0 2px 6px rgba(0,0,0,0.25)';

    // Copie l'apparence de la source (image de fond ou couleur)
    const bg =
      getComputedStyle(sourceElement).backgroundImage ||
      getComputedStyle(sourceElement).backgroundColor ||
      getComputedStyle(cell).backgroundImage ||
      getComputedStyle(cell).backgroundColor;

    clone.style.background = bg;              // Applique le fond copié
    clone.style.backgroundSize = 'contain';   // L'image reste entière
    clone.style.backgroundRepeat = 'no-repeat';
    clone.style.backgroundPosition = 'center';

    // Ajoute le clone dans le document pour qu'il soit visible
    document.body.appendChild(clone);

    // Renvoie le clone et les positions utiles pour l'animation
    return { clone, startY, targetTop: rect.top };
  }

  // Appelée quand une cellule vient d'être remplie par un jeton
  function onCellFilled(cell) {
    try {
      // Crée le clone et récupère les positions
      const { clone, startY, targetTop } = spawnTokenClone(cell);

      // Lance l'animation de chute du clone vers la position finale
      animateFreeFall(clone, startY, targetTop, { g: 1500, restitution: 0.06 })
        .then(() => {
          clone.remove(); // Supprime le clone après l'animation
        })
        .catch(() => {
          clone.remove(); // Supprime le clone même en cas d'erreur
        });
    } catch (e) {
      console.error('physics spawn error', e); // Log d'erreur si problème
    }
  }

  // Installe un MutationObserver pour détecter quand une cellule est remplie
  function installObserver() {
    // Cherche le conteneur de plateau parmi plusieurs classes possibles
    const board =
      document.querySelector('.board-inner, .board-inner-neon, .board-inner-neon-medium, .board-inner-neon-large') ||
      document.querySelector('[class*="board-inner"]');

    if (!board) return; // Si aucun plateau trouvé, on n'installe rien

    // Création du MutationObserver pour suivre les changements dans le plateau
    const mo = new MutationObserver(muts => {
      for (const m of muts) {
        // Cas 1 : changement de classe sur une cellule (ancienne logique)
        if (m.type === 'attributes' && m.attributeName === 'class' && m.target.classList) {
          const target = m.target;
          // Si la cellule devient p1 ou p2 et n'a pas encore été animée
          if ((target.classList.contains('p1') || target.classList.contains('p2')) && !target.dataset._animated) {
            target.dataset._animated = '1'; // Marque comme animée
            onCellFilled(target);          // Lance l'animation
          }
        }

        // Cas 2 : ajout de nœuds (ex: insertion d'un <div class="token...">)
        if (m.type === 'childList' && m.addedNodes.length) {
          m.addedNodes.forEach(node => {
            if (node.nodeType !== 1) return; // On ne gère que les éléments
            const el = node;

            // Liste des éléments qui ressemblent à des jetons
            const possibleTokens = [];

            // Si le nœud lui-même est un token
            if (
              el.matches &&
              (el.matches('.token') ||
                el.matches('.token-p1') ||
                el.matches('.token-p2') ||
                el.matches('[class*="token-p1"]') ||
                el.matches('[class*="token-p2"]'))
            ) {
              possibleTokens.push(el);
            }

            // Cherche des jetons dans ses descendants
            if (el.querySelectorAll) {
              el
                .querySelectorAll('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]')
                .forEach(t => possibleTokens.push(t));
            }

            // Pour chaque jeton trouvé, on cherche la cellule associée et on anime
            possibleTokens.forEach(tokenEl => {
              // Recherche du bouton ou de la "cellule" la plus proche
              const targetCell =
                tokenEl.closest('button, .cell, .hole-neon, .hole-neon-medium, .hole-neon-large') ||
                tokenEl.parentElement;
              if (targetCell && !targetCell.dataset._animated) {
                targetCell.dataset._animated = '1'; // Marque la cellule comme animée
                onCellFilled(targetCell);          // Lance l'animation pour cette cellule
              }
            });
          });
        }
      }
    });

    // Observe les changements de classes et l'ajout d'enfants dans tout le sous-arbre du plateau
    mo.observe(board, { attributes: true, subtree: true, attributeFilter: ['class'], childList: true });
    console.log('physics observer installed'); // Indique dans la console que l'observateur est installé
  }

  // Récupère le joueur courant en analysant le texte de statut affiché
  function getCurrentPlayerFromStatus() {
    try {
      const status = document.querySelector('.status'); // Élément affichant le statut
      if (!status) return 1;                            // Par défaut, joueur 1
      const txt = status.textContent || '';             // Texte du statut
      // Cherche "joueur X" ou "player X" dans le texte
      const m = txt.match(/joueur\s*(\d+)/i) || txt.match(/player\s*(\d+)/i);
      if (m && m[1]) return Number(m[1]);               // Retourne le numéro trouvé
    } catch (e) {
      // Erreurs silencieuses
    }
    return 1; // Valeur par défaut si rien trouvé
  }

  // Indique si la gravité inversée est active d'après l'état du bouton de contrôle
  function isInvertedGravity() {
    try {
      const b = document.querySelector('.gravity-btn[href*="inverted=true"]'); // Bouton qui active la gravité inversée
      return !!(b && b.classList && b.classList.contains('active'));           // Vrai si ce bouton est actif
    } catch (e) {
      return false;
    }
  }

  // Trouve la cellule cible la plus appropriée pour une colonne donnée
  // selon la gravité (normale: case la plus basse vide, inversée: case la plus haute vide)
  function findTargetCellForColumn(col) {
    const gridSelectors = [
      '.game-grid',
      '.game-grid-neon',
      '.game-grid-neon-medium',
      '.game-grid-neon-large',
      '.game-grid-neon',
    ]; // Différentes classes possibles pour la grille
    const inverted = isInvertedGravity(); // Vérifie le mode de gravité

    // Parcourt les différents sélecteurs possibles de grilles
    for (const sel of gridSelectors) {
      const grid = document.querySelector(sel);
      if (!grid) continue; // Si aucune grille trouvée avec ce sélecteur, on continue

      // Récupère tous les boutons de cette colonne dans cette grille
      const colButtons = Array.from(grid.querySelectorAll(`button[name="col"][value="${col}"]`));
      if (!colButtons.length) continue; // Si rien dans cette grille, on essaie la suivante

      if (inverted) {
        // Gravité inversée : on parcourt du haut vers le bas (ordre DOM)
        for (let i = 0; i < colButtons.length; i++) {
          const btn = colButtons[i];
          const hasTokenChild =
            btn.querySelector &&
            btn.querySelector('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]');
          const hasClassToken =
            btn.classList && (btn.classList.contains('p1') || btn.classList.contains('p2'));
          if (!hasTokenChild && !hasClassToken) return btn; // Première cellule vide trouvée
        }
      } else {
        // Gravité normale : on parcourt du bas vers le haut
        for (let i = colButtons.length - 1; i >= 0; i--) {
          const btn = colButtons[i];
          const hasTokenChild =
            btn.querySelector &&
            btn.querySelector('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]');
          const hasClassToken =
            btn.classList && (btn.classList.contains('p1') || btn.classList.contains('p2'));
          if (!hasTokenChild && !hasClassToken) return btn; // Première cellule vide à partir du bas
        }
      }
    }

    // Si aucune grille connue n'a fonctionné, on cherche dans tout le document (fallback)
    const all = Array.from(document.querySelectorAll(`button[name="col"][value="${col}"]`));
    if (inverted) {
      // Gravité inversée : du haut vers le bas
      for (let i = 0; i < all.length; i++) {
        const btn = all[i];
        const hasTokenChild =
          btn.querySelector &&
          btn.querySelector('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]');
        const hasClassToken =
          btn.classList && (btn.classList.contains('p1') || btn.classList.contains('p2'));
        if (!hasTokenChild && !hasClassToken) return btn;
      }
    } else {
      // Gravité normale : du bas vers le haut
      for (let i = all.length - 1; i >= 0; i--) {
        const btn = all[i];
        const hasTokenChild =
          btn.querySelector &&
          btn.querySelector('.token, .token-p1, .token-p2, [class*="token-p1"], [class*="token-p2"]');
        const hasClassToken =
          btn.classList && (btn.classList.contains('p1') || btn.classList.contains('p2'));
        if (!hasTokenChild && !hasClassToken) return btn;
      }
    }
    return null; // Aucune cellule disponible trouvée
  }

  // Crée un clone visuel pour le joueur donné, l'anime jusqu'à targetEl, puis le supprime
  function spawnFallingCloneForPlayer(player, targetEl, opts = {}) {
    const rect = targetEl.getBoundingClientRect(); // Position et taille de la cellule cible
    const clone = document.createElement('div');   // Nouveau div pour le jeton animé
    clone.className = 'token-clone';              // Classe pour style éventuel
    clone.style.position = 'fixed';               // Position fixe (par rapport à l'écran)
    clone.style.width = rect.width + 'px';        // Largeur identique à la cellule
    clone.style.height = rect.height + 'px';      // Hauteur identique à la cellule
    clone.style.left = rect.left + 'px';          // Position horizontale initiale

    // Détermine la position de départ verticale (au-dessus ou en dessous de l'écran si non fournie)
    const startY =
      opts.startY ??
      (isInvertedGravity()
        ? (window.innerHeight + Math.max(80, rect.height * 1.5)) // Gravité inversée: départ sous l'écran
        : -Math.max(80, rect.height * 1.5));                    // Gravité normale: départ au-dessus

    clone.style.top = startY + 'px'; // Applique cette position verticale
    clone.style.zIndex = 9999;       // Devant tout le reste

    // Choisit l'image du jeton selon le joueur (1 = orange, 2 = violet)
    const img =
      player === 1
        ? "url('/static/img/orange_token.png')"
        : "url('/static/img/purple_token.png')";

    clone.style.backgroundImage = img;   // Image du jeton
    clone.style.backgroundSize = 'contain';
    clone.style.backgroundRepeat = 'no-repeat';
    clone.style.backgroundPosition = 'center';

    // Ajoute le clone au document
    document.body.appendChild(clone);

    // Lance l'animation de chute, puis supprime le clone à la fin
    return animateFreeFall(clone, startY, rect.top, {
      g: opts.g ?? 700,
      restitution: opts.restitution ?? 0.06,
    }).then(() => {
      clone.remove();
    });
  }

  // Gestionnaire de clic sur un bouton de contrôle (flèche en haut de chaque colonne)
  function handleControlButtonClick(e) {
    const btn = e.currentTarget; // Bouton cliqué
    if (!btn || btn.disabled) return; // Ignorer si bouton absent ou désactivé

    e.preventDefault(); // Empêche l'envoi immédiat du formulaire

    const col = btn.value; // Récupère la colonne associée à ce bouton
    if (col == null) {
      // Si pas de valeur, on envoie simplement le formulaire
      btn.form && btn.form.submit();
      return;
    }

    // Cherche la cellule cible où le jeton doit atterrir dans cette colonne
    const target = findTargetCellForColumn(col);
    if (!target) {
      // Si la colonne est pleine, petit effet visuel de "rebond"
      btn.animate(
        [{ transform: 'translateY(-3px)' }, { transform: 'translateY(0)' }],
        { duration: 200 }
      );
      return;
    }

    // Détermine le joueur courant à partir du texte de statut
    const player = getCurrentPlayerFromStatus();

    // Lance l'animation de chute, puis envoie le formulaire vers le serveur
    spawnFallingCloneForPlayer(player, target, { g: 1500, restitution: 0.06 })
      .then(() => {
        // Une fois l'animation terminée, on soumet le formulaire original
        if (btn.form) {
          if (typeof btn.form.requestSubmit === 'function') {
            btn.form.requestSubmit(btn); // Envoie proprement le formulaire
          } else {
            // Fallback pour anciens navigateurs: ajout d'un input caché
            const hidden = document.createElement('input');
            hidden.type = 'hidden';
            hidden.name = btn.name || 'col';
            hidden.value = btn.value;
            hidden.dataset._physics = '1';
            btn.form.appendChild(hidden);
            btn.form.submit();
            // Nettoie l'input caché après un délai
            setTimeout(() => {
              try {
                hidden.remove();
              } catch (e) {}
            }, 1000);
          }
        } else {
          // Fallback ultime: création d'un formulaire à la volée
          const f = document.createElement('form');
          f.method = 'post';
          f.action = '/play';
          const inp = document.createElement('input');
          inp.type = 'hidden';
          inp.name = 'col';
          inp.value = col;
          f.appendChild(inp);
          document.body.appendChild(f);
          f.submit();
        }
      })
      .catch(() => {
        // En cas d'échec de l'animation, on soumet malgré tout le formulaire
        if (btn.form) {
          try {
            if (typeof btn.form.requestSubmit === 'function') btn.form.requestSubmit(btn);
            else btn.form.submit();
          } catch (e) {
            btn.form.submit();
          }
        }
      });
  }

  // Attache les événements de clic sur les boutons de contrôle de colonnes
  function attachControlListeners() {
    // Sélectionne les boutons de colonnes dans les zones de contrôles
    const allControlBtns = Array.from(
      document.querySelectorAll('button[name="col"]')
    ).filter(b =>
      !!b.closest('.controls, .neon-controls, .neon-controls-medium, .neon-controls-large, .sizes')
    );

    allControlBtns.forEach(b => {
      // Évite de réattacher l'événement plusieurs fois
      if (b.dataset._physicsBound) return;
      b.addEventListener('click', handleControlButtonClick);
      b.dataset._physicsBound = '1'; // Marque ce bouton comme déjà relié
    });
  }

  // Au chargement du DOM, on prépare les images et on installe l'observateur et les événements
  document.addEventListener('DOMContentLoaded', () => {
    try {
      // Précharge les images de jetons pour éviter des latences lors de la première animation
      ['/static/img/orange_token.png', '/static/img/purple_token.png'].forEach(src => {
        const i = new Image();
        i.src = src;
      });
    } catch (e) {
      // Erreurs silencieuses sur le préchargement
    }

    // Installe l'observateur sur le plateau pour réagir aux cellules remplies
    installObserver();

    // Attache les gestionnaires de clic sur les boutons de contrôle
    attachControlListeners();
  });
})(); // Fin de la fonction auto-invoquée
