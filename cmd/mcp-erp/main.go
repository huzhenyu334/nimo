package main

import (
	"bufio"
	"encoding/base64"
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
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
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

// ERP API client
type ERPClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewERPClient(baseURL, token string) *ERPClient {
	return &ERPClient{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *ERPClient) Request(method, path string, body interface{}) ([]byte, error) {
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
	erp   *ERPClient
	token string
}

func NewMCPServer(erpBaseURL, token string) *MCPServer {
	return &MCPServer{
		erp:   NewERPClient(erpBaseURL, token),
		token: token,
	}
}

func (s *MCPServer) GetTools() []Tool {
	return []Tool{
		// ==================== 供应商管理 ====================
		{
			Name:        "erp_list_suppliers",
			Description: "查询供应商列表",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"search": {Type: "string", Description: "搜索关键词"},
			}},
		},
		{
			Name:        "erp_create_supplier",
			Description: "创建供应商",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"name":           {Type: "string", Description: "供应商名称"},
				"contact_person": {Type: "string", Description: "联系人"},
				"phone":          {Type: "string", Description: "电话"},
				"email":          {Type: "string", Description: "邮箱"},
				"address":        {Type: "string", Description: "地址"},
			}, Required: []string{"name"}},
		},
		{
			Name:        "erp_get_supplier",
			Description: "获取供应商详情",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "供应商ID"},
			}, Required: []string{"id"}},
		},
		{
			Name:        "erp_update_supplier",
			Description: "更新供应商信息",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id":             {Type: "string", Description: "供应商ID"},
				"name":           {Type: "string", Description: "供应商名称"},
				"contact_person": {Type: "string", Description: "联系人"},
				"phone":          {Type: "string", Description: "电话"},
				"email":          {Type: "string", Description: "邮箱"},
				"address":        {Type: "string", Description: "地址"},
			}, Required: []string{"id"}},
		},

		// ==================== 采购管理 ====================
		{
			Name:        "erp_list_purchase_orders",
			Description: "查询采购订单列表",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"status": {Type: "string", Description: "状态筛选: draft/submitted/approved/received"},
			}},
		},
		{
			Name:        "erp_create_purchase_order",
			Description: "创建采购订单",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"supplier_id":   {Type: "string", Description: "供应商ID"},
				"items":         {Type: "string", Description: "订单项JSON数组，如 [{\"material_id\":\"xxx\",\"quantity\":100,\"unit_price\":10.5}]"},
				"delivery_date": {Type: "string", Description: "交货日期 YYYY-MM-DD"},
				"notes":         {Type: "string", Description: "备注"},
			}, Required: []string{"supplier_id", "items"}},
		},
		{
			Name:        "erp_get_purchase_order",
			Description: "获取采购订单详情",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "采购订单ID"},
			}, Required: []string{"id"}},
		},
		{
			Name:        "erp_submit_po",
			Description: "提交采购订单审批",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "采购订单ID"},
			}, Required: []string{"id"}},
		},
		{
			Name:        "erp_approve_po",
			Description: "审批采购订单（通过）",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id":      {Type: "string", Description: "采购订单ID"},
				"comment": {Type: "string", Description: "审批意见"},
			}, Required: []string{"id"}},
		},
		{
			Name:        "erp_receive_po",
			Description: "采购收货",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id":    {Type: "string", Description: "采购订单ID"},
				"items": {Type: "string", Description: "收货项JSON数组，如 [{\"material_id\":\"xxx\",\"quantity\":100}]"},
			}, Required: []string{"id"}},
		},

		// ==================== 库存管理 ====================
		{
			Name:        "erp_query_inventory",
			Description: "查询库存",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"material_id": {Type: "string", Description: "物料ID（可选，不填则查全部）"},
			}},
		},
		{
			Name:        "erp_inbound",
			Description: "入库操作",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"material_id": {Type: "string", Description: "物料ID"},
				"quantity":    {Type: "number", Description: "数量"},
				"warehouse":   {Type: "string", Description: "仓库"},
				"reference":   {Type: "string", Description: "参考单号"},
				"notes":       {Type: "string", Description: "备注"},
			}, Required: []string{"material_id", "quantity"}},
		},
		{
			Name:        "erp_outbound",
			Description: "出库操作",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"material_id": {Type: "string", Description: "物料ID"},
				"quantity":    {Type: "number", Description: "数量"},
				"warehouse":   {Type: "string", Description: "仓库"},
				"reference":   {Type: "string", Description: "参考单号"},
				"notes":       {Type: "string", Description: "备注"},
			}, Required: []string{"material_id", "quantity"}},
		},
		{
			Name:        "erp_inventory_alerts",
			Description: "库存预警查询",
			InputSchema: InputSchema{Type: "object"},
		},

		// ==================== MRP ====================
		{
			Name:        "erp_run_mrp",
			Description: "执行MRP计算（物料需求计划）",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"product_id": {Type: "string", Description: "产品ID"},
				"quantity":   {Type: "number", Description: "需求数量"},
				"due_date":   {Type: "string", Description: "需求日期 YYYY-MM-DD"},
			}, Required: []string{"product_id", "quantity", "due_date"}},
		},
		{
			Name:        "erp_get_mrp_result",
			Description: "查询MRP计算结果",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"run_id": {Type: "string", Description: "MRP运行ID"},
			}, Required: []string{"run_id"}},
		},
		{
			Name:        "erp_apply_mrp",
			Description: "确认MRP结果（自动创建采购需求和工单）",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"run_id": {Type: "string", Description: "MRP运行ID"},
			}, Required: []string{"run_id"}},
		},
		{
			Name:        "erp_list_mrp_runs",
			Description: "查询MRP运行历史",
			InputSchema: InputSchema{Type: "object"},
		},

		// ==================== 生产管理 ====================
		{
			Name:        "erp_list_work_orders",
			Description: "查询工单列表",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"status": {Type: "string", Description: "状态筛选: planned/released/in_progress/completed"},
			}},
		},
		{
			Name:        "erp_create_work_order",
			Description: "创建生产工单",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"product_id":    {Type: "string", Description: "产品ID"},
				"quantity":      {Type: "number", Description: "计划数量"},
				"planned_start": {Type: "string", Description: "计划开始日期 YYYY-MM-DD"},
				"planned_end":   {Type: "string", Description: "计划结束日期 YYYY-MM-DD"},
				"priority":      {Type: "string", Description: "优先级: low/normal/high/urgent"},
				"notes":         {Type: "string", Description: "备注"},
			}, Required: []string{"product_id", "quantity"}},
		},
		{
			Name:        "erp_get_work_order",
			Description: "获取工单详情",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "工单ID"},
			}, Required: []string{"id"}},
		},
		{
			Name:        "erp_release_work_order",
			Description: "下达工单（开始生产）",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "工单ID"},
			}, Required: []string{"id"}},
		},
		{
			Name:        "erp_report_work",
			Description: "生产报工",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id":                 {Type: "string", Description: "工单ID"},
				"quantity_completed": {Type: "number", Description: "完成数量"},
				"notes":             {Type: "string", Description: "备注"},
			}, Required: []string{"id", "quantity_completed"}},
		},
		{
			Name:        "erp_complete_work_order",
			Description: "完工（结束生产）",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "工单ID"},
			}, Required: []string{"id"}},
		},

		// ==================== 销售管理 ====================
		{
			Name:        "erp_list_sales_orders",
			Description: "查询销售订单列表",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"status": {Type: "string", Description: "状态筛选: draft/confirmed/shipped/completed"},
			}},
		},
		{
			Name:        "erp_create_sales_order",
			Description: "创建销售订单",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"customer_id":   {Type: "string", Description: "客户ID"},
				"items":         {Type: "string", Description: "订单项JSON数组，如 [{\"product_id\":\"xxx\",\"quantity\":10,\"unit_price\":100}]"},
				"delivery_date": {Type: "string", Description: "交货日期 YYYY-MM-DD"},
				"notes":         {Type: "string", Description: "备注"},
			}, Required: []string{"customer_id", "items"}},
		},
		{
			Name:        "erp_confirm_sales_order",
			Description: "确认销售订单",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "销售订单ID"},
			}, Required: []string{"id"}},
		},
		{
			Name:        "erp_ship_sales_order",
			Description: "销售订单发货",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"id": {Type: "string", Description: "销售订单ID"},
			}, Required: []string{"id"}},
		},

		// ==================== 售后管理 ====================
		{
			Name:        "erp_list_service_orders",
			Description: "查询服务工单列表",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "erp_create_service_order",
			Description: "创建服务工单",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{
				"customer_id": {Type: "string", Description: "客户ID"},
				"type":        {Type: "string", Description: "服务类型: warranty/repair/complaint/return"},
				"description": {Type: "string", Description: "问题描述"},
				"priority":    {Type: "string", Description: "优先级: low/normal/high/urgent"},
			}, Required: []string{"customer_id", "type", "description"}},
		},

		// ==================== 通用 ====================
		{
			Name:        "erp_health_check",
			Description: "检查ERP系统健康状态",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "erp_get_current_user",
			Description: "获取当前登录用户信息（从JWT Token解析）",
			InputSchema: InputSchema{Type: "object"},
		},
	}
}

func (s *MCPServer) CallTool(name string, args map[string]interface{}) (string, error) {
	apiPrefix := "/api/v1/erp"

	switch name {
	// ==================== 供应商管理 ====================
	case "erp_list_suppliers":
		path := apiPrefix + "/suppliers"
		if search, ok := args["search"].(string); ok && search != "" {
			path += "?search=" + search
		}
		resp, err := s.erp.Request("GET", path, nil)
		return string(resp), err

	case "erp_create_supplier":
		resp, err := s.erp.Request("POST", apiPrefix+"/suppliers", args)
		return string(resp), err

	case "erp_get_supplier":
		id := args["id"].(string)
		resp, err := s.erp.Request("GET", apiPrefix+"/suppliers/"+id, nil)
		return string(resp), err

	case "erp_update_supplier":
		id := args["id"].(string)
		// Remove id from args so it's not sent in body
		body := make(map[string]interface{})
		for k, v := range args {
			if k != "id" {
				body[k] = v
			}
		}
		resp, err := s.erp.Request("PUT", apiPrefix+"/suppliers/"+id, body)
		return string(resp), err

	// ==================== 采购管理 ====================
	case "erp_list_purchase_orders":
		path := apiPrefix + "/purchase-orders"
		if status, ok := args["status"].(string); ok && status != "" {
			path += "?status=" + status
		}
		resp, err := s.erp.Request("GET", path, nil)
		return string(resp), err

	case "erp_create_purchase_order":
		// Parse items from JSON string to actual array
		body := make(map[string]interface{})
		for k, v := range args {
			if k == "items" {
				if itemsStr, ok := v.(string); ok {
					var items []interface{}
					if err := json.Unmarshal([]byte(itemsStr), &items); err == nil {
						body[k] = items
						continue
					}
				}
			}
			body[k] = v
		}
		resp, err := s.erp.Request("POST", apiPrefix+"/purchase-orders", body)
		return string(resp), err

	case "erp_get_purchase_order":
		id := args["id"].(string)
		resp, err := s.erp.Request("GET", apiPrefix+"/purchase-orders/"+id, nil)
		return string(resp), err

	case "erp_submit_po":
		id := args["id"].(string)
		resp, err := s.erp.Request("POST", apiPrefix+"/purchase-orders/"+id+"/submit", nil)
		return string(resp), err

	case "erp_approve_po":
		id := args["id"].(string)
		body := make(map[string]interface{})
		if comment, ok := args["comment"].(string); ok {
			body["comment"] = comment
		}
		resp, err := s.erp.Request("POST", apiPrefix+"/purchase-orders/"+id+"/approve", body)
		return string(resp), err

	case "erp_receive_po":
		id := args["id"].(string)
		body := make(map[string]interface{})
		if itemsStr, ok := args["items"].(string); ok {
			var items []interface{}
			if err := json.Unmarshal([]byte(itemsStr), &items); err == nil {
				body["items"] = items
			}
		}
		resp, err := s.erp.Request("POST", apiPrefix+"/purchase-orders/"+id+"/receive", body)
		return string(resp), err

	// ==================== 库存管理 ====================
	case "erp_query_inventory":
		path := apiPrefix + "/inventory"
		if materialID, ok := args["material_id"].(string); ok && materialID != "" {
			path += "?material_id=" + materialID
		}
		resp, err := s.erp.Request("GET", path, nil)
		return string(resp), err

	case "erp_inbound":
		resp, err := s.erp.Request("POST", apiPrefix+"/inventory/inbound", args)
		return string(resp), err

	case "erp_outbound":
		resp, err := s.erp.Request("POST", apiPrefix+"/inventory/outbound", args)
		return string(resp), err

	case "erp_inventory_alerts":
		resp, err := s.erp.Request("GET", apiPrefix+"/inventory/alerts", nil)
		return string(resp), err

	// ==================== MRP ====================
	case "erp_run_mrp":
		resp, err := s.erp.Request("POST", apiPrefix+"/mrp/run", args)
		return string(resp), err

	case "erp_get_mrp_result":
		runID := args["run_id"].(string)
		resp, err := s.erp.Request("GET", apiPrefix+"/mrp/result?run_id="+runID, nil)
		return string(resp), err

	case "erp_apply_mrp":
		resp, err := s.erp.Request("POST", apiPrefix+"/mrp/apply", args)
		return string(resp), err

	case "erp_list_mrp_runs":
		resp, err := s.erp.Request("GET", apiPrefix+"/mrp/runs", nil)
		return string(resp), err

	// ==================== 生产管理 ====================
	case "erp_list_work_orders":
		path := apiPrefix + "/work-orders"
		if status, ok := args["status"].(string); ok && status != "" {
			path += "?status=" + status
		}
		resp, err := s.erp.Request("GET", path, nil)
		return string(resp), err

	case "erp_create_work_order":
		resp, err := s.erp.Request("POST", apiPrefix+"/work-orders", args)
		return string(resp), err

	case "erp_get_work_order":
		id := args["id"].(string)
		resp, err := s.erp.Request("GET", apiPrefix+"/work-orders/"+id, nil)
		return string(resp), err

	case "erp_release_work_order":
		id := args["id"].(string)
		resp, err := s.erp.Request("POST", apiPrefix+"/work-orders/"+id+"/release", nil)
		return string(resp), err

	case "erp_report_work":
		id := args["id"].(string)
		body := make(map[string]interface{})
		if qty, ok := args["quantity_completed"]; ok {
			body["quantity_completed"] = qty
		}
		if notes, ok := args["notes"].(string); ok {
			body["notes"] = notes
		}
		resp, err := s.erp.Request("POST", apiPrefix+"/work-orders/"+id+"/report", body)
		return string(resp), err

	case "erp_complete_work_order":
		id := args["id"].(string)
		resp, err := s.erp.Request("POST", apiPrefix+"/work-orders/"+id+"/complete", nil)
		return string(resp), err

	// ==================== 销售管理 ====================
	case "erp_list_sales_orders":
		path := apiPrefix + "/sales-orders"
		if status, ok := args["status"].(string); ok && status != "" {
			path += "?status=" + status
		}
		resp, err := s.erp.Request("GET", path, nil)
		return string(resp), err

	case "erp_create_sales_order":
		body := make(map[string]interface{})
		for k, v := range args {
			if k == "items" {
				if itemsStr, ok := v.(string); ok {
					var items []interface{}
					if err := json.Unmarshal([]byte(itemsStr), &items); err == nil {
						body[k] = items
						continue
					}
				}
			}
			body[k] = v
		}
		resp, err := s.erp.Request("POST", apiPrefix+"/sales-orders", body)
		return string(resp), err

	case "erp_confirm_sales_order":
		id := args["id"].(string)
		resp, err := s.erp.Request("POST", apiPrefix+"/sales-orders/"+id+"/confirm", nil)
		return string(resp), err

	case "erp_ship_sales_order":
		id := args["id"].(string)
		resp, err := s.erp.Request("POST", apiPrefix+"/sales-orders/"+id+"/ship", nil)
		return string(resp), err

	// ==================== 售后管理 ====================
	case "erp_list_service_orders":
		resp, err := s.erp.Request("GET", apiPrefix+"/service-orders", nil)
		return string(resp), err

	case "erp_create_service_order":
		resp, err := s.erp.Request("POST", apiPrefix+"/service-orders", args)
		return string(resp), err

	// ==================== 通用 ====================
	case "erp_health_check":
		// Health check is at root, not under /api/v1/erp
		resp, err := s.erp.Request("GET", "/health/live", nil)
		return string(resp), err

	case "erp_get_current_user":
		return s.parseCurrentUser()

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// parseCurrentUser decodes user info from JWT token payload
func (s *MCPServer) parseCurrentUser() (string, error) {
	if s.token == "" {
		return "", fmt.Errorf("no token configured")
	}

	parts := strings.Split(s.token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT token format")
	}

	// Decode the payload (second part)
	payload := parts[1]
	// Add padding if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// Try standard encoding
		decoded, err = base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return "", fmt.Errorf("failed to decode JWT payload: %v", err)
		}
	}

	// Pretty-print the JSON
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(decoded, &prettyJSON); err != nil {
		return string(decoded), nil
	}

	result, err := json.MarshalIndent(prettyJSON, "", "  ")
	if err != nil {
		return string(decoded), nil
	}

	return string(result), nil
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
				Name:    "nimo-erp-mcp",
				Version: "1.0.0",
			},
		}

	case "notifications/initialized":
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
	baseURL := os.Getenv("ERP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8081"
	}

	token := os.Getenv("ERP_TOKEN")

	server := NewMCPServer(baseURL, token)
	server.Run()
}
