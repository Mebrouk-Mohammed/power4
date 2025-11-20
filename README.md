Puissance 4 — Version Web

Un projet complet, structuré et évolutif permettant de jouer au Puissance 4 directement dans un navigateur.
Le système intègre l’authentification, les profils utilisateurs, un classement basé sur l’ELO, plusieurs styles d’interface et une architecture solide pensée pour évoluer vers un véritable jeu en ligne.

Présentation du projet

Ce projet implémente le Puissance 4 en version web avec :

Authentification complète (inscription, connexion, suppression)

Profils utilisateurs : avatar, email, statistiques

Classement dynamique basé sur l’ELO

Système de rangs (Bronze, Silver, Gold, Platine, Diamant, Master)

Progression visuelle vers le prochain rang

Leaderboard trié par ELO

Différents plateaux de jeu (Small, Medium, Large)

Détection automatique des alignements gagnants

Architecture modulable permettant d’ajouter une IA, du multijoueur en ligne ou un lobby

Technologies utilisées
Backend

Go (Golang)

HTML Templates

Cookies / Sessions

MySQL via Go (driver officiel)

Frontend

HTML5

CSS (thèmes neon)

JavaScript (animations et timers)

Base de données

MySQL hébergé dans XAMPP

Gestion via phpMyAdmin :
http://localhost:80/phpmyadmin

Outils

VS Code

Git

XAMPP

Installation
1. Cloner le projet
git clone https://github.com/Mebrouk-Mohammed/power4.git
cd power4

2. Lancer le serveur Go
go run main.go

3. Accéder au jeu

Ouvrir le navigateur sur :
http://localhost:8080/register

Architecture du projet
POWER4/
├── auth/                  # Authentification, profils, leaderboard, DB
├── CSS/                   # Styles des plateaux
├── img/                   # Jetons
├── js/                    # Scripts divers
├── source/                # Logique du serveur
├── static/avatars/        # Avatars des utilisateurs
├── templates/             # Templates HTML
├── project_documentation/ # Documents PDF
├── scripts/               # Scripts PowerShell
├── go.mod
├── go.sum
├── main.go
└── README.md

Fonctionnalités principales
Gestion du plateau

Initialisation du plateau selon la taille choisie

Détection automatique des victoires

Trois tailles de plateau : small, medium, large

Profils utilisateurs

Avatar personnalisable

Email modifiable

Statistiques : parties jouées, victoires, défaites, nuls

Rang affiché + barre de progression

Classement et ELO

Système ELO pour classer les joueurs

Rang attribué automatiquement selon le score :

Bronze

Silver

Gold

Platine

Diamant

Master

Leaderboard trié par ELO décroissant

Authentification

Inscription

Connexion

Déconnexion

Suppression de compte

Cookies sécurisés pour conserver la session

Décisions d’architecture

Séparation nette entre logique métier, interface et accès base de données

Repository abstrait permettant d’utiliser soit MySQL, soit un repository en mémoire si la DB n’est pas accessible

Templates HTML pour une interface flexible et personnalisable

Go choisi pour sa stabilité, sa performance et sa simplicité réseau

Structure pensée pour ajouter plus tard :

Multijoueur en ligne via WebSockets

IA contre le joueur

Salles de jeu

Historique des parties

Système social (amis, chat)

Limites actuelles et pistes d’amélioration
Limites

Le jeu n’est pas encore public

Pas d’IA intégrée

Pas de sécurisation avancée des URLs

Pas d’espace communautaire (forum, chat)

Améliorations possibles

Déploiement du jeu en ligne

Mode spectateur

IA configurable (easy / medium / hard)

Mode classé et parties rapides

Chat en temps réel

Tournois automatiques

Crédits

Développé par :

Humbert Chloé
Mebrouk Mohammed
Leneveu Flavien
