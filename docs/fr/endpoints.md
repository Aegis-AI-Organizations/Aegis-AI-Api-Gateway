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

## Configuration de Sécurité
En tant que pilier de l'infrastructure Zero Trust, l'API Gateway est confinée par une stricte **Cilium Network Policy**, l'empêchant d'effectuer la moindre requête sortante hormis vers le **Brain Aegis** (Port 50051).
