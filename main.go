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

// ==========================================
// 3. MAIN EXECUTION ENTRYPOINT
// ==========================================

func main() {
	// Define arguments at the very beginning. Returns pointers (*int, *string) rather than raw values.
	limitPtr := flag.Int("limit", 3, "The maximum number of VM workloads to display in the telemetry output")
	filterPtr := flag.String("filter", "running", "Filter VMs by status (options: running, stopped, all)")

	// Establish concurrent file logging descriptor pipelines with standard system write access overrides
	logFile, err := os.OpenFile("cluster_health.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Fork standard console outputs into file streams concurrently via multi-writer decoupling wrappers
	mw := io.MultiWriter(os.Stdout, logFile)

	// Execute parsing explicitly after the multi-writer is wired but before connection bootstrapping
	flag.Parse()

	// Encapsulated setup: Initialize client instance directly by extracting decoupled environment contexts
	client, err := NewPveClientFromEnv()
	if err != nil {
		fmt.Fprintf(mw, "[CRITICAL] Initialization architecture failed: %v\n", err)
		return
	}

	// ==========================================
	// NODE MONITORING SUBSYSTEM PIPELINE
	// ==========================================
	fmt.Fprintln(mw, "Fetching PVE cluster nodes...")
	statusCode, rawJson, err := client.GetNodes()
	if err != nil {
		if pveErr, ok := err.(*PveApiError); ok {
			fmt.Fprintf(mw, "[CRITICAL] PVE Server Refused Request! Status: %d | Reason: %s\n", pveErr.StatusCode, pveErr.Message)
		} else {
			fmt.Fprintf(mw, "[NETWORK ERROR] Generic connection fault: %v\n", err)
		}
		return
	}

	var nodesResp PveNodesResponse
	err = json.Unmarshal([]byte(rawJson), &nodesResp)
	if err != nil {
		fmt.Fprintf(mw, "[ERROR] Failed to parse Nodes JSON: %v\n", err)
		return
	}

	fmt.Fprintf(mw, "HTTP Status Code (Cluster): %d\n", statusCode)
	fmt.Fprintln(mw, "\n[Success] Fetched PVE Cluster Nodes Successfully!")
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")

	for index, nodeItem := range nodesResp.Data {
		_, statusJson, err := client.GetNodeStatus(nodeItem.Node)
		if err != nil {
			fmt.Fprintf(mw, "[%d] Node: %s | [ERROR] Failed to fetch real-time metrics: %v\n", index, nodeItem.Node, err)
			continue
		}

		var statusResp PveNodeStatusResponse
		err = json.Unmarshal([]byte(statusJson), &statusResp)
		if err != nil {
			fmt.Fprintf(mw, "[%d] Node: %s | [ERROR] Failed to parse status JSON\n", index, nodeItem.Node)
			continue
		}

		memUsedGB := float64(statusResp.Data.Memory.Used) / 1024 / 1024 / 1024
		memTotalGB := float64(statusResp.Data.Memory.Total) / 1024 / 1024 / 1024
		memPercent := (float64(statusResp.Data.Memory.Used) / float64(statusResp.Data.Memory.Total)) * 100

		fmt.Fprintf(mw, "[%d] Node: %-6s | Status: %-6s | Real CPU: %5.2f%% (%d Cores) | Mem: %5.2fGB / %5.2fGB (%5.2f%%)\n",
			index, nodeItem.Node, nodeItem.Status, statusResp.Data.Cpu*100, statusResp.Data.MaxCpu, memUsedGB, memTotalGB, memPercent)
	}

	// ==========================================
	// VM WORKLOAD MONITORING SUBSYSTEM PIPELINE
	// ==========================================
	fmt.Fprintln(mw, "\nFetching PVE cluster virtual machines...")
	vmStatusCode, vmRawJson, err := client.GetVms()
	if err != nil {
		fmt.Fprintf(mw, "[ERROR] Failed to fetch VMs: %v\n", err)
		return
	}

	var vmsResp PveVmsResponse
	err = json.Unmarshal([]byte(vmRawJson), &vmsResp)
	if err != nil {
		fmt.Fprintf(mw, "[ERROR] Failed to parse VMs JSON: %v\n", err)
		return
	}

	fmt.Fprintf(mw, "HTTP Status Code (VMs): %d\n", vmStatusCode)
	fmt.Fprintln(mw, "\n[Success] Fetched VM List Successfully!")
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")

	// ADDED: Real-time global telemetry auditing BEFORE sorting and slicing.
	// This captures the true cluster-wide metrics instead of being biased by the CLI display limit.
	statusCounter := make(map[string]int)
	for _, vmItem := range vmsResp.Data {
		if vmItem.Type == "qemu" {
			statusCounter[vmItem.Status]++
		}
	}

	// Dynamic type conversion to trigger our custom sort.Interface blueprint
	vmsToSort := VMList(vmsResp.Data)
	sort.Sort(vmsToSort)

	// Dereference our parameter pointer securely to fetch custom boundaries dynamically
	limit := *limitPtr
	maxDisplay := limit
	if len(vmsResp.Data) < limit {
		maxDisplay = len(vmsResp.Data)
	}
	limitedVMs := vmsResp.Data[0:maxDisplay]

	// MODIFIED: Explicitly indicate that the output is prioritized by high-risk resource consumption
	fmt.Fprintf(mw, "\n[Telemetry] Displaying top %d high-risk VM workloads (Sorted by CPU):\n", maxDisplay)

	// Dereference status filter string natively
	filterStatus := *filterPtr

	for _, vmItem := range limitedVMs {
		if vmItem.Type == "qemu" {
			statusIcon := ""
			switch vmItem.Status {
			case "running":
				statusIcon = "🟢 RUNNING"
			case "stopped":
				statusIcon = "🔴 STOPPED"
			default:
				statusIcon = "⚪ UNKNOWN"
			}

			// Decoupled filtering switch using runtime arguments instead of compile-time hardcoding
			if filterStatus != "all" && vmItem.Status != filterStatus {
				continue
			}

			// MODIFIED: Injected real-time CPU metric evaluation (%5.2f%%) into the console layout to aid infrastructure troubleshooting
			fmt.Fprintf(mw, "VM ID: %-5d | VM Name: %-15s | CPU: %5.2f%% | Status: %s\n",
				vmItem.Vmid, vmItem.Name, vmItem.CPU*100, statusIcon)
		}
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	// MODIFIED: Summary now accurately reflects total hypervisor status metrics globally rather than a sliced view
	fmt.Fprintf(mw, "\n[Cluster Workload Health Summary] - Generated at: %s\n", formattedTime)
	fmt.Fprintf(mw, "Total Running VMs in Cluster : %d 🟢\n", statusCounter["running"])
	fmt.Fprintf(mw, "Total Stopped VMs in Cluster : %d 🔴\n", statusCounter["stopped"])
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")
}
