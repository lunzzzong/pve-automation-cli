package main

import (
	"encoding/json"
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
}

// =========================================================================
// pvectl doctor: Health Check Assertion Engine
// =========================================================================
func runDoctorRoutine(mw io.Writer, client *PveClient) {
	// ANSI TERMINAL ESCAPE CODES FOR HIGH-IMPACT VISUALIZATION
	const (
		ColorGreen  = "\033[32m"
		ColorRed    = "\033[31m"
		ColorYellow = "\033[33m"
		ColorReset  = "\033[0m"
	)

	// Render colored sub-section headers for terminal clarity
	fmt.Fprintf(mw, "\n%sCluster%s\n", ColorYellow, ColorReset)
	fmt.Fprintln(mw, "--------")
	fmt.Fprintf(mw, "%s✔%s Quorum (Simulation)\n", ColorGreen, ColorReset)
	fmt.Fprintf(mw, "%s✔%s Corosync (Simulation)\n", ColorGreen, ColorReset)

	// Dispatch synchronized telemetry broker request to poll live hypervisor topology
	nodesData, err := client.FetchAndParseNodesDetailed()
	if err != nil {
		fmt.Fprintf(mw, "%s❌%s All Nodes Unreachable (API Connectivity Failed: %v)\n", ColorRed, ColorReset, err)
		return
	}

	allNodesHealthy := true
	var deadNodes []string
	healthScore := 100

	// Local data structure to buffer distributed node-level storage reports prior to UI rendering
	type storageReportItem struct {
		output string
		active bool
	}
	var storageReports []storageReportItem

	// Enumerate hypervisor clusters to evaluate host runtime operational telemetry
	for _, nodeItem := range nodesData {
		if nodeItem.Status != "online" {
			allNodesHealthy = false
			deadNodes = append(deadNodes, nodeItem.Name)
			// Deduct 20 points dynamically for each distressed host
			healthScore -= 20
			continue // FAULT BOUNDARY: Skip storage probing if the host itself is down
		}

		// INTERSECT LIVE NODE STORAGE: Intercept single node storage pool metrics to bypass 501 restrictions
		sStatus, sJson, sErr := client.GetNodeStorage(nodeItem.Name)
		if sErr == nil && sStatus == 200 {
			var nodeStorageResp PveStorageResponse
			if json.Unmarshal([]byte(sJson), &nodeStorageResp) == nil {
				for _, storeItem := range nodeStorageResp.Data {
					// Assert if the targeted local storage volume is flagged as active
					if storeItem.Active != 1 {
						report := fmt.Sprintf("%s❌ %s/%s (%s) [DISTRESSED / OFFLINE]%s\n", ColorRed, nodeItem.Name, storeItem.Storage, storeItem.Type, ColorReset)
						storageReports = append(storageReports, storageReportItem{output: report, active: false})
						// Deduct 15 points dynamically for each distressed storage backend
						healthScore -= 15
					} else {
						report := fmt.Sprintf("%s✔%s %s/%s (%s)\n", ColorGreen, ColorReset, nodeItem.Name, storeItem.Storage, storeItem.Type)
						storageReports = append(storageReports, storageReportItem{output: report, active: true})
					}
				}
			}
		}
	}

	// Render Node telemetry status with explicit color mappings
	if allNodesHealthy {
		fmt.Fprintf(mw, "%s✔%s All Nodes Reachable\n", ColorGreen, ColorReset)
	} else {
		fmt.Fprintf(mw, "%s❌ Nodes Distressed! Dead Nodes Cluster: %v%s\n", ColorRed, deadNodes, ColorReset)
	}

	// =========================================================================
	// STORAGE PRESENTATION LAYER: LIVE PHYSICAL NODE STORAGE INFRASTRUCTURE
	// =========================================================================
	fmt.Fprintf(mw, "\n%sStorage%s\n", ColorYellow, ColorReset)
	fmt.Fprintln(mw, "--------")

	// Flush buffered telemetry stream outputs into the terminal console
	if len(storageReports) == 0 {
		fmt.Fprintf(mw, "%s❌%s No storage pools discovered or nodes unreachable\n", ColorRed, ColorReset)
	} else {
		for _, r := range storageReports {
			fmt.Print(r.output)
		}
	}

	// Defensive Guardrail bounds checks to block negative telemetry numbers
	if healthScore < 0 {
		healthScore = 0
	}

	// =========================================================================
	// SUMMARY PRESENTATION LAYER WITH DYNAMIC WEIGHT COLORATION
	// =========================================================================
	fmt.Fprintf(mw, "\n%sSummary%s\n", ColorYellow, ColorReset)
	fmt.Fprintln(mw, "--------")

	// Dynamically transition score representation color strings according to risk thresholds
	scoreColor := ColorGreen
	if healthScore < 60 {
		scoreColor = ColorRed
	} else if healthScore < 90 {
		scoreColor = ColorYellow
	}

	// Render fully dynamic, decoupled system metric scores safely without simulation tags
	fmt.Fprintf(mw, "Health Score: %s%d%s/100\n", scoreColor, healthScore, ColorReset)
}

// =========================================================================
// pvectl diagnose: Deep Triage & Resource Contention Engine
// =========================================================================
func runDiagnoseRoutine(mw io.Writer, client *PveClient, limit int, filterStatus string) {
	fmt.Fprintln(mw, "Collecting Comprehensive Infrastructure Telemetry...")

	nodesData, err := client.FetchAndParseNodesDetailed()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] Node Telemetry retrieval failed: %v\n", err)
		return
	}

	fmt.Fprintln(mw, "\n[Success] Fetched Cluster Nodes Successfully!")
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")
	for index, nodeItem := range nodesData {
		fmt.Fprintf(mw, "[%d] Node: %-6s | Status: %-6s | Real CPU: %5.2f%% (%d Cores) | Mem: %5.2fGB / %5.2fGB (%5.2f%%)\n",
			index, nodeItem.Name, nodeItem.Status, nodeItem.CpuPercent, nodeItem.CpuCores, nodeItem.MemUsedGB, nodeItem.MemTotalGB, nodeItem.MemPercent)
	}

	vmsData, err := client.FetchAndParseVms()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] VM Telemetry retrieval failed: %v\n", err)
		return
	}

	const MaxExpectedClusterWorkloads = 150
	statusCounter := make(map[string]int, MaxExpectedClusterWorkloads)
	for _, vmItem := range vmsData {
		if vmItem.Type == "qemu" {
			statusCounter[vmItem.Status]++
		}
	}

	sort.Sort(vmsData)

	maxDisplay := limit
	if len(vmsData) < limit {
		maxDisplay = len(vmsData)
	}
	limitedVMs := vmsData[0:maxDisplay]

	fmt.Fprintf(mw, "\n[Telemetry] Displaying top %d high-risk VM workloads (Sorted by CPU):\n", maxDisplay)
	for _, vmItem := range limitedVMs {
		if vmItem.Type == "qemu" {
			if filterStatus != "all" && vmItem.Status != filterStatus {
				continue
			}
			fmt.Fprintf(mw, "VM ID: %-5d | VM Name: %-15s | CPU: %5.2f%% | Status: %s\n",
				vmItem.Vmid, vmItem.Name, vmItem.CPU*100, vmItem.GetStatusIcon())
		}
	}

	formattedTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(mw, "\n[Cluster Workload Health Summary] - Generated at: %s\n", formattedTime)
	fmt.Fprintf(mw, "Total Running VMs in Cluster : %d 🟢\n", statusCounter["running"])
	fmt.Fprintf(mw, "Total Stopped VMs in Cluster : %d 🔴\n", statusCounter["stopped"])
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")
}
