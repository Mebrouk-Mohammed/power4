// Attend que tout le DOM soit chargé avant d’exécuter le script
document.addEventListener('DOMContentLoaded', () => {
  // Formulaire principal pour jouer un coup (envoyé sur /play)
  const playForm = document.querySelector('form[action="/play"]');
  // Élément qui affiche le temps restant (texte du compte à rebours)
  const timeLeftEl = document.getElementById('time-left');
  // Conteneur du timer (peut servir pour styliser ou masquer le timer)
  const timerContainer = document.getElementById('turn-timer');
  // Durée maximale par tour en secondes
  const TIMEOUT = 10;
  // Identifiant du timer (setInterval)
  let timer = null;
  // Nombre de secondes restantes pour le tour en cours
  let remaining = TIMEOUT;

  // Active ou désactive les contrôles de jeu (boutons du formulaire /play)
  function disablePlayControls(disabled) {
    if (!playForm) return; // Si aucun formulaire, rien à faire

    // Sélectionne tous les boutons et submit dans le formulaire
    const buttons = playForm.querySelectorAll('button, input[type="submit"]');
    buttons.forEach(b => {
      b.disabled = disabled; // Active/désactive chaque bouton
    });

    // Ajoute ou retire une classe CSS pour indiquer que le formulaire est bloqué
    if (disabled) {
      playForm.classList.add('blocked-by-timer');
    } else {
      playForm.classList.remove('blocked-by-timer');
    }
  }

  // Fonction appelée quand le temps est écoulé
  function onTimeout() {
    // Bloque tous les contrôles de jeu
    disablePlayControls(true);

    // Arrête le timer s'il tourne encore
    if (timer) {
      clearInterval(timer);
      timer = null;
    }

    // Met à jour l'affichage du temps restant à 0
    if (timeLeftEl) timeLeftEl.textContent = '0';

    // Envoie une requête au serveur pour jouer un coup aléatoire
    fetch('/random_move', { method: 'POST' })
      .then(res => {
        if (!res.ok) throw new Error('random move failed'); // Erreur si réponse non OK
        // Recharge la page pour afficher le nouveau plateau après le coup aléatoire
        window.location.reload();
      })
      .catch(err => {
        console.error('Erreur random_move:', err);
        // En cas d'erreur, on laisse l'interface bloquée pour éviter un état incohérent
      });
  }

  // Lance ou relance le compte à rebours
  function startTimer() {
    if (!timeLeftEl) return; // Si pas d'affichage de temps, on ne fait rien

    // Réinitialise le temps restant
    remaining = TIMEOUT;
    timeLeftEl.textContent = remaining.toString(); // Affiche la valeur initiale

    // Réactive les contrôles de jeu au début du tour
    disablePlayControls(false);

    // Supprime un ancien timer s'il existe
    if (timer) clearInterval(timer);

    // Crée un nouveau timer qui s'exécute toutes les secondes
    timer = setInterval(() => {
      remaining -= 1; // Décrémente le temps restant
      timeLeftEl.textContent = remaining.toString(); // Met à jour l'affichage

      // Si le temps est écoulé, on arrête le timer et on déclenche onTimeout
      if (remaining <= 0) {
        clearInterval(timer);
        timer = null;
        onTimeout();
      }
    }, 1000);
  }

  // Si le formulaire de jeu existe, on met en place le timer
  if (playForm) {
    // Démarre le timer dès le chargement de la page
    startTimer();

    // Quand l'utilisateur soumet un coup, on arrête le timer
    playForm.addEventListener('submit', () => {
      if (timer) {
        clearInterval(timer);
        timer = null;
      }
      // Petit délai pour laisser la requête partir avant que la page ne se recharge
      setTimeout(() => {
        // Au retour (après redirection), le script se relancera et redémarrera le timer
      }, 200);
    });

    // Si l'utilisateur change de taille de plateau (nouvelle partie), on arrête aussi le timer
    const newForm = document.querySelector('form[action="/new"]');
    if (newForm) {
      newForm.addEventListener('submit', () => {
        if (timer) {
          clearInterval(timer);
          timer = null;
        }
      });
    }
  }
});
