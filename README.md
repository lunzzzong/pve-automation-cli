# Proxmox VE Automation CLI (`pve-automation-cli`)

A production-grade, highly defense-engineered Command Line Interface (CLI) utility built in Go (Golang) designed to monitor and audit Proxmox VE cluster infrastructure workloads seamlessly. 

The application implements enterprise software patterns including **decoupled multi-stream logging**, **custom API diagnostic state introspection**, **runtime defense slicing bounds checking**, and **dynamic argument filtering**.

---

## High-Performance Architecture Features

- **Decoupled Architecture (`pve.go` & `main.go`)**: Strict separation of concerns (SoC). API interaction, data modeling, and validation logic are fully separated from the main runner routine to maximize code reusability.
- **Enterprise Defense Engineering**: 
  - Array slicing features proactive boundary runtime checking (`maxDisplay` fallback) to safely mitigate out-of-bounds runtime panic vectors.
  - Custom error handling (`PveApiError`) leveraging Go's structural type assertion to isolate HTTP fault boundaries from transport network anomalies.
- **Dynamic CLI Arguments Control**: Built-in native concurrency-safe argument parsing flags (`-limit` and `-filter`) using asynchronous pointer referencing/dereferencing models.
- **Dual-Stream Telemetry Recording**: Utilizing `io.MultiWriter` pipelines to fork stdout diagnostics directly onto system files (`cluster_health.log`) simultaneously in real-time.

---

## Getting Started

### 1. Prerequisites
- Go Language compiler (v1.18+ recommended)
- A running Proxmox VE instance with active API Token credentials

### 2. Infrastructure Environment Matrix (.env)
Create a `.env` configuration template file in the project's root directory:
```env
PVE_API_URL="https://YOUR_PVE_IP_OR_DOMAIN:8006/api2/json"
PVE_TOKEN_ID="root@pve!your-token-name"
PVE_SECRET="your-proxmox-api-token-secret-uuid"
