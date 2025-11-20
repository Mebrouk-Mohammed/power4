-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Hôte : 127.0.0.1
-- Généré le : jeu. 20 nov. 2025 à 14:43
-- Version du serveur : 10.4.32-MariaDB
-- Version de PHP : 8.2.12

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Base de données : `power4`
--

-- --------------------------------------------------------

--
-- Structure de la table `games`
--

CREATE TABLE `games` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `status` enum('pending','active','finished','abandoned') NOT NULL DEFAULT 'pending',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `player1_id` bigint(20) UNSIGNED NOT NULL,
  `player2_id` bigint(20) UNSIGNED DEFAULT NULL,
  `winner_id` bigint(20) UNSIGNED DEFAULT NULL,
  `rows_count` tinyint(3) UNSIGNED NOT NULL DEFAULT 6,
  `cols_count` tinyint(3) UNSIGNED NOT NULL DEFAULT 7,
  `connect_n` tinyint(3) UNSIGNED NOT NULL DEFAULT 4,
  `privacy` enum('public','private') NOT NULL DEFAULT 'public',
  `player_to_move` bigint(20) UNSIGNED DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Structure de la table `moves`
--

CREATE TABLE `moves` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `game_id` bigint(20) UNSIGNED NOT NULL,
  `move_no` int(10) UNSIGNED NOT NULL,
  `player_id` bigint(20) UNSIGNED NOT NULL,
  `column_index` tinyint(3) UNSIGNED NOT NULL,
  `row_index` tinyint(3) UNSIGNED NOT NULL,
  `disc_color` enum('R','Y') NOT NULL,
  `played_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Structure de la table `users`
--

CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `username` varchar(32) NOT NULL,
  `email` varchar(255) DEFAULT NULL,
  `password_hash` varchar(255) NOT NULL,
  `avatar_url` varchar(512) DEFAULT NULL,
  `is_admin` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `last_login_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `users`
--

INSERT INTO `users` (`id`, `username`, `email`, `password_hash`, `avatar_url`, `is_admin`, `created_at`, `last_login_at`) VALUES
(4, 'Test', 'filef32930@erynka.com', '$2y$10$hFznTw6WfDGzyz7iim.q.O5bLE6hPhua1CZYzwLpfWdLoUOqVUy5.', 'assets/avatar8.png', 0, '2025-10-30 13:51:25', NULL),
(5, 'Flavien', 'rijime1998@ametitas.com', '$2y$10$R48aRan3.PDlTe0GWf.V/uwV2piecjVKDVN4j2HSYUtIrYMGrPs0m', NULL, 0, '2025-10-30 16:41:56', NULL);

-- --------------------------------------------------------

--
-- Structure de la table `user_ratings`
--

CREATE TABLE `user_ratings` (
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `rating` int(11) NOT NULL DEFAULT 1200,
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `user_ratings`
--

INSERT INTO `user_ratings` (`user_id`, `rating`, `updated_at`) VALUES
(4, 800, '2025-10-30 14:23:44');

-- --------------------------------------------------------

--
-- Doublure de structure pour la vue `v_user_ranking`
-- (Voir ci-dessous la vue réelle)
--
CREATE TABLE `v_user_ranking` (
`user_id` bigint(20) unsigned
,`username` varchar(32)
,`rating` int(11)
);

-- --------------------------------------------------------

--
-- Structure de la vue `v_user_ranking`
--
DROP TABLE IF EXISTS `v_user_ranking`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `v_user_ranking`  AS SELECT `u`.`id` AS `user_id`, `u`.`username` AS `username`, `ur`.`rating` AS `rating` FROM (`users` `u` join `user_ratings` `ur` on(`ur`.`user_id` = `u`.`id`)) ;

--
-- Index pour les tables déchargées
--

--
-- Index pour la table `games`
--
ALTER TABLE `games`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_games_p1` (`player1_id`),
  ADD KEY `fk_games_p2` (`player2_id`),
  ADD KEY `fk_games_w` (`winner_id`),
  ADD KEY `fk_games_turn` (`player_to_move`),
  ADD KEY `ix_games_status` (`status`),
  ADD KEY `ix_games_created` (`created_at`);

--
-- Index pour la table `moves`
--
ALTER TABLE `moves`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_moves` (`game_id`,`move_no`),
  ADD KEY `ix_moves_game` (`game_id`),
  ADD KEY `ix_moves_player` (`player_id`);

--
-- Index pour la table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uq_users_username` (`username`),
  ADD UNIQUE KEY `uq_users_email` (`email`);

--
-- Index pour la table `user_ratings`
--
ALTER TABLE `user_ratings`
  ADD PRIMARY KEY (`user_id`);

--
-- AUTO_INCREMENT pour les tables déchargées
--

--
-- AUTO_INCREMENT pour la table `games`
--
ALTER TABLE `games`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT pour la table `moves`
--
ALTER TABLE `moves`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT pour la table `users`
--
ALTER TABLE `users`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;

--
-- Contraintes pour les tables déchargées
--

--
-- Contraintes pour la table `games`
--
ALTER TABLE `games`
  ADD CONSTRAINT `fk_games_p1` FOREIGN KEY (`player1_id`) REFERENCES `users` (`id`),
  ADD CONSTRAINT `fk_games_p2` FOREIGN KEY (`player2_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `fk_games_turn` FOREIGN KEY (`player_to_move`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  ADD CONSTRAINT `fk_games_w` FOREIGN KEY (`winner_id`) REFERENCES `users` (`id`) ON DELETE SET NULL;

--
-- Contraintes pour la table `moves`
--
ALTER TABLE `moves`
  ADD CONSTRAINT `fk_moves_game` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_moves_player` FOREIGN KEY (`player_id`) REFERENCES `users` (`id`);

--
-- Contraintes pour la table `user_ratings`
--
ALTER TABLE `user_ratings`
  ADD CONSTRAINT `fk_r_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
