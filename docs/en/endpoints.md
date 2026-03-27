# Aegis AI API Gateway

The API Gateway is the central entry point for all HTTP clients (Dashboard, CLI, etc.) into the Aegis AI ecosystem. As of the MVP v2 architecture, it operates as a **Pure Microservice Proxy**, meaning it contains zero business logic, does not connect to the database, and does not orchestrate Temporal workflows.

## Endpoints mapped to gRPC

All REST HTTP traffic is securely mapped and forwarded via `aegis.v2` gRPC protocols directly to the **Brain** backend.

### `POST /scans`
- **Description:** Initiates a new security scan against target infrastructure.
- **Payload:** `{"target_image": "nginx:latest"}`
- **Proxied to:** `aegis.v2.ScanService.StartScan`

### `GET /scans/{id}`
- **Description:** Retrieves the live status of an ongoing or completed scan.
- **Proxied to:** `aegis.v2.ScanService.GetScanStatus`

### `GET /scans`
- **Description:** Lists all historical and active scans.
- **Proxied to:** `aegis.v2.ScanService.ListScans`

### `GET /scans/{id}/vulnerabilities`
- **Description:** Fetches discovered vulnerabilities for a given scan ID.
- **Proxied to:** `aegis.v2.VulnerabilityService.GetVulnerabilities`

### `GET /vulnerabilities/{id}/evidences`
- **Description:** Fetches cryptographic proofs and raw payloads used to exploit a specific vulnerability.
- **Proxied to:** `aegis.v2.VulnerabilityService.GetEvidences`

### `GET /scans/{id}/report`
- **Description:** Downloads a comprehensive generated PDF report summarizing the pentest execution.
- **Proxied to:** `aegis.v2.ScanService.GetScanReport`

## Security configuration
As part of the Zero Trust infrastructure, the API Gateway runs under a strict **Cilium Network Policy** and is restricted from executing egress traffic to anything other than the **Aegis Brain** (Port 50051).
