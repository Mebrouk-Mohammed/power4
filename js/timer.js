document.addEventListener('DOMContentLoaded', () => {
  const playForm = document.querySelector('form[action="/play"]');
  const timeLeftEl = document.getElementById('time-left');
  const timerContainer = document.getElementById('turn-timer');
  const TIMEOUT = 10; // secondes
  let timer = null;
  let remaining = TIMEOUT;

  function disablePlayControls(disabled) {
    if (!playForm) return;
    const buttons = playForm.querySelectorAll('button, input[type="submit"]');
    buttons.forEach(b => { b.disabled = disabled; });
    if (disabled) {
      playForm.classList.add('blocked-by-timer');
    } else {
      playForm.classList.remove('blocked-by-timer');
    }
  }

  function onTimeout() {
    // Bloquer l'UI
    disablePlayControls(true);
    if (timer) { clearInterval(timer); timer = null; }
    if (timeLeftEl) timeLeftEl.textContent = '0';

    // Demander au serveur de poser un jeton aléatoire
    fetch('/random_move', { method: 'POST' })
      .then(res => {
        if (!res.ok) throw new Error('random move failed');
        // recharger la page pour afficher le nouveau plateau
        window.location.reload();
      })
      .catch(err => {
        console.error('Erreur random_move:', err);
        // en cas d'erreur on laisse l'UI bloquée pour éviter de contester l'état
      });
  }

  function startTimer() {
    if (!timeLeftEl) return;
    remaining = TIMEOUT;
    timeLeftEl.textContent = remaining;
    disablePlayControls(false);
    if (timer) clearInterval(timer);
    timer = setInterval(() => {
      remaining -= 1;
      timeLeftEl.textContent = remaining;
      if (remaining <= 0) {
        clearInterval(timer);
        timer = null;
        onTimeout();
      }
    }, 1000);
  }

  // Si le playForm existe, démarrer le timer
  if (playForm) {
    // Démarrer au chargement
    startTimer();

    // Si l'utilisateur clique pour jouer, on laisse la soumission normale (submit -> redirect)
    // mais on peut arrêter le timer pour éviter double envoi
    playForm.addEventListener('submit', () => {
      if (timer) { clearInterval(timer); timer = null; }
      // un léger délai pour laisser la requête partir
      setTimeout(() => {
        // lorsque la page revient (redirect), le script repartira et démarrera un nouveau timer
      }, 200);
    });

    // Aussi, si l'utilisateur change la taille (nouvelle partie), on redémarre le timer
    const newForm = document.querySelector('form[action="/new"]');
    if (newForm) {
      newForm.addEventListener('submit', () => {
        if (timer) { clearInterval(timer); timer = null; }
      });
    }
  }
});