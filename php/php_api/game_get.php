<?php
require_once __DIR__.'/../api_boot.php';
require_api_key();

$gid = (int)($_GET['id'] ?? 0);
if ($gid<=0) out(['error'=>'id required'],400);

$g = $pdo->prepare("SELECT * FROM games WHERE id=?");
$g->execute([$gid]); $game = $g->fetch();
if (!$game) out(['error'=>'not found'],404);

$m = $pdo->prepare("SELECT move_no,player_id,column_index,row_index,disc_color,played_at
                    FROM moves WHERE game_id=? ORDER BY move_no");
$m->execute([$gid]); $moves = $m->fetchAll();

out(['game'=>$game,'moves'=>$moves]);
