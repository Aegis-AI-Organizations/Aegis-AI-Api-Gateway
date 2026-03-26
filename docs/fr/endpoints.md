# [FR] # Endpoints | Aegis-AI-Api-Gateway

Ce document liste les endpoints HTTP exposés par la passerelle API et la méthode gRPC correspondante sur le composant Brain.

L'API Gateway opère désormais sous forme de **Pure Proxy gRPC**. Il n'y a aucune logique d'orchestration ou de traitement de base de données effectuée localement.

## Cartographie REST -> gRPC

- `POST /scans`
  - Reçoit les informations de la cible (`target_image`).
  - RPC Délégué : `ScanService.StartScan`
  - Retourne l'Identifiant de Scan (`scan_id`) et son statut initial.

- `GET /scans`
  - Liste l'historique de tous les scans.
  - RPC Délégué : `ScanService.ListScans`

- `GET /scans/{id}`
  - Récupère le statut et les timestamps d'un scan spécifique.
  - RPC Délégué : `ScanService.ListScans` (Affiné côté serveur) ou `ScanService.GetScanStatus`.

- `GET /scans/{id}/report`
  - Télécharge le fichier PDF du rapport final de pentest généré par le Brain.
  - RPC Délégué : `ScanService.GetScanReport`
  - Gestion automatique des erreurs gRPC `codes.NotFound` -> `404 HTTP`.

- `GET /scans/{id}/vulnerabilities`
  - Récupère la liste des vulnérabilités découvertes avec leur sévérité.
  - RPC Délégué : `VulnerabilityService.GetVulnerabilities`

- `GET /vulnerabilities/{id}/evidences`
  - Fournit les preuves chiffrées/logs JSON ainsi que les payloads utilisés.
  - RPC Délégué : `VulnerabilityService.GetEvidences`
