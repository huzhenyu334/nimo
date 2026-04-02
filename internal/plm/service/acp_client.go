package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
)

// ACPClient 调用 ACP REST API 获取流程和实例数据
type ACPClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewACPClient 创建 ACP 客户端
func NewACPClient(baseURL, token string) *ACPClient {
	return &ACPClient{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// --- ACP 数据结构（仅 PLM 需要的字段）---

// ACPStepDef 流程步骤定义（从 YAML 解析）
type ACPStepDef struct {
	ID              string                `yaml:"id" json:"id"`
	Name            string                `yaml:"name,omitempty" json:"name,omitempty"`
	Executor        string                `yaml:"executor,omitempty" json:"executor,omitempty"`
	Control         string                `yaml:"control,omitempty" json:"control,omitempty"`
	DependsOn       []string              `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	PlannedDuration string                `yaml:"planned_duration,omitempty" json:"planned_duration,omitempty"`
	Type            string                `yaml:"type,omitempty" json:"type,omitempty"` // execution | approval
	Process         string                `yaml:"process,omitempty" json:"process,omitempty"` // subprocess 引用的流程 slug
	Input           map[string]any        `yaml:"input,omitempty" json:"input,omitempty"`

	// 控制节点字段
	ConditionExpr string            `yaml:"condition,omitempty" json:"condition,omitempty"`
	Expression    string            `yaml:"expression,omitempty" json:"expression,omitempty"`
	Mode          string            `yaml:"mode,omitempty" json:"mode,omitempty"`
	Items         string            `yaml:"items,omitempty" json:"items,omitempty"`

	// 嵌套：if/switch 的分支 + loop/foreach 的 body
	Branches map[string][]ACPStepDef `yaml:"branches,omitempty" json:"-"`
	Steps    []ACPStepDef            `yaml:"steps,omitempty" json:"steps,omitempty"`
}

// ACPWorkflowDef 流程定义（从 YAML 解析）
type ACPWorkflowDef struct {
	Name  string       `yaml:"name"`
	Steps []ACPStepDef `yaml:"steps"`
}

// ACPProcess ACP 流程 API 响应
type ACPProcess struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	YAMLContent string `json:"yaml_content"`
	Status      string `json:"status"`
	Version     int    `json:"version"`
}

// ACPTaskStatus ACP 任务执行状态
type ACPTaskStatus struct {
	ID          string     `json:"id"`
	InstanceID  string     `json:"instance_id"`
	StepID      string     `json:"step_id"`
	Status      string     `json:"status"`
	Type        string     `json:"type"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Round       int        `json:"round"`
	IsCurrent   bool       `json:"is_current"`
	Error       *string    `json:"error"`
}

// ACPInstance ACP 流程实例
type ACPInstance struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	ProcessID        string     `json:"process_id"`
	Status           string     `json:"status"`
	ParentInstanceID string     `json:"parent_instance_id"`
	ParentStepName   string     `json:"parent_step_name"`
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

// --- API 方法 ---

// GetProcessDef 获取流程定义并解析 YAML 得到 steps
func (c *ACPClient) GetProcessDef(processID string) (*ACPWorkflowDef, string, error) {
	url := fmt.Sprintf("%s/api/processes/%s", c.BaseURL, processID)

	body, err := c.doGet(url)
	if err != nil {
		return nil, "", fmt.Errorf("get process: %w", err)
	}

	var resp struct {
		Process ACPProcess `json:"process"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, "", fmt.Errorf("unmarshal process: %w", err)
	}

	// 解析 YAML
	var def ACPWorkflowDef
	if err := yaml.Unmarshal([]byte(resp.Process.YAMLContent), &def); err != nil {
		return nil, "", fmt.Errorf("parse yaml: %w", err)
	}

	return &def, resp.Process.Name, nil
}

// GetInstanceTasks 获取流程实例下所有 task 的执行状态
func (c *ACPClient) GetInstanceTasks(instanceID string) ([]ACPTaskStatus, error) {
	url := fmt.Sprintf("%s/api/tasks?instance_id=%s&page_size=500", c.BaseURL, instanceID)

	body, err := c.doGet(url)
	if err != nil {
		return nil, fmt.Errorf("get instance tasks: %w", err)
	}

	var resp struct {
		Tasks []ACPTaskStatus `json:"tasks"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}

	return resp.Tasks, nil
}

// GetProcessBySlug 通过 slug 获取流程定义（用于 subprocess 递归）
func (c *ACPClient) GetProcessBySlug(slug string) (*ACPWorkflowDef, error) {
	url := fmt.Sprintf("%s/api/processes?keyword=%s&page_size=1", c.BaseURL, slug)

	body, err := c.doGet(url)
	if err != nil {
		return nil, fmt.Errorf("search process by slug: %w", err)
	}

	var resp struct {
		Processes []ACPProcess `json:"processes"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal processes: %w", err)
	}

	if len(resp.Processes) == 0 {
		return nil, fmt.Errorf("process not found: %s", slug)
	}

	var def ACPWorkflowDef
	if err := yaml.Unmarshal([]byte(resp.Processes[0].YAMLContent), &def); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	return &def, nil
}

// GetProcessDefByID 通过流程 ID 获取定义（subprocess 常用 ID 引用而非 slug）
func (c *ACPClient) GetProcessDefByID(processID string) (*ACPWorkflowDef, string, error) {
	return c.GetProcessDef(processID)
}

// GetSubprocessInstances 获取某实例下特定 step 的子流程实例
func (c *ACPClient) GetSubprocessInstances(parentInstanceID string) ([]ACPInstance, error) {
	url := fmt.Sprintf("%s/api/instances?parent_instance_id=%s&page_size=100", c.BaseURL, parentInstanceID)

	body, err := c.doGet(url)
	if err != nil {
		return nil, fmt.Errorf("get subprocess instances: %w", err)
	}

	var resp struct {
		Runs []ACPInstance `json:"runs"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal runs: %w", err)
	}

	return resp.Runs, nil
}

// SubprocessProcessID 从 step 中提取子流程 ID
// 优先使用 step.Process 字段，其次检查 input.process
func SubprocessProcessID(step ACPStepDef) string {
	if step.Process != "" {
		return step.Process
	}
	if step.Input != nil {
		if pid, ok := step.Input["process"]; ok {
			if s, ok := pid.(string); ok {
				return s
			}
		}
	}
	return ""
}

// doGet 执行 GET 请求
func (c *ACPClient) doGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ACP API returned %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
