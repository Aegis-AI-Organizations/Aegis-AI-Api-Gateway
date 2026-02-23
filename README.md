# 🛡️ Aegis AI - API Gateway

**Project ID:** AEGIS-CORE-2026

## 🏗️ System Architecture & Role
The **Aegis API Gateway** is the core entrypoint of the compute layer within the Aegis Kubernetes cluster. It sits securely behind the Nginx Ingress Controller (`aegis-gateway` namespace) and handles all incoming gRPC-Web and REST requests from the private Admin Console.

* **Tech Stack:** Go (Gin). Uses Goroutines to handle massive concurrency to backend APIs.
* **Role:** Acts as the primary router, interacting directly with the Data Layer (PostgreSQL, ClickHouse, Neo4j, Redis) and the Brain Cluster (Temporal).
* **Architecture Justification:** Go provides the lowest latency and optimal concurrent networking performance needed for a high-intensity API Gateway routing requests at massive scale.

## 🔐 Security & DevSecOps Mandates
* **No Plain-Text Secrets:** Secrets injected dynamically at runtime via Infisical. `.env` files are STRICTLY FORBIDDEN.
* **Network Isolation:** Resides in the `aegis-core` namespace. Only allows ingress from `aegis-gateway` and strictly regulates egress solely to `aegis-data` and `kube-apiserver`.
* **Authentication:** Mandates OIDC with PKCE validity checks for all incoming requests.

## 🐳 Kubernetes / Docker Deployment
Packaged into an ultra-lean distroless/scratch container for minimal attack surface.

```bash
docker pull ghcr.io/aegis-ai/aegis-api-gateway:latest

# Deployed via Kubernetes Deployments, strictly read-only and unprivileged
infisical run --env=prod -- docker run -d \
  --name aegis-api-gateway \
  --read-only \
  --cap-drop=ALL \
  --security-opt no-new-privileges:true \
  --user 10001:10001 \
  -p 8080:8080 \
  -e INFISICAL_TOKEN=$INFISICAL_TOKEN \
  ghcr.io/aegis-ai/aegis-api-gateway:latest
```
