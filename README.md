# 🛡️ Aegis AI - API Gateway

**Project ID:** AEGIS-CORE-2026

## 🏗️ System Architecture & Role
The **Aegis API Gateway** is the high-performance entry point of the Aegis platform. It handles all incoming traffic (REST/SSE) from the Dashboard and translates it into secure internal gRPC calls to the Brain orchestrator.

* **Tech Stack:** Go 1.22+, **Gin Framework**, gRPC-Go.
* **Role:** Primary router, Authentication provider (JWT), and SSE (Server-Sent Events) broadcaster.
* **Orchestration:** Built to handle massive concurrency using native Go routines for low-latency request mapping.

---

## 🚀 Key Features

- **Gin Router**: Ultra-fast RESTful endpoints with professional middleware (Auth, Logging, Recovery).
- **Internal mTLS**: Bi-directional TLS authentication for all communication with the `Aegis-AI-Brain`.
- **SSE Streams**: Real-time push updates for scan statuses via `/scans/stream`.
- **JWT Auth**: Full session management with HTTP-only refresh token support.

---

## 🔐 Security & DevSecOps Mandates

- **Zero Trust**: No clear-text internal communication. Both client and server certificates are validated.
- **Secret Injection**: No `.env` files in production. Secrets are injected via Kubernetes Secrets or Infisical.
- **Network Isolation**: Only allows ingress from the `nginx-ingress` and regulated egress to the `Aegis-AI-Brain` and databases.

---

## 🐳 Deployment (Kubernetes)

The Gateway is deployed as a lean, distroless container for a minimal attack surface.

```yaml
# Helm values example
image:
  repository: ghcr.io/aegis-ai/aegis-api-gateway
  tag: latest
service:
  port: 80
  targetPort: 8080
tls:
  enabled: true
  caCert: "/etc/tls/ca.crt"
  clientCert: "/etc/tls/client.crt"
  clientKey: "/etc/tls/client.key"
```

---

## 🛠️ Development & Testing

```bash
# Run locally
go run cmd/api/main.go

# Run unit tests
go test ./internal/...
```

*Aegis AI — Security Engineering — 2026*
