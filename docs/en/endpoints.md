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

## Authentication & Session

The API uses JWTs for access security and HTTP-only cookies for session persistence.

### `POST /auth/login`
- **Description:** Authenticates a user and returns an access token. Sets a `refresh_token` cookie.
### `POST /auth/refresh`
- **Description:** Regenerates an access token using the session cookie.
### `POST /auth/logout`
- **Description:** Revokes the session and clears cookies.
### `GET /auth/me`
- **Description:** Retrieves the logged-in user's profile (ID, email, role, company).

## Profile Management

### `PUT /users/me/profile`
- **Description:** Updates profile information (name).
### `DELETE /users/me/profile/avatar`
- **Description:** Removes the current profile picture.
### `PUT /users/me/email`
- **Description:** Updates the email address (requires password verification).
### `PUT /users/me/password`
- **Description:** Updates the password.

## Company Management (Admin/Commercial Roles)

### `GET /companies`
- **Description:** Lists all companies on the platform.
### `POST /companies/onboard`
- **Description:** **Post-payment onboarding workflow.** Creates the company and owner account in `pending_activation`, then triggers the first-login invitation email.

### `POST /auth/setup-password`
- **Description:** Activates the invited owner account, sets the password, starts the session, and returns the clear agent token once.

## Agent Management

### `GET /agents`
- **Description:** Lists the persistent agents attached to the authenticated company.
- **Proxied to:** `aegis.v2.AgentService.ListAgents`

### `GET /agents/status`
- **Description:** Returns aggregated agent status counters for the dashboard (`total_agents`, `active_agents`, `inactive_agents`, `last_seen`).
- **Proxied to:** `aegis.v2.AgentService.GetAgentStatusSummary`

### `POST /agents/register`
- **Description:** Registers a deployed agent using the company deployment token.
- **Auth:** Agent deployment token in `Authorization: Bearer <ag_...>`.
- **Proxied to:** `aegis.v2.AgentService.RegisterAgent`

### `POST /agents/{id}/status`
- **Description:** Receives the agent heartbeat/status update.
- **Auth:** Agent operational secret in `Authorization: Bearer <agent_secret>`.
- **Proxied to:** `aegis.v2.AgentService.UpdateAgentStatus`

## Real-time Flux (Streaming)

### `GET /scans/stream`
- **Description:** Global SSE stream for updates across all scans.
### `GET /scans/:id/stream`
- **Description:** Scan-specific SSE stream.

## Security configuration
As part of the Zero Trust infrastructure, the API Gateway runs under a strict **Cilium Network Policy** and is restricted from executing egress traffic to anything other than the **Aegis Brain** (Port 50051).
