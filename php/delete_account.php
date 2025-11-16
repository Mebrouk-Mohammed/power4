<?php
<<<<<<< HEAD
require_once 'db.php';
session_start();
if (empty($_SESSION['user_id'])) { header('Location: login.php'); exit; }

$userId = (int)$_SESSION['user_id'];
if (empty($_SESSION['csrf'])) { $_SESSION['csrf'] = bin2hex(random_bytes(16)); }
$csrf = $_SESSION['csrf'];

$info = $pdo->prepare("SELECT id, username, email FROM users WHERE id=?");
$info->execute([$userId]); $user = $info->fetch();
if (!$user) { session_destroy(); header('Location: login.php'); exit; }

$msg=''; $err='';

if ($_SERVER['REQUEST_METHOD']==='POST') {
  if (!hash_equals($csrf, $_POST['csrf'] ?? '')) { $err='CSRF invalide.'; }
  else {
    $action = $_POST['action'] ?? 'delete';
    if ($action === 'delete') {
      try {
        $st = $pdo->prepare("DELETE FROM users WHERE id=?");
        $st->execute([$userId]);
        session_destroy();
        header('Location: register.php'); exit;
      } catch (Throwable $e) {
        // souvent erreur FK (player1_id ON DELETE RESTRICT)
        $err = "Impossible de supprimer le compte car il est lié à des parties. Tu peux l'anonymiser ci-dessous.";
      }
    } elseif ($action === 'anonymize') {
      $pdo->beginTransaction();
      try {
        $newName = 'deleted_user_'.$userId;
        $st = $pdo->prepare("UPDATE users SET username=?, email=NULL, avatar_url=NULL, is_admin=0 WHERE id=?");
        $st->execute([$newName, $userId]);
        $pdo->commit();
        session_destroy();
        header('Location: login.php'); exit;
      } catch (Throwable $e) {
        $pdo->rollBack();
        $err = 'Erreur lors de l’anonymisation: '.$e->getMessage();
      }
    }
  }
}
?>
<!doctype html><meta charset="utf-8">
<h1>Supprimer mon compte</h1>
<p>Compte: <strong><?=htmlspecialchars($user['username'])?></strong></p>

<?php if ($msg): ?><p style="color:#080"><?=$msg?></p><?php endif; ?>
<?php if ($err): ?><p style="color:#b00"><?=$err?></p><?php endif; ?>

<form method="post" onsubmit="return confirm('Supprimer définitivement ton compte ? Cette action est irréversible.');">
  <input type="hidden" name="csrf" value="<?=$csrf?>">
  <input type="hidden" name="action" value="delete">
  <button style="color:#fff;background:#b00;padding:8px 12px;border:0;border-radius:6px;">Supprimer définitivement</button>
</form>

<h2>Alternative : anonymiser</h2>
<p>Si tu as déjà joué des parties, la suppression peut être bloquée. Tu peux anonymiser ton compte (pseudo remplacé, email/avatars supprimés) :</p>
<form method="post" onsubmit="return confirm('Anonymiser ton compte ?');">
  <input type="hidden" name="csrf" value="<?=$csrf?>">
  <input type="hidden" name="action" value="anonymize">
  <button>Anonymiser mon compte</button>
</form>

<p style="margin-top:16px"><a href="profile.php">Retour profil</a></p>
=======
// ======================================================
// 1️⃣ DÉMARRAGE DE LA SESSION ET VÉRIFICATION
// ======================================================
session_start();

// Si l'utilisateur n'est pas connecté, on le renvoie au login
if (empty($_SESSION['user_id'])) {
    header('Location: login.php');
    exit;
}

// ======================================================
// 2️⃣ CONNEXION À LA BASE DE DONNÉES
// ======================================================
try {
    $pdo = new PDO(
        'mysql:host=127.0.0.1;dbname=connect4;charset=utf8mb4',
        'root',
        '',
        [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
    );
} catch (PDOException $e) {
    http_response_code(500);
    exit('Erreur de connexion à la base de données.');
}

// ======================================================
// 3️⃣ SUPPRESSION APRÈS CONFIRMATION
// ======================================================

// Si le joueur a cliqué sur “Confirmer la suppression”
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['confirm_delete'])) {

    $user_id = (int)$_SESSION['user_id'];

    // 🔸 Étape 1 : supprimer le profil associé
    $pdo->prepare("DELETE FROM profiles WHERE user_id = ?")->execute([$user_id]);

    // 🔸 Étape 2 : supprimer l’utilisateur
    $pdo->prepare("DELETE FROM users WHERE id = ?")->execute([$user_id]);

    // 🔸 Étape 3 : fermer la session et rediriger
    session_destroy();
    header('Location: login.php?deleted=1');
    exit;
}
?>
<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8" />
  <title>Suppression du compte — Puissance 4</title>
  <style>
    /* ===== STYLE GLOBAL ===== */
    body {
      background: #0f1115;
      color: #eaeaea;
      font-family: "Poppins", system-ui, Arial, sans-serif;
      margin: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      height: 100vh;
    }

    /* ===== CARTE CENTRALE ===== */
    .card {
      background: #161a22;
      border: 1px solid #232839;
      border-radius: 14px;
      padding: 32px;
      width: 360px;
      text-align: center;
      box-shadow: 0 0 25px #000a;
      animation: fadeIn 0.4s ease forwards;
    }

    h1 {
      color: #ff6b6b;
      margin-bottom: 12px;
    }

    p {
      opacity: 0.9;
      line-height: 1.4;
    }

    /* ===== BOUTONS ===== */
    .buttons {
      margin-top: 24px;
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    button {
      padding: 12px;
      border-radius: 8px;
      border: none;
      cursor: pointer;
      font-size: 15px;
      transition: 0.2s;
    }

    .confirm {
      background: #ff6b6b;
      color: #111;
    }

    .confirm:hover {
      background: #ff8787;
      transform: scale(1.03);
    }

    .cancel {
      background: #232839;
      border: 1px solid #2e344a;
      color: #eaeaea;
    }

    .cancel:hover {
      background: #ffa94d;
      color: #111;
      transform: scale(1.03);
    }

    @keyframes fadeIn {
      from { opacity: 0; transform: translateY(10px); }
      to { opacity: 1; transform: translateY(0); }
    }
  </style>
</head>
<body>

  <div class="card">
    <h1>⚠️ Suppression du compte</h1>
    <p>Cette action est <strong>irréversible</strong>.<br>
    Ton profil et toutes tes données seront supprimés définitivement.</p>

    <form method="POST" class="buttons">
      <!-- Bouton rouge : confirmer la suppression -->
      <button type="submit" name="confirm_delete" class="confirm">🗑️ Supprimer mon compte</button>
      <!-- Bouton gris/orangé : retour au menu -->
      <button type="button" class="cancel" onclick="window.location.href='index.php'">⬅️ Annuler</button>
    </form>
  </div>

</body>
</html>
>>>>>>> bda513fd2bb1669761e3605ed8a5539f7056ce17
