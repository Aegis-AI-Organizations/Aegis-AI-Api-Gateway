# Passerelle API Aegis AI

L'API Gateway est le point d'entrée central pour les clients HTTP externes (Dashboard, CLI) vers l'écosystème Aegis AI. Depuis l'architecture MVP v2, il opère comme un **Proxy gRPC**, ce qui implique qu'il ne contient aucune logique métier, il n'est connecté à aucune base de données et il n'orchestre plus les flux Temporal.

## Routes mappées sur gRPC

L'intégralité du trafic REST HTTP est converti et retransmis en toute sécurité via le protocole gRPC `aegis.v2` directement au **Brain** backend.

### `POST /scans`
- **Description :** Initie un nouveau scan de sécurité ciblant une infrastructure.
- **Payload :** `{"target_image": "nginx:latest"}`
- **Routé vers :** `aegis.v2.ScanService.StartScan`

### `GET /scans/{id}`
- **Description :** Récupère le statut en direct d'un scan en cours ou terminé.
- **Routé vers :** `aegis.v2.ScanService.GetScanStatus`

### `GET /scans`
- **Description :** Liste tous les scans historiques et actifs.
- **Routé vers :** `aegis.v2.ScanService.ListScans`

### `GET /scans/{id}/vulnerabilities`
- **Description :** Récupère la liste des vulnérabilités découvertes.
- **Routé vers :** `aegis.v2.VulnerabilityService.GetVulnerabilities`

### `GET /vulnerabilities/{id}/evidences`
- **Description :** Récupère les preuves cryptographiques et historiques de frappe associés aux vulnérabilités exploitées.
- **Routé vers :** `aegis.v2.VulnerabilityService.GetEvidences`

### `GET /scans/{id}/report`
- **Description :** Télécharge le rapport PDF complet résumant l'audit de sécurité effectué.
- **Routé vers :** `aegis.v2.ScanService.GetScanReport`

## Authentification & Session

L'API utilise des JWT pour sécuriser l'accès et des cookies HTTP-only pour la persistance des sessions.

### `POST /auth/login`
- **Description :** Authentifie un utilisateur et retourne un jeton d'accès. Définit un cookie `refresh_token`.
### `POST /auth/refresh`
- **Description :** Régénère un jeton d'accès en utilisant le cookie de session.
### `POST /auth/logout`
- **Description :** Révoque la session et efface les cookies.
### `GET /auth/me`
- **Description :** Récupère le profil de l'utilisateur connecté (ID, email, rôle, entreprise).

## Gestion du Profil

### `PUT /users/me/profile`
- **Description :** Met à jour les informations du profil (nom).
### `DELETE /users/me/profile/avatar`
- **Description :** Supprime la photo de profil actuelle.
### `PUT /users/me/email`
- **Description :** Modifie l'adresse email (nécessite le mot de passe).
### `PUT /users/me/password`
- **Description :** Modifie le mot de passe.

## Gestion des Entreprises (Roles Admin/Commercial)

### `GET /companies`
- **Description :** Liste toutes les entreprises de la plateforme.
### `POST /companies/onboard`
- **Description :** **Workflow d'onboarding complet.** Crée l'entreprise, génère un token de déploiement et crée le compte propriétaire en une seule étape.

## Flux Temps-Réel (Streaming)

### `GET /scans/stream`
- **Description :** Flux SSE global pour les mises à jour de tous les scans.
### `GET /scans/:id/stream`
- **Description :** Flux SSE spécifique à un scan.

## Configuration de Sécurité
En tant que pilier de l'infrastructure Zero Trust, l'API Gateway est confinée par une stricte **Cilium Network Policy**, l'empêchant d'effectuer la moindre requête sortante hormis vers le **Brain Aegis** (Port 50051).
