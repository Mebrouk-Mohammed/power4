<?php
// À adapter selon la configuration de la base de données
$DB_HOST = '127.0.0.1';
$DB_NAME = 'power4';
$DB_USER = 'root';
$DB_PASS = '';
$charset = 'utf8mb4';

$dsn = "mysql:host=$DB_HOST;dbname=$DB_NAME;charset=$charset";
$options = [
  PDO::ATTR_ERRMODE            => PDO::ERRMODE_EXCEPTION,
  PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
  PDO::ATTR_EMULATE_PREPARES   => false,
];

try {
  $pdo = new PDO($dsn, $DB_USER, $DB_PASS, $options);
} catch (PDOException $e) {
  http_response_code(500);
  die('Erreur connexion BDD: '.$e->getMessage());
}
