# AI员工与飞书深度集成设计 v1.0

**版本**: v1.0  
**日期**: 2026-02-05  
**项目**: nimo智能眼镜数字化系统  
**定位**: PLM/ERP系统PRD补充文档

---

## 一、设计理念

### 1.1 核心思想

```
传统模式：人类员工 → 登录系统 → 完成工作
nimo模式：人类员工 + AI员工 → 在飞书协作 → 系统是能力中心
```

### 1.2 设计原则

1. **飞书是统一工作台**：人类和AI都在飞书工作，不用另学一套系统
2. **AI员工 = 飞书成员**：AI在飞书有账号，能收任务、做审批、发消息
3. **多维表格优先**：能用表格解决的就不写前端代码
4. **人机无缝协作**：任务分配、审批流程对人和AI一视同仁
5. **渐进式自主**：AI从辅助开始，逐步提升自主级别

### 1.3 架构定位

```
┌─────────────────────────────────────────────────────┐
│                    飞书工作台                        │
│    (消息、任务、审批、多维表格、文档、会议)           │
└─────────────────────────┬───────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          │               │               │
          v               v               v
    ┌──────────┐   ┌──────────┐   ┌──────────┐
    │ 人类员工  │   │  AI员工  │   │ 后端服务 │
    │ (10-15人) │   │(OpenClaw)│   │(PLM/ERP) │
    └──────────┘   └──────────┘   └──────────┘
                          │               │
                          └───────┬───────┘
                                  v
                         ┌──────────────┐
                         │   业务数据    │
                         │  PostgreSQL  │
                         └──────────────┘
```

---

## 二、飞书组织架构设计

### 2.1 组织结构（人机混合）

```
比特幻境 Bitfantasy（飞书组织）
│
├── 【管理层】
│   └── CEO（陈泽斌）
│
├── 【研发中心】
│   ├── 硬件研发部
│   │   ├── 硬件负责人（人类）
│   │   ├── 硬件工程师 x2（人类）
│   │   └── 🤖 硬件助理（AI）
│   │
│   ├── 软件研发部
│   │   ├── 软件负责人（人类）
│   │   ├── 软件工程师 x2（人类）
│   │   └── 🤖 软件助理（AI）
│   │
│   ├── 结构光学部
│   │   ├── 结构工程师（人类）
│   │   ├── 光学工程师（人类）
│   │   └── 🤖 结构光学助理（AI）
│   │
│   ├── 项目管理部
│   │   ├── 项目经理（人类）
│   │   └── 🤖 项目督导（AI）
│   │
│   └── 质量部
│       ├── 质量工程师（人类）
│       └── 🤖 质量分析师（AI）
│
├── 【供应链中心】
│   ├── 采购部
│   │   ├── 采购负责人（人类）
│   │   └── 🤖 采购专员（AI）
│   │
│   ├── 仓储部
│   │   ├── 仓库负责人（人类）
│   │   └── 🤖 库存管家（AI）
│   │
│   └── 生产部
│       ├── 生产负责人（人类）
│       └── 🤖 生产调度员（AI）
│
├── 【市场销售中心】
│   ├── 市场部
│   │   ├── 市场负责人（人类）
│   │   └── 🤖 市场分析师（AI）
│   │
│   ├── 销售部
│   │   ├── 销售负责人（人类）
│   │   └── 🤖 订单处理员（AI）
│   │
│   └── 售后部
│       └── 🤖 客服专员（AI）
│
└── 【职能中心】
    ├── 财务部
    │   ├── 财务负责人（人类）
    │   └── 🤖 财务助理（AI）
    │
    └── 行政人事部
        ├── HR/行政（人类）
        └── 🤖 行政助理（AI）
```

### 2.2 人员配置统计

| 类型 | 数量 | 说明 |
|-----|------|------|
| 人类员工 | 15-18人 | 核心决策和专业工作 |
| AI员工 | 12-15个 | 执行、协调、分析工作 |
| 人机比例 | 约1:1 | 每个核心岗位配AI助手 |

### 2.3 飞书成员类型

| 成员类型 | 说明 | 飞书配置 |
|---------|------|---------|
| 人类正式员工 | 公司正式员工 | 正式成员 |
| 人类外包/实习 | 临时人员 | 外部成员 |
| AI员工 | OpenClaw Agent | 机器人成员（自定义头像、名称） |

---

## 三、AI员工详细设计

### 3.1 AI员工清单

| AI员工名称 | 所属部门 | 核心职责 | 技能标签 |
|-----------|---------|---------|---------|
| 🤖 PLM管家 | 项目管理部 | 产品数据维护、BOM管理、版本控制 | PLM、BOM、文档 |
| 🤖 项目督导 | 项目管理部 | 任务跟踪、进度预警、会议纪要、报告生成 | 项目管理、提醒 |
| 🤖 硬件助理 | 硬件研发部 | 元器件查询、规格对比、设计文档整理 | 硬件、元器件 |
| 🤖 质量分析师 | 质量部 | 测试数据分析、质量趋势、异常预警 | 测试、数据分析 |
| 🤖 采购专员 | 采购部 | 询价、下单、跟单、供应商管理 | 采购、供应商 |
| 🤖 库存管家 | 仓储部 | 库存监控、预警、盘点、出入库 | 库存、预警 |
| 🤖 生产调度员 | 生产部 | 工单管理、排产建议、进度跟踪 | 生产、排程 |
| 🤖 订单处理员 | 销售部 | 订单确认、发货安排、物流跟踪 | 订单、物流 |
| 🤖 客服专员 | 售后部 | 售后工单、问题诊断、用户回访 | 售后、客服 |
| 🤖 财务助理 | 财务部 | 对账、发票、报表、成本分析 | 财务、报表 |
| 🤖 行政助理 | 行政部 | 日程管理、会议安排、通知发布 | 行政、日程 |
| 🤖 市场分析师 | 市场部 | 竞品分析、市场数据、报告生成 | 市场、分析 |

### 3.2 AI员工自主级别定义

| 级别 | 名称 | 说明 | 适用场景 |
|-----|------|------|---------|
| 1 | 仅建议 | AI提供建议，人类决策和执行 | 重大决策、敏感操作 |
| 2 | 人工确认 | AI准备好操作，人类确认后执行 | 财务操作、对外发送 |
| 3 | 人工监督 | AI自动执行，人类可随时干预 | 日常业务操作 |
| 4 | 事后审计 | AI自主执行，定期人工审计 | 常规任务、提醒通知 |
| 5 | 完全自主 | AI完全自主，仅异常时通知人类 | 信息查询、状态更新 |

### 3.3 AI员工权限矩阵

| AI员工 | 自主级别 | 可执行操作 | 需人工确认 | 人类监督人 |
|-------|---------|-----------|-----------|-----------|
| 🤖 PLM管家 | 3 | 产品信息更新、BOM编辑、文档上传 | BOM发布、产品状态变更 | 硬件负责人 |
| 🤖 项目督导 | 4 | 任务状态更新、提醒发送、报告生成 | 任务创建、里程碑变更 | PM |
| 🤖 采购专员 | 3 | 询价、PO创建、跟单 | PO审批提交、付款申请 | 采购负责人 |
| 🤖 库存管家 | 4 | 库存查询、预警发送、盘点记录 | 库存调整、报废处理 | 仓库负责人 |
| 🤖 订单处理员 | 3 | 订单确认、发货安排 | 退款处理、大额订单 | 销售负责人 |
| 🤖 客服专员 | 3 | 工单创建、问题诊断 | 退换货审批、赔偿处理 | 售后负责人 |
| 🤖 财务助理 | 2 | 对账、报表生成 | 付款执行、发票开具 | 财务负责人 |

### 3.4 AI员工配置模板

```yaml
# AI员工配置示例：🤖 采购专员

ai_agent:
  # 基本信息
  id: "agent_procurement_01"
  name: "采购专员"
  avatar: "🤖"
  feishu_user_id: "ou_agent_procurement_01"
  department: "采购部"
  
  # OpenClaw配置
  openclaw:
    gateway_url: "https://gateway.nimo.internal"
    session_type: "isolated"  # main/isolated
    model: "anthropic/claude-sonnet-4-5"
    
  # 监督配置
  supervision:
    human_supervisor: "ou_xxx"  # 采购负责人飞书ID
    autonomy_level: 3
    escalation_timeout: "4h"  # 超时升级
    
  # 能力配置
  capabilities:
    # 系统API权限
    api_permissions:
      - "supplier.*"
      - "purchase_order.*"
      - "material.read"
      - "inventory.read"
    # 飞书权限
    feishu_permissions:
      - "message.send"
      - "task.create"
      - "task.update"
      - "approval.create"
      - "bitable.read"
      - "bitable.write"
      
  # 工作规则
  work_rules:
    # 自动执行
    auto_execute:
      - action: "send_inquiry"
        condition: "amount < 10000"
      - action: "create_po"
        condition: "approved_pr exists"
      - action: "send_reminder"
        condition: "delivery_overdue"
    # 需人工确认
    require_confirm:
      - action: "submit_po_approval"
        notify: "@supervisor"
      - action: "change_supplier"
        notify: "@supervisor"
        
  # 事件订阅
  event_subscriptions:
    - event: "inventory.below_safety_stock"
      action: "analyze_and_suggest_procurement"
    - event: "pr.approved"
      action: "create_purchase_order"
    - event: "po.delivery_delayed"
      action: "notify_stakeholders_and_followup"
    - event: "feishu.message.mention_me"
      action: "handle_message"
    - event: "feishu.task.assigned_to_me"
      action: "handle_task"
```

---

## 四、飞书深度集成设计

### 4.1 集成架构总览

```
┌─────────────────────────────────────────────────────────────────┐
│                      飞书开放平台                                │
├──────────┬──────────┬──────────┬──────────┬──────────┬─────────┤
│   通讯录  │   消息   │   任务   │   审批   │  多维表格 │   文档  │
│   API    │   API    │   API    │   API    │   API    │   API   │
└────┬─────┴────┬─────┴────┬─────┴────┬─────┴────┬─────┴────┬────┘
     │          │          │          │          │          │
     v          v          v          v          v          v
┌─────────────────────────────────────────────────────────────────┐
│                    集成服务层 (integration-service)              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐  │
│  │ 身份同步 │ │ 消息网关 │ │ 任务同步 │ │ 审批引擎 │ │ 表格同步 │  │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
     │          │          │          │          │
     v          v          v          v          v
┌─────────────────────────────────────────────────────────────────┐
│                      业务服务层 (PLM/ERP)                        │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 身份认证设计

#### 4.2.1 统一身份原则

```yaml
identity_design:
  principle: "系统不维护独立用户表，完全依赖飞书身份"
  
  user_source: "飞书通讯录"
  user_identifier: "feishu_user_id"
  
  user_types:
    human:
      source: "飞书正式成员/外部成员"
      auth_method: "飞书OAuth扫码登录"
      token_type: "JWT (基于飞书access_token)"
      
    ai_agent:
      source: "飞书机器人成员"
      auth_method: "机器人Token + Agent签名"
      token_type: "Service Token"
```

#### 4.2.2 登录流程

**人类员工登录（Web端）：**
```
1. 访问 plm.nimo.com
2. 点击"飞书登录"
3. 跳转飞书OAuth授权页
4. 用户扫码/输入账号密码
5. 飞书回调，携带authorization_code
6. 后端用code换access_token
7. 后端用token获取用户信息
8. 后端生成JWT，返回前端
9. 用户进入系统
```

**人类员工使用（多维表格）：**
```
1. 在飞书中打开多维表格
2. 无需登录（飞书自动识别身份）
3. 根据飞书身份确定数据权限
```

**AI员工调用（API）：**
```
1. OpenClaw Agent启动
2. 使用预配置的飞书机器人Token
3. 附加Agent签名（防止伪造）
4. 调用后端API
5. 后端验证Token和签名
6. 确认Agent身份和权限
7. 执行操作
```

#### 4.2.3 数据库设计

```sql
-- 用户扩展表（仅存储飞书未提供的信息）
CREATE TABLE user_extensions (
    feishu_user_id VARCHAR(64) PRIMARY KEY,
    user_type ENUM('HUMAN', 'AI_AGENT') NOT NULL,
    preferences JSONB,  -- 用户偏好设置
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- AI员工配置表
CREATE TABLE ai_agent_configs (
    feishu_user_id VARCHAR(64) PRIMARY KEY,
    agent_code VARCHAR(50) UNIQUE NOT NULL,
    agent_name VARCHAR(100) NOT NULL,
    department_id VARCHAR(64),
    
    -- OpenClaw配置
    openclaw_gateway_url VARCHAR(200),
    openclaw_session_type ENUM('main', 'isolated') DEFAULT 'isolated',
    openclaw_model VARCHAR(100),
    
    -- 权限配置
    capabilities JSONB NOT NULL,  -- 能力列表
    api_permissions JSONB NOT NULL,  -- API权限
    feishu_permissions JSONB NOT NULL,  -- 飞书权限
    
    -- 监督配置
    autonomy_level INTEGER DEFAULT 3,
    human_supervisor_id VARCHAR(64),
    escalation_timeout_minutes INTEGER DEFAULT 240,
    
    -- 工作规则
    work_rules JSONB,
    event_subscriptions JSONB,
    
    -- 状态
    status ENUM('ACTIVE', 'PAUSED', 'DISABLED') DEFAULT 'ACTIVE',
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- AI操作日志表
CREATE TABLE ai_agent_action_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(64) NOT NULL,
    
    -- 操作信息
    action_type VARCHAR(50) NOT NULL,  -- API_CALL, DECISION, MESSAGE, APPROVAL
    action_name VARCHAR(200) NOT NULL,
    action_detail JSONB,
    
    -- 上下文
    trigger_source VARCHAR(50),  -- EVENT, TASK, MESSAGE, SCHEDULE
    trigger_detail JSONB,
    input_context JSONB,
    
    -- 结果
    output_result JSONB,
    success BOOLEAN,
    error_message TEXT,
    
    -- AI推理
    reasoning TEXT,  -- AI的推理过程
    confidence DECIMAL(5,4),  -- 置信度 0-1
    
    -- 人工审核
    require_human_review BOOLEAN DEFAULT FALSE,
    human_reviewed BOOLEAN DEFAULT FALSE,
    human_reviewer_id VARCHAR(64),
    human_review_result VARCHAR(20),  -- APPROVED, REJECTED, MODIFIED
    human_review_comment TEXT,
    human_reviewed_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_ai_logs_agent ON ai_agent_action_logs(agent_id);
CREATE INDEX idx_ai_logs_action_type ON ai_agent_action_logs(action_type);
CREATE INDEX idx_ai_logs_created ON ai_agent_action_logs(created_at);
CREATE INDEX idx_ai_logs_need_review ON ai_agent_action_logs(require_human_review, human_reviewed);
```

### 4.3 任务同步设计

#### 4.3.1 同步策略

```yaml
task_sync:
  direction: "bidirectional"
  
  # 飞书任务 → 系统任务
  feishu_to_system:
    trigger: "飞书Webhook"
    events:
      - task.created
      - task.updated
      - task.completed
      - task.deleted
    mapping:
      feishu.summary → system.name
      feishu.description → system.description
      feishu.due.timestamp → system.planned_end
      feishu.assignee → system.assignee_id
      feishu.collaborators → system.collaborators
      feishu.completed_at → system.actual_end
      
  # 系统任务 → 飞书任务
  system_to_feishu:
    trigger: "系统事件"
    events:
      - task.created
      - task.updated
      - task.status_changed
    mapping:
      system.name → feishu.summary
      system.description → feishu.description
      system.planned_end → feishu.due.timestamp
      system.assignee_id → feishu.assignee
      system.status == 'COMPLETED' → feishu.completed_at
```

#### 4.3.2 任务分配规则

```yaml
task_assignment:
  # 可以分配给AI的任务类型
  ai_assignable:
    - type: "data_maintenance"
      example: "更新BOM成本信息"
      assign_to: "🤖 PLM管家"
    - type: "progress_tracking"
      example: "跟踪订单发货状态"
      assign_to: "🤖 订单处理员"
    - type: "report_generation"
      example: "生成周报"
      assign_to: "🤖 项目督导"
    - type: "procurement"
      example: "询价并创建采购订单"
      assign_to: "🤖 采购专员"
      
  # 必须人类处理的任务类型
  human_only:
    - type: "design_decision"
      example: "确定产品外观方案"
    - type: "strategic_decision"
      example: "选择新供应商"
    - type: "negotiation"
      example: "与供应商谈判价格"
```

### 4.4 审批流程设计

#### 4.4.1 审批场景

| 审批场景 | 发起者 | 审批人 | AI参与方式 |
|---------|-------|-------|-----------|
| BOM发布 | 工程师 | 技术负责人→PM | 🤖 PLM管家 提供成本分析 |
| 采购订单 | 🤖 采购专员 | 采购负责人→财务 | AI发起，人类审批 |
| 工程变更 | 工程师 | 评审委员会 | 🤖 PLM管家 评估影响 |
| 付款申请 | 🤖 财务助理 | 财务负责人 | AI准备资料，人类审批 |
| 退换货 | 🤖 客服专员 | 售后负责人 | AI初审，人类终审 |

#### 4.4.2 审批规则配置

```yaml
approval_rules:
  # 采购订单审批
  purchase_order:
    conditions:
      - amount: "<5000"
        flow: 
          - role: "AI_AGENT"  # AI可自动审批
            auto_approve: true
            log_for_audit: true
      - amount: "5000-50000"
        flow:
          - role: "采购负责人"
            timeout: "24h"
      - amount: ">50000"
        flow:
          - role: "采购负责人"
            timeout: "24h"
          - role: "财务负责人"
            timeout: "24h"
          
  # BOM发布审批
  bom_release:
    flow:
      - role: "技术负责人"
        timeout: "48h"
        ai_assist: "🤖 PLM管家 自动生成成本对比报告"
      - role: "PM"
        timeout: "24h"
        
  # 退换货审批
  return_exchange:
    conditions:
      - amount: "<500"
        flow:
          - role: "🤖 客服专员"
            auto_approve: true
            rules: "符合7天无理由退货"
      - amount: ">=500"
        flow:
          - role: "售后负责人"
            timeout: "12h"
```

#### 4.4.3 审批表单设计

```yaml
approval_forms:
  # 采购订单审批表单
  purchase_order_approval:
    fields:
      - id: "po_code"
        label: "订单编号"
        type: "text"
        readonly: true
      - id: "supplier_name"
        label: "供应商"
        type: "text"
        readonly: true
      - id: "total_amount"
        label: "订单金额"
        type: "number"
        readonly: true
      - id: "items_summary"
        label: "采购明细"
        type: "table"
        readonly: true
      - id: "ai_analysis"
        label: "AI分析"
        type: "rich_text"
        readonly: true
        description: "🤖 采购专员 自动生成的采购建议"
      - id: "approval_comment"
        label: "审批意见"
        type: "textarea"
        required: false
```

### 4.5 多维表格集成设计

#### 4.5.1 多维表格清单

| 表格名称 | 用途 | 同步模式 | 主要用户 |
|---------|------|---------|---------|
| 物料主数据 | 物料信息管理 | 双向同步 | 硬件工程师、🤖 PLM管家 |
| 供应商管理 | 供应商信息 | 双向同步 | 采购、🤖 采购专员 |
| 采购订单 | 采购订单跟踪 | 双向同步 | 采购、🤖 采购专员 |
| 库存台账 | 库存实时数据 | 系统→表格 | 仓库、🤖 库存管家 |
| 销售订单 | 订单管理 | 双向同步 | 销售、🤖 订单处理员 |
| 售后工单 | 售后服务跟踪 | 双向同步 | 售后、🤖 客服专员 |
| 产品清单 | 产品信息 | 系统→表格 | 全员 |
| 项目看板 | 项目进度 | 系统→表格 | PM、🤖 项目督导 |

#### 4.5.2 同步架构

```
┌─────────────┐      Webhook      ┌─────────────┐
│  飞书多维表格 │ ───────────────> │  集成服务    │
│             │ <─────────────── │             │
└─────────────┘    API推送        └──────┬──────┘
                                         │
                                         v
                                  ┌─────────────┐
                                  │  PostgreSQL │
                                  └─────────────┘
```

#### 4.5.3 同步配置示例

```yaml
bitable_sync_configs:
  # 物料主数据表
  materials:
    bitable_app_token: "bascnXXXXXXXX"
    table_id: "tblXXXXXXXX"
    sync_mode: "bidirectional"
    
    # 字段映射
    field_mapping:
      - db_field: "material_code"
        bitable_field: "物料编码"
        bitable_type: "text"
        sync_direction: "both"
        primary_key: true
        
      - db_field: "name"
        bitable_field: "物料名称"
        bitable_type: "text"
        sync_direction: "both"
        
      - db_field: "category"
        bitable_field: "分类"
        bitable_type: "single_select"
        sync_direction: "both"
        options: ["电子", "光学", "结构", "包装", "辅料"]
        
      - db_field: "specification"
        bitable_field: "规格"
        bitable_type: "text"
        sync_direction: "both"
        
      - db_field: "standard_cost"
        bitable_field: "标准成本"
        bitable_type: "number"
        sync_direction: "both"
        editable_by: ["采购负责人", "财务负责人"]
        
      - db_field: "status"
        bitable_field: "状态"
        bitable_type: "single_select"
        sync_direction: "both"
        options: ["启用", "停用", "淘汰"]
        
      - db_field: "updated_at"
        bitable_field: "更新时间"
        bitable_type: "datetime"
        sync_direction: "db_to_bitable"
        
    # 同步规则
    sync_rules:
      # 从数据库同步到表格
      db_to_bitable:
        trigger: "data_change"
        delay: "realtime"  # 或 "batch_5min"
        
      # 从表格同步到数据库
      bitable_to_db:
        trigger: "webhook"
        validation: true  # 执行业务校验
        audit_log: true  # 记录变更日志
        
    # 权限规则
    permissions:
      view: ["*"]  # 所有人可查看
      edit: ["硬件研发部", "采购部", "🤖 PLM管家", "🤖 采购专员"]
      
  # 采购订单表
  purchase_orders:
    bitable_app_token: "bascnYYYYYYYY"
    table_id: "tblYYYYYYYY"
    sync_mode: "bidirectional"
    
    field_mapping:
      - db_field: "po_code"
        bitable_field: "订单号"
        primary_key: true
        
      - db_field: "supplier_id"
        bitable_field: "供应商"
        bitable_type: "link"  # 关联供应商表
        link_table_id: "tblSUPPLIER"
        
      - db_field: "status"
        bitable_field: "状态"
        bitable_type: "single_select"
        options: ["草稿", "待审批", "已审批", "已发送", "部分到货", "已完成"]
        # 状态变更触发通知
        on_change:
          - value: "已发送"
            notify: ["🤖 采购专员"]
            message: "采购订单 {po_code} 已发送供应商，请跟进"
          - value: "部分到货"
            notify: ["🤖 库存管家"]
            message: "采购订单 {po_code} 部分到货，请安排入库"
```

### 4.6 消息与通知设计

#### 4.6.1 通知场景

| 场景 | 触发条件 | 通知对象 | 消息类型 |
|-----|---------|---------|---------|
| 任务分配 | 任务被分配 | 任务负责人 | 飞书任务通知 |
| 任务到期提醒 | 距离截止不足24h | 任务负责人 | 飞书消息 |
| 任务延期 | 任务超过截止日期 | 负责人+监督人 | 飞书消息（加急） |
| 审批待办 | 新审批到达 | 审批人 | 飞书审批通知 |
| 库存预警 | 库存低于安全线 | 采购负责人+🤖采购专员 | 飞书消息+卡片 |
| 订单状态变更 | 订单状态变化 | 相关人员 | 飞书消息 |
| AI操作完成 | AI完成重要操作 | 监督人 | 飞书消息 |
| AI需要协助 | AI遇到无法处理的情况 | 监督人 | 飞书消息（@提醒） |

#### 4.6.2 消息卡片模板

```json
{
  "msg_type": "interactive",
  "card": {
    "header": {
      "title": {
        "tag": "plain_text",
        "content": "📦 采购订单已创建"
      },
      "template": "blue"
    },
    "elements": [
      {
        "tag": "div",
        "fields": [
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**订单号：**PO-2026-0001"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**供应商：**供应商A"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**金额：**¥12,500.00"
            }
          },
          {
            "is_short": true,
            "text": {
              "tag": "lark_md",
              "content": "**状态：**待审批"
            }
          }
        ]
      },
      {
        "tag": "div",
        "text": {
          "tag": "lark_md",
          "content": "**🤖 AI分析：**\n该订单基于MRP计算结果自动生成，供应商A历史交期准时率98%，建议批准。"
        }
      },
      {
        "tag": "hr"
      },
      {
        "tag": "action",
        "actions": [
          {
            "tag": "button",
            "text": {
              "tag": "plain_text",
              "content": "查看详情"
            },
            "type": "primary",
            "url": "https://plm.nimo.com/po/PO-2026-0001"
          },
          {
            "tag": "button",
            "text": {
              "tag": "plain_text",
              "content": "在表格中查看"
            },
            "type": "default",
            "url": "https://xxx.feishu.cn/base/xxx"
          }
        ]
      }
    ]
  }
}
```

---

## 五、AI员工工作场景设计

### 5.1 场景1：自动补货流程

```
触发：库存低于安全线
参与：🤖 库存管家、🤖 采购专员、采购负责人

流程：
┌─────────────────────────────────────────────────────────────────┐
│ 1. 🤖 库存管家 检测到库存预警                                    │
│    ↓                                                            │
│ 2. 🤖 库存管家 发送群消息：                                      │
│    "⚠️ 库存预警：[MAT-001 主控芯片]                             │
│     当前库存：50个，安全库存：100个                              │
│     @🤖采购专员 请处理"                                         │
│    ↓                                                            │
│ 3. 🤖 采购专员 收到消息，开始分析：                              │
│    - 查询历史采购记录                                           │
│    - 查询供应商报价                                             │
│    - 计算经济批量                                               │
│    ↓                                                            │
│ 4. 🤖 采购专员 在群里回复建议：                                  │
│    "已分析完成：                                                │
│     推荐供应商：供应商A（准时率98%）                             │
│     建议数量：200个，金额：¥3,100                               │
│     @张三（采购负责人）请确认"                                  │
│    ↓                                                            │
│ 5. 张三回复："确认下单"                                         │
│    ↓                                                            │
│ 6. 🤖 采购专员 自动执行：                                        │
│    - 创建采购订单                                               │
│    - 提交审批（金额<5000，自动通过）                            │
│    - 发送给供应商                                               │
│    - 更新多维表格                                               │
│    - 设置到货提醒                                               │
│    ↓                                                            │
│ 7. 🤖 采购专员 发送确认消息：                                    │
│    "✅ 采购订单 PO-2026-0001 已创建并发送供应商                  │
│     预计到货：2026-02-10"                                       │
└─────────────────────────────────────────────────────────────────┘
```

### 5.2 场景2：项目日报生成

```
触发：每日定时（早上9:00）
参与：🤖 项目督导

流程：
┌─────────────────────────────────────────────────────────────────┐
│ 1. 🤖 项目督导 定时任务触发（每日09:00）                         │
│    ↓                                                            │
│ 2. 🤖 项目督导 收集数据：                                        │
│    - 从PLM系统获取所有进行中项目的任务                          │
│    - 从飞书任务获取最新状态                                     │
│    - 对比昨日数据，识别变化                                     │
│    ↓                                                            │
│ 3. 🤖 项目督导 分析风险：                                        │
│    - 检查延期任务                                               │
│    - 检查即将到期任务                                           │
│    - 检查资源冲突                                               │
│    ↓                                                            │
│ 4. 🤖 项目督导 在项目群发送日报：                                │
│    "📊 G1圆框眼镜项目日报 (2026-02-05)                          │
│                                                                 │
│     ✅ 昨日完成：                                               │
│     - EVT-001-01 电路原理图设计（张三）                         │
│                                                                 │
│     🔄 进行中：                                                 │
│     - EVT-001-02 PCB布局（进度60%）                             │
│                                                                 │
│     ⚠️ 风险预警：                                               │
│     - EVT-001-04 光学设计 已延期2天                             │
│       @李四 请更新进度                                          │
│                                                                 │
│     📅 本周里程碑：周五 EVT-002 样机制作开始"                   │
│    ↓                                                            │
│ 5. 如有严重延期，私聊PM并抄送CEO                                 │
│    ↓                                                            │
│ 6. 更新项目甘特图数据                                            │
└─────────────────────────────────────────────────────────────────┘
```

### 5.3 场景3：BOM变更评估

```
触发：工程师提交BOM变更
参与：工程师、🤖 PLM管家、技术负责人、PM

流程：
┌─────────────────────────────────────────────────────────────────┐
│ 1. 工程师在系统提交BOM变更申请                                   │
│    ↓                                                            │
│ 2. 系统自动分配任务给 🤖 PLM管家：                               │
│    "评估BOM变更影响"                                            │
│    ↓                                                            │
│ 3. 🤖 PLM管家 自动分析：                                         │
│    - 对比新旧BOM差异                                            │
│    - 计算成本变化                                               │
│    - 检查库存影响（现有物料是否报废）                           │
│    - 检查供应商影响（是否需要新供应商）                         │
│    - 检查生产影响（工艺是否变化）                               │
│    ↓                                                            │
│ 4. 🤖 PLM管家 生成评估报告并附加到审批：                         │
│    "📋 BOM变更影响评估报告                                      │
│                                                                 │
│     变更内容：替换主控芯片                                      │
│     旧物料：MAT-001 → 新物料：MAT-002                           │
│                                                                 │
│     💰 成本影响：+¥5.2/台（+3.4%）                              │
│     📦 库存影响：MAT-001库存120个需消耗                         │
│     🏭 供应商影响：无变化（同一供应商）                         │
│     ⚙️ 生产影响：SMT贴片程序需更新                              │
│                                                                 │
│     建议：批准变更，但建议先消耗MAT-001库存"                    │
│    ↓                                                            │
│ 5. 技术负责人在飞书审批，看到AI报告，做出决策                    │
│    ↓                                                            │
│ 6. 审批通过后，🤖 PLM管家 自动：                                 │
│    - 更新BOM版本                                                │
│    - 通知相关人员                                               │
│    - 创建后续任务（更新SMT程序）                                │
└─────────────────────────────────────────────────────────────────┘
```

### 5.4 场景4：售后问题自动诊断

```
触发：客户报修
参与：客户、🤖 客服专员、售后负责人

流程：
┌─────────────────────────────────────────────────────────────────┐
│ 1. 客户通过售后渠道提交问题：                                    │
│    "眼镜连接手机经常断开"                                       │
│    ↓                                                            │
│ 2. 🤖 客服专员 收到工单，开始分析：                              │
│    - 解析问题关键词：蓝牙、断连                                 │
│    - 查询产品序列号，获取生产批次                               │
│    - 查询该批次历史故障记录                                     │
│    - 匹配知识库中的解决方案                                     │
│    ↓                                                            │
│ 3. 🤖 客服专员 发现：                                            │
│    "该批次（2026-01批）蓝牙断连投诉率2.3%，                      │
│     高于正常水平（0.5%），                                      │
│     已知原因：蓝牙模块固件问题，                                │
│     解决方案：OTA升级固件V2.1"                                  │
│    ↓                                                            │
│ 4. 🤖 客服专员 自动回复客户：                                    │
│    "您好，感谢反馈。                                            │
│     经诊断，您的问题可通过固件升级解决。                        │
│     请按以下步骤操作：                                          │
│     1. 打开nimo APP                                             │
│     2. 进入设置-固件更新                                        │
│     3. 升级到V2.1版本                                           │
│     如仍有问题请回复，我们将安排换货。"                         │
│    ↓                                                            │
│ 5. 同时，🤖 客服专员 在内部群通知：                              │
│    "@张三（质量工程师）2026-01批次蓝牙问题投诉+1，              │
│     累计投诉23例，建议关注"                                     │
│    ↓                                                            │
│ 6. 如果客户回复问题未解决：                                      │
│    🤖 客服专员 创建退换货工单，                                  │
│    提交给售后负责人审批                                         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 六、技术实现方案

### 6.1 OpenClaw Agent架构

```
┌─────────────────────────────────────────────────────────────────┐
│                    OpenClaw Gateway                             │
│                   (nimo.openclaw.ai)                            │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        v                     v                     v
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│  🤖 PLM管家   │    │  🤖 采购专员  │    │  🤖 项目督导  │
│  Session      │    │  Session      │    │  Session      │
└───────────────┘    └───────────────┘    └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
                              v
┌─────────────────────────────────────────────────────────────────┐
│                      共享工具层                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │ 飞书工具  │ │ PLM工具  │ │ ERP工具  │ │ 分析工具  │          │
│  │ feishu_* │ │ plm_*    │ │ erp_*    │ │ analytics │          │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### 6.2 Agent工具定义

```yaml
# OpenClaw Agent 工具定义

tools:
  # ===== 飞书工具 =====
  - name: feishu_send_message
    description: "发送飞书消息给指定用户或群组"
    parameters:
      target: "用户ID或群组ID"
      message: "消息内容"
      message_type: "text/card/image"
      
  - name: feishu_create_task
    description: "创建飞书任务"
    parameters:
      summary: "任务标题"
      description: "任务描述"
      assignee: "负责人ID"
      due_date: "截止日期"
      
  - name: feishu_update_task
    description: "更新飞书任务状态"
    parameters:
      task_id: "任务ID"
      status: "任务状态"
      progress: "进度百分比"
      
  - name: feishu_create_approval
    description: "发起飞书审批"
    parameters:
      approval_code: "审批定义代码"
      form_data: "表单数据"
      
  - name: feishu_bitable_query
    description: "查询飞书多维表格数据"
    parameters:
      app_token: "表格应用Token"
      table_id: "表格ID"
      filter: "过滤条件"
      
  - name: feishu_bitable_update
    description: "更新飞书多维表格记录"
    parameters:
      app_token: "表格应用Token"
      table_id: "表格ID"
      record_id: "记录ID"
      fields: "字段数据"

  # ===== PLM工具 =====
  - name: plm_get_product
    description: "获取产品信息"
    parameters:
      product_id: "产品ID或SKU"
      
  - name: plm_get_bom
    description: "获取BOM信息"
    parameters:
      product_id: "产品ID"
      bom_type: "BOM类型"
      version: "版本号"
      
  - name: plm_update_bom
    description: "更新BOM"
    parameters:
      bom_id: "BOM ID"
      items: "BOM物料明细"
      
  - name: plm_submit_bom_approval
    description: "提交BOM审批"
    parameters:
      bom_id: "BOM ID"
      comment: "提交说明"
      
  - name: plm_get_project_tasks
    description: "获取项目任务列表"
    parameters:
      project_id: "项目ID"
      include_subtasks: "是否包含子任务"
      
  - name: plm_update_task_progress
    description: "更新任务进度"
    parameters:
      task_id: "任务ID"
      progress: "进度百分比"
      comment: "进度说明"

  # ===== ERP工具 =====
  - name: erp_get_inventory
    description: "查询库存信息"
    parameters:
      material_id: "物料ID"
      warehouse_id: "仓库ID（可选）"
      
  - name: erp_get_inventory_alerts
    description: "获取库存预警列表"
    parameters:
      alert_type: "预警类型"
      
  - name: erp_create_purchase_order
    description: "创建采购订单"
    parameters:
      supplier_id: "供应商ID"
      items: "采购明细"
      
  - name: erp_get_supplier_performance
    description: "获取供应商绩效"
    parameters:
      supplier_id: "供应商ID"
      
  - name: erp_create_work_order
    description: "创建生产工单"
    parameters:
      product_id: "产品ID"
      quantity: "数量"
      planned_date: "计划日期"
      
  - name: erp_get_sales_orders
    description: "查询销售订单"
    parameters:
      status: "订单状态"
      date_range: "日期范围"

  # ===== 分析工具 =====
  - name: analytics_cost_compare
    description: "BOM成本对比分析"
    parameters:
      bom_id_1: "BOM 1"
      bom_id_2: "BOM 2"
      
  - name: analytics_supplier_recommend
    description: "供应商推荐"
    parameters:
      material_id: "物料ID"
      quantity: "采购数量"
      
  - name: analytics_quality_trend
    description: "质量趋势分析"
    parameters:
      product_id: "产品ID"
      time_range: "时间范围"
```

### 6.3 事件处理配置

```yaml
# Agent事件订阅配置

event_handlers:
  # 🤖 库存管家 - 库存预警处理
  inventory_agent:
    - event: "inventory.below_safety_stock"
      handler: |
        1. 获取物料信息和当前库存
        2. 计算建议采购量
        3. 发送群消息通知 @采购专员
        4. 记录预警日志
        
  # 🤖 采购专员 - 采购建议处理
  procurement_agent:
    - event: "message.mention_me"
      filter: "contains('库存预警')"
      handler: |
        1. 解析消息中的物料信息
        2. 查询供应商报价
        3. 分析历史采购数据
        4. 生成采购建议
        5. 回复群消息，@采购负责人确认
        
    - event: "message.reply"
      filter: "from_supervisor AND contains('确认')"
      handler: |
        1. 创建采购订单
        2. 提交审批
        3. 更新多维表格
        4. 发送确认消息
        
  # 🤖 项目督导 - 定时任务
  project_agent:
    - event: "schedule.daily_9am"
      handler: |
        1. 获取所有进行中项目
        2. 检查任务状态变化
        3. 识别风险任务
        4. 生成日报
        5. 发送到项目群
        
    - event: "task.overdue"
      handler: |
        1. 识别延期任务
        2. 发送提醒给负责人
        3. 如延期超过3天，通知PM
```

---

## 七、安全与审计设计

### 7.1 AI操作审计

```yaml
audit_requirements:
  # 必须记录的操作
  mandatory_logging:
    - category: "data_modification"
      operations: ["create", "update", "delete"]
      log_fields: ["operator", "timestamp", "before_value", "after_value", "reasoning"]
      
    - category: "approval"
      operations: ["submit", "approve", "reject"]
      log_fields: ["operator", "timestamp", "decision", "reasoning", "confidence"]
      
    - category: "external_communication"
      operations: ["send_message", "send_email", "create_po"]
      log_fields: ["operator", "timestamp", "recipient", "content_summary"]
      
  # 审计报告
  audit_reports:
    - name: "AI操作日报"
      frequency: "daily"
      content: ["操作统计", "异常操作", "人工干预统计"]
      recipients: ["系统管理员"]
      
    - name: "AI操作周报"
      frequency: "weekly"
      content: ["操作趋势", "自主级别分析", "改进建议"]
      recipients: ["CEO", "系统管理员"]
```

### 7.2 权限边界

```yaml
permission_boundaries:
  # AI员工不可执行的操作
  forbidden_actions:
    - "删除产品/BOM/订单"  # 只能标记为废弃
    - "修改历史数据"  # 只能追加
    - "直接付款执行"  # 必须人工确认
    - "修改权限配置"  # 仅管理员
    - "关闭安全预警"  # 必须人工处理
    
  # 金额限制
  amount_limits:
    - agent: "🤖 采购专员"
      max_single_po: 50000  # 单笔采购上限
      max_daily_po: 200000  # 日累计上限
      
    - agent: "🤖 客服专员"
      max_refund: 500  # 自动退款上限
```

### 7.3 人工干预机制

```yaml
human_intervention:
  # 强制人工干预场景
  mandatory_review:
    - condition: "单笔金额 > 50000"
      action: "暂停操作，通知监督人"
    - condition: "置信度 < 0.7"
      action: "请求人工确认"
    - condition: "连续3次相似操作"
      action: "暂停并请求审核"
    - condition: "操作影响 > 10个关联对象"
      action: "生成影响报告，请求确认"
      
  # 紧急停止
  emergency_stop:
    - trigger: "人工发送'停止'命令"
    - action: "立即暂停所有AI操作"
    - notify: ["系统管理员", "CEO"]
```

---

## 八、实施路线图

### 8.1 阶段规划

| 阶段 | 时间 | 目标 | 交付物 |
|-----|------|------|-------|
| P0 | 第1-2周 | 基础架构 | 飞书应用配置、Agent身份体系 |
| P1 | 第3-6周 | 核心集成 | 任务同步、审批集成、多维表格同步 |
| P2 | 第7-10周 | 首批AI员工 | 🤖 PLM管家、🤖 项目督导 上线 |
| P3 | 第11-14周 | 供应链AI | 🤖 采购专员、🤖 库存管家 上线 |
| P4 | 第15-18周 | 销售售后AI | 🤖 订单处理员、🤖 客服专员 上线 |
| P5 | 第19-20周 | 优化迭代 | 性能优化、规则调优 |

### 8.2 首批AI员工上线计划

```yaml
phase_2_rollout:
  week_7:
    - deploy: "🤖 PLM管家"
      initial_scope: "只读操作 + 提醒通知"
      autonomy_level: 4
      
  week_8:
    - upgrade: "🤖 PLM管家"
      new_scope: "BOM编辑 + 审批提交"
      autonomy_level: 3
      
  week_9:
    - deploy: "🤖 项目督导"
      initial_scope: "日报生成 + 进度提醒"
      autonomy_level: 4
      
  week_10:
    - review: "首批AI员工效果评估"
      metrics: ["操作准确率", "人工干预率", "用户满意度"]
      decision: "是否进入P3阶段"
```

---

## 九、成功指标

### 9.1 效率指标

| 指标 | 当前（人工） | 目标（AI辅助） | 提升 |
|-----|------------|--------------|------|
| 日报生成时间 | 30分钟/天 | 自动生成 | 100% |
| 采购订单创建 | 15分钟/单 | 2分钟/单 | 87% |
| 库存预警响应 | 4小时 | 10分钟 | 96% |
| BOM变更评估 | 2小时 | 10分钟 | 92% |
| 售后问题诊断 | 30分钟 | 5分钟 | 83% |

### 9.2 质量指标

| 指标 | 目标 |
|-----|------|
| AI操作准确率 | > 95% |
| 人工干预率 | < 10% |
| 审计异常率 | < 1% |
| 用户满意度 | > 4.5/5 |

---

**文档状态**：v1.0 完成  
**下次评审**：2026-02-06
