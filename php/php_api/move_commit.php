<?php
require_once __DIR__.'/../api_boot.php';
require_api_key();

/*
JSON attendu (côté Go) :
{
  "game_id": 1,
  "player_id": 123,
  "column_index": 3,
  "row_index": 2,           // calculé par Go
  "next_player_id": 456,    // qui doit jouer APRÈS ce coup
  "finish": {               // optionnel : si la partie se termine sur CE coup
    "status": "finished",   // "finished" | "abandoned"
    "winner_id": 123        // ou null pour un nul
  }
}
*/

$in = json_input();
$gid  = (int)($in['game_id'] ?? 0);
$pid  = (int)($in['player_id'] ?? 0);
$col  = (int)($in['column_index'] ?? -1);
$row  = (int)($in['row_index'] ?? -1);
$next = isset($in['next_player_id']) ? (int)$in['next_player_id'] : null;
$fin  = $in['finish'] ?? null;

if ($gid<=0 || $pid<=0 || $col<0 || $row<0) out(['error'=>'missing/invalid fields'],400);

try {
  $pdo->beginTransaction();

  // Verrouille la partie
  $st = $pdo->prepare("SELECT id,status,player_to_move,rows_count,cols_count,player1_id,player2_id
                       FROM games WHERE id=? FOR UPDATE");
  $st->execute([$gid]);
  $g = $st->fetch();
  if (!$g) throw new RuntimeException('game not found');
  if ($g['status']!=='active') throw new RuntimeException('game not active');

  // Cohérence minimale (sécurité côté BDD, mais la logique reste en Go)
  if ((int)$g['player_to_move'] !== $pid) throw new RuntimeException('turn mismatch');
  if ($col >= (int)$g['cols_count']) throw new RuntimeException('bad column');
  if ($row >= (int)$g['rows_count']) throw new RuntimeException('bad row');

  // Vérifie que row_index correspond au nombre de pions déjà dans la colonne
  $c = $pdo->prepare("SELECT COUNT(*) FROM moves WHERE game_id=? AND column_index=?");
  $c->execute([$gid, $col]);
  $expectedRow = (int)$c->fetchColumn();
  if ($row !== $expectedRow) throw new RuntimeException('row_index mismatch');

  // Prochain numéro de coup
  $mno = $pdo->prepare("SELECT COALESCE(MAX(move_no),0)+1 FROM moves WHERE game_id=?");
  $mno->execute([$gid]);
  $moveNo = (int)$mno->fetchColumn();

  // Couleur pour l'historique (optionnel)
  $disc = ($pid === (int)$g['player1_id']) ? 'R' : 'Y';

  // Insère le coup
  $ins = $pdo->prepare("INSERT INTO moves (game_id, move_no, player_id, column_index, row_index, disc_color)
                        VALUES (?,?,?,?,?,?)");
  $ins->execute([$gid, $moveNo, $pid, $col, $row, $disc]);

  // Si fin de partie fournie par Go → cloture + Elo
  if (is_array($fin) && ($fin['status'] ?? '') === 'finished') {
    $winner = array_key_exists('winner_id',$fin) && $fin['winner_id']!==null ? (int)$fin['winner_id'] : null;

    // Met à jour la partie
    $upd = $pdo->prepare("UPDATE games SET status='finished', finished_at=NOW(), winner_id=?, player_to_move=NULL WHERE id=?");
    $upd->execute([$winner, $gid]);

    // Elo (K variable) — même logique que ton finish_game amélioré
    $p1 = (int)$g['player1_id']; $p2 = (int)$g['player2_id'];
    if ($p2) {
      // ratings actuels
      $getR = $pdo->prepare("SELECT user_id, rating FROM user_ratings WHERE user_id IN (?,?)");
      $getR->execute([$p1,$p2]); $r1=1200; $r2=1200;
      foreach ($getR as $rowR) {
        if ((int)$rowR['user_id']===$p1) $r1=(int)$rowR['rating'];
        if ((int)$rowR['user_id']===$p2) $r2=(int)$rowR['rating'];
      }
      $pdo->prepare("INSERT IGNORE INTO user_ratings (user_id, rating) VALUES (?,1200)")->execute([$p1]);
      $pdo->prepare("INSERT IGNORE INTO user_ratings (user_id, rating) VALUES (?,1200)")->execute([$p2]);

      // parties jouées (avant ce coup)
      $cnt = $pdo->prepare("SELECT COUNT(*) AS c FROM games WHERE status='finished' AND (player1_id=? OR player2_id=?)");
      $cnt->execute([$p1,$p1]); $c1 = (int)$cnt->fetch()['c'];
      $cnt->execute([$p2,$p2]); $c2 = (int)$cnt->fetch()['c'];

      $K1 = ($c1 < 30) ? 40 : (($r1 < 2000) ? 20 : 10);
      $K2 = ($c2 < 30) ? 40 : (($r2 < 2000) ? 20 : 10);

      // scores
      if ($winner === null) { $S1=0.5; $S2=0.5; }
      else if ($winner === $p1) { $S1=1.0; $S2=0.0; }
      else if ($winner === $p2) { $S1=0.0; $S2=1.0; }
      else throw new RuntimeException('winner_id invalid');

      // espérance
      $E1 = 1.0 / (1.0 + pow(10.0, ($r2 - $r1)/400.0));
      $E2 = 1.0 / (1.0 + pow(10.0, ($r1 - $r2)/400.0));

      // nouveaux ratings
      $new1 = max((int)round($r1 + $K1 * ($S1 - $E1)), 100);
      $new2 = max((int)round($r2 + $K2 * ($S2 - $E2)), 100);

      $pdo->prepare("UPDATE user_ratings SET rating=? WHERE user_id=?")->execute([$new1,$p1]);
      $pdo->prepare("UPDATE user_ratings SET rating=? WHERE user_id=?")->execute([$new2,$p2]);

      $pdo->commit();
      out(['ok'=>true,'move_no'=>$moveNo,'row_index'=>$row,'disc'=>$disc,'finished'=>true,
           'elo'=>['p1'=>['id'=>$p1,'old'=>$r1,'new'=>$new1],'p2'=>['id'=>$p2,'old'=>$r2,'new'=>$new2]]]);
    }
  }

  // Sinon, partie continue → on met à jour le trait avec la valeur envoyée par Go
  if ($next === null) throw new RuntimeException('next_player_id required when not finished');
  $pdo->prepare("UPDATE games SET player_to_move=? WHERE id=?")->execute([$next,$gid]);

  $pdo->commit();
  out(['ok'=>true,'move_no'=>$moveNo,'row_index'=>$row,'disc'=>$disc,'finished'=>false]);
} catch (Throwable $e) {
  if ($pdo->inTransaction()) $pdo->rollBack();
  out(['error'=>$e->getMessage()],400);
}
