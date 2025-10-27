<?php
// On inclut le fichier bootstrap.php qui initialise la connexion PDO, les sessions, etc.
require __DIR__ . '/includes/bootstrap.php';

// On récupère une instance de connexion à la base via la fonction pdo()
$pdo = pdo();

// On récupère les paramètres envoyés en POST : l'ID de la partie et l'ID du gagnant (ou null si égalité)
$gameId   = (int)($_POST['game_id'] ?? 0);
$winnerId = isset($_POST['winner_id']) && $_POST['winner_id'] !== '' ? (int)$_POST['winner_id'] : null;

// Si aucun ID de partie valide n’est fourni, on renvoie une erreur HTTP 400
if ($gameId <= 0) { 
  http_response_code(400); 
  exit('bad game_id'); 
}

// On prépare une requête pour récupérer la partie correspondante, seulement si elle n’est pas encore finie
$g = $pdo->prepare("SELECT id, player1_id, player2_id FROM games WHERE id=? AND status<>'finished' FOR UPDATE");

// On démarre une transaction SQL pour s’assurer que toutes les modifications soient atomiques
$pdo->beginTransaction();

// On exécute la requête avec l’ID de la partie
$g->execute([$gameId]);

// On récupère la ligne de la partie
$game = $g->fetch();

// Si la partie n’existe pas ou est déjà finie, on annule la transaction et on arrête le script
if (!$game) { 
  $pdo->rollBack(); 
  http_response_code(404); 
  exit('game not found'); 
}

// On extrait les IDs des deux joueurs
$p1 = (int)$game['player1_id'];
$p2 = (int)$game['player2_id'];

// On récupère les informations de profil (ELO et nombre de parties) pour ces deux joueurs
$st = $pdo->prepare("SELECT user_id, rating_elo, games_played FROM profiles WHERE user_id IN (?, ?) ORDER BY user_id");
$st->execute([$p1, $p2]);

// On récupère les résultats sous forme de tableau
$rows = $st->fetchAll();

// Si on ne trouve pas deux profils (un pour chaque joueur), on annule la transaction
if (count($rows) !== 2) { 
  $pdo->rollBack(); 
  exit('profiles missing'); 
}

// On crée une table associative (clé = user_id) pour accéder rapidement aux ELO et parties
$map = [];
foreach ($rows as $r) 
  $map[(int)$r['user_id']] = ['elo'=>(int)$r['rating_elo'], 'gp'=>(int)$r['games_played']];

// On attribue les ELO et le nombre de parties de chaque joueur
$Ra = $map[$p1]['elo']; // ELO du joueur 1
$Rb = $map[$p2]['elo']; // ELO du joueur 2
$gpa = $map[$p1]['gp']; // Parties jouées par le joueur 1
$gpb = $map[$p2]['gp']; // Parties jouées par le joueur 2

// Détermination du coefficient K selon l’expérience (plus un joueur a de parties, moins K est élevé)
$Ka = $gpa < 10 ? 40 : ($gpa < 30 ? 24 : 16);
$Kb = $gpb < 10 ? 40 : ($gpb < 30 ? 24 : 16);

// Fonction interne : calcule la probabilité de victoire attendue selon la formule ELO
function expected($Rself, $Ropp){ 
  return 1.0 / (1.0 + pow(10.0, ($Ropp - $Rself)/400.0)); 
}

// Calcul des attentes de victoire (probabilités) pour chaque joueur
$Ea = expected($Ra, $Rb); // chance de victoire de A contre B
$Eb = expected($Rb, $Ra); // chance de victoire de B contre A

// Attribution des scores réels selon le résultat du match
if (is_null($winnerId)) { // cas d’égalité
  $Sa = 0.5; $Sb = 0.5; 
  $winA = 0; $winB = 0; 
  $drawA = 1; $drawB = 1;
}
elseif ($winnerId === $p1) { // joueur 1 gagne
  $Sa = 1.0; $Sb = 0.0; 
  $winA = 1; $winB = 0; 
  $drawA = 0; $drawB = 0;
}
else { // joueur 2 gagne
  $Sa = 0.0; $Sb = 1.0; 
  $winA = 0; $winB = 1; 
  $drawA = 0; $drawB = 0;
}

// Calcul du nouveau ELO pour les deux joueurs avec la formule : new = old + K × (score réel − score attendu)
$newRa = round($Ra + $Ka * ($Sa - $Ea));
$newRb = round($Rb + $Kb * ($Sb - $Eb));

// On empêche le score de descendre en dessous de 800 pour éviter les ELO trop faibles
$newRa = max(800, $newRa);
$newRb = max(800, $newRb);

// Préparation de la requête SQL pour mettre à jour le profil des joueurs
$upd = $pdo->prepare("
  UPDATE profiles
  SET rating_elo = ?, 
      games_played = games_played + 1, 
      wins = wins + ?, 
      losses = losses + ?, 
      draws = draws + ?
  WHERE user_id = ?
");

// Mise à jour du joueur 1
$upd->execute([
  $newRa,                  // nouveau ELO
  $winA,                   // +1 si victoire
  ($Sa===0.0?1:0),         // +1 si défaite
  $drawA,                  // +1 si match nul
  $p1                      // ID du joueur
]);

// Mise à jour du joueur 2
$upd->execute([
  $newRb,
  $winB,
  ($Sb===0.0?1:0),
  $drawB,
  $p2
]);

// On clôture la partie dans la table "games" : on marque comme terminée et on indique le gagnant
$close = $pdo->prepare("UPDATE games SET status='finished', winner_user_id=?, finished_at=NOW() WHERE id=?");
$close->execute([$winnerId, $gameId]);

// Validation finale de toutes les modifications (commit de la transaction)
$pdo->commit();

// On renvoie un objet JSON avec les anciens et nouveaux ELO pour les deux joueurs
echo json_encode([
  'ok' => true,
  'p1' => ['id'=>$p1, 'old'=>$Ra, 'new'=>$newRa],
  'p2' => ['id'=>$p2, 'old'=>$Rb, 'new'=>$newRb]
]);
