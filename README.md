# pvectl

An operations-focused automation CLI for Proxmox VE designed to simplify infrastructure management and accelerate troubleshooting during incidents.

## Why pvectl?

Proxmox VE is powerful, but its native command-line ecosystem is deeply fragmented. During a high-stress infrastructure incident, engineers are forced to context-switch between disparate utilities:
* `pvecm` for cluster operations
* `qm` for QEMU Virtual Machines
* `pct` for LXC Containers
* `pvesm` for storage backends

**pvectl** unifies these fragmented control planes into a single, cohesive, binary engine. It transitions your operational workflow from a passive "API Wrapper" into an active "Expert Diagnostic System" tailored for SREs and Infrastructure Engineers managing high-density environments.

---

## Core Command Hierarchy Architecture

To accelerate triage during disaster recovery scenarios (such as power outages or ELK observability blindness), `pvectl` organizes its capabilities into specialized execution modes:

### 1. `pvectl doctor` (The Telemetry Health Assessment)
Performs automated assertions across the entire cluster topography to generate a holistic health scoring report.
* **Triage Scope**: Corosync connectivity, Quorum validation, Node reachability, and Replication statuses.
* **Operational Value**: Instantly exposes split-brain or network isolation issues without manual cross-referencing.

### 2. `pvectl diagnose` (Deep-Dive Triage Engine)
Collects and aggregates real-time telemetry from physical hypervisors down to individual workloads.
* **Triage Scope**: Compiles Compute (CPU/Mem Overcommit), Storage pools, Bridges, and VM/CT status metrics.
* **Output Formats**: Supports standard ANSI tables, `--json`, and `--yaml` for seamless workflow integration.

### 3. `pvectl workflow recover` (Autonomous Disaster Recovery)
An active automation runbook that sequentially validates the cluster state and gracefully orchestrates workload recovery post-incident.

---

## Roadmap

- [x] Pre-allocated resource maps to eliminate inner-loop runtime memory rehashing.
- [x] Thread-safe decoupled parsing engine for localized node telemetry.
- [ ] Implement `pvectl doctor` automated assertion engine.
- [ ] Integrate Cobra CLI framework for advanced subcommand profiles and shell auto-completion.
- [ ] Support structured structured logging and JSON output configurations.