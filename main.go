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

// PveApiError is a custom error struct engineered to capture both HTTP context and system faults
type PveApiError struct {
	StatusCode int
	Message    string
}

// Error implements the standard Go error interface to enable specialized debugging telemetry
func (e *PveApiError) Error() string {
	return fmt.Sprintf("PVE API Fault (Status: %d): %s", e.StatusCode, e.Message)
}

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
	Data []PveNodeData `json:"data"`
}

// PveNodeMemory represents the detailed memory metrics inside a specific node status
type PveNodeMemory struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
}

// PveNodeStatusData represents the actual real-time metrics of a specific node
type PveNodeStatusData struct {
	Cpu    float64       `json:"cpu"`
	MaxCpu int           `json:"cpus"`
	Memory PveNodeMemory `json:"memory"`
	Uptime int64         `json:"uptime"`
}

// PveNodeStatusResponse represents the top-level structure for the /nodes/{node}/status API
type PveNodeStatusResponse struct {
	Data PveNodeStatusData `json:"data"`
}

// PveVmData represents the structure of an individual VM
type PveVmData struct {
	Vmid   int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

// PveVmsResponse represents the top-level structure for cluster resources
type PveVmsResponse struct {
	Data []PveVmData `json:"data"`
}

// NewPveClient initializes a new client with insecure TLS bypass for homelabs
func NewPveClient(apiUrl, tokenID, secret string) *PveClient {
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

// GetNodes connects to /nodes endpoint and evaluates response correctness natively
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

	// 🛠️ Milestone 6: Intercept failed HTTP requests and return explicit custom PveApiError
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, "", &PveApiError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %v", err)
	}
	return resp.StatusCode, string(body), nil
}

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
	_ = godotenv.Load()

	logFile, err := os.OpenFile("cluster_health.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create log file: %v\n", err)
		return
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)

	apiUrl := os.Getenv("PVE_API_URL")
	tokenID := os.Getenv("PVE_TOKEN_ID")
	secret := os.Getenv("PVE_SECRET")

	if apiUrl == "" || tokenID == "" || secret == "" {
		fmt.Fprintln(mw, "[ERROR] Missing required environment variables: PVE_API_URL, PVE_TOKEN_ID, or PVE_SECRET")
		return
	}

	fmt.Fprintln(mw, "Initializing Proxmox VE API client...")
	client := NewPveClient(apiUrl, tokenID, secret)

	fmt.Fprintln(mw, "Fetching PVE cluster nodes...")
	statusCode, rawJson, err := client.GetNodes()
	if err != nil {
		// 🛠️ Milestone 6: Execute type assertion switch to unpack rich error telemetry context
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

	statusCounter := make(map[string]int)
	filterStatus := "running"

	for _, vmItem := range vmsResp.Data {
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

			fmt.Fprintf(mw, "VM ID: %-5d | VM Name: %-15s | Status: %s\n",
				vmItem.Vmid, vmItem.Name, statusIcon)
		}
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	fmt.Fprintf(mw, "\n[Cluster Workload Health Summary] - Generated at: %s\n", formattedTime)
	fmt.Fprintf(mw, "Total Running VMs : %d 🟢\n", statusCounter["running"])
	fmt.Fprintf(mw, "Total Stopped VMs : %d 🔴\n", statusCounter["stopped"])
	fmt.Fprintln(mw, "--------------------------------------------------------------------------------")
}
