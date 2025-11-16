<?php
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
