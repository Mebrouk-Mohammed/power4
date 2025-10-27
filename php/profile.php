<?php
session_start();
if (empty($_SESSION['user_id'])) { header("Location: login.php"); exit; }

$pdo = new PDO('mysql:host=127.0.0.1;dbname=connect4;charset=utf8mb4','root','',[PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]);
$userId = (int)$_SESSION['user_id'];
$flash = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST' && ($_POST['__action'] ?? '') === 'pick_avatar') {
  $picked = trim($_POST['picked'] ?? '');
  $galleryDir = realpath(__DIR__ . '/assets/avatars');
  $realPicked = $picked !== '' ? realpath(__DIR__ . '/' . $picked) : false;
  if ($picked !== '' && $galleryDir && $realPicked && str_starts_with($realPicked, $galleryDir)) {
    $pdo->prepare("UPDATE profiles SET avatar_url=? WHERE user_id=?")->execute([$picked, $userId]);
    $flash = "✅ Avatar mis à jour.";
  } else { $flash = "❌ Avatar invalide."; }
}

if ($_SERVER['REQUEST_METHOD'] === 'POST' && ($_POST['__action'] ?? '') === 'infos') {
  $display = trim($_POST['display_name'] ?? '');
  $bio     = trim($_POST['bio'] ?? '');
  if (mb_strlen($display) > 50) $display = mb_substr($display,0,50);
  if (mb_strlen($bio) > 280)    $bio     = mb_substr($bio,0,280);
  $pdo->prepare("UPDATE profiles SET display_name=?, bio=? WHERE user_id=?")->execute([$display !== '' ? $display : null, $bio !== '' ? $bio : null, $userId]);
  $flash = "✅ Profil mis à jour.";
}

$stmt = $pdo->prepare("
  SELECT u.email, p.username, p.display_name, p.rating_elo, p.last_elo_change,
         p.games_played, p.wins, p.losses, p.draws, p.avatar_url, p.bio
  FROM users u LEFT JOIN profiles p ON p.user_id=u.id WHERE u.id=?");
$stmt->execute([$userId]);
$p = $stmt->fetch(PDO::FETCH_ASSOC);

$avatar   = $p['avatar_url'] ?: 'https://via.placeholder.com/160?text=Avatar';
$elo      = (int)($p['rating_elo'] ?? 1000);
$deltaElo = (int)($p['last_elo_change'] ?? 0);

function rankMeta(int $elo): array {
  $tiers = [
    ['name'=>'Bronze',      'min'=>800,  'max'=>1099, 'color'=>'#b08d57'],
    ['name'=>'Silver',      'min'=>1100, 'max'=>1299, 'color'=>'#c0c8d0'],
    ['name'=>'Gold',        'min'=>1300, 'max'=>1499, 'color'=>'#f2c94c'],
    ['name'=>'Platinum',    'min'=>1500, 'max'=>1699, 'color'=>'#5bd1c9'],
    ['name'=>'Diamond',     'min'=>1700, 'max'=>1899, 'color'=>'#6ea8ff'],
    ['name'=>'Master',      'min'=>1900, 'max'=>2099, 'color'=>'#c37dff'],
    ['name'=>'Grandmaster', 'min'=>2100, 'max'=>9999, 'color'=>'#ff6b6b'],
  ];
  foreach ($tiers as $i => $t) if ($elo >= $t['min'] && $elo <= $t['max']) {
    $next = $i < count($tiers)-1 ? $tiers[$i+1]['min'] : null;
    return ['name'=>$t['name'],'color'=>$t['color'],'min'=>$t['min'],'max'=>$t['max'],'next'=>$next];
  }
  return ['name'=>'Bronze','color'=>'#b08d57','min'=>800,'max'=>1099,'next'=>1100];
}
$rank = rankMeta($elo);
$progressPct = 100;
if (!is_null($rank['next'])) {
  $progressPct = max(0, min(100, (int)round(($elo - $rank['min'])/max(1, $rank['next'] - $rank['min']) * 100)));
}

$gallery = glob(__DIR__ . '/assets/avatars/*.{png,jpg,jpeg,webp}', GLOB_BRACE);
sort($gallery);
$galleryRel = array_map(fn($abs) => 'assets/avatars/' . basename($abs), $gallery);
?>
<!doctype html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <title>Mon profil — Puissance 4</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    :root{
      --bg:#0f1115; --text:#e8e8e8; --panel:#161a22; --border:#232839;
      --accent:#ffa94d; --hover:#2a2f3a;
    }
    *{box-sizing:border-box}
    body{background:var(--bg);color:var(--text);font-family:system-ui,Arial;margin:0}
    .wrap{max-width:1000px;margin:40px auto 40px;padding:0 20px}

    /* Bouton discret (surbrillance légère) */
    .btn-ghost{
      display:inline-flex; align-items:center; gap:8px;
      background:#161922; color:var(--text);
      border:1px solid #23283a; border-radius:10px;
      padding:8px 12px; text-decoration:none; transition:.18s ease;
    }
    .btn-ghost:hover, .btn-ghost:focus-visible{
      background:var(--hover); outline:none;
      box-shadow:0 0 0 2px #ffffff22, 0 8px 20px #0006;
      transform:translateY(-1px);
    }

    .card{background:var(--panel);border:1px solid var(--border);border-radius:14px;padding:20px;display:flex;gap:20px;align-items:flex-start}
    .avatar{width:160px;height:160px;border-radius:14px;object-fit:cover;border:1px solid #2a3042}
    .grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px}
    .stat{background:#0f1320;border:1px solid var(--border);border-radius:10px;padding:12px}
    .ok{color:#7bd88f}.err{color:#ff6b6b}
    input,textarea,button{background:#232839;color:var(--text);border:1px solid #2f3650;border-radius:8px;padding:8px 12px;width:100%}
    textarea{min-height:90px;resize:vertical}
    form{margin-top:10px}
    .col{flex:1}
    .badge{display:inline-block;margin-left:8px;padding:4px 10px;border-radius:999px;font-size:12px;border:1px solid transparent}
    .bar{height:8px;background:#232839;border:1px solid #2f3650;border-radius:999px;overflow:hidden;margin-top:6px}
    .bar > span{display:block;height:100%}
    .gal-wrap{margin-top:16px}
    .gal{display:grid;grid-template-columns:repeat(6, 1fr);gap:10px}
    .gal button{background:transparent;border:0;padding:0;cursor:pointer}
    .gal img{width:64px;height:64px;border-radius:50%;border:2px solid transparent;transition:.2s}
    .gal img:hover{transform:scale(1.06);border-color:var(--accent)}
    .gal img.sel{border-color:var(--accent)}
    .choose-title{margin:14px 0 6px 0;font-weight:600}
    h1{margin:0 0 14px 0;color:var(--accent)}
  </style>
</head>
<body>
  <div class="wrap">
    <div style="margin-bottom:12px">
      <a class="btn-ghost" href="index.php">⬅️ Retour au menu</a>
    </div>

    <h1>Mon profil</h1>

    <?php if ($flash): ?>
      <p class="<?= str_starts_with($flash,'✅') ? 'ok' : 'err' ?>"><?= htmlspecialchars($flash) ?></p>
    <?php endif; ?>

    <div class="card">
      <div>
        <img src="<?= htmlspecialchars($avatar) ?>" alt="Avatar" class="avatar">

        <div class="gal-wrap">
          <div class="choose-title">Choisir un avatar</div>
          <div class="gal">
            <?php foreach ($galleryRel as $rel): ?>
              <form method="post" style="display:inline">
                <input type="hidden" name="__action" value="pick_avatar">
                <input type="hidden" name="picked" value="<?= htmlspecialchars($rel) ?>">
                <button type="submit" title="Choisir cet avatar">
                  <img src="<?= htmlspecialchars($rel) ?>" alt="avatar" class="<?= ($avatar === $rel) ? 'sel' : '' ?>">
                </button>
              </form>
            <?php endforeach; ?>
          </div>
        </div>
      </div>

      <div class="col">
        <h2 style="margin-top:0;"><?= htmlspecialchars($p['username'] ?: '(sans pseudo)') ?></h2>
        <p>Email : <strong><?= htmlspecialchars($p['email']) ?></strong></p>

        <div class="grid" style="margin:12px 0">
          <div class="stat">
            <strong>ELO</strong><br>
            <span style="font-size:22px"><?= $elo ?></span>
            <span class="badge"
                  style="border-color:<?= htmlspecialchars($rank['color']) ?>; color:<?= htmlspecialchars($rank['color']) ?>; background:<?= htmlspecialchars($rank['color']) ?>20;">
              <?= htmlspecialchars($rank['name']) ?>
            </span>
            <?php if(!is_null($rank['next'])): ?>
              <div style="margin-top:8px">Palier suivant à <strong><?= $rank['next'] ?></strong></div>
              <div class="bar"><span style="width:<?= $progressPct ?>%; background:<?= htmlspecialchars($rank['color']) ?>;"></span></div>
              <div style="font-size:12px;opacity:.7;margin-top:4px;"><?= $progressPct ?>% vers le palier suivant</div>
            <?php endif; ?>
            <?php if($deltaElo !== 0): ?>
              <div style="margin-top:6px;font-size:14px;color:<?= $deltaElo > 0 ? '#7bd88f' : '#ff6b6b' ?>;">
                <?= $deltaElo > 0 ? '🟢 +' : '🔴 ' ?><?= $deltaElo ?> ELO
              </div>
            <?php endif; ?>
          </div>

          <div class="stat"><strong>Parties</strong><br><?= (int)$p['games_played'] ?></div>
          <div class="stat"><strong>Victoires</strong><br><?= (int)$p['wins'] ?></div>
          <div class="stat"><strong>Défaites</strong><br><?= (int)$p['losses'] ?></div>
          <div class="stat"><strong>Nuls</strong><br><?= (int)$p['draws'] ?></div>
        </div>

        <form method="post">
          <input type="hidden" name="__action" value="infos">
          <label>Nom public (50 caractères max)</label>
          <input name="display_name" maxlength="50" value="<?= htmlspecialchars($p['display_name'] ?? '') ?>">
          <div style="height:8px"></div>
          <label>Bio (280 caractères max)</label>
          <textarea name="bio" maxlength="280" placeholder="Présente-toi en quelques mots..."><?= htmlspecialchars($p['bio'] ?? '') ?></textarea>
          <div style="height:8px"></div>
          <button type="submit">Enregistrer les infos</button>
        </form>
      </div>
    </div>
  </div>
</body>
</html>
