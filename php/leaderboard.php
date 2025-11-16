<?php
<<<<<<< HEAD
require_once 'db.php';
session_start();

/* map Elo -> rang */
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

/* vérifie si la table user_ratings existe */
$hasRatings = false;
try { $pdo->query("SELECT 1 FROM user_ratings LIMIT 1"); $hasRatings = true; } catch(Throwable $e){}

/* on récupère Elo + stats basiques */
$sql = "
  SELECT
    u.id AS user_id,
    u.username,
    COALESCE(ur.rating, 1200) AS rating,
    COUNT(DISTINCT g.id) AS games_participated,
    SUM(CASE WHEN g.winner_id = u.id THEN 1 ELSE 0 END) AS wins
  FROM users u
  ".($hasRatings ? "LEFT JOIN user_ratings ur ON ur.user_id = u.id" : "LEFT JOIN (SELECT 0 AS user_id, 1200 AS rating) ur ON ur.user_id = u.id")."
  LEFT JOIN games g ON (g.player1_id = u.id OR g.player2_id = u.id)
  GROUP BY u.id, u.username, ur.rating
  ORDER BY rating DESC, wins DESC, games_participated DESC, username ASC
";
$rows = $pdo->query($sql)->fetchAll();
?>
<!doctype html><meta charset="utf-8">
<h1>Classement (Elo)</h1>
<table border="1" cellpadding="6" cellspacing="0">
  <tr><th>#</th><th>Joueur</th><th>Elo</th><th>Rang</th><th>Parties</th><th>Victoires</th></tr>
  <?php $i=1; foreach($rows as $r): $rank = rank_from_rating((int)$r['rating']); ?>
    <tr>
      <td><?= $i++ ?></td>
      <td><a href="public_profile.php?id=<?= (int)$r['user_id'] ?>"><?= htmlspecialchars($r['username']) ?></a></td>
      <td><?= (int)$r['rating'] ?></td>
      <td><?= $rank ?></td>
      <td><?= (int)($r['games_participated'] ?? 0) ?></td>
      <td><?= (int)($r['wins'] ?? 0) ?></td>
    </tr>
  <?php endforeach; ?>
</table>
<p><a href="index.php">Menu</a></p>

=======
declare(strict_types=1);
ini_set('display_errors','1');
error_reporting(E_ALL);

function h($s){ return htmlspecialchars((string)$s, ENT_QUOTES|ENT_SUBSTITUTE, 'UTF-8'); }

try {
  $pdo = new PDO('mysql:host=127.0.0.1;dbname=connect4;charset=utf8mb4','root','',[
    PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
    PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
  ]);
} catch (Throwable $e) {
  http_response_code(500);
  echo "Erreur DB: ".h($e->getMessage());
  exit;
}

$sql = "SELECT username, display_name, avatar_url, rating_elo, games_played, wins, losses, draws FROM profiles";
try {
  $hasPrivacy = $pdo->query("SHOW COLUMNS FROM profiles LIKE 'is_public'")->fetch();
  if ($hasPrivacy) $sql .= " WHERE is_public = 1";
} catch (Throwable $e) { /* ignore */ }
$sql .= " ORDER BY rating_elo DESC, wins DESC, games_played DESC LIMIT 100";
$players = $pdo->query($sql)->fetchAll();
?>
<!doctype html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <title>Classement — Puissance 4</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    :root{
      --bg:#0f1115; --panel:#14161a; --panel2:#1b1d22; --text:#eaeaea;
      --accent:#ffa94d; --hover: #2a2f3a;
    }
    *{box-sizing:border-box}
    body{font-family:system-ui, Arial; background:var(--bg); color:var(--text); margin:0}
    .wrap{max-width:1000px;margin:40px auto;padding:20px}

    /* ===== LIENS/BOUTONS FINS (surbrillance légère) ===== */
    .btn-ghost{
      display:inline-flex; align-items:center; gap:8px;
      background:#161922; color:var(--text);
      border:1px solid #23283a; border-radius:10px;
      padding:8px 12px; text-decoration:none; transition:.18s ease;
    }
    .btn-ghost:hover, .btn-ghost:focus-visible{
      background:var(--hover); outline:none;
      box-shadow:0 0 0 2px #ffffff22, 0 8px 20px #0006;
      transform:translateY(-1px);
    }
    .btn-row{display:flex; gap:10px; justify-content:flex-start; margin-bottom:14px}

    header{background:var(--panel);padding:16px 20px;border:1px solid #23283a;border-radius:12px}
    h1{margin:0;font-size:22px;color:var(--accent)}
    table{width:100%;border-collapse:collapse;margin-top:16px}
    th,td{padding:12px;text-align:left}
    th{background:var(--panel2);border-bottom:2px solid var(--accent)}
    tr:nth-child(even){background:#14161a}
    tr:hover{background:#1c1f26}
    .avatar{width:40px;height:40px;border-radius:8px;object-fit:cover;vertical-align:middle;margin-right:10px}
    a{color:var(--accent);text-decoration:none}
  </style>
</head>
<body>
  <div class="wrap">
    <div class="btn-row">
      <a class="btn-ghost" href="index.php">⬅️ Retour au menu</a>
    </div>

    <header><h1>🏆 Classement des joueurs — Puissance 4</h1></header>

    <table>
      <tr>
        <th>#</th><th>Joueur</th><th>ELO</th><th>Parties</th><th>Victoires</th><th>Défaites</th><th>Nuls</th>
      </tr>
      <?php if (empty($players)): ?>
        <tr><td colspan="7" style="text-align:center;padding:20px;">Aucun joueur trouvé.</td></tr>
      <?php else: ?>
        <?php foreach ($players as $i => $p): ?>
          <?php $avatar = $p['avatar_url'] ?: 'https://via.placeholder.com/40?text=%F0%9F%91%A4'; ?>
          <tr>
            <td><?= $i + 1 ?></td>
            <td>
              <a href="public_profile.php?u=<?= h($p['username']) ?>">
                <img src="<?= h($avatar) ?>" class="avatar" alt="">
                <?= h($p['display_name'] ?: $p['username']) ?>
              </a>
            </td>
            <td><?= (int)$p['rating_elo'] ?></td>
            <td><?= (int)$p['games_played'] ?></td>
            <td><?= (int)$p['wins'] ?></td>
            <td><?= (int)$p['losses'] ?></td>
            <td><?= (int)$p['draws'] ?></td>
          </tr>
        <?php endforeach; ?>
      <?php endif; ?>
    </table>
  </div>
</body>
</html>
>>>>>>> bda513fd2bb1669761e3605ed8a5539f7056ce17
