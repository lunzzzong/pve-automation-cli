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

// =========================================================================
// 1. DATA STRUCTS & CUSTOM ERROR BOUNDARIES
// =========================================================================

// NodeResult defines the clean operational domain model exposed to the presentation layer
type NodeResult struct {
	Name       string
	Status     string
	CpuPercent float64
	CpuCores   int
	MemUsedGB  float64
	MemTotalGB float64
	MemPercent float64
}

// PveApiError wraps HTTP fault boundaries for advanced diagnostic state introspection
type PveApiError struct {
	StatusCode int
	Message    string
}

// Error converts the internal telemetry tracking state into a standard error string payload
func (e *PveApiError) Error() string {
	return fmt.Sprintf("PVE API Fault (Status: %d): %s", e.StatusCode, e.Message)
}

// PveClient abstracts the cluster transport pipeline connection parameters
type PveClient struct {
	ApiUrl  string
	TokenID string
	Secret  string
	HttpCli *http.Client
}

type PveNodeData struct {
	Node   string `json:"node"`
	Status string `json:"status"`
}

type PveNodesResponse struct {
	Data []PveNodeData `json:"data"`
}

type PveNodeMemory struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
}

type PveNodeStatusData struct {
	Cpu    float64       `json:"cpu"`
	MaxCpu int           `json:"cpus"`
	Memory PveNodeMemory `json:"memory"`
	Uptime int64         `json:"uptime"`
}

type PveNodeStatusResponse struct {
	Data PveNodeStatusData `json:"data"`
}

type PveVmData struct {
	Vmid   int     `json:"vmid"`
	Name   string  `json:"name"`
	Status string  `json:"status"`
	Type   string  `json:"type"`
	CPU    float64 `json:"cpu"`
}

type PveVmsResponse struct {
	Data []PveVmData `json:"data"`
}

// VMList implements sort.Interface to unlock native high-risk classification ranking
type VMList []PveVmData

// PveStorageData models the JSON schema for individual cluster storage allocations
type PveStorageData struct {
	Storage string `json:"storage"`
	Type    string `json:"type"`
	Active  int    `json:"active"` // 1 indicates active/online, 0 indicates inactive/unreachable
}

// PveStorageResponse encapsulates the root JSON data payload returned by cluster storage endpoints
type PveStorageResponse struct {
	Data []PveStorageData `json:"data"`
}

// =========================================================================
// 2. CONSTRUCTORS & API TRANSPORT LAYER IMPLEMENTATIONS
// =========================================================================

// NewPveClient acts as a low-level constructor allocating a custom transport pipeline with TLS bypass
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

// NewPveClientFromEnv automates environment configuration parsing with native boundary sanity checks
func NewPveClientFromEnv() (*PveClient, error) {
	_ = godotenv.Load()
	apiUrl := os.Getenv("PVE_API_URL")
	tokenID := os.Getenv("PVE_TOKEN_ID")
	secret := os.Getenv("PVE_SECRET")

	// Defensive engineering: Validate that all critical identity variables are present
	if apiUrl == "" || tokenID == "" || secret == "" {
		return nil, fmt.Errorf("missing required environment variables: PVE_API_URL, PVE_TOKEN_ID, or PVE_SECRET")
	}

	// Dynamic delegation: Propagate verified configs to the underlying connection constructor
	return NewPveClient(apiUrl, tokenID, secret), nil
}

// GetNodes fetches the cluster top-level machine nodes index map securely
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

// GetNodeStatus tracks real-time compute hardware consumption metrics on specific targets
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

// GetVms evaluates system hypervisor compute resource assets distributed inside the infrastructure
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

// Len returns the cardinality bounds dynamically for index calculations
func (v VMList) Len() int {
	return len(v)
}

// Swap exchanges elements out-of-order in-place securely using native tuple decoupling
func (v VMList) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// Less evaluates ordering parameters to achieve descending risk allocation metrics
func (v VMList) Less(i, j int) bool {
	return v[i].CPU > v[j].CPU
}

// FetchAndParseVms acts as a high-level domain orchestrator that encapsulates
// the entire raw network transport layer and JSON unmarshalling guardrails.
func (c *PveClient) FetchAndParseVms() (VMList, error) {
	statusCode, rawJson, err := c.GetVms()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cluster resources: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, &PveApiError{
			StatusCode: statusCode,
			Message:    "PVE cluster infrastructure returned an unexpected status code",
		}
	}

	var vmsResp PveVmsResponse
	err = json.Unmarshal([]byte(rawJson), &vmsResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal internal PVE VMs data payload: %w", err)
	}

	return vmsResp.Data, nil
}

// FetchAndParseNodesDetailed acts as an aggregation orchestrator that dispatches
// N+1 infrastructure telemetry inquiries across the cluster.
func (c *PveClient) FetchAndParseNodesDetailed() ([]NodeResult, error) {
	_, rawJson, err := c.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cluster inventory nodes list: %w", err)
	}

	var nodesResp PveNodesResponse
	if err := json.Unmarshal([]byte(rawJson), &nodesResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal infrastructure nodes list payload: %w", err)
	}

	var finalResults []NodeResult

	for _, nodeItem := range nodesResp.Data {
		_, statusJson, err := c.GetNodeStatus(nodeItem.Node)

		// 🌟 STARS INDICATOR: EXPLICIT FAULT ISOLATION HANDLER
		if err != nil {
			deadNode := NodeResult{
				Name:       nodeItem.Node,
				Status:     "offline",
				CpuPercent: 0.0,
				CpuCores:   0,
				MemUsedGB:  0.0,
				MemTotalGB: 0.0,
				MemPercent: 0.0,
			}
			finalResults = append(finalResults, deadNode)
			continue
		}

		var statusResp PveNodeStatusResponse
		if err := json.Unmarshal([]byte(statusJson), &statusResp); err != nil {
			anomalousNode := NodeResult{
				Name:   nodeItem.Node,
				Status: "error",
			}
			finalResults = append(finalResults, anomalousNode)
			continue
		}

		memUsed := float64(statusResp.Data.Memory.Used) / 1024 / 1024 / 1024
		memTotal := float64(statusResp.Data.Memory.Total) / 1024 / 1024 / 1024
		memPercent := (memUsed / memTotal) * 100

		singleNode := NodeResult{
			Name:       nodeItem.Node,
			Status:     nodeItem.Status,
			CpuPercent: statusResp.Data.Cpu * 100,
			CpuCores:   statusResp.Data.MaxCpu,
			MemUsedGB:  memUsed,
			MemTotalGB: memTotal,
			MemPercent: memPercent,
		}

		finalResults = append(finalResults, singleNode)
	}

	return finalResults, nil
}

// GetClusterStorage dispatches a transport token to securely fetch global infrastructure storage pools
// Added on 2026-06-30 to bridge the physical storage tracking layer
func (c *PveClient) GetNodeStorage(nodeName string) (int, string, error) {
	reqUrl := fmt.Sprintf("%s/nodes/%s/storage", c.ApiUrl, nodeName)

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create node storage request: %v", err)
	}

	authHeader := fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret)
	req.Header.Add("Authorization", authHeader)

	resp, err := c.HttpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to PVE Node Storage API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read node storage response body: %v", err)
	}
	return resp.StatusCode, string(body), nil
}

// GetStatusIcon maps the raw lowercase VM status string to a production-ready console indicator emoji
func (vm PveVmData) GetStatusIcon() string {
	switch vm.Status {
	case "running":
		return "🟢 RUNNING"
	case "stopped":
		return "🔴 STOPPED"
	default:
		return "⚪ UNKNOWN"
	}
}
