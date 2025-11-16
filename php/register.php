<?php
require_once 'db.php';
session_start();

$error = '';
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = trim($_POST['username'] ?? '');
    $email = trim($_POST['email'] ?? '');
    $password = $_POST['password'] ?? '';

    if ($username === '' || $password === '') {
        $error = 'Pseudo et mot de passe requis.';
    } else {
        // Vérifie si le pseudo existe déjà
        $stmt = $pdo->prepare("SELECT id FROM users WHERE username = ?");
        $stmt->execute([$username]);
        if ($stmt->fetch()) {
            $error = 'Ce pseudo existe déjà.';
        } else {
            $hash = password_hash($password, PASSWORD_BCRYPT);
            $stmt = $pdo->prepare("INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)");
            $stmt->execute([$username, $email ?: null, $hash]);

            $_SESSION['user_id'] = $pdo->lastInsertId();
            header('Location: login.php');
            exit;
        }
    }
}
?>
<!doctype html>
<html lang="fr">
<head>
    <meta charset="utf-8">
    <title>Inscription - Power4</title>
</head>
<body>
    <h1>Créer un compte</h1>
    <?php if ($error): ?>
        <p style="color:red"><?= htmlspecialchars($error) ?></p>
    <?php endif; ?>
    <form method="post">
        <label>Pseudo : <input name="username" required></label><br>
        <label>Email : <input name="email" type="email"></label><br>
        <label>Mot de passe : <input name="password" type="password" required></label><br>
        <button type="submit">S'inscrire</button>
    </form>
</body>
</html>
