<?php
require_once 'db.php';
session_start();
if (empty($_SESSION['user_id'])) { header('Location: login.php'); exit; }

/* ⚠️ ICI : on pointe vers php/assets/ (et plus avatars/) */
$avatarDir = __DIR__ . '/assets';
$avatarWeb = 'assets'; // chemin web

if (empty($_SESSION['csrf'])) { $_SESSION['csrf'] = bin2hex(random_bytes(16)); }
$csrf = $_SESSION['csrf'];

$allowedExt = ['png','jpg','jpeg','gif','webp'];
$files = is_dir($avatarDir) ? array_values(array_filter(scandir($avatarDir), function($f) use ($avatarDir, $allowedExt){
  if ($f[0] === '.') return false;
  $ext = strtolower(pathinfo($f, PATHINFO_EXTENSION));
  if (!in_array($ext, $allowedExt)) return false;
  return is_file($avatarDir . '/' . $f);
})) : [];

$err=''; $ok='';
if ($_SERVER['REQUEST_METHOD']==='POST') {
  if (!hash_equals($csrf, $_POST['csrf'] ?? '')) { $err = 'Action non autorisée (CSRF).'; }
  else {
    $choice = $_POST['avatar'] ?? '';
    if (!in_array($choice, $files, true)) { $err = 'Avatar invalide.'; }
    else {
      $relPath = $avatarWeb . '/' . $choice;
      $st = $pdo->prepare("UPDATE users SET avatar_url=? WHERE id=?");
      $st->execute([$relPath, $_SESSION['user_id']]);
      $ok = 'Avatar mis à jour 👍';
    }
  }
}

$st = $pdo->prepare("SELECT username, avatar_url FROM users WHERE id=?");
$st->execute([$_SESSION['user_id']]);
$user = $st->fetch();
?>
<!doctype html>
<meta charset="utf-8">
<title>Choisir un avatar</title>
<style>
  body{font-family:system-ui,Segoe UI,Roboto,Arial,sans-serif;padding:16px;}
  .grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(110px,1fr));gap:14px;max-width:920px}
  .card{border:1px solid #ddd;border-radius:10px;padding:10px;text-align:center;cursor:pointer;transition:.15s}
  .card:hover{box-shadow:0 3px 10px rgba(0,0,0,.08)}
  .card input{display:none}
  .card img{width:88px;height:88px;object-fit:cover;border-radius:8px;display:block;margin:0 auto 8px}
  .selected{outline:3px solid #7a5cff}
  .top{display:flex;align-items:center;gap:12px;margin-bottom:14px}
  .top img{width:48px;height:48px;border-radius:8px;object-fit:cover;border:1px solid #ddd}
  button{padding:10px 16px;border:0;border-radius:8px;background:#7a5cff;color:#fff;font-weight:600;cursor:pointer}
  .msg-ok{color:#0a0}
  .msg-err{color:#b00}
</style>

<h1>Choisir un avatar</h1>

<div class="top">
  <strong><?=htmlspecialchars($user['username'])?></strong>
  <?php if (!empty($user['avatar_url'])): ?>
    <img src="<?=htmlspecialchars($user['avatar_url'])?>" alt="avatar actuel">
  <?php else: ?>
    <span style="color:#666">Aucun avatar actuel</span>
  <?php endif; ?>
</div>

<?php if ($ok): ?><p class="msg-ok"><?=$ok?></p><?php endif; ?>
<?php if ($err): ?><p class="msg-err"><?=$err?></p><?php endif; ?>

<?php if (!$files): ?>
  <p class="msg-err">Aucun avatar disponible. Mets des images dans <code>php/assets/</code>.</p>
<?php else: ?>
<form method="post" id="avatarForm">
  <input type="hidden" name="csrf" value="<?=$csrf?>">
  <div class="grid" id="grid">
    <?php foreach ($files as $f): $src = $avatarWeb . '/' . rawurlencode($f); ?>
    <label class="card">
      <img src="<?=$src?>" alt="avatar">
      <input type="radio" name="avatar" value="<?=htmlspecialchars($f)?>">
      <div><?=htmlspecialchars(pathinfo($f, PATHINFO_FILENAME))?></div>
    </label>
    <?php endforeach; ?>
  </div>
  <p style="margin-top:14px"><button type="submit">Enregistrer</button>
     <a href="profile.php" style="margin-left:10px">Retour</a></p>
</form>
<script>
  const grid=document.getElementById('grid');
  grid.addEventListener('change',e=>{
    [...grid.querySelectorAll('.card')].forEach(c=>c.classList.remove('selected'));
    const card=e.target.closest('.card'); if(card) card.classList.add('selected');
  });
</script>
<?php endif; ?>
