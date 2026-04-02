# PLM Module v2.0 — 与 PLM 项目数据模型看齐

> 版本：v1.0 | 日期：2026-03-23
> 模块：`acp-module-plm`（Go package: `github.com/bitfantasy/acp-module-plm`）
> 目标：将 ACP PLM Module 的数据模型和 command 能力从 v1.0 基础版升级到与 PLM 项目（`internal/plm/entity/`）核心功能对齐

---

## 一、背景与目标

### 1.1 现状

ACP PLM Module v1.0 覆盖了基础的项目/任务/产品/物料/BOM/文档/ECN 管理，共 16 张表、30 个 commands。但与 PLM 项目（端口 8080，已生产运行）相比，在以下核心领域存在显著差距：

| 领域 | PLM 项目 | ACP PLM Module v1.0 | 差距 |
|------|----------|---------------------|------|
| BOM 体系 | 三级 BOM（EBOM/PBOM/MBOM）+ 品类属性模板 + ERP 快照 + 草稿 | 单层 BOM，固定列 | **大** |
| SKU/CMF | 7 张表覆盖配色方案 → CMF 设计 → BOM 变体 | 完全缺失 | **大** |
| 制造工艺 | 工艺路线 + 工艺步骤 + 工序物料 | 完全缺失 | **大** |
| ECN 流程 | 审批链 + 执行任务 + 操作历史 + BOM ECN | 仅基本信息 + 受影响项 | **中** |
| 零件图纸 | 版本化图纸管理 | 完全缺失 | **中** |
| 阶段交付物 | 交付物定义 + 状态追踪 | 完全缺失 | **中** |

### 1.2 目标

补齐以上 6 个领域，新增约 **15 张表**、**40+ 个 commands**，使 ACP PLM Module 具备与 PLM 项目等同的数据承载能力，支撑硬件开发流程（plm-hardware-dev）的完整数据链路。

### 1.3 设计原则

| 原则 | 说明 |
|------|------|
| **数据模型对齐** | 表结构尽量与 PLM 项目一致，便于未来数据迁移 |
| **Command 风格统一** | 遵循现有 `plm` module 的 command 命名和 I/O 风格 |
| **向下兼容** | 不修改已有 16 张表结构，仅新增表和 commands |
| **JSONB 扩展** | BOM Item 扩展属性用 JSONB，品类属性模板驱动动态列 |
| **Module SDK 标准** | 新增 Entity/View/Nav/Route 遵循 `sdk.Module` 接口 |

---

## 二、Feature 1 — BOM 体系升级

### 2.1 概述

将现有单层 BOM 升级为三级 BOM 体系（EBOM → PBOM → MBOM），支持品类驱动的动态属性、BOM 草稿暂存、ERP 发布快照。

### 2.2 现有表改造

#### `plm_boms` 表 — 增加字段

现有表已有 `BOMType` 字段，但缺少上游 BOM 引用和审批流程字段。

| 新增字段 | 类型 | 说明 |
|---------|------|------|
| `source_bom_id` | string(100) | 上游 BOM ID（PBOM→EBOM, MBOM→EBOM） |
| `source_version` | string(20) | 上游 BOM 版本号 |
| `submitted_at` | *time.Time | 提交时间 |
| `reviewed_at` | *time.Time | 审核时间 |
| `approved_at` | *time.Time | 批准时间 |
| `frozen_by` | string(100) | 冻结人 |
| `created_by` | string(100) | 创建人 |

#### `plm_bom_items` 表 — 增加字段

| 新增字段 | 类型 | 说明 |
|---------|------|------|
| `sub_category` | string(32) | 小类（component/pcb/connector/housing 等） |
| `supplier_id` | string(100) | 供应商 ID（关联 SRM） |
| `manufacturer_id` | string(100) | 制造商 ID（电子料） |
| `process_step_id` | string(100) | 工艺步骤 ID（PBOM/MBOM 专用） |
| `scrap_rate` | float64 | 报废率（MBOM） |
| `effective_date` | *time.Time | 生效日期 |
| `expire_date` | *time.Time | 失效日期 |
| `attachments` | jsonb | 附件列表 |
| `thumbnail_url` | string(512) | 缩略图 |

> `extended_attrs` 字段已存在（text 类型），需升级为 `jsonb` 类型。

### 2.3 新增表

#### `plm_category_attr_templates` — 品类属性模板

驱动 BOM Item 的动态扩展属性。每个品类 + 小类定义一组专属字段。

```go
type PlmCategoryAttrTemplate struct {
    ID           string    `gorm:"primaryKey;size:100"`
    Category     string    `gorm:"size:32;not null;index:idx_cat_sub"`    // electronic/structural/optical...
    SubCategory  string    `gorm:"size:32;not null;index:idx_cat_sub"`    // component/pcb/housing...
    BOMType      string    `gorm:"size:16;not null;default:EBOM"`         // EBOM/PBOM — 属性归属哪种 BOM
    FieldKey     string    `gorm:"size:64;not null"`                      // 属性 key（存入 extended_attrs）
    FieldName    string    `gorm:"size:64;not null"`                      // 显示名
    FieldType    string    `gorm:"size:16;not null"`                      // text/number/select/boolean/file/thumbnail
    Unit         string    `gorm:"size:16"`                               // 单位
    Required     bool      `gorm:"default:false"`
    Options      jsonb     `gorm:"type:jsonb"`                            // select 选项 {"values":["A","B"]}
    Validation   jsonb     `gorm:"type:jsonb"`                            // {min, max, pattern}
    DefaultValue string    `gorm:"size:128"`
    SortOrder    int       `gorm:"default:0"`
    ShowInTable  bool      `gorm:"default:true"`                          // 是否在列表中显示
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**用途示例**：电子/元器件（electronic/component）类定义 `package_size`(text)、`voltage_rating`(number)、`datasheet_url`(file)；光学/波导片（optical/waveguide）类定义 `fov_degrees`(number)、`transmittance`(number)。

#### `plm_bom_drafts` — BOM 草稿暂存

```go
type PlmBOMDraft struct {
    ID        string    `gorm:"primaryKey;size:100"`
    BOMID     string    `gorm:"size:100;not null;uniqueIndex"`
    DraftData jsonb     `gorm:"type:jsonb;not null"`           // 临时修改的 BOM 数据（完整 items 快照）
    CreatedBy string    `gorm:"size:100;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### `plm_bom_releases` — BOM 发布快照（ERP 对接）

```go
type PlmBOMRelease struct {
    ID           string     `gorm:"primaryKey;size:100"`
    BOMID        string     `gorm:"size:100;not null"`
    ProjectID    string     `gorm:"size:100;not null"`
    BOMType      string     `gorm:"size:16;not null"`
    Version      string     `gorm:"size:20;not null"`
    SnapshotJSON jsonb      `gorm:"type:jsonb;not null"`     // 发布时刻的 BOM 完整快照
    Status       string     `gorm:"size:16;default:pending"` // pending/synced/failed
    SyncedAt     *time.Time
    CreatedAt    time.Time
}
```

### 2.4 新增 Commands

| Command | Input | Output | 说明 |
|---------|-------|--------|------|
| `convert_bom` | bom_id, target_type(pbom/mbom) | BOMOutput | 从 EBOM 转换生成 PBOM/MBOM |
| `freeze_bom` | bom_id, comment | BOMOutput | 冻结 BOM（不可再编辑） |
| `save_bom_draft` | bom_id, draft_data(json) | DraftOutput | 保存 BOM 草稿 |
| `load_bom_draft` | bom_id | DraftOutput | 加载 BOM 草稿 |
| `delete_bom_draft` | bom_id | DeleteOutput | 删除 BOM 草稿 |
| `release_bom_snapshot` | bom_id, comment | ReleaseOutput | 生成 ERP 发布快照 |
| `list_bom_releases` | bom_id | ListReleasesOutput | 查询发布历史 |
| `list_category_templates` | category, sub_category, bom_type | TemplateListOutput | 查询品类属性模板 |
| `create_category_template` | category, sub_category, field_key, field_name, field_type, ... | TemplateOutput | 创建属性模板 |
| `update_category_template` | template_id, ... | TemplateOutput | 更新属性模板 |
| `delete_category_template` | template_id | DeleteOutput | 删除属性模板 |

### 2.5 Entity UI

新增 `categoryAttrTemplateEntity` 注册到 `Entities()` 返回值，支持在前端管理品类属性模板。

新增 `bomDraftEntity` 和 `bomReleaseEntity` 作为 BOM 的关联子 entity。

---

## 三、Feature 2 — SKU / CMF 系统

### 3.1 概述

智能眼镜产品有多种配色方案（SKU），每种 SKU 影响 BOM 中外观件的材料/颜色/表面处理选择。需要建立 SKU → BOM Item → CMF 变体的完整链路。

### 3.2 新增表

#### `plm_product_skus` — 产品配色/型号方案

```go
type PlmProductSKU struct {
    ID          string    `gorm:"primaryKey;size:100"`
    ProjectID   string    `gorm:"size:100;not null;index"`
    Name        string    `gorm:"size:128;not null"`              // "曜石黑"、"星空灰"
    Code        string    `gorm:"size:32"`                        // SKU 编码
    Description string    `gorm:"type:text"`
    Status      string    `gorm:"size:16;default:active"`         // active/discontinued
    SortOrder   int       `gorm:"default:0"`
    CreatedBy   string    `gorm:"size:100;not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### `plm_sku_cmf_configs` — SKU × BOM外观件 → CMF 配置

```go
type PlmSKUCMFConfig struct {
    ID               string    `gorm:"primaryKey;size:100"`
    SKUID            string    `gorm:"size:100;not null;index"`
    BOMItemID        string    `gorm:"size:100;not null"`          // 对应 BOM 中的外观件
    Color            string    `gorm:"size:64"`                    // 颜色名
    ColorCode        string    `gorm:"size:32"`                    // 色号
    SurfaceTreatment string    `gorm:"size:128"`                   // 表面处理工艺
    ProcessParams    jsonb     `gorm:"type:jsonb"`                 // 工艺参数 JSON
    Notes            string    `gorm:"type:text"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

#### `plm_sku_bom_items` — SKU 与 BOM 零件关联

```go
type PlmSKUBOMItem struct {
    ID           string    `gorm:"primaryKey;size:100"`
    SKUID        string    `gorm:"size:100;not null;index"`
    BOMItemID    string    `gorm:"size:100;not null;index"`
    CMFVariantID string    `gorm:"size:100"`                  // 选用哪个 CMF 变体
    Quantity     float64   `gorm:"type:numeric(15,4);default:0"` // 0 = 使用 BOM 默认数量
    Notes        string    `gorm:"type:text"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### `plm_cmf_designs` — 外观零件 CMF 方案

```go
type PlmCMFDesign struct {
    ID                  string    `gorm:"primaryKey;size:100"`
    ProjectID           string    `gorm:"size:100;not null;index"`
    TaskID              string    `gorm:"size:100;index"`           // 关联开发任务
    BOMItemID           string    `gorm:"size:100;not null;index"`  // 对应 BOM 外观件
    SchemeName          string    `gorm:"size:64"`                  // 方案名
    Color               string    `gorm:"size:64"`
    ColorCode           string    `gorm:"size:64"`
    GlossLevel          string    `gorm:"size:32"`                  // 高光/半哑/哑光/丝光/镜面
    SurfaceTreatment    string    `gorm:"size:128"`
    TexturePattern      string    `gorm:"size:64"`                  // 皮纹/磨砂/拉丝
    CoatingType         string    `gorm:"size:64"`                  // UV漆/PU漆/粉末涂装
    RenderImageFileID   string    `gorm:"size:100"`                 // 渲染图文件 ID
    RenderImageFileName string    `gorm:"size:256"`
    Notes               string    `gorm:"type:text"`
    SortOrder           int       `gorm:"default:0"`
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
```

#### `plm_cmf_drawings` — CMF 工艺图纸

```go
type PlmCMFDrawing struct {
    ID          string    `gorm:"primaryKey;size:100"`
    CMFDesignID string    `gorm:"size:100;not null;index"`
    DrawingType string    `gorm:"size:32"`                    // 丝印/激光雷雕/UV转印/移印/烫金
    FileID      string    `gorm:"size:100;not null"`
    FileName    string    `gorm:"size:256;not null"`
    Notes       string    `gorm:"type:text"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### `plm_bom_item_cmf_variants` — BOM Item CMF 变体

一个外观件可有多种 CMF 方案（颜色/材质/表面处理），供 SKU 选用。

```go
type PlmBOMItemCMFVariant struct {
    ID                   string    `gorm:"primaryKey;size:100"`
    BOMItemID            string    `gorm:"size:100;not null;index"`
    VariantIndex         int       `gorm:"not null;default:1"`
    MaterialCode         string    `gorm:"size:50"`
    ColorName            string    `gorm:"size:100"`
    ColorHex             string    `gorm:"size:7"`                  // #RRGGBB
    Material             string    `gorm:"size:200"`                // 材质描述
    Finish               string    `gorm:"size:200"`                // 表面处理
    Texture              string    `gorm:"size:200"`                // 纹理
    Coating              string    `gorm:"size:200"`                // 涂层
    PantoneCode          string    `gorm:"size:50"`
    GlossLevel           string    `gorm:"size:32"`
    ReferenceImageFileID string    `gorm:"size:100"`
    ReferenceImageURL    string    `gorm:"size:500"`
    ProcessDrawingType   string    `gorm:"size:50"`
    ProcessDrawings      jsonb     `gorm:"type:jsonb;default:'[]'"` // 工艺图纸列表
    Notes                string    `gorm:"type:text"`
    Status               string    `gorm:"size:20;default:draft"`   // draft/approved/obsolete
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
```

#### `plm_bom_item_lang_variants` — 包装件语言变体

包装类 BOM Item 需要按市场/语言出不同版本的印刷设计文件。

```go
type PlmBOMItemLangVariant struct {
    ID             string    `gorm:"primaryKey;size:100"`
    BOMItemID      string    `gorm:"size:100;not null;index"`
    VariantIndex   int       `gorm:"not null;default:1"`
    MaterialCode   string    `gorm:"size:50"`
    LanguageCode   string    `gorm:"size:10;not null"`        // zh-CN, en-US, ja-JP
    LanguageName   string    `gorm:"size:50;not null"`        // 简体中文, English
    DesignFileID   string    `gorm:"size:100"`
    DesignFileName string    `gorm:"size:200"`
    DesignFileURL  string    `gorm:"size:500"`
    Notes          string    `gorm:"type:text"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### 3.3 新增 Commands

| Command | Input | Output | 说明 |
|---------|-------|--------|------|
| `create_sku` | project_id, name, code, description | SKUOutput | 创建 SKU |
| `update_sku` | sku_id, name, description, status | SKUOutput | 更新 SKU |
| `delete_sku` | sku_id | DeleteOutput | 删除 SKU |
| `list_skus` | project_id | SKUListOutput | 列出项目所有 SKU |
| `set_sku_cmf` | sku_id, bom_item_id, color, surface_treatment, ... | SKUCMFOutput | 设置 SKU 的 CMF 配置 |
| `set_sku_bom_item` | sku_id, bom_item_id, cmf_variant_id, quantity | SKUBOMItemOutput | 设置 SKU 的 BOM 零件关联 |
| `create_cmf_design` | project_id, bom_item_id, scheme_name, color, ... | CMFDesignOutput | 创建 CMF 设计方案 |
| `update_cmf_design` | design_id, ... | CMFDesignOutput | 更新 CMF 设计方案 |
| `delete_cmf_design` | design_id | DeleteOutput | 删除 CMF 设计方案 |
| `add_cmf_drawing` | cmf_design_id, drawing_type, file_id, file_name | CMFDrawingOutput | 添加 CMF 工艺图纸 |
| `create_cmf_variant` | bom_item_id, color_name, color_hex, material, finish, ... | CMFVariantOutput | 创建 BOM Item CMF 变体 |
| `update_cmf_variant` | variant_id, ... | CMFVariantOutput | 更新 CMF 变体 |
| `delete_cmf_variant` | variant_id | DeleteOutput | 删除 CMF 变体 |
| `create_lang_variant` | bom_item_id, language_code, language_name, design_file_id | LangVariantOutput | 创建语言变体 |
| `update_lang_variant` | variant_id, ... | LangVariantOutput | 更新语言变体 |
| `delete_lang_variant` | variant_id | DeleteOutput | 删除语言变体 |

### 3.4 Entity UI

新增以下 Entity 注册：
- `productSKUEntity` — 导航入口 `/m/plm/skus`
- `cmfDesignEntity` — 作为 BOM Item 的关联子页面
- `cmfVariantEntity` — 嵌入 BOM Item 详情页
- `langVariantEntity` — 嵌入包装类 BOM Item 详情页

新增 View：
- `sku_matrix` — 自定义组件，展示 SKU × 外观件 的 CMF 配置矩阵

---

## 四、Feature 3 — 制造工艺

### 4.1 概述

为 PBOM/MBOM 提供工艺路线定义能力。工艺路线由多个工艺步骤组成，每个步骤可关联物料消耗。

### 4.2 新增表

#### `plm_process_routes` — 工艺路线

```go
type PlmProcessRoute struct {
    ID          string    `gorm:"primaryKey;size:100"`
    ProjectID   string    `gorm:"size:100;not null"`
    BOMID       string    `gorm:"size:100"`               // 关联 BOM
    Name        string    `gorm:"size:128;not null"`
    Version     string    `gorm:"size:16;default:v1.0"`
    Status      string    `gorm:"size:16;default:draft"`   // draft/active/obsolete
    Description string    `gorm:"type:text"`
    TotalSteps  int       `gorm:"default:0"`
    CreatedBy   string    `gorm:"size:100;not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### `plm_process_steps` — 工艺步骤

```go
type PlmProcessStep struct {
    ID             string    `gorm:"primaryKey;size:100"`
    RouteID        string    `gorm:"size:100;not null"`
    StepNumber     int       `gorm:"not null"`
    Name           string    `gorm:"size:128;not null"`
    WorkCenter     string    `gorm:"size:64"`              // 工作中心
    Description    string    `gorm:"type:text"`
    StdTimeMinutes float64   `gorm:"type:numeric(10,2)"`   // 标准工时（分钟）
    SetupMinutes   float64   `gorm:"type:numeric(10,2)"`   // 换线时间
    LaborCost      float64   `gorm:"type:numeric(15,4)"`   // 人工成本
    SortOrder      int       `gorm:"default:0"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

#### `plm_process_step_materials` — 工序物料关联

```go
type PlmProcessStepMaterial struct {
    ID         string    `gorm:"primaryKey;size:100"`
    StepID     string    `gorm:"size:100;not null"`
    MaterialID string    `gorm:"size:100"`
    Name       string    `gorm:"size:128"`
    Category   string    `gorm:"size:32;not null"`       // tooling（工装）/consumable（耗材）/service（外协）
    Quantity   float64   `gorm:"type:numeric(15,4);default:1"`
    Unit       string    `gorm:"size:16;default:pcs"`
    Notes      string    `gorm:"type:text"`
    CreatedAt  time.Time
}
```

### 4.3 新增 Commands

| Command | Input | Output | 说明 |
|---------|-------|--------|------|
| `create_process_route` | project_id, bom_id, name, description | RouteOutput | 创建工艺路线 |
| `update_process_route` | route_id, name, status, description | RouteOutput | 更新工艺路线 |
| `delete_process_route` | route_id | DeleteOutput | 删除工艺路线 |
| `add_process_step` | route_id, name, step_number, work_center, std_time_minutes, ... | StepOutput | 添加工艺步骤 |
| `update_process_step` | step_id, ... | StepOutput | 更新工艺步骤 |
| `delete_process_step` | step_id | DeleteOutput | 删除工艺步骤 |
| `add_step_material` | step_id, material_id, name, category, quantity | StepMaterialOutput | 添加工序物料 |
| `delete_step_material` | material_id | DeleteOutput | 删除工序物料 |

### 4.4 Entity UI

新增 `processRouteEntity` 注册，导航入口 `/m/plm/routes`。

---

## 五、Feature 4 — ECN 完整流程

### 5.1 概述

将 ECN 从"基本信息 + 受影响项"扩展为完整的变更管理流程：审批链 → 执行任务 → 操作历史 → BOM 变更关联。

### 5.2 新增表

#### `plm_ecn_approvals` — ECN 审批记录

```go
type PlmECNApproval struct {
    ID         string     `gorm:"primaryKey;size:100"`
    ECNID      string     `gorm:"size:100;not null"`
    ApproverID string     `gorm:"size:100;not null"`
    Sequence   int        `gorm:"not null"`                  // 审批顺序（串行审批时使用）
    Status     string     `gorm:"size:16;default:pending"`   // pending/approved/rejected
    Decision   string     `gorm:"size:16"`                   // approve/reject
    Comment    string     `gorm:"type:text"`
    DecidedAt  *time.Time
    CreatedAt  time.Time
}
```

#### `plm_ecn_tasks` — ECN 执行任务

ECN 批准后需要落地执行（修改 BOM、更新图纸、通知供应商等），每项工作是一个 ECN Task。

```go
type PlmECNTask struct {
    ID          string     `gorm:"primaryKey;size:100"`
    ECNID       string     `gorm:"size:100;not null;index"`
    Type        string     `gorm:"size:32;not null"`         // bom_update/drawing_update/supplier_notify/process_update
    Title       string     `gorm:"size:256;not null"`
    Description string     `gorm:"type:text"`
    AssigneeID  string     `gorm:"size:100"`
    DueDate     *time.Time
    Status      string     `gorm:"size:16;default:pending"`  // pending/in_progress/completed/cancelled
    CompletedAt *time.Time
    CompletedBy string     `gorm:"size:100"`
    Metadata    jsonb      `gorm:"type:jsonb"`               // 任务附加数据
    SortOrder   int        `gorm:"default:0"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### `plm_ecn_histories` — ECN 操作历史

```go
type PlmECNHistory struct {
    ID        string    `gorm:"primaryKey;size:100"`
    ECNID     string    `gorm:"size:100;not null;index"`
    Action    string    `gorm:"size:32;not null"`         // created/submitted/approved/rejected/task_completed/implemented
    UserID    string    `gorm:"size:100;not null"`
    Detail    jsonb     `gorm:"type:jsonb"`               // 操作详情
    CreatedAt time.Time
}
```

#### `plm_bom_ecns` — BOM 变更通知

直接关联 BOM 和变更记录，存储变更 diff。

```go
type PlmBOMECN struct {
    ID            string     `gorm:"primaryKey;size:100"`
    ECNNumber     string     `gorm:"size:32;not null;uniqueIndex"` // ECN-2026-0001
    BOMID         string     `gorm:"size:100;not null;index"`
    Title         string     `gorm:"size:256;not null"`
    Description   string     `gorm:"type:text"`
    ChangeSummary jsonb      `gorm:"type:jsonb;not null"`          // 变更 diff（before/after）
    Status        string     `gorm:"size:16;default:pending"`      // pending/approved/rejected
    CreatedBy     string     `gorm:"size:100;not null"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    ApprovedBy    string     `gorm:"size:100"`
    ApprovedAt    *time.Time
    RejectedBy    string     `gorm:"size:100"`
    RejectedAt    *time.Time
    RejectionNote string    `gorm:"type:text"`
}
```

### 5.3 现有 `plm_ecns` 表增加字段

| 新增字段 | 类型 | 说明 |
|---------|------|------|
| `approval_mode` | string(16) | serial（串行）/ parallel（并行） |
| `sop_impact` | jsonb | SOP 影响分析 |
| `requested_at` | *time.Time | 提交时间 |
| `rejection_reason` | text | 驳回原因 |
| `implemented_at` | *time.Time | 实施完成时间 |

### 5.4 新增 Commands

| Command | Input | Output | 说明 |
|---------|-------|--------|------|
| `reject_ecn` | ecn_id, reason | ECNOutput | 驳回 ECN |
| `implement_ecn` | ecn_id, comment | ECNOutput | 标记 ECN 实施完成 |
| `add_ecn_approval` | ecn_id, approver_id, sequence | ApprovalOutput | 添加审批人 |
| `decide_ecn_approval` | approval_id, decision(approve/reject), comment | ApprovalOutput | 审批决定 |
| `create_ecn_task` | ecn_id, type, title, assignee_id, due_date | ECNTaskOutput | 创建 ECN 执行任务 |
| `complete_ecn_task` | task_id, comment | ECNTaskOutput | 完成 ECN 任务 |
| `list_ecn_history` | ecn_id | HistoryListOutput | 查询 ECN 操作历史 |
| `create_bom_ecn` | bom_id, title, change_summary | BOMECNOutput | 创建 BOM 变更通知 |
| `approve_bom_ecn` | ecn_id, comment | BOMECNOutput | 批准 BOM 变更 |
| `reject_bom_ecn` | ecn_id, reason | BOMECNOutput | 驳回 BOM 变更 |

---

## 六、Feature 5 — 零件图纸管理

### 6.1 新增表

#### `plm_part_drawings` — 零件图纸版本

```go
type PlmPartDrawing struct {
    ID                string    `gorm:"primaryKey;size:100"`
    BOMItemID         string    `gorm:"size:100;not null;index"`  // 关联 BOM Item
    DrawingType       string    `gorm:"size:4;not null"`          // 2D / 3D
    Version           string    `gorm:"size:16;not null"`         // v1, v2, v3
    FileID            string    `gorm:"size:100;not null"`
    FileName          string    `gorm:"size:256;not null"`
    FileSize          int64     `gorm:"default:0"`
    FileURL           string    `gorm:"size:512"`
    ChangeDescription string    `gorm:"type:text"`
    ChangeReason      string    `gorm:"size:256"`
    UploadedBy        string    `gorm:"size:100;not null"`
    CreatedAt         time.Time
}
```

### 6.2 新增 Commands

| Command | Input | Output | 说明 |
|---------|-------|--------|------|
| `add_part_drawing` | bom_item_id, drawing_type(2D/3D), version, file_id, file_name, ... | DrawingOutput | 上传零件图纸 |
| `list_part_drawings` | bom_item_id, drawing_type | DrawingListOutput | 查询零件图纸 |
| `delete_part_drawing` | drawing_id | DeleteOutput | 删除图纸 |

---

## 七、Feature 6 — 阶段交付物管理

### 7.1 新增表

#### `plm_phase_deliverables` — 阶段交付物

```go
type PlmPhaseDeliverable struct {
    ID              string     `gorm:"primaryKey;size:100"`
    PhaseID         string     `gorm:"size:100;not null"`
    Name            string     `gorm:"size:128;not null"`
    DeliverableType string     `gorm:"size:16;default:document"` // document/bom/review
    ResponsibleRole string     `gorm:"size:32"`                  // 负责角色
    IsRequired      bool       `gorm:"default:true"`
    Status          string     `gorm:"size:16;default:pending"`  // pending/submitted/approved
    DocumentID      string     `gorm:"size:100"`                 // 关联文档
    BOMID           string     `gorm:"size:100"`                 // 关联 BOM
    SubmittedAt     *time.Time
    SubmittedBy     string     `gorm:"size:100"`
    ApprovedAt      *time.Time
    ApprovedBy      string     `gorm:"size:100"`
    SortOrder       int        `gorm:"default:0"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### 7.2 新增 Commands

| Command | Input | Output | 说明 |
|---------|-------|--------|------|
| `create_deliverable` | phase_id, name, deliverable_type, responsible_role, is_required | DeliverableOutput | 创建交付物定义 |
| `submit_deliverable` | deliverable_id, document_id/bom_id | DeliverableOutput | 提交交付物 |
| `approve_deliverable` | deliverable_id, comment | DeliverableOutput | 审批交付物 |
| `list_deliverables` | phase_id | DeliverableListOutput | 查询阶段交付物 |

---

## 八、实现汇总

### 8.1 新增表（15 张）

| # | 表名 | Feature |
|---|------|---------|
| 1 | `plm_category_attr_templates` | BOM 体系 |
| 2 | `plm_bom_drafts` | BOM 体系 |
| 3 | `plm_bom_releases` | BOM 体系 |
| 4 | `plm_product_skus` | SKU/CMF |
| 5 | `plm_sku_cmf_configs` | SKU/CMF |
| 6 | `plm_sku_bom_items` | SKU/CMF |
| 7 | `plm_cmf_designs` | SKU/CMF |
| 8 | `plm_cmf_drawings` | SKU/CMF |
| 9 | `plm_bom_item_cmf_variants` | SKU/CMF |
| 10 | `plm_bom_item_lang_variants` | SKU/CMF |
| 11 | `plm_process_routes` | 制造工艺 |
| 12 | `plm_process_steps` | 制造工艺 |
| 13 | `plm_process_step_materials` | 制造工艺 |
| 14 | `plm_ecn_approvals` | ECN 流程 |
| 15 | `plm_ecn_tasks` | ECN 流程 |
| 16 | `plm_ecn_histories` | ECN 流程 |
| 17 | `plm_bom_ecns` | ECN 流程 |
| 18 | `plm_part_drawings` | 零件图纸 |
| 19 | `plm_phase_deliverables` | 阶段交付物 |

### 8.2 现有表改造（2 张）

| 表 | 改动 |
|----|------|
| `plm_boms` | 新增 7 个字段（source_bom_id 等） |
| `plm_bom_items` | 新增 9 个字段 + ExtendedAttrs 类型 text→jsonb |
| `plm_ecns` | 新增 5 个字段（approval_mode 等） |

### 8.3 新增 Commands（~49 个）

| Feature | Commands 数量 |
|---------|--------------|
| BOM 体系升级 | 11 |
| SKU/CMF 系统 | 16 |
| 制造工艺 | 8 |
| ECN 完整流程 | 10 |
| 零件图纸 | 3 |
| 阶段交付物 | 4 |

### 8.4 新增 Entity（~8 个）

`categoryAttrTemplateEntity`, `bomDraftEntity`, `bomReleaseEntity`, `productSKUEntity`, `cmfDesignEntity`, `cmfVariantEntity`, `processRouteEntity`, `phaseDeliverableEntity`

### 8.5 新增 Nav 入口

| Key | Label | Icon |
|-----|-------|------|
| `/m/plm/skus` | 产品 SKU | SkinOutlined |
| `/m/plm/routes` | 工艺路线 | ToolOutlined |
| `/m/plm/deliverables` | 阶段交付物 | CheckSquareOutlined |

### 8.6 Migrate 更新

在 `plm.go` 的 `Migrate()` 方法中追加所有新表的 `AutoMigrate` 调用。已有表的字段新增通过 GORM AutoMigrate 自动处理（仅 ADD COLUMN，不会 DROP）。

`ExtendedAttrs` 从 text 升级为 jsonb 需要手动 migration：

```sql
ALTER TABLE plm_bom_items ALTER COLUMN extended_attrs TYPE jsonb USING extended_attrs::jsonb;
```

---

## 九、实现优先级建议

| 优先级 | Feature | 理由 |
|--------|---------|------|
| **P0** | BOM 体系升级 | 硬件开发流程的核心数据结构，三级 BOM 是所有下游功能的基础 |
| **P0** | ECN 完整流程 | 设计变更是硬件开发的高频操作，需要完整审批链 |
| **P1** | SKU/CMF 系统 | 智能眼镜多配色方案的关键能力，但可在 BOM 稳定后接入 |
| **P1** | 零件图纸 | 与 BOM Item 紧密关联，实现简单 |
| **P2** | 制造工艺 | PVT/MP 阶段才需要，可后期补齐 |
| **P2** | 阶段交付物 | 增强型功能，可用 human/form 暂时替代 |

---

## 十、与硬件开发流程的关系

本 PRD 补齐的能力直接增强 `plm-hardware-dev` 流程的数据链路：

| 流程步骤 | 当前状态 | 补齐后 |
|---------|---------|--------|
| EVT BOM 编制 | human/form 手动填写成本数据 | 可用 `plm.create_bom` + `plm.add_bom_item` 在 PLM 中创建真实 BOM |
| DVT 设计冻结 | human/approval 确认冻结 | 可用 `plm.freeze_bom` 冻结 BOM |
| DVT 供应商定点 | human/form 手动记录 | BOM Item 的 `supplier_id` 关联 SRM 供应商 |
| PVT BOM 锁定 | human/form 填写成本 | `plm.release_bom_snapshot` 生成 ERP 快照 |
| 设计变更 | 无 ECN 流程 | 完整 ECN 审批→执行→追溯 |
| SKU 配色 | 无 | SKU → CMF 变体 → BOM 外观件选型 |
| 工艺路线 | 无 | MBOM 关联工艺步骤 + 工序物料 |
