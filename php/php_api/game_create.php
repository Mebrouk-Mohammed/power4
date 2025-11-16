<?php
require_once __DIR__.'/../api_boot.php';
require_api_key();

$in = json_input(); // { "player1_id": 1, "player2_id": 2 }
$p1 = (int)($in['player1_id'] ?? 0);
$p2 = isset($in['player2_id']) && $in['player2_id']!=='' ? (int)$in['player2_id'] : null;
if ($p1 <= 0) out(['error'=>'player1_id required'], 400);

try {
  $st = $pdo->prepare("INSERT INTO games (status, player1_id, player2_id, player_to_move)
                       VALUES ('active', ?, ?, ?)");
  $st->execute([$p1, $p2, $p1]);
  out(['ok'=>true, 'game_id'=>(int)$pdo->lastInsertId()]);
} catch (Throwable $e) {
  out(['error'=>$e->getMessage()], 400);
}
