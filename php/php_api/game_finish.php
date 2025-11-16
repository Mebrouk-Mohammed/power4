<?php
require_once __DIR__.'/../api_boot.php';
require_api_key();

$in = json_input(); // { "game_id":1, "status":"finished", "winner_id":1 | null }
$_POST = $in; // on réutilise le code ci-dessous tel quel

// === à partir d’ici, c’est la version finish_game "K variable" convertie en JSON ===
$gid    = (int)($_POST['game_id'] ?? 0);
$status = $_POST['status'] ?? 'finished';
$winner = isset($_POST['winner_id']) && $_POST['winner_id'] !== '' ? (int)$_POST['winner_id'] : null;

if ($gid <= 0) out(['error'=>'game_id requis'],400);

try {
  $pdo->beginTransaction();

  $g = $pdo->prepare("SELECT player1_id, player2_id, status FROM games WHERE id=? FOR UPDATE");
  $g->execute([$gid]);
  $game = $g->fetch();
  if (!$game) throw new RuntimeException('Partie introuvable');

  $p1 = (int)$game['player1_id'];
  $p2 = (int)$game['player2_id'];

  $upd = $pdo->prepare("UPDATE games
                        SET status=?, finished_at=NOW(), winner_id=?, player_to_move=NULL
                        WHERE id=?");
  $upd->execute([$status, $winner, $gid]);

  if ($status !== 'finished' || !$p2) {
    $pdo->commit(); out(['ok'=>true,'elo_updated'=>false]); exit;
  }

  // ratings
  $getR = $pdo->prepare("SELECT user_id, rating FROM user_ratings WHERE user_id IN (?,?)");
  $getR->execute([$p1, $p2]);
  $r1 = 1200; $r2 = 1200;
  foreach ($getR as $row) {
    if ((int)$row['user_id'] === $p1) $r1 = (int)$row['rating'];
    if ((int)$row['user_id'] === $p2) $r2 = (int)$row['rating'];
  }
  $insR = $pdo->prepare("INSERT IGNORE INTO user_ratings (user_id, rating) VALUES (?,1200)");
  $insR->execute([$p1]);
  $insR->execute([$p2]);

  // parties jouées
  $cnt1 = $pdo->prepare("SELECT COUNT(*) AS c FROM games WHERE status='finished' AND (player1_id=? OR player2_id=?)");
  $cnt2 = $pdo->prepare("SELECT COUNT(*) AS c FROM games WHERE status='finished' AND (player1_id=? OR player2_id=?)");
  $cnt1->execute([$p1, $p1]); $c1 = (int)$cnt1->fetch()['c'];
  $cnt2->execute([$p2, $p2]); $c2 = (int)$cnt2->fetch()['c'];

  $K1 = ($c1 < 30) ? 40 : (($r1 < 2000) ? 20 : 10);
  $K2 = ($c2 < 30) ? 40 : (($r2 < 2000) ? 20 : 10);

  if ($winner === null) { $S1 = 0.5; $S2 = 0.5; }
  else if ($winner === $p1) { $S1 = 1.0; $S2 = 0.0; }
  else if ($winner === $p2) { $S1 = 0.0; $S2 = 1.0; }
  else { throw new RuntimeException('winner_id invalide'); }

  $E1 = 1.0 / (1.0 + pow(10.0, ($r2 - $r1)/400.0));
  $E2 = 1.0 / (1.0 + pow(10.0, ($r1 - $r2)/400.0));

  $new1 = max((int)round($r1 + $K1 * ($S1 - $E1)), 100);
  $new2 = max((int)round($r2 + $K2 * ($S2 - $E2)), 100);

  $upR = $pdo->prepare("UPDATE user_ratings SET rating=? WHERE user_id=?");
  $upR->execute([$new1, $p1]);
  $upR->execute([$new2, $p2]);

  $pdo->commit();
  out([
    'ok'=>true, 'elo_updated'=>true,
    'p1'=>['id'=>$p1,'old'=>$r1,'new'=>$new1],
    'p2'=>['id'=>$p2,'old'=>$r2,'new'=>$new2]
  ]);
} catch (Throwable $e) {
  if ($pdo->inTransaction()) $pdo->rollBack();
  out(['error'=>$e->getMessage()], 400);
}
