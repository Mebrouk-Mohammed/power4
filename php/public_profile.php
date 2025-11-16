<?php
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

