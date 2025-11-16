<?php
require_once __DIR__.'/../api_boot.php';
require_api_key();

$in = json_input(); // { "game_id":1, "player_id":1, "column_index":3 }
$gid = (int)($in['game_id'] ?? 0);
$pid = (int)($in['player_id'] ?? 0);
$col = (int)($in['column_index'] ?? -1);
if ($gid<=0 || $pid<=0) out(['error'=>'missing ids'],400);

try {
  $pdo->beginTransaction();

  $st = $pdo->prepare("SELECT id,status,player_to_move,rows_count,cols_count,player1_id,player2_id
                       FROM games WHERE id=? FOR UPDATE");
  $st->execute([$gid]); $g = $st->fetch();
  if (!$g || $g['status']!=='active') throw new RuntimeException('invalid game');
  if ((int)$g['player_to_move'] !== $pid) throw new RuntimeException('not your turn');
  if ($col < 0 || $col >= (int)$g['cols_count']) throw new RuntimeException('bad column');

  $c = $pdo->prepare("SELECT COUNT(*) FROM moves WHERE game_id=? AND column_index=?");
  $c->execute([$gid, $col]);
  $rowIndex = (int)$c->fetchColumn();
  if ($rowIndex >= (int)$g['rows_count']) throw new RuntimeException('column full');

  $mno = $pdo->prepare("SELECT COALESCE(MAX(move_no),0)+1 FROM moves WHERE game_id=?");
  $mno->execute([$gid]); $moveNo = (int)$mno->fetchColumn();

  $disc = ($pid === (int)$g['player1_id']) ? 'R' : 'Y';
  $next = ($pid === (int)$g['player1_id']) ? (int)$g['player2_id'] : (int)$g['player1_id'];

  $ins = $pdo->prepare("INSERT INTO moves (game_id, move_no, player_id, column_index, row_index, disc_color)
                        VALUES (?,?,?,?,?,?)");
  $ins->execute([$gid, $moveNo, $pid, $col, $rowIndex, $disc]);

  // on passe juste le tour ici (la victoire est décidée côté Go)
  $up = $pdo->prepare("UPDATE games SET player_to_move=? WHERE id=?");
  $up->execute([$next, $gid]);

  $pdo->commit();
  out(['ok'=>true, 'move_no'=>$moveNo, 'row_index'=>$rowIndex, 'disc'=>$disc]);
} catch (Throwable $e) {
  if ($pdo->inTransaction()) $pdo->rollBack();
  out(['error'=>$e->getMessage()], 400);
}
