<?php
require_once __DIR__ . '/api_boot.php';
require_api_key();
session_start();

if (empty($_SESSION['user_id'])) {
  http_response_code(401);
  echo json_encode(['error' => 'not_logged_in']);
  exit;
}

$out = [
  'ok' => true,
  'user_id' => (int)$_SESSION['user_id'],
  'username' => $_SESSION['username'] ?? null
];

echo json_encode($out);
