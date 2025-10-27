<?php
// 1. SESSION & REDIRECTION SI DÉJÀ CONNECTÉ
session_start();

if (!empty($_SESSION['user_id'])) {
  header('Location: index.php');
  exit;
}

// 2. CONNEXION À LA BASE DE DONNÉES
try {
  $pdo = new PDO(
    'mysql:host=127.0.0.1;dbname=power4_db;charset=utf8mb4',
    'root',
    '',
    [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
  );
} catch (PDOException $e) {
  http_response_code(500);
  exit('Erreur de connexion DB.');
}

// 3. TRAITEMENT DU FORMULAIRE DE CONNEXION
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {

  $email = trim($_POST['email'] ?? '');
  $password = $_POST['password'] ?? '';

  if ($email === '' || $password === '') {
    $message = '❌ Merci de remplir tous les champs.';
  } else {
    $st = $pdo->prepare("SELECT * FROM users WHERE email = ?");
    $st->execute([$email]);
    $user = $st->fetch(PDO::FETCH_ASSOC);

    if ($user && password_verify($password, $user['password_hash'])) {
      // OK -> connexion
      $_SESSION['user_id'] = (int)$user['id'];

      header('Location: index.php');
      exit;
    } else {
      $message = '❌ Email ou mot de passe incorrect.';
    }
  }
}

// 4. MESSAGES SPÉCIAUX (INSCRIPTION, SUPPRESSION)
if (!$message && isset($_GET['created'])) {
  $message = '✅ Compte créé. Connecte-toi !';
}
if (!$message && isset($_GET['deleted'])) {
  $message = '✅ Ton compte a bien été supprimé.';
}
?>
<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8" />
  <title>Connexion — Puissance 4</title>

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

    .card {
      background: #161a22;
      border: 1px solid #232839;
      border-radius: 14px;
      padding: 24px;
      width: 360px;
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
  </style>
</head>

<body>
  <div class="card">
    <h1>🔐 Connexion</h1>

    <?php if ($message): ?>
      <div class="msg"><?= htmlspecialchars($message) ?></div>
    <?php endif; ?>

    <form method="post" autocomplete="on">
      <label>Email</label>
      <input type="email" name="email" required autofocus>

      <label>Mot de passe</label>
      <input type="password" name="password" required>

      <button type="submit">Se connecter</button>
    </form>

    <div class="footer-link">
      Pas de compte ? <a href="register.php">Créer un compte</a>
    </div>
  </div>
</body>
</html>
