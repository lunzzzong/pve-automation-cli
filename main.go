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
	// CLI Flags
	limitPtr := flag.Int("limit", 3, "The maximum number of VM workloads to display")
	filterPtr := flag.String("filter", "running", "Filter VMs by status (running, stopped, all)")

	// Generate a dynamic log filename appended with the current calendar date
	logFilename := fmt.Sprintf("cluster_health_%s.log", time.Now().Format("2006-01-02"))
	// Setup logging descriptor
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		fmt.Printf("[ERROR] Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Fork console output to log file
	mw := io.MultiWriter(os.Stdout, logFile)
	flag.Parse()

	// Init PVE Client from environment variables
	client, err := NewPveClientFromEnv()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] Initialization failed: %v\n", err)
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

	// Calculate cluster-wide VM stats before sorting/slicing
	statusCounter := make(map[string]int)
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
