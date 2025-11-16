<?php
require_once 'db.php';
session_start();

$error = '';
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = trim($_POST['username'] ?? '');
    $password = $_POST['password'] ?? '';

    if ($username === '' || $password === '') {
        $error = 'Pseudo et mot de passe requis.';
    } else {
        $st = $pdo->prepare("SELECT id, password_hash FROM users WHERE username = ?");
        $st->execute([$username]);
        $u = $st->fetch();

        if (!$u || !password_verify($password, $u['password_hash'])) {
            $error = 'Identifiants invalides.';
        } else {
            $_SESSION['user_id'] = (int)$u['id'];
            header('Location: index.php'); // page d’accueil minimale juste après
            exit;
        }
    }
}
?>
<!doctype html>
<meta charset="utf-8">
<h1>Connexion</h1>
<?php if ($error): ?><p style="color:red"><?=htmlspecialchars($error)?></p><?php endif; ?>
<form method="post">
  <label>Pseudo : <input name="username" required></label><br>
  <label>Mot de passe : <input type="password" name="password" required></label><br>
  <button>Se connecter</button>
</form>
<p><a href="register.php">Créer un compte</a></p>
