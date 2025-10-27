<?php
// ======================================================
// 1️⃣ DÉMARRAGE DE LA SESSION ET CONNEXION À LA BASE
// ======================================================
session_start();

if (empty($_SESSION['user_id'])) {
    header('Location: login.php');
    exit;
}

try {
    $pdo = new PDO(
        'mysql:host=127.0.0.1;dbname=connect4;charset=utf8mb4',
        'root',
        '',
        [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
    );
} catch (PDOException $e) {
    http_response_code(500);
    exit('Erreur de connexion à la base.');
}

// ======================================================
// 2️⃣ SI L’UTILISATEUR CLIQUE SUR UN AVATAR
// ======================================================
if (isset($_POST['avatar'])) {
    $avatar = $_POST['avatar'];
    $user_id = (int)$_SESSION['user_id'];

    // Met à jour l’avatar choisi dans la table "profiles"
    $stmt = $pdo->prepare("UPDATE profiles SET avatar = ? WHERE user_id = ?");
    $stmt->execute([$avatar, $user_id]);

    header('Location: profile.php?avatar_updated=1');
    exit;
}

// ======================================================
// 3️⃣ LISTE DES IMAGES DANS LE DOSSIER assets/avatars
// ======================================================
$avatars = glob('assets/avatars/*.png'); // Récupère tous les fichiers PNG
?>
<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<title>Choisir un avatar — Puissance 4</title>
<style>
  body {
    background: #0f1115;
    color: #eaeaea;
    font-family: system-ui, Arial;
    margin: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
  }
  .card {
    background: #161a22;
    border: 1px solid #232839;
    border-radius: 14px;
    padding: 30px;
    width: 420px;
    box-shadow: 0 0 25px #000a;
    text-align: center;
  }
  h1 {
    color: #ffa94d;
    margin-bottom: 20px;
  }
  .avatars {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 15px;
    justify-items: center;
  }
  .avatars img {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    border: 2px solid transparent;
    cursor: pointer;
    transition: .2s;
  }
  .avatars img:hover {
    transform: scale(1.1);
    border-color: #ffa94d;
  }
  button {
    margin-top: 20px;
    padding: 10px 20px;
    border-radius: 8px;
    background: #ffa94d;
    border: none;
    color: #111;
    cursor: pointer;
    font-weight: bold;
  }
</style>
</head>
<body>

<div class="card">
  <h1>Choisis ton avatar 😎</h1>
  <form method="post">
    <div class="avatars">
      <?php foreach ($avatars as $a): ?>
        <label>
          <input type="radio" name="avatar" value="<?= htmlspecialchars($a) ?>" required hidden>
          <img src="<?= htmlspecialchars($a) ?>" onclick="selectAvatar(this)">
        </label>
      <?php endforeach; ?>
    </div>
    <button type="submit">✅ Valider mon choix</button>
  </form>
</div>

<script>
function selectAvatar(img) {
  // Réinitialise tous les cadres
  document.querySelectorAll('.avatars img').forEach(i => i.style.borderColor = 'transparent');
  // Sélection visuelle
  img.style.borderColor = '#ffa94d';
  // Active le radio caché correspondant
  img.previousElementSibling.checked = true;
}
</script>

</body>
</html>
