package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// ==========================================
// 1. STRUCT & CONSTRUCTOR DEFINITIONS
// ==========================================

// PveClient stores connection details for Proxmox VE API
type PveClient struct {
	ApiUrl  string
	TokenID string
	Secret  string
	HttpCli *http.Client
}

// PveNodeData represents the structure of an individual node from the cluster list
type PveNodeData struct {
	Node   string `json:"node"`
	Status string `json:"status"`
}

// PveNodesResponse represents the top-level structure for the /nodes cluster list API
type PveNodesResponse struct {
	Data []PveNodeData `json:"data"` // A slice (array) to hold multiple cluster nodes
}

// PveNodeMemory represents the detailed memory metrics inside a specific node status
type PveNodeMemory struct {
	Used  int64 `json:"used"` // Memory bytes can be huge, using int64 to avoid overflow
	Total int64 `json:"total"`
}

// PveNodeStatusData represents the actual real-time metrics of a specific node
type PveNodeStatusData struct {
	Cpu    float64       `json:"cpu"`
	MaxCpu int           `json:"cpus"`   // Mapped to 'cpus' based on PVE 9.1.9 JSON payload
	Memory PveNodeMemory `json:"memory"` // Nested memory struct for resource breakdown
	Uptime int64         `json:"uptime"`
}

// PveNodeStatusResponse represents the top-level structure for the /nodes/{node}/status API
type PveNodeStatusResponse struct {
	Data PveNodeStatusData `json:"data"` // Single object containing real-time status metrics
}

// PveVmData represents the structure of an individual VM
type PveVmData struct {
	Vmid   int    `json:"vmid"`   // VM ID is an integer
	Name   string `json:"name"`   // VM Name is a string
	Status string `json:"status"` // VM Status (running/stopped)
	Type   string `json:"type"`   // Resource type (qemu/lxc/storage...)
}

// PveVmsResponse represents the top-level structure for cluster resources
type PveVmsResponse struct {
	Data []PveVmData `json:"data"` // A slice holding clustered workload environments
}

// NewPveClient initializes and returns a new PveClient instance with insecure TLS safety bypass
func NewPveClient(apiUrl, tokenID, secret string) *PveClient {
	// Bypass self-signed certificate verification for homelab or internal cluster safety
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &PveClient{
		ApiUrl:  apiUrl,
		TokenID: tokenID,
		Secret:  secret,
		HttpCli: &http.Client{Transport: tr, Timeout: 10 * time.Second},
	}
}

// ==========================================
// 2. METHOD DEFINITIONS (API Actions)
// ==========================================

// GetNodes connects to /nodes endpoint and fetches the initial list of all cluster nodes
func (c *PveClient) GetNodes() (int, string, error) {
	reqUrl := c.ApiUrl + "/nodes"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %v", err)
	}

	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
	req.Header.Add("Authorization", authHeader)

	resp, err := c.HttpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to PVE API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %v", err)
	}

	return resp.StatusCode, string(body), nil
}

// GetNodeStatus connects to /nodes/{node}/status to query full real-time metrics dynamically
func (c *PveClient) GetNodeStatus(nodeName string) (int, string, error) {
	reqUrl := fmt.Sprintf("%s/nodes/%s/status", c.ApiUrl, nodeName)

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request for node %s: %v", nodeName, err)
	}

	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
	req.Header.Add("Authorization", authHeader)

	resp, err := c.HttpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to PVE node status API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read node status response body: %v", err)
	}

	return resp.StatusCode, string(body), nil
}

// GetVms connects to /cluster/resources and fetches all cluster resources (including VMs)
func (c *PveClient) GetVms() (int, string, error) {
	reqUrl := c.ApiUrl + "/cluster/resources"

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request for VMs: %v", err)
	}

	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
	req.Header.Add("Authorization", authHeader)

	resp, err := c.HttpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to PVE resource API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read resource response body: %v", err)
	}

	return resp.StatusCode, string(body), nil
}

// ==========================================
// 3. MAIN EXECUTION ENTRYPOINT
// ==========================================

func main() {
	// Initialize environment configuration pipeline using godotenv
	_ = godotenv.Load()
	apiUrl := os.Getenv("PVE_API_URL")
	tokenID := os.Getenv("PVE_TOKEN_ID")
	secret := os.Getenv("PVE_SECRET")

	if apiUrl == "" || tokenID == "" || secret == "" {
		fmt.Println("[ERROR] Missing required environment variables: PVE_API_URL, PVE_TOKEN_ID, or PVE_SECRET")
		return
	}

	fmt.Println("Initializing Proxmox VE API client...")
	client := NewPveClient(apiUrl, tokenID, secret)

	fmt.Println("Fetching PVE cluster nodes...")
	statusCode, rawJson, err := client.GetNodes()
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	var nodesResp PveNodesResponse
	err = json.Unmarshal([]byte(rawJson), &nodesResp)
	if err != nil {
		fmt.Printf("[ERROR] Failed to parse Nodes JSON: %v\n", err)
		return
	}

	fmt.Printf("HTTP Status Code (Cluster): %d\n", statusCode)
	fmt.Println("\n[Success] Fetched PVE Cluster Nodes Successfully!")
	fmt.Println("--------------------------------------------------------------------------------")

	for index, nodeItem := range nodesResp.Data {
		_, statusJson, err := client.GetNodeStatus(nodeItem.Node)
		if err != nil {
			fmt.Printf("[%d] Node: %s | [ERROR] Failed to fetch real-time metrics: %v\n", index, nodeItem.Node, err)
			continue
		}

		var statusResp PveNodeStatusResponse
		err = json.Unmarshal([]byte(statusJson), &statusResp)
		if err != nil {
			fmt.Printf("[%d] Node: %s | [ERROR] Failed to parse status JSON\n", index, nodeItem.Node)
			continue
		}

		memUsedGB := float64(statusResp.Data.Memory.Used) / 1024 / 1024 / 1024
		memTotalGB := float64(statusResp.Data.Memory.Total) / 1024 / 1024 / 1024
		memPercent := (float64(statusResp.Data.Memory.Used) / float64(statusResp.Data.Memory.Total)) * 100

		fmt.Printf("[%d] Node: %-6s | Status: %-6s | Real CPU: %5.2f%% (%d Cores) | Mem: %5.2fGB / %5.2fGB (%5.2f%%)\n",
			index,
			nodeItem.Node,
			nodeItem.Status,
			statusResp.Data.Cpu*100,
			statusResp.Data.MaxCpu,
			memUsedGB,
			memTotalGB,
			memPercent,
		)
	}

	fmt.Println("\nFetching PVE cluster virtual machines...")
	vmStatusCode, vmRawJson, err := client.GetVms()
	if err != nil {
		fmt.Printf("[ERROR] Failed to fetch VMs: %v\n", err)
		return
	}

	var vmsResp PveVmsResponse
	err = json.Unmarshal([]byte(vmRawJson), &vmsResp)
	if err != nil {
		fmt.Printf("[ERROR] Failed to parse VMs JSON: %v\n", err)
		return
	}

	fmt.Printf("HTTP Status Code (VMs): %d\n", vmStatusCode)
	fmt.Println("\n[Success] Fetched VM List Successfully!")
	fmt.Println("--------------------------------------------------------------------------------")

	// Initialize metrics mapping table to accumulate real-time workload quantities
	statusCounter := make(map[string]int)

	// Local filtering configuration flag to scope runtime workload display
	filterStatus := "running"

	for _, vmItem := range vmsResp.Data {
		if vmItem.Type == "qemu" {
			statusIcon := ""

			// Map strict runtime state strings into visualized telemetry indicators
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

			// Enforce active control-flow constraint filter to dynamically bypass unmatched objects
			if filterStatus != "all" && vmItem.Status != filterStatus {
				continue
			}

			fmt.Printf("VM ID: %-5d | VM Name: %-15s | Status: %s\n",
				vmItem.Vmid, vmItem.Name, statusIcon)
		}
	}

	// Dynamic temporal execution tracking: capturing precise execution runtime telemetry
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	// Optimized layout: Merging time context into a consolidated execution header
	fmt.Printf("\n[Cluster Workload Health Summary] - Generated at: %s\n", formattedTime)
	fmt.Printf("Total Running VMs : %d 🟢\n", statusCounter["running"])
	fmt.Printf("Total Stopped VMs : %d 🔴\n", statusCounter["stopped"])
	fmt.Println("--------------------------------------------------------------------------------")
}
