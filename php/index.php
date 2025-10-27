<?php
session_start();

// 1. Si l'utilisateur n'est pas connecté -> retourne à login
if (empty($_SESSION['user_id'])) {
    header('Location: login.php');
    exit;
}

// === CONFIG FACILE À CHANGER ===
// URL vers ton serveur Go (tel que tu le lances avec `go run main.go`)
$GO_BASE = 'http://127.0.0.1:8080';

// 2. Connexion à ta nouvelle base power4_db
try {
    $pdo = new PDO(
        'mysql:host=127.0.0.1;dbname=power4_db;charset=utf8mb4', // <-- IMPORTANT: ta nouvelle base
        'root',
        '',
        [
            PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
            PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
        ]
    );
} catch (Throwable $e) {
    http_response_code(500);
    exit('Erreur de connexion à la base.');
}

// 3. Récup infos joueur connecté
$userId = $_SESSION['user_id'];

$st = $pdo->prepare("
    SELECT p.username, p.rating_elo
    FROM profiles p
    WHERE p.user_id = ?
");
$st->execute([$userId]);
$me = $st->fetch();

$username = $me['username'] ?? 'Joueur';
$elo      = isset($me['rating_elo']) ? (int)$me['rating_elo'] : 1000;
?>
<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8" />
  <title>Menu principal — Puissance 4</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <style>
    :root { --accent: #ffa94d; }

    html, body { height:100%; }
    body {
      margin: 0;
      font-family: system-ui, "Poppins", Arial, sans-serif;
      background: radial-gradient(circle at 30% -20%, #1a1d24, #0f1115 55%);
      color: #f5f5f5;
      display: grid;
      place-items: center;
      min-height: 100vh;
    }

    .menu {
      width: 360px;
      background: #161a22;
      border: 1px solid #22283a;
      border-radius: 16px;
      box-shadow: 0 10px 32px rgba(0,0,0,.45);
      padding: 34px 26px;
      text-align: center;
    }

    h1 {
      margin: 0 0 6px;
      font-size: 26px;
      color: var(--accent);
    }

    .hello {
      opacity: .85;
      margin-bottom: 4px;
      line-height: 1.4;
    }

    .elo {
      font-size: 14px;
      opacity: .75;
      margin-bottom: 20px;
    }

    /* Boutons = <a> cliquables */
    .btn {
      display:block;
      width:100%;
      margin: 10px 0;
      padding: 12px 14px;
      border-radius: 10px;
      border: 1px solid #2e344a;
      background: #232839;
      color: #f5f5f5;
      text-decoration: none;
      font-size: 16px;
      transition:
        transform .16s ease,
        box-shadow .16s ease,
        background-color .16s ease,
        color .16s ease,
        border-color .16s ease;
      position: relative;
      overflow: hidden;
    }

    .btn:hover,
    .btn:focus-visible {
      background: #2a3044;
      border-color: #3a4260;
      transform: translateY(-1px);
      box-shadow:
        0 0 0 2px #ffa94d33,
        0 12px 26px rgba(0,0,0,.45);
      outline: none;
    }

    .btn:active {
      transform: translateY(0);
      box-shadow:
        0 0 0 2px #ffa94d55,
        0 8px 18px rgba(0,0,0,.45);
    }

    .btn.danger {
      background:#3a1b1b;
      border-color:#5a2222;
      color:#ff6b6b;
    }

    .btn.danger:hover,
    .btn.danger:focus-visible {
      background:#ff6b6b;
      color:#111;
      box-shadow:
        0 0 0 2px #ff6b6b55,
        0 12px 26px rgba(0,0,0,.45);
    }

    .section {
      margin-top: 14px;
      display:none;
    }

    .toggle {
      display:flex;
      align-items:center;
      justify-content:center;
      gap:10px;
    }

    .toggle input[type="checkbox"] {
      transform: scale(1.3);
      cursor: pointer;
    }

    small {
      display:block;
      margin-top: 16px;
      opacity:.7;
      font-size: 12px;
    }
  </style>
</head>
<body>
  <div class="menu">
    <h1>Menu principal 🎮</h1>

    <div class="hello">
      Bonjour,
      <strong><?= htmlspecialchars($username) ?></strong>
    </div>

    <div class="elo">
      Ton ELO : <strong><?= $elo ?></strong>
    </div>

    <!-- Lancer une partie → serveur Go -->
    <a class="btn" href="<?= htmlspecialchars($GO_BASE) ?>/game">▶️ Lancer une partie</a>

    <!-- Liens internes PHP -->
    <a class="btn" href="profile.php">👤 Voir mon profil</a>
    <a class="btn" href="leaderboard.php">🏆 Classement</a>

    <!-- Paramètres (toggle simple) -->
    <a class="btn" href="#" onclick="toggleSettings();return false;">⚙️ Paramètres</a>
    <div id="settings" class="section">
      <div class="toggle">
        <label for="sound">🔊 Son :</label>
        <input type="checkbox" id="sound" checked>
      </div>

      <a
        class="btn danger"
        href="delete_account.php"
        onclick="return confirm('⚠️ Supprimer ton compte ? Cette action est irréversible.');"
      >🗑️ Supprimer mon compte</a>
    </div>

    <a class="btn" href="logout.php">🚪 Se déconnecter</a>

    <small>
      Connecté à Puissance 4<br>
      (PHP + Go + power4_db)
    </small>
  </div>

  <script>
    function toggleSettings() {
      const s = document.getElementById('settings');
      s.style.display = (s.style.display === 'block') ? 'none' : 'block';
    }
  </script>
</body>
</html>
