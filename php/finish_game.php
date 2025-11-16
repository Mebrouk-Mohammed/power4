<?php
require_once 'db.php';
session_start();

/* Appelé par bouton ou par Go/serveur quand la partie se termine
   POST:
     - game_id (int, obligatoire)
     - status: 'finished' | 'abandoned' (par défaut 'finished')
     - winner_id: int | '' (vide => match nul)
*/
$gid    = (int)($_POST['game_id'] ?? 0);
$status = $_POST['status'] ?? 'finished';
$winner = isset($_POST['winner_id']) && $_POST['winner_id'] !== '' ? (int)$_POST['winner_id'] : null;

if ($gid <= 0) { http_response_code(400); die('game_id requis'); }

try {
  $pdo->beginTransaction();

  // Verrouille la partie
  $g = $pdo->prepare("SELECT player1_id, player2_id, status FROM games WHERE id=? FOR UPDATE");
  $g->execute([$gid]);
  $game = $g->fetch();
  if (!$game) { throw new RuntimeException('Partie introuvable'); }

  $p1 = (int)$game['player1_id'];
  $p2 = (int)$game['player2_id'];

  // Met à jour l'état de la partie (winner_id peut être NULL pour un nul)
  $upd = $pdo->prepare("UPDATE games
                        SET status=?, finished_at=NOW(), winner_id=?, player_to_move=NULL
                        WHERE id=?");
  $upd->execute([$status, $winner, $gid]);

  // Pas d'Elo si partie abandonnée ou si pas d'adversaire
  if ($status !== 'finished' || !$p2) {
    $pdo->commit();
    echo "OK";
    exit;
  }

  // -------- Elo --------

  // 1) Récup ratings actuels (fallback 1200) + crée si manquants
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

  // 2) Nombre de parties classées jouées par chacun (pour K)
  $cnt = $pdo->prepare("
    SELECT
      SUM(CASE WHEN (player1_id=? OR player2_id=?) AND status='finished' THEN 1 ELSE 0 END) AS c1,
      SUM(CASE WHEN (player1_id=? OR player2_id=?) AND status='finished' THEN 1 ELSE 0 END) AS c2
    FROM games
  ");
  // petite astuce : on exécute deux requêtes pour éviter ambiguïté ; sinon on peut faire 2 SELECT séparés
  $cnt1 = $pdo->prepare("SELECT COUNT(*) AS c FROM games WHERE status='finished' AND (player1_id=? OR player2_id=?)");
  $cnt2 = $pdo->prepare("SELECT COUNT(*) AS c FROM games WHERE status='finished' AND (player1_id=? OR player2_id=?)");
  $cnt1->execute([$p1, $p1]); $c1 = (int)$cnt1->fetch()['c'];
  $cnt2->execute([$p2, $p2]); $c2 = (int)$cnt2->fetch()['c'];

  // 3) K-factor individuel (progression logique)
  $K1 = ($c1 < 30) ? 40 : (($r1 < 2000) ? 20 : 10);
  $K2 = ($c2 < 30) ? 40 : (($r2 < 2000) ? 20 : 10);

  // 4) Scores S1/S2 (nul si winner_id absent)
  if ($winner === null) { $S1 = 0.5; $S2 = 0.5; }
  else if ($winner === $p1) { $S1 = 1.0; $S2 = 0.0; }
  else if ($winner === $p2) { $S1 = 0.0; $S2 = 1.0; }
  else { throw new RuntimeException('winner_id invalide'); }

  // 5) Espérances (logistique)
  //    Si tu bats un joueur plus fort, E est faible => gros gain.
  //    Si tu bats un joueur plus faible, E est grand => petit gain.
  $E1 = 1.0 / (1.0 + pow(10.0, ($r2 - $r1) / 400.0));
  $E2 = 1.0 / (1.0 + pow(10.0, ($r1 - $r2) / 400.0));

  // 6) Nouveaux ratings (arrondis)
  $new1 = (int)round($r1 + $K1 * ($S1 - $E1));
  $new2 = (int)round($r2 + $K2 * ($S2 - $E2));

  // (optionnel) clamp minimum à 100 si tu veux éviter de descendre trop bas
  $new1 = max($new1, 100);
  $new2 = max($new2, 100);

  // 7) Persiste
  $upR = $pdo->prepare("UPDATE user_ratings SET rating=? WHERE user_id=?");
  $upR->execute([$new1, $p1]);
  $upR->execute([$new2, $p2]);

  $pdo->commit();
  echo "OK";
} catch (Throwable $e) {
  if ($pdo->inTransaction()) $pdo->rollBack();
  http_response_code(400);
  echo "Erreur: ".$e->getMessage();
}
