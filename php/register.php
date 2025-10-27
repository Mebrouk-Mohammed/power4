<?php
// =======================================
// 1. CONNEXION À LA BASE DE DONNÉES
// =======================================

$pdo = null;
$dbError = '';

try {
    // IMPORTANT : même base que Go => power4_db
    $pdo = new PDO(
        'mysql:host=127.0.0.1;dbname=power4_db;charset=utf8mb4',
        'root',
        '',
        [
            PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
            // Afficher les résultats sous forme de tableau associatif
            PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
            // Activer le mode de simulation des requêtes
        ]
    );
} catch (PDOException $e) {
    // On capture l'erreur exacte pour l'afficher en haut de page
    $dbError = $e->getMessage();
}

// Message à afficher dans le formulaire (erreurs de validation etc.)
$message = '';

// =======================================
// 2. TRAITEMENT DU FORMULAIRE (POST)
// =======================================

if ($_SERVER['REQUEST_METHOD'] === 'POST' && $dbError === '') {
    $email    = trim($_POST['email']    ?? '');
    $username = trim($_POST['username'] ?? '');
    $password = $_POST['password']      ?? '';

    // Vérifications basiques
    if ($email === '' || $username === '' || $password === '') {
        $message = "❌ Tous les champs sont obligatoires.";
    } elseif (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
        $message = "❌ Adresse email invalide.";
    } elseif (mb_strlen($username) < 2 || mb_strlen($username) > 24) {
        $message = "❌ Le pseudo doit contenir entre 2 et 24 caractères.";
    } else {
        // Vérifier si l'email est déjà utilisé
        $stmt = $pdo->prepare("SELECT id FROM users WHERE email = ?");
        $stmt->execute([$email]);

        if ($stmt->fetch()) {
            $message = "⚠️ Cet email est déjà utilisé.";
        } else {
            // Hasher le mot de passe
            $hash = password_hash($password, PASSWORD_DEFAULT);

            // Insérer dans users
            $pdo->prepare(
                "INSERT INTO users (email, password_hash) VALUES (?, ?)"
            )->execute([$email, $hash]);

            // Récupérer l'id du nouvel utilisateur
            $user_id = $pdo->lastInsertId();

            // Créer le profil avec un ELO de départ à 1000
            $pdo->prepare(
                "INSERT INTO profiles (user_id, username, rating_elo)
                 VALUES (?, ?, 1000)"
            )->execute([$user_id, $username]);

            // Redirection vers login avec message de succès
            header('Location: login.php?created=1');
            exit;
        }
    }
}

?>
<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8" />
  <title>Inscription — Puissance 4</title>
  <style>
    body {
      background: #0f1115;
      color: #eaeaea;
      font-family: system-ui, Arial;
      margin: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
    }

    .wrap {
      width: 100%;
      max-width: 400px;
    }

    .debug {
      font-family: system-ui, Consolas, monospace;
      font-size: 13px;
      border-radius: 8px;
      padding: 10px 12px;
      margin-bottom: 16px;
      line-height: 1.4;
      white-space: pre-wrap;
      word-break: break-word;
    }
    .debug.ok {
      background: #113a1a;
      border: 1px solid #2b7a2f;
      color: #7dff94;
    }
    .debug.err {
      background: #3a1111;
      border: 1px solid #7a2f2f;
      color: #ff7d7d;
    }

    .card {
      background: #161a22;
      border: 1px solid #232839;
      border-radius: 14px;
      padding: 24px;
      width: 100%;
      box-shadow: 0 0 25px #000a;
    }

    h1 {
      color: #ffa94d;
      margin: 0 0 12px 0;
      font-size: 1.4rem;
      display: flex;
      align-items: center;
      gap: .5rem;
    }

    label {
      display: block;
      margin: 10px 0 4px;
      font-size: 0.9rem;
    }

    input {
      width: 100%;
      padding: 10px;
      border-radius: 8px;
      border: 1px solid #2f3650;
      background: #232839;
      color: #eaeaea;
      font-size: 0.95rem;
    }

    button {
      width: 100%;
      margin-top: 12px;
      padding: 12px;
      border-radius: 8px;
      border: 1px solid #2e344a;
      background: #ffa94d;
      color: #111;
      cursor: pointer;
      font-weight: bold;
      font-size: 0.95rem;
      transition: 0.2s;
    }

    button:hover {
      background: #ffb957;
      transform: scale(1.02);
    }

    .msg {
      margin-bottom: 10px;
      text-align: center;
      font-size: 0.9rem;
    }

    .footer-link {
      margin-top:10px;
      opacity:.8;
      font-size:14px;
      text-align:center;
    }

    a {
      color: #ffa94d;
      text-decoration: none;
    }
    a:hover {
      text-decoration: underline;
    }
  </style>
</head>

<body>
  <div class="wrap">

    <!-- Bandeau debug connexion DB -->
    <?php if ($dbError === ''): ?>
      <div class="debug ok">
        ✅ DB OK — connectée à power4_db
      </div>
    <?php else: ?>
      <div class="debug err">
        ❌ DB FAIL — erreur PDO :
        <?= htmlspecialchars($dbError) ?>
      </div>
    <?php endif; ?>

    <div class="card">
      <h1>📝 Créer un compte</h1>

      <?php if ($message): ?>
        <div class="msg"><?= htmlspecialchars($message) ?></div>
      <?php endif; ?>

      <form method="POST">
        <label>Email :</label>
        <input type="email" name="email" required>

        <label>Nom d’utilisateur :</label>
        <input type="text" name="username" required minlength="2" maxlength="24">

        <label>Mot de passe :</label>
        <input type="password" name="password" required minlength="6">

        <button type="submit" <?= $dbError !== '' ? 'disabled style="opacity:.4;cursor:not-allowed;"' : '' ?>>
          Créer mon compte
        </button>
      </form>

      <div class="footer-link">
        Déjà un compte ? <a href="login.php">Se connecter</a>
      </div>
    </div>
  </div>
</body>
</html>