<?php
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

