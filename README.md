# homelabd

`homelabd` is a lightweight, API-first control plane for Linux servers and homelabs. 

It provides a secure, fully-documented REST API to monitor and manage your host, Docker Engine, and Docker Compose stacks. It is designed to run anywhere, statically linked without any external dependencies, using SQLite for persistence.

## Features

- **System Monitoring**: CPU, memory, load average, disk storage, network, and processes.
- **Docker Management**: List containers, stream logs, view stats, and operate containers (start/stop/restart).
- **Compose Management**: Dynamically discovers Docker Compose stacks and safely allows operations (`up -d`, `down`, `pull`).
- **Power Management**: Secure system reboot and shutdown capabilities with explicit confirmation requirements.
- **Real-Time Events**: Server-Sent Events (SSE) streaming for immediate updates when containers change state or resources fluctuate.
- **Security First**: 
  - Token-based API authentication.
  - Granular Role-Based Access Control (RBAC) scopes.
  - Comprehensive Audit Logging stored in SQLite for every destructive operation.

## Installation

### Docker (Recommended)

```bash
docker run -d \
  --name homelabd \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v homelabd_data:/data \
  -p 8080:8080 \
  ghcr.io/madhavbiju/homelabd:latest
```

### Native Binary (Systemd)

Download the binary from the [Releases](#) page and configure `systemd`. See [docs/systemd.md](docs/systemd.md) for a hardened native configuration.

## Bootstrapping Auth

On first run, the database is empty and requires an initial Admin token. You can generate one via the API without authentication since the system detects it is uninitialized:

```bash
curl -X POST http://localhost:8080/api/v1/auth/tokens \
  -H "Content-Type: application/json" \
  -d '{"name": "bootstrap", "scopes": ["admin"]}'
```

Save the generated `token` securely. For all subsequent requests, pass it as a Bearer token:

```bash
curl http://localhost:8080/api/v1/system -H "Authorization: Bearer hl_..."
```

## API Documentation

See [openapi.yaml](openapi.yaml) for the full OpenAPI 3.0 specification.

## Security

Please report any security issues by following our [Security Policy](SECURITY.md). 
All power operations require explicit JSON confirmation (`{"confirm": true}`) to prevent accidental execution via CSRF or loose scripts.
