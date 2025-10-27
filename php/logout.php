<?php
// Démarre la session actuelle (ou la reprend si elle existe déjà)
session_start();

// Détruit toutes les données stockées dans la session
// → cela déconnecte l'utilisateur (son ID, son email, etc. sont supprimés de $_SESSION)
session_destroy();

// Redirige immédiatement l’utilisateur vers la page de connexion
header('Location: login.php');

// "exit" stoppe l’exécution du script après la redirection
// (sinon le code continuerait à s’exécuter inutilement)
exit;
