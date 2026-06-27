package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// =========================================================================
// MAIN EXECUTION ENTRYPOINT
// =========================================================================

func main() {
	// Define infrastructure engineering constraints to throttle dynamic memory churning
	const MaxExpectedClusterWorkloads = 150

	// Register command-line flags under global registry context for downstream propagation
	limitPtr := flag.Int("limit", 3, "The maximum number of VM workloads to display")
	filterPtr := flag.String("filter", "running", "Filter VMs by status (running, stopped, all)")

	// Generate a dynamic log filename appended with the current ISO calendar date
	logFilename := fmt.Sprintf("cluster_health_%s.log", time.Now().Format("2006-01-02"))

	// Initialize standard localized file descriptor for immutable operational logging
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Establish synchronized MultiWriter pipeline to fork output streams telemetry
	mw := io.MultiWriter(os.Stdout, logFile)

	// =========================================================================
	// CONTROL PLANE ARCHITECTURE: Subcommand Triage Routing Engine
	// =========================================================================

	// Enforce defensive programming guardrail: Halt execution if no tactical subcommand is declared
	if len(os.Args) < 2 {
		fmt.Fprintln(mw, "[ERROR] Missing triage subcommand context.")
		fmt.Fprintln(mw, "Usage: pvectl [doctor | diagnose | workflow] [options]")
		os.Exit(1)
	}

	// Intercept and extract primary operational context token
	subcommand := os.Args[1]

	// Initialize secure Proxmox API client backend from local environment state variables
	// We instantiate the client here so it's ready for any upstream routine consumption
	client, err := NewPveClientFromEnv()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] Infrastructure controller initialization failed: %v\n", err)
		return
	}

	switch subcommand {
	case "doctor":
		// Advanced Slice Slicing: Strip subcommand routing tokens prior to flag parsing
		flag.CommandLine.Parse(os.Args[2:])
		runDoctorRoutine(mw, client)

	case "diagnose":
		// Advanced Slice Slicing: Strip subcommand routing tokens prior to flag parsing
		flag.CommandLine.Parse(os.Args[2:])
		runDiagnoseRoutine(mw, client, *limitPtr, *filterPtr)

	default:
		fmt.Fprintf(mw, "[ERROR] Unknown tactical subcommand '%s'. Execute 'pvectl' for operational help.\n", subcommand)
		os.Exit(1)
	}
} // PERFECT MIGRATION: main() function now gracefully concludes here!

// =========================================================================
// pvectl doctor: Health Check Assertion Engine
// =========================================================================
func runDoctorRoutine(mw io.Writer, client *PveClient) {
	fmt.Fprintln(mw, "\nCluster")
	fmt.Fprintln(mw, "--------")
	fmt.Fprintln(mw, "✔ Quorum (Simulation)")
	fmt.Fprintln(mw, "✔ Corosync (Simulation)")

	// Dispatch synchronized telemetry broker request to poll live hypervisor topology
	nodesData, err := client.FetchAndParseNodesDetailed()
	if err != nil {
		fmt.Fprintf(mw, "❌ All Nodes Unreachable (API Connectivity Failed: %v)\n", err)
		return
	}

	// Initialize localized state flags to capture physical host failure domains
	allNodesHealthy := true
	var deadNodes []string

	// DYNAMIC METRICS: Initialize a perfect baseline score
	healthScore := 100

	// Enumerate hypervisor clusters to evaluate host runtime operational telemetry
	for _, nodeItem := range nodesData {
		// Assert node availability state against the upstream orchestration standard
		if nodeItem.Status != "online" {
			allNodesHealthy = false
			deadNodes = append(deadNodes, nodeItem.Name)

			// 💡 DYNAMIC METRICS: Deduct 20 points dynamically for each distressed host
			healthScore -= 20
		}
	}

	// DEFENSIVE GUARDRAIL: Prevent the telemetry score from dropping into negative bounds
	if healthScore < 0 {
		healthScore = 0
	}

	// Evaluate systemic state machine flag to formulate control plane diagnostics
	if allNodesHealthy {
		fmt.Fprintln(mw, "✔ All Nodes Reachable")
	} else {
		// Log explicit fault isolation data containing high-risk distressed targets
		fmt.Fprintf(mw, "❌ Nodes Distressed! Dead Nodes Cluster: %v\n", deadNodes)
	}

	// ---- Downstream Subsystems (Temporary Static Mock Framework) ----
	fmt.Fprintln(mw, "\nStorage")
	fmt.Fprintln(mw, "--------")
	fmt.Fprintln(mw, "✔ local-lvm (Simulation)")
	fmt.Fprintln(mw, "✔ nfs-storage (Simulation)")

	// PLUG DYNAMIC SCORE: Render metrics driven output summary
	fmt.Fprintln(mw, "\nSummary")
	fmt.Fprintln(mw, "--------")
	fmt.Fprintf(mw, "Health Score: %d/100\n", healthScore)
}

// =========================================================================
// pvectl diagnose: Deep Triage & Resource Contention Engine
// =========================================================================
func runDiagnoseRoutine(mw io.Writer, client *PveClient, limit int, filterStatus string) {
	fmt.Fprintln(mw, "Collecting Comprehensive Infrastructure Telemetry...")

	// 1. Dispatch telemetry request via the parsing broker to fetch detailed hypervisor states
	nodesData, err := client.FetchAndParseNodesDetailed()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] Node Telemetry retrieval failed: %v\n", err)
		return
	}

	// 2. Render structured hypervisor cluster matrix
	fmt.Fprintln(mw, "\n[Success] Fetched Cluster Nodes Successfully!")
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")
	for index, nodeItem := range nodesData {
		fmt.Fprintf(mw, "[%d] Node: %-6s | Status: %-6s | Real CPU: %5.2f%% (%d Cores) | Mem: %5.2fGB / %5.2fGB (%5.2f%%)\n",
			index, nodeItem.Name, nodeItem.Status, nodeItem.CpuPercent, nodeItem.CpuCores, nodeItem.MemUsedGB, nodeItem.MemTotalGB, nodeItem.MemPercent)
	}

	// 3. Query centralized client for virtualized workload telemetry volume
	vmsData, err := client.FetchAndParseVms()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] VM Telemetry retrieval failed: %v\n", err)
		return
	}

	// 4. Enforce pre-allocated map allocations to block runtime hashmap growth rehash spikes
	const MaxExpectedClusterWorkloads = 150
	statusCounter := make(map[string]int, MaxExpectedClusterWorkloads)
	for _, vmItem := range vmsData {
		if vmItem.Type == "qemu" {
			statusCounter[vmItem.Status]++
		}
	}

	// 5. Orchestrate high-risk classification by sorting workloads natively via CPU footprints
	sort.Sort(vmsData)

	// 6. Calculate safe boundaries for display windows to dodge slice out-of-bounds runtime panics
	maxDisplay := limit
	if len(vmsData) < limit {
		maxDisplay = len(vmsData)
	}
	limitedVMs := vmsData[0:maxDisplay]

	// 7. Render high-risk virtualized computing instances dashboard
	fmt.Fprintf(mw, "\n[Telemetry] Displaying top %d high-risk VM workloads (Sorted by CPU):\n", maxDisplay)
	for _, vmItem := range limitedVMs {
		if vmItem.Type == "qemu" {
			// Apply operational runtime state filtering constraints
			if filterStatus != "all" && vmItem.Status != filterStatus {
				continue
			}
			fmt.Fprintf(mw, "VM ID: %-5d | VM Name: %-15s | CPU: %5.2f%% | Status: %s\n",
				vmItem.Vmid, vmItem.Name, vmItem.CPU*100, vmItem.GetStatusIcon())
		}
	}

	// 8. Generate execution timestamp metrics and formulate cluster health summaries
	formattedTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(mw, "\n[Cluster Workload Health Summary] - Generated at: %s\n", formattedTime)
	fmt.Fprintf(mw, "Total Running VMs in Cluster : %d 🟢\n", statusCounter["running"])
	fmt.Fprintf(mw, "Total Stopped VMs in Cluster : %d 🔴\n", statusCounter["stopped"])
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")
}
