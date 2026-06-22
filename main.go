package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// ==========================================
// MAIN EXECUTION ENTRYPOINT
// ==========================================

func main() {
	// Define infrastructure engineering constraints to throttle memory churning
	const MaxExpectedClusterWorkloads = 150

	// Register command-line flags under global registry context
	limitPtr := flag.Int("limit", 3, "The maximum number of VM workloads to display")
	filterPtr := flag.String("filter", "running", "Filter VMs by status (running, stopped, all)")

	// Generate a dynamic log filename appended with the current calendar date
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
	// 20260622 CONTROL PLANE ARCHITECTURE: Subcommand Triage Routing Engine
	// =========================================================================

	// Enforce defensive programming guardrail: Halt execution if no subcommand is declared
	if len(os.Args) < 2 {
		fmt.Fprintln(mw, "[ERROR] Missing triage subcommand context.")
		fmt.Fprintln(mw, "Usage: pvectl [doctor | diagnose | workflow] [options]")
		os.Exit(1)
	}

	// Intercept and extract primary operational context token
	subcommand := os.Args[1]

	switch subcommand {
	case "doctor", "diagnose":
		fmt.Fprintf(mw, "Launching pvectl [%s] triage routine...\n", subcommand)

		// Advanced Slice Slicing: Strip away binary path and subcommand context,
		// routing only downstream flags to the legacy parsing controller.
		flag.CommandLine.Parse(os.Args[2:])

	default:
		fmt.Fprintf(mw, "[ERROR] Unknown tactical subcommand '%s'. Execute 'pvectl' for operational help.\n", subcommand)
		os.Exit(1)
	}

	// =========================================================================

	// Initialize secure Proxmox API client backend from local environment state variables
	client, err := NewPveClientFromEnv()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] Infrastructure controller initialization failed: %v\n", err)
		return
	}

	// ==========================================
	// NODE MONITORING SUBSYSTEM
	// ==========================================
	fmt.Fprintln(mw, "Fetching PVE cluster nodes...")

	// Fetch aggregated node metrics
	nodesData, err := client.FetchAndParseNodesDetailed()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] Node Telemetry retrieval failed: %v\n", err)
		return
	}

	fmt.Fprintln(mw, "\n[Success] Fetched Cluster Nodes Successfully!")
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")

	// Print node list
	for index, nodeItem := range nodesData {
		fmt.Fprintf(mw, "[%d] Node: %-6s | Status: %-6s | Real CPU: %5.2f%% (%d Cores) | Mem: %5.2fGB / %5.2fGB (%5.2f%%)\n",
			index, nodeItem.Name, nodeItem.Status, nodeItem.CpuPercent, nodeItem.CpuCores, nodeItem.MemUsedGB, nodeItem.MemTotalGB, nodeItem.MemPercent)
	}

	// ==========================================
	// VM WORKLOAD MONITORING SUBSYSTEM
	// ==========================================
	fmt.Fprintln(mw, "\nFetching PVE cluster virtual machines...")

	// Fetch all VMs from client wrapper
	vmsData, err := client.FetchAndParseVms()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] VM Telemetry retrieval failed: %v\n", err)
		return
	}

	fmt.Fprintln(mw, "\n[Success] Fetched VM List Successfully!")
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")

	// Allocate map capacity dynamically based on total fetched telemetry volume
	// to prevent inner-loop memory resizing overhead for high-density clusters
	statusCounter := make(map[string]int, MaxExpectedClusterWorkloads)
	for _, vmItem := range vmsData {
		if vmItem.Type == "qemu" {
			statusCounter[vmItem.Status]++
		}
	}

	// Sort VMs by CPU usage (triggers Custom sort.Interface)
	sort.Sort(vmsData)

	// Get display limits securely
	limit := *limitPtr
	filterStatus := *filterPtr

	maxDisplay := limit
	if len(vmsData) < limit {
		maxDisplay = len(vmsData)
	}

	// Slice top high-risk workloads
	limitedVMs := vmsData[0:maxDisplay]

	fmt.Fprintf(mw, "\n[Telemetry] Displaying top %d high-risk VM workloads (Sorted by CPU):\n", maxDisplay)

	for _, vmItem := range limitedVMs {
		if vmItem.Type == "qemu" {
			// Apply runtime status filter
			if filterStatus != "all" && vmItem.Status != filterStatus {
				continue
			}

			fmt.Fprintf(mw, "VM ID: %-5d | VM Name: %-15s | CPU: %5.2f%% | Status: %s\n",
				vmItem.Vmid, vmItem.Name, vmItem.CPU*100, vmItem.GetStatusIcon())
		}
	}

	// currentTime := time.Now()
	// formattedTime := currentTime.Format("2006-01-02 15:04:05")
	// Format current system timestamp natively using Method Chaining
	formattedTime := time.Now().Format("2006-01-02 15:04:05")

	// Render cluster health summary
	fmt.Fprintf(mw, "\n[Cluster Workload Health Summary] - Generated at: %s\n", formattedTime)
	fmt.Fprintf(mw, "Total Running VMs in Cluster : %d 🟢\n", statusCounter["running"])
	fmt.Fprintf(mw, "Total Stopped VMs in Cluster : %d 🔴\n", statusCounter["stopped"])
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")
}
