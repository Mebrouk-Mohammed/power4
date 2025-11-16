<?php
// api_boot.php
header('Content-Type: application/json; charset=utf-8');
header('Access-Control-Allow-Origin: *'); // à restreindre si besoin
header('Access-Control-Allow-Methods: GET,POST,OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, X-Api-Key');

if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') { http_response_code(204); exit; }

require_once __DIR__.'/db.php';

// 🔐 clé API très simple pour autoriser ton binaire Go
const API_KEY = 'cle-api-fuefijefe524895'; // change-la !

function require_api_key(): void {
  $k = $_SERVER['HTTP_X_API_KEY'] ?? '';
  if ($k !== API_KEY) {
    http_response_code(401);
    echo json_encode(['error'=>'unauthorized']); exit;
  }
}

function json_input(): array {
  $raw = file_get_contents('php://input');
  if (!$raw) return [];
  $data = json_decode($raw, true);
  return is_array($data) ? $data : [];
}
function out($data, int $code=200): void {
  http_response_code($code);
  echo json_encode($data); exit;
}
