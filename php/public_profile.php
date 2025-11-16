<?php
<<<<<<< HEAD
require_once 'db.php';
session_start();

/* rechercher par id (prioritaire) ou par username ?u=... */
$user = null;
if (isset($_GET['id'])) {
  $st = $pdo->prepare("SELECT id, username, avatar_url, created_at FROM users WHERE id=?");
  $st->execute([(int)$_GET['id']]); $user = $st->fetch();
} elseif (isset($_GET['u'])) {
  $st = $pdo->prepare("SELECT id, username, avatar_url, created_at FROM users WHERE username=?");
  $st->execute([$_GET['u']]); $user = $st->fetch();
}
if (!$user) { http_response_code(404); die('Utilisateur introuvable'); }

/* Elo + rang (fallback 1200) */
function rank_from_rating(int $r): string {
  return match (true) {
    $r < 1000 => 'Bronze',
    $r < 1200 => 'Silver',
    $r < 1400 => 'Gold',
    $r < 1600 => 'Platinum',
    $r < 1800 => 'Diamond',
    $r < 2000 => 'Master',
    $r < 2200 => 'Grandmaster',
    default   => 'Challenger',
  };
}
$rating = 1200;
try {
  $rr = $pdo->prepare("SELECT rating FROM user_ratings WHERE user_id=?");
  $rr->execute([$user['id']]);
  $row = $rr->fetch();
  if ($row) $rating = (int)$row['rating'];
} catch (Throwable $e) {}
$rank = rank_from_rating($rating);

/* stats */
$stats = $pdo->prepare("
  SELECT COUNT(DISTINCT g.id) AS games_participated,
         SUM(CASE WHEN g.winner_id = ? THEN 1 ELSE 0 END) AS wins
  FROM games g
  WHERE g.player1_id = ? OR g.player2_id = ?
");
$stats->execute([$user['id'], $user['id'], $user['id']]);
$s = $stats->fetch();
?>
<!doctype html><meta charset="utf-8">
<h1>Profil de <?= htmlspecialchars($user['username']) ?></h1>
<p><img src="<?= htmlspecialchars($user['avatar_url'] ?: 'https://via.placeholder.com/96') ?>" alt="avatar" width="96" height="96" style="border-radius:8px"></p>
<p><strong>Elo :</strong> <?= $rating ?> &nbsp;|&nbsp; <strong>Rang :</strong> <?= $rank ?></p>
<p><strong>Membre depuis :</strong> <?= htmlspecialchars($user['created_at']) ?></p>
<p><strong>Parties :</strong> <?= (int)$s['games_participated'] ?> |
   <strong>Victoires :</strong> <?= (int)$s['wins'] ?></p>
<p><a href="leaderboard.php">Retour classement</a> | <a href="index.php">Menu</a></p>

=======
// -------------------------------
// public_profile.php
// Ce fichier affiche la page de profil publique d’un joueur (accessible via une URL comme :
// http://localhost/connect4_test/public_profile.php?u=Test)
// -------------------------------

// --- DEBUG ---
// Ces lignes activent l’affichage des erreurs PHP directement dans le navigateur.
// Très utile pendant le développement, mais à désactiver en production.
ini_set('display_errors', '1');
ini_set('display_startup_errors', '1');
error_reporting(E_ALL);

// Petite fonction utilitaire "h()" pour sécuriser tout texte affiché à l’écran.
// Elle empêche les failles XSS en convertissant les caractères spéciaux (ex: <, >, &, ", ') en entités HTML.
function h($s){ 
  return htmlspecialchars((string)$s, ENT_QUOTES|ENT_SUBSTITUTE, 'UTF-8'); 
}

// -------------------------------
//  Connexion à la base de données
// -------------------------------
try {
  // Création d'une connexion PDO à MySQL.
  $pdo = new PDO(
    // DSN = Data Source Name (type de base, adresse, nom de base, encodage)
    'mysql:host=127.0.0.1;dbname=connect4;charset=utf8mb4',
    // Identifiant MySQL (root = par défaut sur Laragon)
    'root',
    // Mot de passe (vide sur Laragon)
    '',
    [
      // Si une erreur SQL survient, on lance une exception
      PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
      // Les résultats seront sous forme de tableaux associatifs (clé = nom de colonne)
      PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
    ]
  );
} catch (Throwable $e) {
  // Si la connexion échoue :
  http_response_code(500); // → erreur HTTP 500 (erreur serveur)
  echo "❌ Erreur de connexion DB : " . h($e->getMessage());
  exit; // on arrête tout
}

// -------------------------------
//  Lecture du paramètre "u" dans l’URL
// Exemple : public_profile.php?u=Test → $_GET['u'] = "Test"
// -------------------------------
$username = isset($_GET['u']) ? trim($_GET['u']) : ''; // on supprime les espaces autour

// Vérifie que le paramètre "u" est valide :
// - contient uniquement lettres, chiffres, tiret ou underscore
// - longueur entre 2 et 40 caractères
if ($username === '' || !preg_match('/^[A-Za-z0-9_-]{2,40}$/', $username)) {
  http_response_code(400); // → mauvais paramètre
  echo "Profil invalide (username attendu : lettres/chiffres/_/-).";
  exit;
}

// -------------------------------
//  Lecture du profil dans la base de données
// -------------------------------
try {
  // On prépare une requête SQL pour chercher un utilisateur par son username
  $stmt = $pdo->prepare("
    SELECT
      p.user_id, p.username, p.display_name, p.bio, p.avatar_url,
      p.rating_elo, p.games_played, p.wins, p.losses, p.draws,
      u.created_at
    FROM profiles p
    JOIN users u ON u.id = p.user_id
    WHERE p.username = ?
    LIMIT 1
  ");

  // Exécution de la requête SQL avec la valeur du pseudo
  $stmt->execute([$username]);

  // On récupère le profil sous forme de tableau associatif
  $profile = $stmt->fetch();

  // Si aucun profil trouvé → erreur 404
  if (!$profile) {
    http_response_code(404);
    echo "Profil introuvable.";
    exit;
  }

  // -------------------------------
  //  Vérification de la confidentialité (is_public)
  // -------------------------------
  $isPublic = true; // par défaut, on considère le profil comme public

  try {
    // On vérifie si la colonne "is_public" existe dans la table "profiles"
    $check = $pdo->query("SHOW COLUMNS FROM profiles LIKE 'is_public'")->fetch();

    if ($check) {
      // Si la colonne existe, on lit sa valeur pour cet utilisateur
      $stmt2 = $pdo->prepare("SELECT is_public FROM profiles WHERE user_id=?");
      $stmt2->execute([$profile['user_id']]);
      // Si la colonne vaut 1 → profil public ; sinon privé
      $isPublic = (int)$stmt2->fetchColumn() === 1;
    }
  } catch (Throwable $e) {
    // Si une erreur survient, on ignore (pas grave)
  }

  // Si le profil est privé → on empêche l’accès
  if (!$isPublic) {
    http_response_code(403); // → accès interdit
    echo "Ce profil est privé.";
    exit;
  }

} catch (Throwable $e) {
  // Si une erreur SQL survient pendant la lecture du profil
  http_response_code(500);
  echo "❌ Erreur requête : " . h($e->getMessage());
  exit;
}

// Si aucun avatar n’est défini, on met une image par défaut
$avatar = $profile['avatar_url'] ?: 'https://via.placeholder.com/240?text=Avatar';
?>

<!-- -------------------------------
 HTML : affichage du profil public
-------------------------------- -->
<!doctype html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <title>Profil de <?= h($profile['username']) ?> — Puissance4</title>
  <!-- Meta description pour le SEO -->
  <meta name="description" content="<?= h(substr((string)($profile['bio'] ?? ''), 0, 140)) ?>">
  <meta name="viewport" content="width=device-width,initial-scale=1">

  <!-- --- CSS interne --- -->
  <style>
    /* Style global */
    body{
      font-family:system-ui, Arial;
      background:#0f1115;
      color:#eaeaea;
      padding:24px
    }

    /* Conteneur principal */
    .wrap{
      max-width:900px;
      margin:0 auto
    }

    /* Carte du profil */
    .card{
      background:#14161a;
      border-radius:12px;
      padding:20px;
      display:flex;
      gap:20px;
      align-items:flex-start
    }

    /* Avatar */
    .avatar{
      width:160px;
      height:160px;
      border-radius:10px;
      object-fit:cover;
      border:1px solid #222
    }

    /* Partie droite du profil */
    .meta{
      flex:1
    }

    /* Nom affiché */
    .meta h1{
      margin:0 0 6px 0
    }

    /* Bloc des statistiques (ELO, victoires, etc.) */
    .stats{
      display:flex;
      gap:10px;
      margin-top:12px
    }

    /* Style de chaque case de stats */
    .stat{
      background:#0f1216;
      padding:8px 12px;
      border-radius:8px;
      border:1px solid #222
    }

    /* Couleur des liens */
    a{
      color:#ffa94d;
      text-decoration:none
    }
  </style>
</head>

<body>
  <div class="wrap">
    <!-- Lien retour vers l'accueil -->
    <a href="index.php">← Accueil</a>

    <!-- Carte principale du profil -->
    <div class="card" style="margin-top:12px">
      <div>
        <!-- Affichage de l’avatar -->
        <img src="<?= h($avatar) ?>" alt="Avatar <?= h($profile['username']) ?>" class="avatar">
      </div>

      <div class="meta">
        <!-- Nom d’affichage (ou username s’il n’y en a pas) -->
        <h1><?= h($profile['display_name'] ?: $profile['username']) ?></h1>

        <!-- Nom d’utilisateur (avec @ devant) -->
        <div style="opacity:.8"><?= '@' . h($profile['username']) ?></div>

        <!-- Biographie -->
        <!-- nl2br() convertit les retours à la ligne en <br> -->
        <p style="margin-top:12px;white-space:pre-wrap;">
          <?= nl2br(h($profile['bio'] ?? '')) ?>
        </p>

        <!-- Statistiques du joueur -->
        <div class="stats">
          <div class="stat"><strong>ELO</strong><div><?= (int)$profile['rating_elo'] ?></div></div>
          <div class="stat"><strong>Parties</strong><div><?= (int)$profile['games_played'] ?></div></div>
          <div class="stat"><strong>V</strong><div><?= (int)$profile['wins'] ?></div></div>
          <div class="stat"><strong>D</strong><div><?= (int)$profile['losses'] ?></div></div>
          <div class="stat"><strong>Nuls</strong><div><?= (int)$profile['draws'] ?></div></div>
        </div>

        <!-- Date d’inscription -->
        <p style="margin-top:12px;opacity:.7">
          Membre depuis : <?= h($profile['created_at']) ?>
        </p>
      </div>
    </div>
  </div>
</body>
</html>
>>>>>>> bda513fd2bb1669761e3605ed8a5539f7056ce17
