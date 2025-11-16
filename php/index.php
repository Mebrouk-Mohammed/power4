<?php
require_once 'db.php';
session_start();

/* Utilisateur connecté ? */
$user = null;
if (!empty($_SESSION['user_id'])) {
  $st = $pdo->prepare("SELECT id, username, avatar_url FROM users WHERE id=?");
  $st->execute([$_SESSION['user_id']]);
  $user = $st->fetch();
}
?>
<!doctype html>
<html lang="fr">
<head>
<meta charset="utf-8">
<title>Menu</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>
  :root{--bg:#0f1115;--card:#171a21;--muted:#9aa4b2;--acc:#7a5cff;}
  *{box-sizing:border-box} body{margin:0;font-family:system-ui,Segoe UI,Roboto,Arial,sans-serif;background:var(--bg);color:#fff}
  .wrap{max-width:980px;margin:32px auto;padding:0 16px}
  .top{display:flex;align-items:center;justify-content:space-between;margin-bottom:20px}
  .user{display:flex;align-items:center;gap:12px}
  .user img{width:40px;height:40px;border-radius:8px;object-fit:cover;border:1px solid #2a2f3a}
  .muted{color:var(--muted)}
  .grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:14px}
  .card{background:var(--card);border:1px solid #232838;border-radius:14px;padding:16px;transition:.15s;text-decoration:none;color:#fff;display:block}
  .card:hover{transform:translateY(-2px);box-shadow:0 6px 18px rgba(0,0,0,.25);border-color:#2f3650}
  .card h3{margin:0 0 6px;font-size:18px}
  .card p{margin:0;color:var(--muted);font-size:14px}
  .btn{background:var(--acc);border:0;color:#fff;padding:9px 14px;border-radius:10px;text-decoration:none;display:inline-block}
  .row{display:flex;gap:8px;flex-wrap:wrap;margin-top:10px}
  .sep{height:1px;background:#222938;margin:18px 0}
</style>
</head>
<body>
<div class="wrap">

  <div class="top">
    <h1 style="margin:0;">Menu</h1>
    <?php if ($user): ?>
      <div class="user">
        <?php if ($user['avatar_url']): ?><img src="<?=htmlspecialchars($user['avatar_url'])?>" alt="avatar"><?php endif; ?>
        <div>
          <div>Bonjour, <strong><?=htmlspecialchars($user['username'])?></strong> 👋</div>
          <div class="row">
            <a class="btn" href="logout.php">Se déconnecter</a>
          </div>
        </div>
      </div>
    <?php else: ?>
      <div class="row">
        <a class="btn" href="login.php">Se connecter</a>
        <a class="btn" href="register.php" style="background:#3aa675">S’inscrire</a>
      </div>
    <?php endif; ?>
  </div>

  <div class="sep"></div>

  <div class="grid">
    <?php if ($user): ?>
      <a class="card" href="profile.php">
        <h3>Mon profil</h3>
        <p>Voir mes infos, stats et actions.</p>
      </a>

      <a class="card" href="choose_avatar.php">
        <h3>Choisir un avatar</h3>
        <p>Sélectionner une image prédéfinie.</p>
      </a>

      <a class="card" href="public_profile.php?id=<?=$user['id']?>">
        <h3>Mon profil public</h3>
        <p>Page publique partageable.</p>
      </a>

      <a class="card" href="leaderboard.php">
        <h3>Classement</h3>
        <p>Top joueurs, parties et victoires.</p>
      </a>

      <!-- Ces cartes seront actives quand on ajoutera la partie “jeu” -->
      <a class="card" href="create_game.php">
        <h3>Créer une partie</h3>
        <p>Lancer une nouvelle partie de Power4.</p>
      </a>

      <a class="card" href="game.php">
        <h3>Rejoindre / voir une partie</h3>
        <p>Accéder à une partie (via ID).</p>
      </a>

      <a class="card" href="delete_account.php">
        <h3>Supprimer / Anonymiser</h3>
        <p>Gestion avancée du compte.</p>
      </a>
    <?php else: ?>
      <a class="card" href="login.php">
        <h3>Connexion</h3>
        <p>Accéder à ton compte.</p>
      </a>
      <a class="card" href="register.php">
        <h3>Inscription</h3>
        <p>Créer un nouveau compte.</p>
      </a>
      <a class="card" href="leaderboard.php">
        <h3>Classement</h3>
        <p>Voir les meilleurs joueurs.</p>
      </a>
      <a class="card" href="public_profile.php">
        <h3>Profil public (connecté)</h3>
        <p>Redirigera vers ton profil si connecté.</p>
      </a>
    <?php endif; ?>
  </div>

  <p class="muted" style="margin-top:18px">
    Dossier actuel : <code>php/</code> • Tout est accessible depuis cette page.
  </p>
</div>
</body>
</html>
