<?php
require_once __DIR__.'/db.php';
$row = $pdo->query("SELECT COUNT(*) AS n FROM users")->fetch();
echo "Connexion OK ✅ — users: ".$row['n'];
