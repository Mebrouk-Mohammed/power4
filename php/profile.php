<?php
require_once 'db.php';
session_start();
if (empty($_SESSION['user_id'])) { header('Location: login.php'); exit; }

$st = $pdo->prepare("SELECT id, username, email, avatar_url, created_at, last_login_at FROM users WHERE id=?");
$st->execute([$_SESSION['user_id']]);
$me = $st->fetch();
if (!$me) { session_destroy(); header('Location: login.php'); exit; }

/* Elo + rang (fallback 1200 si pas de table/ligne) */
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
  $rr->execute([$me['id']]);
  $row = $rr->fetch();
  if ($row) $rating = (int)$row['rating'];
} catch (Throwable $e) {}
$rank = rank_from_rating($rating);

/* stats rapides */
$stats = $pdo->prepare("
  SELECT
    COUNT(DISTINCT g.id) AS games_participated,
    SUM(CASE WHEN g.winner_id = ? THEN 1 ELSE 0 END) AS wins
  FROM games g
  WHERE g.player1_id = ? OR g.player2_id = ?
");
$stats->execute([$me['id'], $me['id'], $me['id']]);
$s = $stats->fetch();
?>
<!doctype html><meta charset="utf-8">
<h1>Mon profil</h1>
<p><img src="<?= htmlspecialchars($me['avatar_url'] ?: 'https://via.placeholder.com/96') ?>" alt="avatar" width="96" height="96" style="border-radius:8px"></p>
<p><strong>Pseudo :</strong> <?= htmlspecialchars($me['username']) ?></p>
<p><strong>Email :</strong> <?= htmlspecialchars($me['email'] ?? '—') ?></p>
<p><strong>Elo :</strong> <?= $rating ?> &nbsp;|&nbsp; <strong>Rang :</strong> <?= $rank ?></p>
<p><strong>Parties :</strong> <?= (int)$s['games_participated'] ?> |
   <strong>Victoires :</strong> <?= (int)$s['wins'] ?></p>
<p>
  <a href="choose_avatar.php">Changer d’avatar</a> |
  <a href="public_profile.php?id=<?= (int)$me['id'] ?>">Mon profil public</a> |
  <a href="leaderboard.php">Classement</a> |
  <a href="delete_account.php">Supprimer mon compte</a> |
  <a href="index.php">Menu</a>
</p>
