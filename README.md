# pvectl

> A Proxmox VE troubleshooting CLI built by an ops engineer, for ops engineers.

## Why this exists

Proxmox VE's native tooling is fragmented: `pvecm` for cluster ops, `qm` for VMs, `pct` for containers, `pvesm` for storage. When a cluster is misbehaving at 3am, you don't have time to stitch together the picture from four different commands.

`pvectl` isn't trying to replace those tools. It's built around the two questions you actually ask first during an incident:

1. **Is the cluster healthy right now?** (`pvectl doctor`)
2. **If not, where's the problem and how bad is it?** (`pvectl diagnose`)

## Installation

```bash
git clone https://github.com/lunzzzong/pve-automation-cli
cd pve-automation-cli
go build -o pvectl .
```

## Configuration

Create a `.env` file in the project root:

```
PVE_API_URL=https://your-pve-host:8006/api2/json
PVE_TOKEN_ID=user@pve!tokenid
PVE_SECRET=your-token-secret
```

> ⚠️ TLS verification is currently disabled by default (`InsecureSkipVerify: true`), to make local development against self-signed home-lab clusters easier. See "Known Limitations" before using this against production.

## Usage

### `pvectl doctor` — know if the cluster is healthy in 30 seconds

```
$ pvectl doctor

Cluster
--------
✔ Quorum (Simulation)
✔ Corosync (Simulation)
✔ All Nodes Reachable

Storage
--------
✔ pve-node1/local-lvm (lvmthin)
❌ pve-node2/backup-nfs (nfs) [DISTRESSED / OFFLINE]

Summary
--------
Health Score: 85/100
```

Every run is appended to `cluster_health_YYYY-MM-DD.log`, preserving a timestamped record of cluster state for post-incident review.

The Health Score is a simple deduction model: -20 per offline node, -15 per offline storage pool, floored at 0. It's not meant to be a precise SLA calculation — it's meant to let you tell at a glance whether this is a minor blip or something worth waking someone up for.

### `pvectl diagnose` — narrow down what's actually at risk

```
$ pvectl diagnose --limit 5 --filter running

[Success] Fetched Cluster Nodes Successfully!
--------------------------------------------------------------------------------
[0] Node: pve1    | Status: online  | Real CPU: 78.40% (8 Cores) | Mem: 12.30GB / 32.00GB (38.44%)

[Telemetry] Displaying top 5 high-risk VM workloads (Sorted by CPU):
VM ID: 101   | VM Name: web-prod-01    | CPU: 92.10% | Status: 🟢 RUNNING
...

[Cluster Workload Health Summary] - Generated at: 2026-06-30 14:32:01
Total Running VMs in Cluster : 12 🟢
Total Stopped VMs in Cluster : 3 🔴
--------------------------------------------------------------------------------
```

Sorts VMs by CPU usage descending, so you can spot resource contention without manually running `qm status` on every workload.

Supported flags:
- `--limit N`: show the top N highest-CPU VMs (default 3)
- `--filter running|stopped|all`: filter by VM status (default running)

## Known limitations (this is "the CLI that understands what an on-call engineer needs", not "the PVE CLI with the most features")

- **Quorum / Corosync checks are currently simulated output**, not wired up to real `pvecm status` data yet.
- **`pvectl workflow recover` is not implemented yet.** It's the next planned milestone: dry-run-first recovery suggestions based on the `doctor` diagnosis, not auto-execution.
- TLS verification is disabled by default with no flag to toggle it yet — be cautious using this against production.
- VM lifecycle management (create/clone/migrate) is intentionally out of scope. That's Terraform/Ansible territory. `pvectl` stays focused on "what do you need to see when something breaks."

## Roadmap

- [ ] Wire `pvectl doctor` up to real Quorum/Corosync status
- [ ] `pvectl workflow recover --dry-run`
- [ ] Fix `diagnose` so `--filter` is applied before sort/limit, not after
- [ ] Make TLS verification configurable (`--insecure` flag, off by default)
- [ ] Unit tests for health score calculation and JSON parsing logic

## License

MIT
