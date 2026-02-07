package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// MCP JSON-RPC types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP types
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo        `json:"serverInfo"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type ToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PLM API client
type PLMClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewPLMClient(baseURL, token string) *PLMClient {
	return &PLMClient{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *PLMClient) Request(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = strings.NewReader(string(bodyBytes))
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// MCP Server
type MCPServer struct {
	plm *PLMClient
}

func NewMCPServer(plmBaseURL, token string) *MCPServer {
	return &MCPServer{
		plm: NewPLMClient(plmBaseURL, token),
	}
}

func (s *MCPServer) GetTools() []Tool {
	return []Tool{
		// Products
		{
			Name:        "plm_list_products",
			Description: "列出所有产品",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"status": {Type: "string", Description: "状态筛选: draft/active/deprecated"},
				"search": {Type: "string", Description: "搜索关键词"},
			}},
		},
		{
			Name:        "plm_create_product",
			Description: "创建新产品",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"code": {Type: "string", Description: "产品编码"},
				"name": {Type: "string", Description: "产品名称"},
				"description": {Type: "string", Description: "产品描述"},
			}, Required: []string{"code", "name"}},
		},
		{
			Name:        "plm_get_product",
			Description: "获取产品详情",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "产品ID"},
			}, Required: []string{"id"}},
		},

		// Templates
		{
			Name:        "plm_list_templates",
			Description: "列出所有项目模板",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"type": {Type: "string", Description: "模板类型: SYSTEM/CUSTOM"},
			}},
		},
		{
			Name:        "plm_get_template",
			Description: "获取模板详情（含任务列表）",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "模板ID"},
			}, Required: []string{"id"}},
		},

		// Projects
		{
			Name:        "plm_list_projects",
			Description: "列出所有项目",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"status": {Type: "string", Description: "状态筛选"},
			}},
		},
		{
			Name:        "plm_create_project_from_template",
			Description: "从模板创建项目",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"template_id":   {Type: "string", Description: "模板ID"},
				"project_code":  {Type: "string", Description: "项目编码"},
				"project_name":  {Type: "string", Description: "项目名称"},
				"start_date":    {Type: "string", Description: "开始日期 YYYY-MM-DD"},
				"pm_user_id":    {Type: "string", Description: "项目经理用户ID"},
				"skip_weekends": {Type: "boolean", Description: "是否跳过周末"},
			}, Required: []string{"template_id", "project_code", "project_name", "start_date", "pm_user_id"}},
		},
		{
			Name:        "plm_get_project",
			Description: "获取项目详情",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "项目ID"},
			}, Required: []string{"id"}},
		},

		// Tasks
		{
			Name:        "plm_list_project_tasks",
			Description: "列出项目的所有任务",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"project_id": {Type: "string", Description: "项目ID"},
			}, Required: []string{"project_id"}},
		},
		{
			Name:        "plm_complete_task",
			Description: "完成任务（触发自动化）",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"task_id": {Type: "string", Description: "任务ID"},
			}, Required: []string{"task_id"}},
		},
		{
			Name:        "plm_update_task_status",
			Description: "更新任务状态",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"task_id": {Type: "string", Description: "任务ID"},
				"status":  {Type: "string", Description: "新状态: pending/in_progress/completed/blocked"},
			}, Required: []string{"task_id", "status"}},
		},

		// Materials
		{
			Name:        "plm_list_materials",
			Description: "列出所有物料",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"category": {Type: "string", Description: "物料分类"},
				"search":   {Type: "string", Description: "搜索关键词"},
			}},
		},
		{
			Name:        "plm_create_material",
			Description: "创建新物料",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"code":          {Type: "string", Description: "物料编码"},
				"name":          {Type: "string", Description: "物料名称"},
				"specification": {Type: "string", Description: "规格"},
				"unit":          {Type: "string", Description: "单位"},
				"category":      {Type: "string", Description: "分类"},
			}, Required: []string{"code", "name", "unit"}},
		},

		// BOM
		{
			Name:        "plm_get_product_bom",
			Description: "获取产品BOM",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"product_id": {Type: "string", Description: "产品ID"},
			}, Required: []string{"product_id"}},
		},

		// ECN
		{
			Name:        "plm_list_ecns",
			Description: "列出所有ECN变更",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"status": {Type: "string", Description: "状态筛选"},
			}},
		},
		{
			Name:        "plm_create_ecn",
			Description: "创建ECN变更申请",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"title":       {Type: "string", Description: "ECN标题"},
				"description": {Type: "string", Description: "变更描述"},
				"type":        {Type: "string", Description: "类型: design/material/process"},
				"priority":    {Type: "string", Description: "优先级: low/medium/high/urgent"},
				"product_id":  {Type: "string", Description: "关联产品ID"},
			}, Required: []string{"title", "type"}},
		},

		// Documents
		{
			Name:        "plm_list_documents",
			Description: "列出所有文档",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"category": {Type: "string", Description: "文档分类"},
			}},
		},

		// Users
		{
			Name:        "plm_list_users",
			Description: "列出所有用户",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "plm_get_current_user",
			Description: "获取当前登录用户信息",
			InputSchema: InputSchema{Type: "object"},
		},

		// System
		{
			Name:        "plm_health_check",
			Description: "检查PLM系统健康状态",
			InputSchema: InputSchema{Type: "object"},
		},
	}
}

func (s *MCPServer) CallTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	// Products
	case "plm_list_products":
		path := "/api/v1/products"
		if status, ok := args["status"].(string); ok && status != "" {
			path += "?status=" + status
		}
		resp, err := s.plm.Request("GET", path, nil)
		return string(resp), err

	case "plm_create_product":
		resp, err := s.plm.Request("POST", "/api/v1/products", args)
		return string(resp), err

	case "plm_get_product":
		id := args["id"].(string)
		resp, err := s.plm.Request("GET", "/api/v1/products/"+id, nil)
		return string(resp), err

	// Templates
	case "plm_list_templates":
		path := "/api/v1/templates"
		if t, ok := args["type"].(string); ok && t != "" {
			path += "?type=" + t
		}
		resp, err := s.plm.Request("GET", path, nil)
		return string(resp), err

	case "plm_get_template":
		id := args["id"].(string)
		resp, err := s.plm.Request("GET", "/api/v1/templates/"+id, nil)
		return string(resp), err

	// Projects
	case "plm_list_projects":
		resp, err := s.plm.Request("GET", "/api/v1/projects", nil)
		return string(resp), err

	case "plm_create_project_from_template":
		resp, err := s.plm.Request("POST", "/api/v1/projects/create-from-template", args)
		return string(resp), err

	case "plm_get_project":
		id := args["id"].(string)
		resp, err := s.plm.Request("GET", "/api/v1/projects/"+id, nil)
		return string(resp), err

	// Tasks
	case "plm_list_project_tasks":
		projectID := args["project_id"].(string)
		resp, err := s.plm.Request("GET", "/api/v1/projects/"+projectID+"/tasks", nil)
		return string(resp), err

	case "plm_complete_task":
		taskID := args["task_id"].(string)
		resp, err := s.plm.Request("POST", "/api/v1/tasks/"+taskID+"/complete", nil)
		return string(resp), err

	case "plm_update_task_status":
		taskID := args["task_id"].(string)
		resp, err := s.plm.Request("PUT", "/api/v1/tasks/"+taskID+"/status", map[string]string{
			"status": args["status"].(string),
		})
		return string(resp), err

	// Materials
	case "plm_list_materials":
		resp, err := s.plm.Request("GET", "/api/v1/materials", nil)
		return string(resp), err

	case "plm_create_material":
		resp, err := s.plm.Request("POST", "/api/v1/materials", args)
		return string(resp), err

	// BOM
	case "plm_get_product_bom":
		productID := args["product_id"].(string)
		resp, err := s.plm.Request("GET", "/api/v1/products/"+productID+"/bom", nil)
		return string(resp), err

	// ECN
	case "plm_list_ecns":
		resp, err := s.plm.Request("GET", "/api/v1/ecns", nil)
		return string(resp), err

	case "plm_create_ecn":
		resp, err := s.plm.Request("POST", "/api/v1/ecns", args)
		return string(resp), err

	// Documents
	case "plm_list_documents":
		resp, err := s.plm.Request("GET", "/api/v1/documents", nil)
		return string(resp), err

	// Users
	case "plm_list_users":
		resp, err := s.plm.Request("GET", "/api/v1/users", nil)
		return string(resp), err

	case "plm_get_current_user":
		resp, err := s.plm.Request("GET", "/api/v1/auth/me", nil)
		return string(resp), err

	// System
	case "plm_health_check":
		resp, err := s.plm.Request("GET", "/health/live", nil)
		return string(resp), err

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *MCPServer) HandleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			ServerInfo: ServerInfo{
				Name:    "nimo-plm-mcp",
				Version: "1.0.0",
			},
		}

	case "notifications/initialized":
		// No response needed for notifications
		return nil

	case "tools/list":
		resp.Result = ToolsListResult{Tools: s.GetTools()}

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()}
			return resp
		}

		result, err := s.CallTool(params.Name, params.Arguments)
		if err != nil {
			resp.Result = ToolResult{
				Content: []ContentItem{{Type: "text", Text: "Error: " + err.Error()}},
				IsError: true,
			}
		} else {
			resp.Result = ToolResult{
				Content: []ContentItem{{Type: "text", Text: result}},
			}
		}

	default:
		resp.Error = &RPCError{Code: -32601, Message: "Method not found: " + req.Method}
	}

	return resp
}

func (s *MCPServer) Run() {
	reader := bufio.NewReader(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			continue
		}

		resp := s.HandleRequest(&req)
		if resp != nil {
			encoder.Encode(resp)
		}
	}
}

func main() {
	baseURL := os.Getenv("PLM_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}

	token := os.Getenv("PLM_TOKEN")

	server := NewMCPServer(baseURL, token)
	server.Run()
}
