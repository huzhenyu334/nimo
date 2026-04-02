package service

import (
	"fmt"
	"strings"
	"time"
)

// GanttNode 甘特图节点（返回给前端）
type GanttNode struct {
	ID                string      `json:"id"`
	Label             string      `json:"label"`
	Status            string      `json:"status"`
	Depth             int         `json:"depth"`
	Children          []GanttNode `json:"children"`
	StartedAt         *int64      `json:"started_at,omitempty"`
	CompletedAt       *int64      `json:"completed_at,omitempty"`
	DurationMs        *int64      `json:"duration_ms,omitempty"`
	PlannedDurationMs *int64      `json:"planned_duration_ms,omitempty"`
	IsMilestone       bool        `json:"isMilestone,omitempty"`
	IsConditional     bool        `json:"isConditional,omitempty"`
	IsRepeating       bool        `json:"isRepeating,omitempty"`
	Executor          string      `json:"executor,omitempty"`
	Assignee          string      `json:"assignee,omitempty"`
	StepType          string      `json:"step_type,omitempty"`
}

// GanttResponse 甘特图 API 响应
type GanttResponse struct {
	Nodes          []GanttNode `json:"nodes"`
	Mode           string      `json:"mode"`
	ProcessName    string      `json:"process_name"`
	InstanceStatus string      `json:"instance_status,omitempty"`
}

// GanttFilterService 从 ACP 流程中过滤出甘特图节点
type GanttFilterService struct {
	acpClient *ACPClient
}

// NewGanttFilterService 创建甘特图过滤服务
func NewGanttFilterService(acpClient *ACPClient) *GanttFilterService {
	return &GanttFilterService{acpClient: acpClient}
}

// executor 可见性默认规则
var defaultVisibleExecutors = map[string]bool{
	"human":      true,
	"agent":      true,
	"plm":        true,
	"user":       true,
	"subprocess": true, // subprocess 单独处理，这里标记为可见确保不被跳过
}

// BuildProjectGantt 构建项目甘特图数据
func (s *GanttFilterService) BuildProjectGantt(processID, instanceID string) (*GanttResponse, error) {
	// 1. 获取流程定义
	def, processName, err := s.acpClient.GetProcessDef(processID)
	if err != nil {
		return nil, fmt.Errorf("get process def: %w", err)
	}

	// 2. 递归过滤 steps
	nodes := s.filterSteps(def.Steps, 0, false, false, map[string]bool{processID: true})

	// 3. 修复依赖链
	s.fixDependencies(nodes, def.Steps)

	// 4. 标记里程碑
	s.detectMilestones(nodes)

	resp := &GanttResponse{
		Nodes:       nodes,
		Mode:        "plan",
		ProcessName: processName,
	}

	// 5. 如果有运行中实例，合并执行状态
	if instanceID != "" {
		tasks, err := s.acpClient.GetInstanceTasks(instanceID)
		if err == nil && len(tasks) > 0 {
			s.mergeTaskStatus(resp.Nodes, tasks)
			resp.Mode = "execution"

			// 获取 subprocess 实例并合并其 tasks
			subInstances, subErr := s.acpClient.GetSubprocessInstances(instanceID)
			if subErr == nil && len(subInstances) > 0 {
				s.mergeSubprocessInstances(resp.Nodes, subInstances)
			}

			// 尝试判断实例状态
			allCompleted := true
			hasRunning := false
			for _, t := range tasks {
				if t.Status == "running" {
					hasRunning = true
				}
				if t.Status != "completed" && t.Status != "skipped" {
					allCompleted = false
				}
			}
			if allCompleted {
				resp.InstanceStatus = "completed"
			} else if hasRunning {
				resp.InstanceStatus = "running"
			} else {
				resp.InstanceStatus = "pending"
			}
		}
	}

	return resp, nil
}

// filterSteps 递归过滤步骤
func (s *GanttFilterService) filterSteps(steps []ACPStepDef, depth int, conditional, repeating bool, visited map[string]bool) []GanttNode {
	var result []GanttNode

	for _, step := range steps {
		// 控制节点：跳过自身，递归处理子步骤
		if step.Control != "" {
			switch step.Control {
			case "if", "switch":
				for _, branchSteps := range step.Branches {
					children := s.filterSteps(branchSteps, depth, true, repeating, visited)
					result = append(result, children...)
				}
			case "loop", "foreach":
				if len(step.Steps) > 0 {
					children := s.filterSteps(step.Steps, depth, conditional, true, visited)
					result = append(result, children...)
				}
			case "terminate":
				// 跳过
			}
			continue
		}

		// subprocess：递归展开子流程
		if step.Executor == "subprocess" {
			subProcessID := SubprocessProcessID(step)
			if subProcessID != "" && !visited[subProcessID] {
				visited[subProcessID] = true
				subDef, _, err := s.acpClient.GetProcessDefByID(subProcessID)
				if err == nil && len(subDef.Steps) > 0 {
					children := s.filterSteps(subDef.Steps, depth+1, conditional, repeating, visited)
					node := GanttNode{
						ID:            step.ID,
						Label:         stepLabel(step),
						Status:        "pending",
						Depth:         depth,
						Children:      children,
						IsConditional: conditional,
						IsRepeating:   repeating,
						Executor:      "subprocess",
					}
					if step.PlannedDuration != "" {
						if ms := parseDurationMs(step.PlannedDuration); ms > 0 {
							node.PlannedDurationMs = &ms
						}
					}
					result = append(result, node)
					continue
				}
			}
			// 没有子流程ID或获取失败：显示为普通节点
			node := GanttNode{
				ID:            step.ID,
				Label:         stepLabel(step),
				Status:        "pending",
				Depth:         depth,
				Children:      []GanttNode{},
				IsConditional: conditional,
				IsRepeating:   repeating,
				Executor:      "subprocess",
			}
			result = append(result, node)
			continue
		}

		// executor 可见性判断
		if !defaultVisibleExecutors[step.Executor] {
			continue
		}

		node := GanttNode{
			ID:            step.ID,
			Label:         stepLabel(step),
			Status:        "pending",
			Depth:         depth,
			Children:      []GanttNode{},
			IsConditional: conditional,
			IsRepeating:   repeating,
			Executor:      step.Executor,
			StepType:      step.Type,
		}

		if step.PlannedDuration != "" {
			if ms := parseDurationMs(step.PlannedDuration); ms > 0 {
				node.PlannedDurationMs = &ms
			}
		}

		result = append(result, node)
	}

	return result
}

// fixDependencies 修复被过滤节点导致的依赖链断裂
func (s *GanttFilterService) fixDependencies(nodes []GanttNode, originalSteps []ACPStepDef) {
	// 构建可见节点集合
	visibleSet := make(map[string]bool)
	collectVisibleIDs(nodes, visibleSet)

	// 构建原始步骤的 depends_on 查找表
	stepDeps := make(map[string][]string)
	collectAllDeps(originalSteps, stepDeps)

	// 对每个可见节点，穿透不可见的依赖
	// 目前不在 GanttNode 上维护 depends_on 字段，
	// 因为前端甘特图组件是通过树形结构而非 depends_on 来渲染的。
	// 如需要，后续可以添加 DependsOn []string 字段。
	_ = visibleSet
	_ = stepDeps
}

// detectMilestones 标记里程碑（approval 类型 = 里程碑）
func (s *GanttFilterService) detectMilestones(nodes []GanttNode) {
	for i := range nodes {
		if nodes[i].StepType == "approval" {
			nodes[i].IsMilestone = true
		}
		if len(nodes[i].Children) > 0 {
			s.detectMilestones(nodes[i].Children)
		}
	}
}

// mergeTaskStatus 合并执行状态到甘特图节点
func (s *GanttFilterService) mergeTaskStatus(nodes []GanttNode, tasks []ACPTaskStatus) {
	// 构建 step_id → tasks 映射
	taskMap := make(map[string][]ACPTaskStatus)
	for _, t := range tasks {
		taskMap[t.StepID] = append(taskMap[t.StepID], t)
	}

	for i := range nodes {
		s.mergeNodeStatus(&nodes[i], taskMap)
	}
}

func (s *GanttFilterService) mergeNodeStatus(node *GanttNode, taskMap map[string][]ACPTaskStatus) {
	tasks := taskMap[node.ID]

	if len(tasks) > 0 {
		// 循环体步骤可能有多个 task（多轮），按 round 分组
		if node.IsRepeating && len(tasks) > 1 {
			node.Children = s.buildIterationChildren(node, tasks)
		} else {
			// 取最新的 current task
			var current *ACPTaskStatus
			for j := range tasks {
				if tasks[j].IsCurrent {
					current = &tasks[j]
					break
				}
			}
			if current == nil {
				current = &tasks[len(tasks)-1]
			}
			node.Status = current.Status
			if current.StartedAt != nil {
				ms := current.StartedAt.UnixMilli()
				node.StartedAt = &ms
			}
			if current.CompletedAt != nil {
				ms := current.CompletedAt.UnixMilli()
				node.CompletedAt = &ms
			}
			if node.StartedAt != nil && node.CompletedAt != nil {
				dur := *node.CompletedAt - *node.StartedAt
				node.DurationMs = &dur
			}
		}
	}

	// 递归处理子节点
	for i := range node.Children {
		s.mergeNodeStatus(&node.Children[i], taskMap)
	}
}

// buildIterationChildren 为循环步骤构建按轮次的子节点
func (s *GanttFilterService) buildIterationChildren(node *GanttNode, tasks []ACPTaskStatus) []GanttNode {
	// 按 round 分组
	roundMap := make(map[int][]ACPTaskStatus)
	for _, t := range tasks {
		roundMap[t.Round] = append(roundMap[t.Round], t)
	}

	var children []GanttNode
	for round := 1; round <= len(roundMap); round++ {
		roundTasks := roundMap[round]
		if len(roundTasks) == 0 {
			continue
		}

		t := roundTasks[0] // 每轮只有一个 task（每个 step 每轮一个）
		child := GanttNode{
			ID:       fmt.Sprintf("%s-round-%d", node.ID, round),
			Label:    fmt.Sprintf("%s (第%d轮)", node.Label, round),
			Status:   t.Status,
			Depth:    node.Depth + 1,
			Children: []GanttNode{},
			Executor: node.Executor,
			StepType: node.StepType,
		}
		if t.StartedAt != nil {
			ms := t.StartedAt.UnixMilli()
			child.StartedAt = &ms
		}
		if t.CompletedAt != nil {
			ms := t.CompletedAt.UnixMilli()
			child.CompletedAt = &ms
		}
		if child.StartedAt != nil && child.CompletedAt != nil {
			dur := *child.CompletedAt - *child.StartedAt
			child.DurationMs = &dur
		}
		if node.IsMilestone {
			child.IsMilestone = true
		}
		children = append(children, child)
	}

	// 更新父节点状态：取最新轮的状态
	if len(children) > 0 {
		last := children[len(children)-1]
		node.Status = last.Status
		if len(children) > 0 {
			node.StartedAt = children[0].StartedAt
		}
		node.CompletedAt = last.CompletedAt
	}

	return children
}

// mergeSubprocessInstances 合并子流程实例状态到 subprocess 节点
func (s *GanttFilterService) mergeSubprocessInstances(nodes []GanttNode, subInstances []ACPInstance) {
	// 按 parent_step_name 分组
	stepInstMap := make(map[string][]ACPInstance)
	for _, inst := range subInstances {
		stepInstMap[inst.ParentStepName] = append(stepInstMap[inst.ParentStepName], inst)
	}

	for i := range nodes {
		if nodes[i].Executor == "subprocess" {
			instances := stepInstMap[nodes[i].ID]
			if len(instances) > 0 {
				inst := instances[0] // 通常一个 step 只有一个子实例
				nodes[i].Status = inst.Status
				if inst.StartedAt != nil {
					ms := inst.StartedAt.UnixMilli()
					nodes[i].StartedAt = &ms
				}
				if inst.CompletedAt != nil {
					ms := inst.CompletedAt.UnixMilli()
					nodes[i].CompletedAt = &ms
				}
				if nodes[i].StartedAt != nil && nodes[i].CompletedAt != nil {
					dur := *nodes[i].CompletedAt - *nodes[i].StartedAt
					nodes[i].DurationMs = &dur
				}

				// 获取子实例的 tasks，合并到 subprocess 的 children 中
				subTasks, err := s.acpClient.GetInstanceTasks(inst.ID)
				if err == nil && len(subTasks) > 0 && len(nodes[i].Children) > 0 {
					s.mergeTaskStatus(nodes[i].Children, subTasks)
				}
			}
		}
		// 递归处理嵌套子节点
		if len(nodes[i].Children) > 0 {
			s.mergeSubprocessInstances(nodes[i].Children, subInstances)
		}
	}
}

// --- 辅助函数 ---

func stepLabel(step ACPStepDef) string {
	if step.Name != "" {
		return step.Name
	}
	return step.ID
}

func parseDurationMs(s string) int64 {
	// 支持 Go duration 格式: "24h", "30m", "2h30m"
	// 以及自定义格式: "14d", "2w"
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// 自定义：天和周
	if strings.HasSuffix(s, "d") {
		s = strings.TrimSuffix(s, "d")
		var days float64
		if _, err := fmt.Sscanf(s, "%f", &days); err == nil {
			return int64(days * 24 * float64(time.Hour) / float64(time.Millisecond))
		}
	}
	if strings.HasSuffix(s, "w") {
		s = strings.TrimSuffix(s, "w")
		var weeks float64
		if _, err := fmt.Sscanf(s, "%f", &weeks); err == nil {
			return int64(weeks * 7 * 24 * float64(time.Hour) / float64(time.Millisecond))
		}
	}

	// Go 标准 duration
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d.Milliseconds()
}

func collectVisibleIDs(nodes []GanttNode, set map[string]bool) {
	for _, n := range nodes {
		set[n.ID] = true
		collectVisibleIDs(n.Children, set)
	}
}

func collectAllDeps(steps []ACPStepDef, deps map[string][]string) {
	for _, step := range steps {
		if len(step.DependsOn) > 0 {
			deps[step.ID] = step.DependsOn
		}
		// 递归收集分支和body中的步骤
		for _, branchSteps := range step.Branches {
			collectAllDeps(branchSteps, deps)
		}
		if len(step.Steps) > 0 {
			collectAllDeps(step.Steps, deps)
		}
	}
}
