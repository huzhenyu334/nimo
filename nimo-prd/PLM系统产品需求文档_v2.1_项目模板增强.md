# PLM系统产品需求文档 v2.1 - 项目模板与飞书深度集成增强

**版本**: v2.1  
**日期**: 2026-02-06  
**变更**: 项目模板管理、任务自动化、飞书深度集成增强  
**作者**: Claude (AI Assistant)

---

## 变更说明

本文档是对 v2.0 的增强补充，主要新增以下功能：

1. **项目任务模板管理** - 模板CRUD、模板继承、版本管理
2. **从模板创建项目** - 一键从模板复制任务结构
3. **任务自动化引擎** - 前序任务完成自动开启后续任务
4. **飞书深度集成** - 任务双向同步、审批流程、评审会议

---

## 一、项目任务模板管理

### 1.1 模板层级结构

```
模板库
├── 系统模板（预置，只读）
│   ├── 智能眼镜标准研发流程
│   ├── 配件快速开发流程
│   └── 平台升级流程
└── 自定义模板（可编辑）
    ├── nimo Air 系列模板
    └── nimo Pro 系列模板
```

### 1.2 模板数据模型

#### 1.2.1 项目模板表 project_templates

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| code | VARCHAR(50) | 是 | 模板编码（唯一） |
| name | VARCHAR(200) | 是 | 模板名称 |
| description | TEXT | 否 | 模板描述 |
| template_type | ENUM | 是 | SYSTEM（系统）/ CUSTOM（自定义） |
| product_type | ENUM | 否 | 适用产品类型：GLASSES/ACCESSORY/PLATFORM |
| phases | JSONB | 是 | 阶段配置 ["CONCEPT","EVT","DVT","PVT","MP"] |
| estimated_days | INTEGER | 否 | 预估总工期（天） |
| is_active | BOOLEAN | 是 | 是否启用 |
| parent_template_id | UUID | 否 | 继承自哪个模板 |
| version | INTEGER | 是 | 版本号 |
| created_by | VARCHAR(64) | 是 | 创建人 |
| created_at | TIMESTAMP | 是 | 创建时间 |
| updated_at | TIMESTAMP | 是 | 更新时间 |

#### 1.2.2 模板任务表 template_tasks

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| template_id | UUID | 是 | 关联模板 |
| task_code | VARCHAR(50) | 是 | 任务编码（模板内唯一） |
| name | VARCHAR(200) | 是 | 任务名称 |
| description | TEXT | 否 | 任务描述 |
| phase | ENUM | 是 | 所属阶段 |
| parent_task_code | VARCHAR(50) | 否 | 父任务编码（子任务用） |
| task_type | ENUM | 是 | MILESTONE/TASK/SUBTASK |
| default_assignee_role | VARCHAR(50) | 否 | 默认负责角色（如 HW_ENG） |
| estimated_days | INTEGER | 是 | 预估工期（天） |
| is_critical | BOOLEAN | 是 | 是否关键路径 |
| deliverables | JSONB | 否 | 交付物定义 |
| checklist | JSONB | 否 | 完成检查清单 |
| requires_approval | BOOLEAN | 是 | 是否需要审批 |
| approval_type | VARCHAR(50) | 否 | 审批类型：REVIEW_MEETING/FEISHU_APPROVAL |
| sort_order | INTEGER | 是 | 排序序号 |
| created_at | TIMESTAMP | 是 | 创建时间 |
| updated_at | TIMESTAMP | 是 | 更新时间 |

#### 1.2.3 模板任务依赖表 template_task_dependencies

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| template_id | UUID | 是 | 关联模板 |
| task_code | VARCHAR(50) | 是 | 任务编码 |
| depends_on_task_code | VARCHAR(50) | 是 | 依赖的任务编码 |
| dependency_type | ENUM | 是 | FS/SS/FF/SF |
| lag_days | INTEGER | 否 | 延迟天数（默认0） |

### 1.3 模板交付物定义

```json
{
  "deliverables": [
    {
      "code": "D001",
      "name": "电路原理图",
      "type": "DESIGN",
      "format": ["PDF", "SCHEMATIC"],
      "requires_review": true,
      "review_type": "REVIEW_MEETING",
      "reviewers": ["HW_LEAD", "PM"]
    },
    {
      "code": "D002", 
      "name": "PCB设计文件",
      "type": "DESIGN",
      "format": ["GERBER", "PDF"],
      "requires_review": true,
      "review_type": "FEISHU_APPROVAL",
      "reviewers": ["HW_LEAD"]
    }
  ]
}
```

### 1.4 模板管理功能

#### UC-TPL-001：创建任务模板

**主流程：**
1. 用户选择"新建模板"或"从现有模板复制"
2. 填写模板基本信息（名称、描述、适用产品类型）
3. 配置阶段（可增减、调整顺序）
4. 添加任务：
   - 拖拽式添加任务到阶段
   - 设置任务属性（工期、负责角色、交付物）
   - 建立任务依赖关系（连线）
   - 添加子任务
5. 配置审批节点
6. 保存模板

#### UC-TPL-002：从模板创建项目

**主流程：**
1. 用户创建新项目时，选择"从模板创建"
2. 选择目标模板
3. 设置项目基本信息和启动日期
4. 系统自动：
   - 复制所有模板任务到项目
   - 根据启动日期计算各任务日期
   - 根据默认负责角色分配负责人（可后续调整）
   - 建立任务依赖关系
   - 创建各阶段里程碑
5. 用户确认或调整
6. 项目创建完成

**日期计算规则：**
- 无依赖任务：从阶段开始日期起算
- 有依赖任务：根据依赖类型和前置任务计算
- 考虑周末（可配置是否跳过）
- 考虑节假日（可配置）

#### UC-TPL-003：模板版本管理

**主流程：**
1. 编辑已有模板时，系统提示"是否创建新版本"
2. 选择是：创建新版本，保留旧版本历史
3. 选择否：直接覆盖（仅DRAFT状态可覆盖）
4. 已使用的模板版本不可删除

---

## 二、任务自动化引擎

### 2.1 自动化规则

#### 2.1.1 任务状态自动流转

| 触发条件 | 自动动作 | 配置选项 |
|---------|---------|---------|
| 前置任务全部完成 | 后续任务状态变为"待开始" | 可配置是否自动开始 |
| 前置任务全部完成 + 自动开始 | 后续任务状态变为"进行中" | 发送飞书通知 |
| 所有子任务完成 | 父任务自动完成 | 可配置 |
| 任务逾期 | 发送逾期提醒 | 提醒频率可配置 |
| 阶段所有任务完成 | 触发阶段评审 | 自动创建评审任务 |

#### 2.1.2 自动化配置表 automation_rules

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| rule_type | ENUM | 是 | TASK_START/TASK_COMPLETE/OVERDUE/PHASE_COMPLETE |
| trigger_condition | JSONB | 是 | 触发条件 |
| action_type | ENUM | 是 | UPDATE_STATUS/SEND_NOTIFICATION/CREATE_TASK/CREATE_APPROVAL |
| action_config | JSONB | 是 | 动作配置 |
| is_active | BOOLEAN | 是 | 是否启用 |
| project_id | UUID | 否 | 项目级规则（空=全局） |
| template_id | UUID | 否 | 模板级规则 |

#### 2.1.3 任务状态机

```
                    ┌─────────────────────────────────────┐
                    │                                     │
                    v                                     │
PENDING ──依赖满足──> READY ──手动/自动开始──> IN_PROGRESS ──完成──> COMPLETED
    │                  │                          │              │
    │                  │                          v              │
    │                  │                      ON_HOLD ───────────┤
    │                  │                          │              │
    │                  v                          v              │
    └─────────────> CANCELLED <──────────────────┘              │
                                                                │
                    ┌───────────────────────────────────────────┘
                    v
              NEEDS_REVIEW ──审批通过──> APPROVED
                    │
                    v
               审批驳回 ──> IN_PROGRESS（重新修改）
```

**状态说明：**
- PENDING: 等待中（前置任务未完成）
- READY: 就绪（可以开始）
- IN_PROGRESS: 进行中
- ON_HOLD: 暂停
- NEEDS_REVIEW: 待评审
- APPROVED: 已通过
- COMPLETED: 已完成
- CANCELLED: 已取消

### 2.2 依赖检查服务

```go
// 伪代码：任务完成时的依赖检查
func OnTaskCompleted(taskID string) {
    // 1. 获取依赖此任务的所有后续任务
    dependentTasks := GetDependentTasks(taskID)
    
    for _, depTask := range dependentTasks {
        // 2. 检查该任务的所有前置依赖是否满足
        if AllDependenciesMet(depTask.ID) {
            // 3. 更新状态为 READY
            UpdateTaskStatus(depTask.ID, "READY")
            
            // 4. 如果配置了自动开始，则自动开始
            if depTask.AutoStart {
                UpdateTaskStatus(depTask.ID, "IN_PROGRESS")
                StartFeishuTask(depTask)  // 同步到飞书
            }
            
            // 5. 发送通知
            NotifyAssignee(depTask)
        }
    }
    
    // 6. 检查是否所有子任务完成，自动完成父任务
    CheckParentTaskCompletion(taskID)
    
    // 7. 检查阶段是否完成
    CheckPhaseCompletion(taskID)
}
```

---

## 三、飞书深度集成

### 3.1 飞书任务双向同步

#### 3.1.1 同步机制

| 方向 | 同步内容 | 触发时机 |
|-----|---------|---------|
| PLM → 飞书 | 创建任务 | PLM创建任务时 |
| PLM → 飞书 | 更新任务（标题、描述、截止日期） | PLM任务更新时 |
| PLM → 飞书 | 完成任务 | PLM任务完成时 |
| PLM → 飞书 | 分配负责人 | PLM分配/变更负责人时 |
| 飞书 → PLM | 完成状态同步 | 飞书任务完成时（Webhook） |
| 飞书 → PLM | 评论同步 | 飞书任务评论时（Webhook） |

#### 3.1.2 飞书任务关联表 feishu_task_sync

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| task_id | UUID | 是 | PLM任务ID |
| feishu_task_id | VARCHAR(100) | 是 | 飞书任务ID |
| feishu_task_guid | VARCHAR(100) | 否 | 飞书任务GUID |
| sync_status | ENUM | 是 | SYNCED/PENDING/FAILED |
| last_sync_at | TIMESTAMP | 否 | 最后同步时间 |
| sync_error | TEXT | 否 | 同步错误信息 |

#### 3.1.3 飞书任务API调用

```go
// 创建飞书任务
func CreateFeishuTask(task *Task) (string, error) {
    reqBody := map[string]interface{}{
        "summary": task.Name,
        "description": task.Description,
        "due": map[string]interface{}{
            "time": task.DueDate.Unix() * 1000,
            "is_all_day": false,
        },
        "members": []map[string]interface{}{
            {
                "id": task.Assignee.FeishuUserID,
                "role": "assignee",
            },
        },
        "origin": map[string]interface{}{
            "platform_i18n_name": "nimo PLM",
            "href": map[string]interface{}{
                "url": fmt.Sprintf("https://plm.nimo.com/tasks/%s", task.ID),
                "title": task.Name,
            },
        },
        "custom_fields": []map[string]interface{}{
            {
                "guid": "project_field_guid",
                "text_value": task.ProjectName,
            },
            {
                "guid": "phase_field_guid", 
                "text_value": task.Phase,
            },
        },
    }
    
    resp, err := feishuClient.Post("/open-apis/task/v2/tasks", reqBody)
    return resp.Data.TaskID, err
}
```

### 3.2 飞书审批流程集成

#### 3.2.1 审批场景

| 审批场景 | 触发时机 | 审批人 | 审批后动作 |
|---------|---------|-------|----------|
| BOM发布审批 | 提交BOM发布 | HW_LEAD → PM | 通过后BOM状态变RELEASED |
| ECN审批 | 提交ECN | 根据变更类型决定 | 通过后执行变更 |
| 任务产出物审批 | 任务提交审批 | 任务配置的审批人 | 通过后任务完成 |
| 阶段评审审批 | 阶段完成 | PM + 技术负责人 | 通过后进入下一阶段 |

#### 3.2.2 审批流程定义表 approval_definitions

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| code | VARCHAR(50) | 是 | 审批流程编码 |
| name | VARCHAR(200) | 是 | 审批流程名称 |
| feishu_approval_code | VARCHAR(100) | 是 | 飞书审批定义Code |
| approval_type | ENUM | 是 | BOM/ECN/TASK/PHASE |
| form_definition | JSONB | 是 | 表单定义 |
| is_active | BOOLEAN | 是 | 是否启用 |

#### 3.2.3 审批实例表 approval_instances

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| approval_def_id | UUID | 是 | 审批定义ID |
| feishu_instance_code | VARCHAR(100) | 否 | 飞书审批实例Code |
| business_type | ENUM | 是 | BOM/ECN/TASK/PHASE |
| business_id | UUID | 是 | 业务ID |
| status | ENUM | 是 | PENDING/APPROVED/REJECTED/CANCELLED |
| applicant_id | VARCHAR(64) | 是 | 申请人 |
| form_data | JSONB | 是 | 表单数据 |
| created_at | TIMESTAMP | 是 | 创建时间 |
| completed_at | TIMESTAMP | 否 | 完成时间 |

#### 3.2.4 创建飞书审批

```go
// 创建飞书审批实例
func CreateFeishuApproval(approval *ApprovalInstance) (string, error) {
    // 构建表单数据
    formData := buildFormData(approval.FormData, approval.Definition.FormDefinition)
    
    reqBody := map[string]interface{}{
        "approval_code": approval.Definition.FeishuApprovalCode,
        "user_id": approval.ApplicantID,
        "form": formData,
        "node_approver_user_id_list": approval.ApproverIDs,
    }
    
    resp, err := feishuClient.Post("/open-apis/approval/v4/instances", reqBody)
    return resp.Data.InstanceCode, err
}

// 飞书审批回调处理
func HandleApprovalCallback(event *FeishuApprovalEvent) error {
    // 查找对应的审批实例
    instance := FindApprovalByFeishuCode(event.InstanceCode)
    
    switch event.Status {
    case "APPROVED":
        instance.Status = "APPROVED"
        // 执行通过后动作
        executeApprovalAction(instance)
    case "REJECTED":
        instance.Status = "REJECTED"
        // 通知申请人
        notifyRejection(instance, event.Comment)
    }
    
    return UpdateApprovalInstance(instance)
}
```

### 3.3 飞书评审会议集成

#### 3.3.1 评审会议场景

| 评审类型 | 触发时机 | 参会人 | 会议议程 |
|---------|---------|-------|---------|
| 设计评审 | 设计任务提交评审 | 设计者+评审人 | 设计文档讲解、问题讨论、结论 |
| EVT评审 | EVT阶段完成 | 项目组+技术负责人 | 测试结果汇报、问题总结、是否进入DVT |
| DVT评审 | DVT阶段完成 | 项目组+管理层 | 全面评审、量产准备、是否进入PVT |
| BOM评审 | BOM提交评审 | 设计+采购+成本 | BOM结构、成本分析、替代方案 |

#### 3.3.2 评审会议表 review_meetings

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| title | VARCHAR(200) | 是 | 会议主题 |
| meeting_type | ENUM | 是 | DESIGN/PHASE/BOM/ECN |
| project_id | UUID | 否 | 关联项目 |
| task_id | UUID | 否 | 关联任务 |
| feishu_calendar_event_id | VARCHAR(100) | 否 | 飞书日历事件ID |
| feishu_meeting_id | VARCHAR(100) | 否 | 飞书会议ID（视频会议） |
| scheduled_at | TIMESTAMP | 是 | 计划时间 |
| duration_minutes | INTEGER | 是 | 时长（分钟） |
| location | VARCHAR(200) | 否 | 会议地点/会议室 |
| organizer_id | VARCHAR(64) | 是 | 组织者 |
| attendees | JSONB | 是 | 参会人列表 |
| agenda | TEXT | 否 | 会议议程 |
| documents | JSONB | 否 | 评审文档列表 |
| status | ENUM | 是 | SCHEDULED/IN_PROGRESS/COMPLETED/CANCELLED |
| conclusion | TEXT | 否 | 会议结论 |
| action_items | JSONB | 否 | 待办事项 |
| minutes_doc_id | UUID | 否 | 会议纪要文档ID |
| created_at | TIMESTAMP | 是 | 创建时间 |
| updated_at | TIMESTAMP | 是 | 更新时间 |

#### 3.3.3 创建飞书日历事件

```go
// 创建飞书日历事件（评审会议）
func CreateFeishuCalendarEvent(meeting *ReviewMeeting) (string, error) {
    attendeeIDs := make([]map[string]interface{}, len(meeting.Attendees))
    for i, att := range meeting.Attendees {
        attendeeIDs[i] = map[string]interface{}{
            "type": "user",
            "user_id": att.FeishuUserID,
        }
    }
    
    reqBody := map[string]interface{}{
        "summary": meeting.Title,
        "description": buildMeetingDescription(meeting),
        "start_time": map[string]interface{}{
            "timestamp": meeting.ScheduledAt.Unix(),
        },
        "end_time": map[string]interface{}{
            "timestamp": meeting.ScheduledAt.Add(time.Minute * time.Duration(meeting.DurationMinutes)).Unix(),
        },
        "attendee_ability": "can_modify_event",
        "attendees": attendeeIDs,
        "visibility": "default",
        "need_notification": true,
        "reminders": []map[string]interface{}{
            {"minutes": 15},  // 提前15分钟提醒
            {"minutes": 60},  // 提前1小时提醒
        },
    }
    
    // 如果需要视频会议，添加vc_chat
    if meeting.NeedsVideoConference {
        reqBody["vchat"] = map[string]interface{}{
            "vc_type": "vc",  // 使用飞书视频会议
        }
    }
    
    resp, err := feishuClient.Post("/open-apis/calendar/v4/calendars/primary/events", reqBody)
    return resp.Data.EventID, err
}
```

### 3.4 飞书Webhook事件处理

#### 3.4.1 订阅的事件类型

| 事件类型 | 事件名称 | 处理逻辑 |
|---------|---------|---------|
| task.task.updated_v1 | 任务更新 | 同步任务状态到PLM |
| task.task.comment.created_v1 | 任务评论 | 同步评论到PLM |
| approval.approval.updated_v1 | 审批状态变更 | 更新审批状态，执行后续动作 |
| calendar.event.updated_v1 | 日历事件变更 | 同步会议时间变更 |
| contact.user.updated_v1 | 用户信息变更 | 同步用户信息 |

#### 3.4.2 Webhook处理流程

```go
// 飞书Webhook统一入口
func HandleFeishuWebhook(c *gin.Context) {
    var event FeishuEvent
    c.BindJSON(&event)
    
    // 验证签名
    if !verifySignature(event, c.GetHeader("X-Lark-Signature")) {
        c.JSON(401, "Invalid signature")
        return
    }
    
    // 处理URL验证请求
    if event.Type == "url_verification" {
        c.JSON(200, gin.H{"challenge": event.Challenge})
        return
    }
    
    // 异步处理事件
    go processEvent(event)
    
    c.JSON(200, gin.H{"success": true})
}

func processEvent(event FeishuEvent) {
    switch event.Header.EventType {
    case "task.task.updated_v1":
        handleTaskUpdate(event)
    case "approval.approval.updated_v1":
        handleApprovalUpdate(event)
    case "calendar.event.updated_v1":
        handleCalendarEventUpdate(event)
    default:
        log.Printf("Unhandled event type: %s", event.Header.EventType)
    }
}
```

---

## 四、API接口设计（新增）

### 4.1 模板管理接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/v1/templates | 获取模板列表 |
| POST | /api/v1/templates | 创建模板 |
| GET | /api/v1/templates/{id} | 获取模板详情（含任务） |
| PUT | /api/v1/templates/{id} | 更新模板 |
| DELETE | /api/v1/templates/{id} | 删除模板 |
| POST | /api/v1/templates/{id}/duplicate | 复制模板 |
| GET | /api/v1/templates/{id}/versions | 获取模板版本历史 |
| POST | /api/v1/templates/{id}/tasks | 添加模板任务 |
| PUT | /api/v1/templates/{id}/tasks/{taskCode} | 更新模板任务 |
| DELETE | /api/v1/templates/{id}/tasks/{taskCode} | 删除模板任务 |
| POST | /api/v1/templates/{id}/tasks/{taskCode}/dependencies | 添加任务依赖 |
| DELETE | /api/v1/templates/{id}/tasks/{taskCode}/dependencies/{depCode} | 删除任务依赖 |

### 4.2 从模板创建项目

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/projects/create-from-template | 从模板创建项目 |

**请求体：**
```json
{
  "template_id": "uuid",
  "project_name": "nimo Air 3 研发项目",
  "project_code": "PRJ-2026-003",
  "product_id": "uuid",
  "start_date": "2026-02-15",
  "pm_user_id": "ou_xxx",
  "skip_weekends": true,
  "skip_holidays": true,
  "role_assignments": {
    "HW_ENG": "ou_xxx",
    "SW_ENG": "ou_yyy",
    "QA_ENG": "ou_zzz"
  }
}
```

### 4.3 任务自动化接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/tasks/{id}/complete | 完成任务（触发自动化） |
| POST | /api/v1/tasks/{id}/submit-review | 提交任务评审 |
| GET | /api/v1/projects/{id}/automation-rules | 获取项目自动化规则 |
| PUT | /api/v1/projects/{id}/automation-rules | 更新项目自动化规则 |

### 4.4 飞书集成接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/tasks/{id}/sync-to-feishu | 同步任务到飞书 |
| POST | /api/v1/tasks/{id}/create-approval | 创建飞书审批 |
| POST | /api/v1/tasks/{id}/create-review-meeting | 创建评审会议 |
| POST | /api/v1/feishu/webhook | 飞书Webhook回调 |
| GET | /api/v1/feishu/approval-definitions | 获取审批定义列表 |
| POST | /api/v1/feishu/test-connection | 测试飞书连接 |

---

## 五、前端页面设计（新增）

### 5.1 模板管理页面

**页面：/templates**

功能：
- 模板列表（卡片/列表视图切换）
- 筛选：系统模板/自定义模板、产品类型、状态
- 操作：新建、编辑、复制、删除、查看历史

**页面：/templates/{id}/edit**

功能：
- 可视化任务编辑器（类似甘特图）
- 拖拽添加任务
- 连线建立依赖
- 属性面板编辑任务详情
- 实时预览工期计算

### 5.2 从模板创建项目

**对话框：CreateProjectFromTemplate**

步骤：
1. 选择模板（预览模板结构）
2. 填写项目信息
3. 设置启动日期
4. 分配角色负责人
5. 确认并创建

### 5.3 任务详情增强

**页面：/projects/{id}/tasks/{taskId}**

新增功能：
- 交付物上传区域
- 提交评审按钮
- 评审状态展示
- 飞书任务链接
- 飞书审批状态

---

## 六、实施建议

### 6.1 开发优先级

1. **P0 - 核心功能**（第一阶段）
   - 模板数据模型和基础CRUD
   - 从模板创建项目
   - 任务依赖自动化

2. **P1 - 飞书集成**（第二阶段）
   - 飞书任务双向同步
   - 飞书审批流程
   - Webhook事件处理

3. **P2 - 增强功能**（第三阶段）
   - 评审会议管理
   - 可视化模板编辑器
   - 高级自动化规则

### 6.2 飞书开放平台配置

需要在飞书开放平台配置：
1. 任务 API 权限
2. 审批 API 权限
3. 日历 API 权限
4. Webhook 订阅
5. 创建审批定义模板

---

## 七、附录

### 7.1 默认模板：智能眼镜标准研发流程

详见 `templates/smart_glasses_default.json`

### 7.2 飞书审批表单定义示例

详见 `templates/feishu_approval_forms.json`

---

**文档结束**
