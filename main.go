package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ==========================================
// 3. MAIN EXECUTION ENTRYPOINT
// ==========================================

func main() {
	// Establish concurrent file logging descriptor pipelines with standard system write access overrides
	logFile, err := os.OpenFile("cluster_health.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Fork standard console outputs into file streams concurrently via multi-writer decoupling wrappers
	mw := io.MultiWriter(os.Stdout, logFile)

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

	// Intercept array slicing operations to isolate dynamic runtime boundary panics
	limit := 3
	maxDisplay := limit
	if len(vmsResp.Data) < limit {
		maxDisplay = len(vmsResp.Data)
	}
	limitedVMs := vmsResp.Data[0:maxDisplay]

	fmt.Fprintf(mw, "\n[Telemetry] Displaying top %d VM workloads:\n", maxDisplay)

	statusCounter := make(map[string]int)
	filterStatus := "running"

	for _, vmItem := range limitedVMs {
		if vmItem.Type == "qemu" {
			statusIcon := ""
			switch vmItem.Status {
			case "running":
				statusIcon = "🟢 RUNNING"
				statusCounter["running"]++
			case "stopped":
				statusIcon = "🔴 STOPPED"
				statusCounter["stopped"]++
			default:
				statusIcon = "⚪ UNKNOWN"
				statusCounter["unknown"]++
			}

			if filterStatus != "all" && vmItem.Status != filterStatus {
				continue
			}

			fmt.Fprintf(mw, "VM ID: %-5d | VM Name: %-15s | Status: %s\n", vmItem.Vmid, vmItem.Name, statusIcon)
		}
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	fmt.Fprintf(mw, "\n[Cluster Workload Health Summary] - Generated at: %s\n", formattedTime)
	fmt.Fprintf(mw, "Total Running VMs : %d 🟢\n", statusCounter["running"])
	fmt.Fprintf(mw, "Total Stopped VMs : %d 🔴\n", statusCounter["stopped"])
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")
}
