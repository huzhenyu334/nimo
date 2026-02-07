# nimo PLM+ERP MCP接口设计文档 v1.0

> 文档版本：1.0
> 更新日期：2026-02-05
> 编写：OpenClaw AI
> 关联文档：PLM PRD v2.0, ERP PRD v2.0, 技术架构选型文档 v2.0

---

## 一、概述

### 1.1 什么是MCP

**MCP (Model Context Protocol)** 是Anthropic提出的AI Agent工具调用协议标准，定义了Agent与外部系统交互的规范方式。

**核心概念：**
- **Tools（工具）**：Agent可调用的操作，如"创建产品"、"查询库存"
- **Resources（资源）**：Agent可访问的数据，如文档、报表
- **Prompts（提示）**：预定义的复杂操作模板

### 1.2 为什么选择MCP

| 对比项 | 直接REST API | MCP协议 |
|-------|-------------|---------|
| Agent理解成本 | 高（需理解HTTP、Auth、参数格式） | 低（工具抽象） |
| 错误处理 | 需自行解析 | 协议层标准化 |
| 类型安全 | 依赖文档 | Schema强制校验 |
| 可组合性 | 手动编排 | 原生支持 |
| 生态支持 | 通用 | Claude/OpenClaw原生 |

### 1.3 架构设计

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           AI Agent 层                                   │
│                                                                         │
│    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐            │
│    │   OpenClaw   │    │    Claude    │    │   其他Agent  │            │
│    │  (主要用户)   │    │   Desktop    │    │              │            │
│    └──────┬───────┘    └──────┬───────┘    └──────┬───────┘            │
└───────────┼────────────────────┼────────────────────┼────────────────────┘
            │                    │                    │
            │         MCP Protocol (JSON-RPC 2.0)     │
            │                    │                    │
            └────────────────────┼────────────────────┘
                                 │
┌────────────────────────────────┼────────────────────────────────────────┐
│                         MCP Server 层                                   │
│                                │                                        │
│    ┌───────────────────────────┼───────────────────────────┐           │
│    │                    MCP Gateway                         │           │
│    │           (统一入口、认证、日志、限流)                  │           │
│    └───────────────────────────┼───────────────────────────┘           │
│                                │                                        │
│         ┌──────────────────────┼──────────────────────┐                │
│         │                      │                      │                │
│         ▼                      ▼                      ▼                │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐         │
│  │ nimo-plm-mcp │      │ nimo-erp-mcp │      │nimo-report-mcp│         │
│  │              │      │              │      │              │         │
│  │ 17个工具     │      │ 24个工具     │      │ 8个工具      │         │
│  └──────┬───────┘      └──────┬───────┘      └──────┬───────┘         │
└─────────┼──────────────────────┼──────────────────────┼─────────────────┘
          │                      │                      │
          │            Internal HTTP/gRPC              │
          │                      │                      │
┌─────────┼──────────────────────┼──────────────────────┼─────────────────┐
│         ▼                      ▼                      ▼                │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐         │
│  │   PLM 服务   │      │   ERP 服务   │      │  报表服务    │         │
│  │  (Go + Gin)  │      │  (Go + Gin)  │      │  (Go + Gin)  │         │
│  └──────────────┘      └──────────────┘      └──────────────┘         │
│                                                                         │
│                           业务服务层                                    │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 二、MCP Server规划

### 2.1 Server划分

| Server | 职责 | 工具数量 | 优先级 |
|--------|------|---------|--------|
| **nimo-plm-mcp** | 产品生命周期管理 | 17 | P0 |
| **nimo-erp-mcp** | 企业资源管理 | 24 | P0 |
| **nimo-report-mcp** | 报表与数据分析 | 8 | P1 |
| **nimo-workflow-mcp** | 审批流程管理 | 6 | P1 |
| **nimo-integration-mcp** | 外部系统集成 | 5 | P2 |

### 2.2 技术选型

| 组件 | 技术 | 说明 |
|-----|------|------|
| MCP Server框架 | @modelcontextprotocol/sdk (TypeScript) | 官方SDK，成熟稳定 |
| 运行时 | Node.js 20 LTS | 与SDK兼容 |
| 传输协议 | stdio / SSE | stdio用于本地，SSE用于远程 |
| 内部通信 | HTTP REST | 调用Go后端服务 |
| 配置管理 | dotenv + yaml | 环境变量 + 配置文件 |

---

## 三、nimo-plm-mcp 工具定义

### 3.1 产品管理工具

#### plm_search_products
```yaml
name: plm_search_products
description: |
  搜索产品列表。支持按名称、编码、类别、状态等条件筛选。
  返回产品基本信息列表，不包含BOM详情。
  
inputSchema:
  type: object
  properties:
    query:
      type: string
      description: 搜索关键词，匹配产品名称或编码
    category:
      type: string
      enum: [frame, temple, lens, platform, accessory]
      description: 产品类别（镜框/镜腿/镜片/平台/配件）
    status:
      type: string
      enum: [draft, developing, active, discontinued]
      description: 产品状态
    created_after:
      type: string
      format: date
      description: 创建时间起始（ISO日期）
    created_before:
      type: string
      format: date
      description: 创建时间截止
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 20
      maximum: 100

returns:
  type: object
  properties:
    items:
      type: array
      items:
        $ref: "#/definitions/ProductSummary"
    total:
      type: integer
    page:
      type: integer
    page_size:
      type: integer
```

#### plm_get_product
```yaml
name: plm_get_product
description: |
  获取产品详细信息，包含基本信息、当前BOM版本、关联项目等。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
      description: 产品ID
  required: [product_id]

returns:
  $ref: "#/definitions/ProductDetail"
```

#### plm_create_product
```yaml
name: plm_create_product
description: |
  创建新产品。创建后状态为"草稿"，需要完善BOM后才能发布。
  
inputSchema:
  type: object
  properties:
    name:
      type: string
      description: 产品名称
      maxLength: 100
    code:
      type: string
      description: 产品编码（可选，不填则自动生成）
      pattern: "^[A-Z0-9-]+$"
    category:
      type: string
      enum: [frame, temple, lens, platform, accessory]
      description: 产品类别
    description:
      type: string
      description: 产品描述
    specs:
      type: object
      description: 规格参数（自定义JSON）
    base_product_id:
      type: string
      description: 基于哪个产品创建（可选，会复制BOM结构）
  required: [name, category]
```

#### plm_update_product
```yaml
name: plm_update_product
description: |
  更新产品信息。注意：已发布的产品修改需要走ECN流程。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
    name:
      type: string
    description:
      type: string
    specs:
      type: object
  required: [product_id]
```

### 3.2 BOM管理工具

#### plm_get_bom
```yaml
name: plm_get_bom
description: |
  获取产品BOM清单。返回完整的物料树结构。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
      description: 产品ID
    version:
      type: string
      description: BOM版本号（不填则返回最新版本）
    expand_level:
      type: integer
      description: 展开层级，-1表示全部展开
      default: -1
      minimum: -1
      maximum: 10
    include_cost:
      type: boolean
      description: 是否包含成本信息
      default: false
    include_inventory:
      type: boolean
      description: 是否包含当前库存信息
      default: false
  required: [product_id]

returns:
  type: object
  properties:
    product_id:
      type: string
    version:
      type: string
    status:
      type: string
      enum: [draft, released, obsolete]
    items:
      type: array
      items:
        $ref: "#/definitions/BOMItem"
    total_cost:
      type: number
      description: 总成本（仅当include_cost=true时返回）
```

#### plm_add_bom_item
```yaml
name: plm_add_bom_item
description: |
  向BOM添加物料。只能操作草稿状态的BOM。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
    parent_item_id:
      type: string
      description: 父级物料ID（顶级物料不填）
    material_id:
      type: string
      description: 物料ID
    quantity:
      type: number
      description: 用量
      minimum: 0.001
    unit:
      type: string
      description: 单位
      enum: [pcs, kg, m, set]
    position:
      type: string
      description: 位置编号
    notes:
      type: string
      description: 备注
  required: [product_id, material_id, quantity]
```

#### plm_update_bom_item
```yaml
name: plm_update_bom_item
description: |
  更新BOM中的物料信息。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
    item_id:
      type: string
      description: BOM行项ID
    quantity:
      type: number
    position:
      type: string
    notes:
      type: string
  required: [product_id, item_id]
```

#### plm_remove_bom_item
```yaml
name: plm_remove_bom_item
description: |
  从BOM移除物料。如果该物料有子物料，会一并移除。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
    item_id:
      type: string
  required: [product_id, item_id]
```

#### plm_release_bom
```yaml
name: plm_release_bom
description: |
  发布BOM版本。发布后BOM不可修改，如需修改需创建新版本。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
    version:
      type: string
      description: 版本号（如 1.0, 1.1）
    release_notes:
      type: string
      description: 发布说明
  required: [product_id]
```

#### plm_compare_bom
```yaml
name: plm_compare_bom
description: |
  对比两个BOM版本的差异。返回新增、删除、修改的物料清单。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
    version_a:
      type: string
      description: 版本A
    version_b:
      type: string
      description: 版本B
  required: [product_id, version_a, version_b]

returns:
  type: object
  properties:
    added:
      type: array
      description: 新增的物料
    removed:
      type: array
      description: 删除的物料
    modified:
      type: array
      description: 修改的物料（含before/after）
```

### 3.3 项目任务工具

#### plm_list_projects
```yaml
name: plm_list_projects
description: |
  获取项目列表。项目是产品研发的管理单元，包含EVT/DVT/PVT/MP四个阶段。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
      description: 按产品筛选
    status:
      type: string
      enum: [planning, evt, dvt, pvt, mp, completed, cancelled]
      description: 项目状态/阶段
    owner_id:
      type: string
      description: 项目负责人ID
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 20
```

#### plm_get_project
```yaml
name: plm_get_project
description: |
  获取项目详情，包含当前阶段、任务列表、里程碑等。
  
inputSchema:
  type: object
  properties:
    project_id:
      type: string
  required: [project_id]
```

#### plm_create_project
```yaml
name: plm_create_project
description: |
  创建研发项目。可基于模板自动生成任务结构。
  
inputSchema:
  type: object
  properties:
    name:
      type: string
      description: 项目名称
    product_id:
      type: string
      description: 关联产品ID
    template:
      type: string
      enum: [standard, fast_track, custom]
      description: 项目模板（标准/快速/自定义）
      default: standard
    owner_id:
      type: string
      description: 项目负责人ID
    planned_start:
      type: string
      format: date
    planned_end:
      type: string
      format: date
  required: [name, product_id]
```

#### plm_list_tasks
```yaml
name: plm_list_tasks
description: |
  获取项目任务列表。
  
inputSchema:
  type: object
  properties:
    project_id:
      type: string
      description: 项目ID
    phase:
      type: string
      enum: [evt, dvt, pvt, mp]
      description: 阶段筛选
    status:
      type: string
      enum: [pending, in_progress, completed, blocked]
    assignee_id:
      type: string
      description: 执行人ID
    overdue_only:
      type: boolean
      description: 只显示逾期任务
      default: false
  required: [project_id]
```

#### plm_update_task
```yaml
name: plm_update_task
description: |
  更新任务状态或信息。
  
inputSchema:
  type: object
  properties:
    task_id:
      type: string
    status:
      type: string
      enum: [pending, in_progress, completed, blocked]
    progress:
      type: integer
      minimum: 0
      maximum: 100
      description: 完成百分比
    assignee_id:
      type: string
    due_date:
      type: string
      format: date
    notes:
      type: string
  required: [task_id]
```

### 3.4 ECN变更管理工具

#### plm_create_ecn
```yaml
name: plm_create_ecn
description: |
  创建工程变更通知(ECN)。用于已发布产品的设计变更管理。
  
inputSchema:
  type: object
  properties:
    title:
      type: string
      description: 变更标题
    product_id:
      type: string
      description: 变更的产品ID
    change_type:
      type: string
      enum: [design, material, process, document]
      description: 变更类型
    reason:
      type: string
      description: 变更原因
    description:
      type: string
      description: 变更详细描述
    affected_items:
      type: array
      items:
        type: object
        properties:
          item_type:
            type: string
            enum: [bom_item, document, drawing]
          item_id:
            type: string
          change_description:
            type: string
      description: 受影响的项目列表
    urgency:
      type: string
      enum: [low, medium, high, critical]
      default: medium
  required: [title, product_id, change_type, reason]
```

#### plm_get_ecn
```yaml
name: plm_get_ecn
description: |
  获取ECN详情，包含审批状态和历史。
  
inputSchema:
  type: object
  properties:
    ecn_id:
      type: string
  required: [ecn_id]
```

#### plm_submit_ecn
```yaml
name: plm_submit_ecn
description: |
  提交ECN进入审批流程。
  
inputSchema:
  type: object
  properties:
    ecn_id:
      type: string
    approvers:
      type: array
      items:
        type: string
      description: 审批人ID列表（按顺序）
  required: [ecn_id]
```

---

## 四、nimo-erp-mcp 工具定义

### 4.1 供应商管理工具

#### erp_search_suppliers
```yaml
name: erp_search_suppliers
description: |
  搜索供应商列表。
  
inputSchema:
  type: object
  properties:
    query:
      type: string
      description: 搜索关键词（名称、编码）
    category:
      type: string
      enum: [raw_material, component, packaging, service]
      description: 供应商类别
    status:
      type: string
      enum: [active, inactive, blacklisted]
    rating_min:
      type: integer
      minimum: 1
      maximum: 5
      description: 最低评级
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 20
```

#### erp_get_supplier
```yaml
name: erp_get_supplier
description: |
  获取供应商详情，包含联系人、资质、历史交易等。
  
inputSchema:
  type: object
  properties:
    supplier_id:
      type: string
  required: [supplier_id]
```

#### erp_get_supplier_materials
```yaml
name: erp_get_supplier_materials
description: |
  获取供应商可供应的物料清单及报价。
  
inputSchema:
  type: object
  properties:
    supplier_id:
      type: string
    material_category:
      type: string
      description: 物料类别筛选
  required: [supplier_id]
```

### 4.2 采购管理工具

#### erp_create_purchase_request
```yaml
name: erp_create_purchase_request
description: |
  创建采购申请。申请审批通过后可转为采购订单。
  
inputSchema:
  type: object
  properties:
    title:
      type: string
      description: 申请标题
    items:
      type: array
      items:
        type: object
        properties:
          material_id:
            type: string
          quantity:
            type: number
          expected_date:
            type: string
            format: date
          notes:
            type: string
        required: [material_id, quantity]
    reason:
      type: string
      description: 采购原因
    urgency:
      type: string
      enum: [normal, urgent, critical]
      default: normal
  required: [items]
```

#### erp_create_purchase_order
```yaml
name: erp_create_purchase_order
description: |
  创建采购订单。可以从采购申请转换，也可直接创建。
  
inputSchema:
  type: object
  properties:
    supplier_id:
      type: string
      description: 供应商ID
    items:
      type: array
      items:
        type: object
        properties:
          material_id:
            type: string
          quantity:
            type: number
          unit_price:
            type: number
          tax_rate:
            type: number
            default: 0.13
        required: [material_id, quantity, unit_price]
    expected_delivery_date:
      type: string
      format: date
    delivery_address:
      type: string
    payment_terms:
      type: string
      enum: [prepaid, cod, net30, net60]
      default: net30
    notes:
      type: string
    from_request_id:
      type: string
      description: 来源采购申请ID（可选）
  required: [supplier_id, items]
```

#### erp_get_purchase_order
```yaml
name: erp_get_purchase_order
description: |
  获取采购订单详情。
  
inputSchema:
  type: object
  properties:
    order_id:
      type: string
  required: [order_id]
```

#### erp_list_purchase_orders
```yaml
name: erp_list_purchase_orders
description: |
  查询采购订单列表。
  
inputSchema:
  type: object
  properties:
    supplier_id:
      type: string
    status:
      type: string
      enum: [draft, pending, approved, ordered, partial_received, received, cancelled]
    created_after:
      type: string
      format: date
    created_before:
      type: string
      format: date
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 20
```

#### erp_receive_purchase_order
```yaml
name: erp_receive_purchase_order
description: |
  采购订单收货。支持部分收货。
  
inputSchema:
  type: object
  properties:
    order_id:
      type: string
    items:
      type: array
      items:
        type: object
        properties:
          order_item_id:
            type: string
          received_quantity:
            type: number
          warehouse_id:
            type: string
          location:
            type: string
            description: 库位
          quality_status:
            type: string
            enum: [passed, pending_inspection, rejected]
            default: pending_inspection
        required: [order_item_id, received_quantity, warehouse_id]
    receipt_date:
      type: string
      format: date
    notes:
      type: string
  required: [order_id, items]
```

### 4.3 库存管理工具

#### erp_search_inventory
```yaml
name: erp_search_inventory
description: |
  查询库存。可按物料、仓库、库位等条件筛选。
  
inputSchema:
  type: object
  properties:
    material_id:
      type: string
    warehouse_id:
      type: string
    location:
      type: string
      description: 库位编码
    category:
      type: string
      description: 物料类别
    low_stock_only:
      type: boolean
      description: 只显示低于安全库存的
      default: false
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 50
```

#### erp_get_inventory_detail
```yaml
name: erp_get_inventory_detail
description: |
  获取物料库存详情，包含各仓库分布、批次信息、出入库历史。
  
inputSchema:
  type: object
  properties:
    material_id:
      type: string
    include_history:
      type: boolean
      default: false
      description: 是否包含最近出入库记录
    history_days:
      type: integer
      default: 30
      description: 历史记录天数
  required: [material_id]
```

#### erp_check_inventory_alerts
```yaml
name: erp_check_inventory_alerts
description: |
  检查库存预警。返回低于安全库存或即将过期的物料清单。
  
inputSchema:
  type: object
  properties:
    warehouse_id:
      type: string
      description: 仓库ID（不填检查所有仓库）
    alert_type:
      type: string
      enum: [low_stock, overstock, expiring, expired]
      description: 预警类型
    category:
      type: string
      description: 物料类别筛选

returns:
  type: object
  properties:
    alerts:
      type: array
      items:
        type: object
        properties:
          material_id:
            type: string
          material_name:
            type: string
          alert_type:
            type: string
          current_quantity:
            type: number
          threshold:
            type: number
          warehouse:
            type: string
          suggestion:
            type: string
```

#### erp_transfer_inventory
```yaml
name: erp_transfer_inventory
description: |
  库存调拨。在仓库间或库位间移动物料。
  
inputSchema:
  type: object
  properties:
    material_id:
      type: string
    quantity:
      type: number
    from_warehouse:
      type: string
    from_location:
      type: string
    to_warehouse:
      type: string
    to_location:
      type: string
    reason:
      type: string
  required: [material_id, quantity, from_warehouse, to_warehouse]
```

#### erp_adjust_inventory
```yaml
name: erp_adjust_inventory
description: |
  库存调整。用于盘点差异调整、报损等场景。
  
inputSchema:
  type: object
  properties:
    material_id:
      type: string
    warehouse_id:
      type: string
    location:
      type: string
    adjust_quantity:
      type: number
      description: 调整数量（正数增加，负数减少）
    adjust_type:
      type: string
      enum: [inventory_count, damage, return, other]
    reason:
      type: string
      description: 调整原因
  required: [material_id, warehouse_id, adjust_quantity, adjust_type, reason]
```

### 4.4 生产管理工具

#### erp_list_production_orders
```yaml
name: erp_list_production_orders
description: |
  查询生产工单列表。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
    status:
      type: string
      enum: [planned, released, in_progress, completed, cancelled]
    planned_start_after:
      type: string
      format: date
    planned_start_before:
      type: string
      format: date
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 20
```

#### erp_create_production_order
```yaml
name: erp_create_production_order
description: |
  创建生产工单。会自动根据BOM计算物料需求。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
      description: 产品ID
    quantity:
      type: integer
      description: 生产数量
    bom_version:
      type: string
      description: BOM版本（不填用最新版）
    planned_start:
      type: string
      format: date
    planned_end:
      type: string
      format: date
    priority:
      type: string
      enum: [low, normal, high, urgent]
      default: normal
    notes:
      type: string
  required: [product_id, quantity]

returns:
  type: object
  properties:
    order_id:
      type: string
    material_requirements:
      type: array
      description: 物料需求清单
      items:
        type: object
        properties:
          material_id:
            type: string
          material_name:
            type: string
          required_quantity:
            type: number
          available_quantity:
            type: number
          shortage:
            type: number
```

#### erp_get_production_order
```yaml
name: erp_get_production_order
description: |
  获取生产工单详情，包含物料需求、工序进度等。
  
inputSchema:
  type: object
  properties:
    order_id:
      type: string
  required: [order_id]
```

#### erp_update_production_progress
```yaml
name: erp_update_production_progress
description: |
  更新生产进度。报告工序完成情况。
  
inputSchema:
  type: object
  properties:
    order_id:
      type: string
    process_id:
      type: string
      description: 工序ID
    completed_quantity:
      type: integer
    defect_quantity:
      type: integer
      default: 0
    notes:
      type: string
  required: [order_id, process_id, completed_quantity]
```

#### erp_complete_production_order
```yaml
name: erp_complete_production_order
description: |
  完成生产工单。产品入库。
  
inputSchema:
  type: object
  properties:
    order_id:
      type: string
    actual_quantity:
      type: integer
      description: 实际完成数量
    warehouse_id:
      type: string
      description: 入库仓库
    quality_status:
      type: string
      enum: [passed, pending_inspection]
      default: pending_inspection
    notes:
      type: string
  required: [order_id, actual_quantity, warehouse_id]
```

### 4.5 销售管理工具

#### erp_list_sales_orders
```yaml
name: erp_list_sales_orders
description: |
  查询销售订单列表。
  
inputSchema:
  type: object
  properties:
    customer_id:
      type: string
    status:
      type: string
      enum: [draft, confirmed, processing, shipped, delivered, cancelled]
    created_after:
      type: string
      format: date
    created_before:
      type: string
      format: date
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 20
```

#### erp_create_sales_order
```yaml
name: erp_create_sales_order
description: |
  创建销售订单。
  
inputSchema:
  type: object
  properties:
    customer_id:
      type: string
    items:
      type: array
      items:
        type: object
        properties:
          product_id:
            type: string
          quantity:
            type: integer
          unit_price:
            type: number
          discount:
            type: number
            default: 0
        required: [product_id, quantity, unit_price]
    shipping_address:
      type: string
    expected_delivery:
      type: string
      format: date
    payment_method:
      type: string
      enum: [prepaid, cod, credit]
    notes:
      type: string
  required: [customer_id, items]
```

#### erp_ship_sales_order
```yaml
name: erp_ship_sales_order
description: |
  销售订单发货。
  
inputSchema:
  type: object
  properties:
    order_id:
      type: string
    items:
      type: array
      items:
        type: object
        properties:
          order_item_id:
            type: string
          shipped_quantity:
            type: integer
          warehouse_id:
            type: string
        required: [order_item_id, shipped_quantity, warehouse_id]
    carrier:
      type: string
      description: 承运商
    tracking_number:
      type: string
      description: 物流单号
  required: [order_id, items]
```

### 4.6 财务管理工具

#### erp_list_payables
```yaml
name: erp_list_payables
description: |
  查询应付账款列表。
  
inputSchema:
  type: object
  properties:
    supplier_id:
      type: string
    status:
      type: string
      enum: [pending, partial_paid, paid, overdue]
    due_before:
      type: string
      format: date
      description: 到期日在此之前
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 20
```

#### erp_list_receivables
```yaml
name: erp_list_receivables
description: |
  查询应收账款列表。
  
inputSchema:
  type: object
  properties:
    customer_id:
      type: string
    status:
      type: string
      enum: [pending, partial_received, received, overdue]
    due_before:
      type: string
      format: date
    page:
      type: integer
      default: 1
    page_size:
      type: integer
      default: 20
```

#### erp_create_payment
```yaml
name: erp_create_payment
description: |
  创建付款记录。
  
inputSchema:
  type: object
  properties:
    payable_id:
      type: string
      description: 应付单ID
    amount:
      type: number
      description: 付款金额
    payment_method:
      type: string
      enum: [bank_transfer, check, cash]
    payment_date:
      type: string
      format: date
    bank_account:
      type: string
    notes:
      type: string
  required: [payable_id, amount, payment_method]
```

---

## 五、nimo-report-mcp 工具定义

### 5.1 报表工具

#### report_inventory_summary
```yaml
name: report_inventory_summary
description: |
  生成库存汇总报表。
  
inputSchema:
  type: object
  properties:
    warehouse_id:
      type: string
      description: 仓库筛选（不填汇总所有仓库）
    category:
      type: string
      description: 物料类别
    as_of_date:
      type: string
      format: date
      description: 截止日期
    group_by:
      type: string
      enum: [category, warehouse, supplier]
      default: category
    format:
      type: string
      enum: [json, csv, excel]
      default: json

returns:
  type: object
  properties:
    data:
      type: array
    summary:
      type: object
      properties:
        total_items:
          type: integer
        total_value:
          type: number
        low_stock_count:
          type: integer
    file_url:
      type: string
      description: 文件下载链接（当format为csv/excel时）
```

#### report_purchase_analysis
```yaml
name: report_purchase_analysis
description: |
  采购分析报表。分析采购金额、供应商分布、价格趋势等。
  
inputSchema:
  type: object
  properties:
    start_date:
      type: string
      format: date
    end_date:
      type: string
      format: date
    group_by:
      type: string
      enum: [supplier, category, month]
      default: supplier
    format:
      type: string
      enum: [json, csv, excel]
      default: json
  required: [start_date, end_date]
```

#### report_production_efficiency
```yaml
name: report_production_efficiency
description: |
  生产效率报表。分析产能利用率、良品率、工时等。
  
inputSchema:
  type: object
  properties:
    start_date:
      type: string
      format: date
    end_date:
      type: string
      format: date
    product_id:
      type: string
      description: 产品筛选
    format:
      type: string
      enum: [json, csv, excel]
      default: json
  required: [start_date, end_date]
```

#### report_bom_cost
```yaml
name: report_bom_cost
description: |
  BOM成本分析报表。计算产品成本构成。
  
inputSchema:
  type: object
  properties:
    product_id:
      type: string
    bom_version:
      type: string
    include_labor:
      type: boolean
      default: true
    include_overhead:
      type: boolean
      default: true
    format:
      type: string
      enum: [json, csv, excel]
      default: json
  required: [product_id]
```

#### report_project_status
```yaml
name: report_project_status
description: |
  项目状态报表。汇总所有项目进度、风险、里程碑。
  
inputSchema:
  type: object
  properties:
    status:
      type: string
      enum: [active, completed, all]
      default: active
    format:
      type: string
      enum: [json, csv, excel]
      default: json
```

#### report_custom_query
```yaml
name: report_custom_query
description: |
  自定义查询报表。使用自然语言描述需要的数据，系统自动生成。
  注意：仅支持只读查询，不会修改数据。
  
inputSchema:
  type: object
  properties:
    description:
      type: string
      description: 用自然语言描述需要的报表，如"过去30天各供应商的采购金额排名"
    format:
      type: string
      enum: [json, csv, excel]
      default: json
  required: [description]
```

---

## 六、数据类型定义

### 6.1 产品相关

```yaml
definitions:
  ProductSummary:
    type: object
    properties:
      id:
        type: string
      code:
        type: string
      name:
        type: string
      category:
        type: string
      status:
        type: string
      current_bom_version:
        type: string
      created_at:
        type: string
        format: date-time

  ProductDetail:
    type: object
    properties:
      id:
        type: string
      code:
        type: string
      name:
        type: string
      category:
        type: string
      status:
        type: string
      description:
        type: string
      specs:
        type: object
      current_bom:
        $ref: "#/definitions/BOMSummary"
      projects:
        type: array
        items:
          $ref: "#/definitions/ProjectSummary"
      documents:
        type: array
        items:
          $ref: "#/definitions/DocumentRef"
      created_by:
        type: string
      created_at:
        type: string
        format: date-time
      updated_at:
        type: string
        format: date-time

  BOMItem:
    type: object
    properties:
      id:
        type: string
      level:
        type: integer
        description: 层级（0为顶级）
      parent_id:
        type: string
      material:
        $ref: "#/definitions/MaterialRef"
      quantity:
        type: number
      unit:
        type: string
      position:
        type: string
      unit_cost:
        type: number
      total_cost:
        type: number
      inventory_available:
        type: number
      children:
        type: array
        items:
          $ref: "#/definitions/BOMItem"
```

### 6.2 物料相关

```yaml
definitions:
  MaterialRef:
    type: object
    properties:
      id:
        type: string
      code:
        type: string
      name:
        type: string
      category:
        type: string
      unit:
        type: string
      
  InventoryItem:
    type: object
    properties:
      material:
        $ref: "#/definitions/MaterialRef"
      warehouse_id:
        type: string
      warehouse_name:
        type: string
      location:
        type: string
      quantity:
        type: number
      reserved_quantity:
        type: number
      available_quantity:
        type: number
      safety_stock:
        type: number
      unit_cost:
        type: number
      last_movement_at:
        type: string
        format: date-time
```

---

## 七、安全与权限

### 7.1 认证机制

```
┌─────────────────────────────────────────────────────────────────┐
│                        认证流程                                  │
└─────────────────────────────────────────────────────────────────┘

  Agent           MCP Server         Auth Service          业务服务
    │                │                    │                    │
    │ 1. 调用工具     │                    │                    │
    │  (带context)   │                    │                    │
    │ ──────────────>│                    │                    │
    │                │ 2. 验证Agent身份   │                    │
    │                │ ──────────────────>│                    │
    │                │ <──────────────────│                    │
    │                │ 3. 获取用户上下文   │                    │
    │                │ ──────────────────>│                    │
    │                │ <──────────────────│                    │
    │                │ (user_id, roles,   │                    │
    │                │  permissions)      │                    │
    │                │                    │                    │
    │                │ 4. 调用业务API      │                    │
    │                │ ──────────────────────────────────────>│
    │                │ (带用户上下文)      │                    │
    │                │ <──────────────────────────────────────│
    │ <──────────────│                    │                    │
    │  返回结果       │                    │                    │
```

### 7.2 权限控制

**工具级权限：**
```yaml
tool_permissions:
  # PLM工具权限
  plm_search_products: [plm_viewer, plm_editor, plm_admin]
  plm_create_product: [plm_editor, plm_admin]
  plm_release_bom: [plm_admin]
  plm_create_ecn: [plm_editor, plm_admin]
  
  # ERP工具权限
  erp_search_inventory: [erp_viewer, erp_editor, erp_admin]
  erp_create_purchase_order: [erp_purchaser, erp_admin]
  erp_adjust_inventory: [erp_warehouse, erp_admin]
  erp_create_payment: [erp_finance, erp_admin]
```

**数据级权限：**
```yaml
data_permissions:
  # 供应商数据：只能看到自己负责的
  supplier:
    filter: "owner_id = :current_user_id OR :has_admin_role"
  
  # 财务数据：需要财务角色
  finance:
    require_roles: [erp_finance, erp_admin]
```

### 7.3 审计日志

每次工具调用都会记录：

```json
{
  "timestamp": "2026-02-05T21:26:00+08:00",
  "agent_id": "openclaw-main",
  "user_id": "u_123456",
  "user_name": "陈泽斌",
  "tool": "erp_create_purchase_order",
  "input": {
    "supplier_id": "sup_001",
    "items": [...]
  },
  "output": {
    "order_id": "PO-2026-001"
  },
  "duration_ms": 156,
  "status": "success"
}
```

---

## 八、实现指南

### 8.1 项目结构

```
nimo-mcp/
├── packages/
│   ├── plm-server/              # PLM MCP Server
│   │   ├── src/
│   │   │   ├── index.ts         # 入口
│   │   │   ├── server.ts        # Server实现
│   │   │   ├── tools/           # 工具实现
│   │   │   │   ├── product.ts
│   │   │   │   ├── bom.ts
│   │   │   │   ├── project.ts
│   │   │   │   └── ecn.ts
│   │   │   ├── client/          # API客户端
│   │   │   │   └── plm-api.ts
│   │   │   └── types/           # 类型定义
│   │   ├── package.json
│   │   └── tsconfig.json
│   │
│   ├── erp-server/              # ERP MCP Server
│   │   └── (同上结构)
│   │
│   ├── report-server/           # Report MCP Server
│   │   └── (同上结构)
│   │
│   └── shared/                  # 共享代码
│       ├── auth/                # 认证
│       ├── client/              # HTTP客户端基类
│       ├── types/               # 共享类型
│       └── utils/               # 工具函数
│
├── package.json                 # Monorepo配置
├── pnpm-workspace.yaml
└── turbo.json
```

### 8.2 核心代码示例

**Server入口：**
```typescript
// packages/plm-server/src/index.ts
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { PLMServer } from "./server.js";

async function main() {
  const server = new PLMServer();
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("nimo-plm-mcp server running on stdio");
}

main().catch(console.error);
```

**Server实现：**
```typescript
// packages/plm-server/src/server.ts
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { productTools, handleProductTool } from "./tools/product.js";
import { bomTools, handleBOMTool } from "./tools/bom.js";
import { projectTools, handleProjectTool } from "./tools/project.js";
import { PLMApiClient } from "./client/plm-api.js";

export class PLMServer {
  private server: Server;
  private apiClient: PLMApiClient;

  constructor() {
    this.server = new Server(
      {
        name: "nimo-plm-mcp",
        version: "1.0.0",
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.apiClient = new PLMApiClient({
      baseUrl: process.env.PLM_API_URL || "http://localhost:8002",
      apiKey: process.env.PLM_API_KEY,
    });

    this.setupHandlers();
  }

  private setupHandlers() {
    // 列出所有工具
    this.server.setRequestHandler(ListToolsRequestSchema, async () => ({
      tools: [
        ...productTools,
        ...bomTools,
        ...projectTools,
      ],
    }));

    // 处理工具调用
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;

      try {
        // 产品工具
        if (name.startsWith("plm_") && name.includes("product")) {
          return await handleProductTool(name, args, this.apiClient);
        }
        // BOM工具
        if (name.startsWith("plm_") && name.includes("bom")) {
          return await handleBOMTool(name, args, this.apiClient);
        }
        // 项目工具
        if (name.startsWith("plm_") && name.includes("project") || name.includes("task")) {
          return await handleProjectTool(name, args, this.apiClient);
        }

        throw new Error(`Unknown tool: ${name}`);
      } catch (error) {
        return {
          content: [
            {
              type: "text",
              text: `Error: ${error.message}`,
            },
          ],
          isError: true,
        };
      }
    });
  }

  async connect(transport: StdioServerTransport) {
    await this.server.connect(transport);
  }
}
```

**工具实现示例：**
```typescript
// packages/plm-server/src/tools/product.ts
import { Tool } from "@modelcontextprotocol/sdk/types.js";
import { PLMApiClient } from "../client/plm-api.js";

export const productTools: Tool[] = [
  {
    name: "plm_search_products",
    description: "搜索产品列表。支持按名称、编码、类别、状态等条件筛选。",
    inputSchema: {
      type: "object",
      properties: {
        query: {
          type: "string",
          description: "搜索关键词，匹配产品名称或编码",
        },
        category: {
          type: "string",
          enum: ["frame", "temple", "lens", "platform", "accessory"],
          description: "产品类别",
        },
        status: {
          type: "string",
          enum: ["draft", "developing", "active", "discontinued"],
          description: "产品状态",
        },
        page: {
          type: "integer",
          default: 1,
        },
        page_size: {
          type: "integer",
          default: 20,
        },
      },
    },
  },
  {
    name: "plm_get_product",
    description: "获取产品详细信息，包含基本信息、当前BOM版本、关联项目等。",
    inputSchema: {
      type: "object",
      properties: {
        product_id: {
          type: "string",
          description: "产品ID",
        },
      },
      required: ["product_id"],
    },
  },
  {
    name: "plm_create_product",
    description: "创建新产品。创建后状态为草稿，需要完善BOM后才能发布。",
    inputSchema: {
      type: "object",
      properties: {
        name: {
          type: "string",
          description: "产品名称",
        },
        category: {
          type: "string",
          enum: ["frame", "temple", "lens", "platform", "accessory"],
          description: "产品类别",
        },
        description: {
          type: "string",
          description: "产品描述",
        },
        specs: {
          type: "object",
          description: "规格参数",
        },
        base_product_id: {
          type: "string",
          description: "基于哪个产品创建（可选，会复制BOM）",
        },
      },
      required: ["name", "category"],
    },
  },
];

export async function handleProductTool(
  name: string,
  args: Record<string, unknown>,
  client: PLMApiClient
) {
  switch (name) {
    case "plm_search_products": {
      const result = await client.searchProducts(args);
      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(result, null, 2),
          },
        ],
      };
    }

    case "plm_get_product": {
      const result = await client.getProduct(args.product_id as string);
      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(result, null, 2),
          },
        ],
      };
    }

    case "plm_create_product": {
      const result = await client.createProduct(args);
      return {
        content: [
          {
            type: "text",
            text: `产品创建成功！\n\n${JSON.stringify(result, null, 2)}`,
          },
        ],
      };
    }

    default:
      throw new Error(`Unknown product tool: ${name}`);
  }
}
```

### 8.3 OpenClaw集成配置

```yaml
# openclaw.yaml 或 mcp配置
mcp:
  servers:
    - name: nimo-plm
      command: node
      args: 
        - /path/to/nimo-mcp/packages/plm-server/dist/index.js
      env:
        PLM_API_URL: https://plm-api.nimo.internal
        PLM_API_KEY: "${NIMO_PLM_API_KEY}"
      
    - name: nimo-erp
      command: node
      args:
        - /path/to/nimo-mcp/packages/erp-server/dist/index.js
      env:
        ERP_API_URL: https://erp-api.nimo.internal
        ERP_API_KEY: "${NIMO_ERP_API_KEY}"
        
    - name: nimo-report
      command: node
      args:
        - /path/to/nimo-mcp/packages/report-server/dist/index.js
      env:
        REPORT_API_URL: https://report-api.nimo.internal
```

---

## 九、使用场景示例

### 9.1 智能补货

```
用户: 检查库存预警，低于安全库存的物料自动生成采购单

Agent执行流程:
1. erp_check_inventory_alerts(alert_type="low_stock")
   → 返回5个低库存物料
   
2. 对每个物料，查询推荐供应商:
   erp_get_supplier_materials(supplier_id=preferred_supplier)
   → 获取最新报价
   
3. 按供应商分组，创建采购单:
   erp_create_purchase_order(supplier_id, items=[...])
   → 创建3张采购单
   
4. 返回汇总:
   "已检测到5个物料低于安全库存，已自动创建3张采购单：
    - PO-2026-001: 供应商A, 2个物料, 金额¥12,500
    - PO-2026-002: 供应商B, 2个物料, 金额¥8,300
    - PO-2026-003: 供应商C, 1个物料, 金额¥3,200"
```

### 9.2 新产品立项

```
用户: 创建NIMO Air 2代产品，基于Air 1代

Agent执行流程:
1. plm_search_products(query="NIMO Air 1")
   → 找到产品ID: PRD-001
   
2. plm_create_product(
     name="NIMO Air 2",
     category="platform",
     base_product_id="PRD-001"
   )
   → 创建产品并复制BOM，返回PRD-002
   
3. plm_create_project(
     name="NIMO Air 2 研发项目",
     product_id="PRD-002",
     template="standard"
   )
   → 创建项目，自动生成EVT/DVT/PVT/MP任务

4. 返回:
   "NIMO Air 2 产品已创建（PRD-002），BOM已从Air 1代复制。
    研发项目已建立，包含127个任务，预计周期180天。
    项目链接: https://plm.nimo.com/projects/PROJ-002"
```

### 9.3 生产计划

```
用户: 根据下月销售预测，安排NIMO Air Pro的生产计划

Agent执行流程:
1. 查询销售预测（假设通过报表工具）
   report_custom_query("下月NIMO Air Pro销售预测")
   → 预计销售1000台
   
2. 检查当前库存:
   erp_search_inventory(material_id="PRD-AIRPRO")
   → 当前库存200台
   
3. 计算需求并创建生产工单:
   erp_create_production_order(
     product_id="PRD-AIRPRO",
     quantity=800,
     planned_start="2026-02-10",
     planned_end="2026-02-28"
   )
   → 返回物料需求清单
   
4. 检查物料缺口，自动创建采购:
   erp_check_inventory_alerts() → 发现3个物料短缺
   erp_create_purchase_order(...) → 创建补货订单

5. 返回:
   "生产计划已安排：
    - 生产工单: MO-2026-015, 800台NIMO Air Pro
    - 计划周期: 2/10 - 2/28
    - 物料状态: 3个物料需要补货，已创建采购单PO-2026-005
    - 预计产能满足度: 100%"
```

---

## 十、开发里程碑

### Phase 1: MVP (4周)

| 周次 | 任务 |
|-----|------|
| W1 | MCP Server框架搭建，认证集成 |
| W2 | PLM核心工具（产品、BOM） |
| W3 | ERP核心工具（采购、库存） |
| W4 | OpenClaw集成测试，文档 |

**交付物：**
- nimo-plm-mcp: 10个核心工具
- nimo-erp-mcp: 12个核心工具
- OpenClaw集成配置

### Phase 2: 完整功能 (4周)

| 周次 | 任务 |
|-----|------|
| W5 | PLM完整工具（项目、ECN、文档） |
| W6 | ERP完整工具（生产、销售、财务） |
| W7 | 报表MCP Server |
| W8 | 端到端场景测试，性能优化 |

**交付物：**
- 全部49个工具
- nimo-report-mcp
- 性能基准报告

### Phase 3: 增强 (2周)

| 周次 | 任务 |
|-----|------|
| W9 | Resources支持（文档、报表访问） |
| W10 | Prompts模板、复杂场景优化 |

---

*文档结束*
