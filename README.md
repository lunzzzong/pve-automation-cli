# Proxmox VE Cluster Telemetry CLI Tool (`pve-automation-cli`)

A production-grade command-line interface (CLI) monitoring tool engineered in **Go**. This tool interacts directly with the **Proxmox VE (PVE) REST API** to execute chained topology discovery, dynamically fetching real-time CPU utilization, active core counts, and precise memory metrics across all cluster nodes.



## Key Engineering Highlights

* **Chained API Topology Discovery**: Executes a two-tier orchestration workflow—initial global master list compilation fetched from `/nodes`, followed by multi-threaded logical execution querying individual granular status from `/nodes/{node}/status`.
* **Dynamic URL Routing**: Implements dynamic endpoint string interpolation (`fmt.Sprintf`) to securely inject node variables into localized routing paths.
* **Multi-Dimensional JSON Deserialization**: Maps deeply nested, heterogeneous upstream JSON structures into custom nested Go `struct` types using `json.Unmarshal`.
* **Defensive Engineering & Resilience**: Integrates error budget controls (`if err != nil`), active connection release cycles (`defer resp.Body.Close()`), and graceful failure degradation via loop continuation logic.
* **Telemetry Normalization**: Parses low-level computational integer byte payloads from hardware output and dynamically translates them into human-readable Gigabyte (GB) floating-point specifications.

---

## Architecture & Data Workflow

1.  **Authentication**: Injects custom `PVEAPIToken` authorizations into outbound HTTP request headers.
2.  **TLS Security Bypass**: Configures custom `tls.Config` transport wrappers (`InsecureSkipVerify: true`) to seamlessly bypass self-signed homelab/enterprise SSL certificate blocks.
3.  **Data Hydration**:
    ```
    [Main Entry] ──> Fetch Cluster Nodes List (/nodes)
                          │
                          └──> Loop Iterator (pve1, pve2, pve3...)
                                    │
                                    └──> Fetch Real-Time Status (/nodes/{node}/status)
                                              │
                                              └──> JSON Decoupling & Byte-to-GB Conversion
                                                       │
                                                       └──> CLI Live Output Render
    ```

---

## System Prerequisites & Installation

### 1. Configure Environmental Variables
The binary reads operational credentials directly from the host environment to enforce zero-hardcoded-secret security principles.

In your **Windows PowerShell**, export your specific target PVE environment settings:

```powershell
$env:PVE_API_URL="https://<YOUR_PVE_IP>:8006/api2/json"
$env:PVE_TOKEN_ID="root@pam!go-cli"
$env:PVE_SECRET="ddbf9c81-2d87-44e6-8c0b-4510704730d7"