# nimo PLM/SRM 系统代码现状审计报告

**审计日期:** 2026-02-24
**审计范围:** 后端 Go 代码 + 前端 React 代码
**代码位置:**
- 后端 PLM: `/home/claw/.openclaw/workspace/nimo-plm/internal/plm/`
- 后端 SRM: `/home/claw/.openclaw/workspace/internal/srm/`
- 前端: `/home/claw/.openclaw/workspace/nimo-plm-web/src/`
- 路由注册: `/home/claw/.openclaw/workspace/cmd/plm/main.go`

---

## 系统总览

### 技术栈
| 层级 | 技术选型 |
|------|---------|
| 后端框架 | Go + Gin |
| ORM | GORM (PostgreSQL + JSONB) |
| 缓存 | Redis |
| 对象存储 | MinIO |
| 前端框架 | React 18 + TypeScript + Vite |
| UI组件库 | Ant Design |
| 状态管理 | React Query + Context API |
| 实时通讯 | SSE (Server-Sent Events) |
| 企业集成 | 飞书 (OAuth/审批/消息/任务/日历) |

### 代码规模
| 指标 | 数量 |
|------|------|
| 后端 PLM 文件 | ~60 个 .go 文件 (14 entity + 18 handler + 13 service + 15 repository) |
| 后端 SRM 文件 | ~60 个 .go 文件 (13 entity + 18 handler + 13 service + 15 repository) |
| 前端文件 | 93 个 .tsx/.ts 文件 (24 pages + 18 components + 27 API modules) |
| API 端点总数 | ~450+ 个路由 |
| 数据库表 (估) | ~50+ 张表 |

### 架构模式
- 后端: Handler → Service → Repository → GORM (三层架构)
- 前端: Pages → API Modules → Axios Client (组件化架构)
- 认证: 飞书 OAuth + JWT + Redis Token 管理
- 通知: 飞书消息卡片 (异步发送)

---

## 一、项目管理 (Project / Template / Task)

### 后端

#### Entity
- **Project**: ID, Code, Name, Description, ProductID, OwnerID, CurrentPhase, Status(planning/evt/dvt/pvt/mp/completed/cancelled), PlannedStart, PlannedEnd, ActualStart, ActualEnd, Progress, CreatedBy, CreatedAt, UpdatedAt, DeletedAt
- **ProjectPhase**: ID, ProjectID, Code(concept/evt/dvt/pvt/mp), Name, Status(pending/active/completed), ActualStart, ActualEnd, SortOrder
- **Task**: ID, ProjectID, PhaseID, ParentTaskID, Code, Title, Description, TaskType(task/milestone/deliverable), Status(pending/in_progress/completed/confirmed/blocked/cancelled/unassigned/reviewing/rejected), Priority(low/medium/high/critical), AssigneeID, ReviewerID, StartDate, DueDate, CompletedAt, Progress, EstimatedHours, ActualHours, FeishuTaskID, Sequence, Level, Path, AutoStart, RequiresApproval, ApprovalType, ApprovalStatus, DefaultAssigneeRole, AutoCreateFeishuTask, FeishuApprovalCode
- **TaskDependency**: ID, TaskID, DependsOnID, DependencyType(FS/SS/FF/SF), LagDays
- **TaskComment**: ID, TaskID, UserID, Content
- **ProjectRoleAssignment**: ID, ProjectID, Phase, RoleCode, UserID, FeishuUserID, AssignedBy
- **TaskActionLog**: ID, ProjectID, TaskID, Action, FromStatus, ToStatus, OperatorID, OperatorType, EventData(JSONB), Comment
- **ProjectTemplate**: ID, Code, Name, Description, TemplateType(SYSTEM/CUSTOM), ProductType, Phases(JSON), EstimatedDays, IsActive, Version, Status(draft/published), BaseCode
- **TemplateTask**: ID, TemplateID, TaskCode, Name, Phase, ParentTaskCode, TaskType(MILESTONE/TASK/SUBTASK), DefaultAssigneeRole, EstimatedDays, IsCritical, Deliverables(JSON), Checklist(JSON), RequiresApproval, ApprovalType, AutoCreateFeishuTask, SortOrder
- **TemplateTaskDependency**: ID, TemplateID, TaskCode, DependsOnTaskCode, DependencyType, LagDays
- **TemplateTaskOutcome**: ID, TemplateID, TaskCode, OutcomeCode, OutcomeName, OutcomeType(pass/fail_rollback), RollbackToTaskCode, RollbackCascade
- **TaskForm**: ID, TaskID, Name, Description, Fields(JSON)
- **TaskFormSubmission**: ID, FormID, TaskID, Data(JSONB), Files(JSON), SubmittedBy, Version
- **TemplateTaskForm**: ID, TemplateID, TaskCode, Name, Fields(JSON)
- **AutomationRule**: ID, RuleType, TriggerCondition(JSON), ActionType, ActionConfig(JSON), IsActive, ProjectID, TemplateID, Priority
- **FeishuTaskSync**: ID, TaskID, FeishuTaskID, FeishuTaskGUID, SyncStatus, LastSyncAt
- **ReviewMeeting**: ID, Title, MeetingType(DESIGN/PHASE/BOM/ECN), ProjectID, TaskID, FeishuCalendarEventID, ScheduledAt, DurationMinutes, Attendees(JSON), Status
- **ApprovalInstance**: ID, ApprovalDefID, FeishuInstanceCode, BusinessType, BusinessID, Status, ApplicantID, FormData(JSON), Approvers(JSON), CurrentApproverID

#### API 端点 (30+ 路由)
```
GET    /api/v1/projects                              → 项目列表(分页/搜索/筛选)
POST   /api/v1/projects                              → 创建项目
GET    /api/v1/projects/:id                           → 项目详情
PUT    /api/v1/projects/:id                           → 更新项目
DELETE /api/v1/projects/:id                           → 删除项目
PUT    /api/v1/projects/:id/status                    → 更新项目状态
POST   /api/v1/projects/:id/assign-roles              → 分配角色
GET    /api/v1/projects/:id/phases                    → 阶段列表
PUT    /api/v1/projects/:id/phases/:phaseId/status    → 更新阶段状态
GET    /api/v1/projects/:id/tasks                     → 任务列表
POST   /api/v1/projects/:id/tasks                     → 创建任务
GET    /api/v1/projects/:id/tasks/:taskId             → 任务详情
PUT    /api/v1/projects/:id/tasks/:taskId             → 更新任务
DELETE /api/v1/projects/:id/tasks/:taskId             → 删除任务
PUT    /api/v1/projects/:id/tasks/:taskId/status      → 更新任务状态
GET    /api/v1/projects/:id/tasks/:taskId/subtasks    → 子任务列表
GET    /api/v1/projects/:id/tasks/:taskId/comments    → 任务评论
POST   /api/v1/projects/:id/tasks/:taskId/comments    → 添加评论
GET    /api/v1/projects/:id/tasks/:taskId/dependencies → 依赖关系
POST   /api/v1/projects/:id/tasks/:taskId/dependencies → 添加依赖
DELETE /api/v1/projects/:id/tasks/:taskId/dependencies/:depId → 删除依赖
GET    /api/v1/projects/:id/overdue-tasks             → 逾期任务
POST   /api/v1/projects/:id/tasks/:taskId/confirm     → 确认任务
POST   /api/v1/projects/:id/tasks/:taskId/reject      → 驳回任务
GET    /api/v1/my/tasks                               → 我的任务
POST   /api/v1/my/tasks/:taskId/complete              → 完成我的任务
POST   /api/v1/projects/create-from-template          → 从模板创建项目

# 模板管理 (15+ 路由)
GET    /api/v1/templates                              → 模板列表
POST   /api/v1/templates                              → 创建模板
GET    /api/v1/templates/:id                          → 模板详情
PUT    /api/v1/templates/:id                          → 更新模板
DELETE /api/v1/templates/:id                          → 删除模板
POST   /api/v1/templates/:id/duplicate                → 复制模板
POST   /api/v1/templates/:id/tasks                    → 创建模板任务
PUT    /api/v1/templates/:id/tasks/:taskCode          → 更新模板任务
DELETE /api/v1/templates/:id/tasks/:taskCode          → 删除模板任务
PUT    /api/v1/templates/:id/tasks/batch              → 批量保存任务
POST   /api/v1/templates/:id/publish                  → 发布模板
POST   /api/v1/templates/:id/upgrade                  → 升级版本
POST   /api/v1/templates/:id/revert                   → 回退版本
GET    /api/v1/templates/:id/versions                 → 版本列表

# 工作流 (条件注册, 6+ 路由)
POST   /api/v1/projects/:id/tasks/:taskId/assign      → 分配任务
POST   /api/v1/projects/:id/tasks/:taskId/start       → 开始任务
POST   /api/v1/projects/:id/tasks/:taskId/complete     → 完成任务
POST   /api/v1/projects/:id/tasks/:taskId/review       → 提交评审
POST   /api/v1/projects/:id/phases/:phase/assign-roles → 按阶段分配角色
GET    /api/v1/projects/:id/tasks/:taskId/history      → 操作历史
```

#### Service 层核心逻辑
- **ProjectService**: 项目CRUD、从模板创建带默认阶段(evt/dvt/pvt/mp)、任务管理、依赖关系、任务完成工作流(含表单提交+角色分配+自动审批)、进度计算
- **TemplateService**: 模板版本管理(BaseCode分组)、任务依赖图构建、日期自动计算(支持跳过周末)、从模板创建项目(含完整任务+依赖+表单复制)
- **AutomationService**: 任务完成后自动触发依赖任务(支持FS/SS/FF/SF四种依赖)、父任务自动完成、阶段完成检测、飞书通知
- **WorkflowService**: 状态机引擎驱动、任务分配/开始/完成/评审全流程、评审结果支持回滚(含级联)、操作日志审计

#### 完成度评估: **基本可用**

### 前端

#### 页面
- `Projects.tsx`: 项目列表页，卡片式展示，搜索+状态筛选，从模板创建项目入口
- `ProjectDetail.tsx`: 项目详情页，阶段/任务展示
- `Templates.tsx`: 模板管理页，版本分组，含代号管理集成，从模板创建项目流程(选代号→填信息→分配角色)
- `TemplateDetail.tsx`: 模板详情页
- `MyTasks.tsx`: 我的任务列表，SSE 实时更新，状态筛选
- `Dashboard.tsx`: 首页仪表盘，展示活跃项目/待办任务/进度统计

#### 核心功能
- 项目列表查看、搜索、状态筛选
- 从模板创建项目(含代号选择、项目经理分配、开始日期设定)
- 任务查看、评论
- SSE 实时通知
- 移动端适配

#### 缺失功能
- **甘特图**: 无 (当前仅有列表视图)
- **看板视图**: 项目管理无看板 (SRM 有独立看板)
- **任务拖拽排序**: 无
- **里程碑可视化**: 无专门视图
- **资源视图**: 无 (无法看谁负载过重)
- **项目仪表盘/报表**: 基础级别，无项目级深度分析
- **批量操作**: 无批量修改任务功能
- **项目导出**: 无导出为 Excel/PDF 能力

#### 完成度评估: **基本可用**

### 综合评估
- **当前状态**: 项目管理核心流程完整 (模板→项目→阶段→任务→依赖→审批→自动化)，是系统最成熟的模块之一
- **与"产品化"的差距**: 缺甘特图/看板/资源视图等可视化能力；缺项目级报表和仪表盘；缺项目复制、归档功能；需要更完善的权限控制（目前大部分 API 无细粒度权限校验）

---

## 二、BOM 管理 (BOM / Project BOM / BOM ECN)

### 后端

#### Entity
- **BOMHeader** (产品级BOM): ID, ProductID, Version, Status(draft/released/obsolete), Description, TotalItems, TotalCost, MaxLevel, ReleasedBy, ReleasedAt, ReleaseNotes, CreatedBy
- **BOMItem** (产品级BOM行项): ID, BOMHeaderID, ParentItemID, MaterialID, Level, Sequence, Quantity, Unit, Position, Reference, Notes, UnitCost, ExtendedCost
- **ProjectBOM** (项目级研发BOM): ID, ProjectID, PhaseID, BOMType(EBOM/SBOM/OBOM/FWBOM), Version, Name, Status(draft/pending_review/published/frozen/rejected), Description, SubmittedBy, SubmittedAt, ReviewedBy, ReviewedAt, ReviewComment, ApprovedBy, ApprovedAt, FrozenAt, FrozenBy, ParentBOMID, TotalItems, EstimatedCost, CreatedBy
- **ProjectBOMItem** (项目级BOM行项): ID, BOMID, ItemNumber, MaterialID, Category, Name, Specification, Quantity, Unit, Reference, Manufacturer, ManufacturerPN, Supplier, UnitPrice, LeadTimeDays, IsCritical, IsAlternative, AlternativeFor, Notes

#### API 端点 (40+ 路由)
```
# 项目BOM管理
GET    /api/v1/projects/:id/bom-permissions            → BOM权限查询
GET    /api/v1/projects/:id/boms                       → BOM列表(按类型/状态筛选)
POST   /api/v1/projects/:id/boms                       → 创建BOM
GET    /api/v1/projects/:id/boms/:bomId                → BOM详情
PUT    /api/v1/projects/:id/boms/:bomId                → 更新BOM
DELETE /api/v1/projects/:id/boms/:bomId                → 删除BOM
POST   /api/v1/projects/:id/boms/:bomId/submit         → 提交审核
POST   /api/v1/projects/:id/boms/:bomId/approve        → 审核通过
POST   /api/v1/projects/:id/boms/:bomId/reject         → 审核驳回
POST   /api/v1/projects/:id/boms/:bomId/freeze         → 冻结BOM
POST   /api/v1/projects/:id/boms/:bomId/items          → 添加行项
POST   /api/v1/projects/:id/boms/:bomId/items/batch    → 批量添加行项
PUT    /api/v1/projects/:id/boms/:bomId/items/:itemId  → 更新行项
DELETE /api/v1/projects/:id/boms/:bomId/items/:itemId  → 删除行项
POST   /api/v1/projects/:id/boms/:bomId/reorder        → 行项重排序
GET    /api/v1/projects/:id/boms/:bomId/export         → 导出BOM
POST   /api/v1/projects/:id/boms/:bomId/import         → 导入BOM
POST   /api/v1/projects/:id/boms/:bomId/release        → 发布BOM
POST   /api/v1/projects/:id/boms/create-from           → 从已有BOM创建
POST   /api/v1/projects/:id/boms/:bomId/convert-to-mbom → 转换为MBOM
POST   /api/v1/projects/:id/boms/:bomId/convert-to-pbom → 转换为PBOM
GET    /api/v1/projects/:id/boms/:bomId/category-tree  → 分类树视图

# 全局BOM操作
GET    /api/v1/bom-template                            → 下载BOM模板
GET    /api/v1/bom-items/search                        → 搜索BOM行项
GET    /api/v1/bom-items/search-paginated              → 分页搜索
GET    /api/v1/bom-items/global                        → 全局搜索
GET    /api/v1/bom-cost-summary                        → 成本汇总
POST   /api/v1/bom/parse                               → 解析BOM文件
GET    /api/v1/bom-compare                             → BOM对比

# BOM属性模板
GET    /api/v1/bom-attr-templates                      → 属性模板列表
POST   /api/v1/bom-attr-templates                      → 创建属性模板
PUT    /api/v1/bom-attr-templates/:id                  → 更新属性模板
DELETE /api/v1/bom-attr-templates/:id                  → 删除属性模板
POST   /api/v1/bom-attr-templates/seed                 → 初始化属性模板

# ERP集成
GET    /api/v1/erp/bom-releases                        → BOM发布记录列表
POST   /api/v1/erp/bom-releases/:id/ack                → 确认BOM发布

# BOM ECN (变更管理)
POST   /api/v1/projects/:id/boms/:bomId/edit           → 开始编辑(进入草稿)
POST   /api/v1/projects/:id/boms/:bomId/draft          → 保存草稿
GET    /api/v1/projects/:id/boms/:bomId/draft          → 获取草稿
DELETE /api/v1/projects/:id/boms/:bomId/draft           → 放弃草稿
POST   /api/v1/projects/:id/boms/:bomId/ecn            → 提交ECN变更
GET    /api/v1/bom-ecn                                 → ECN列表
GET    /api/v1/bom-ecn/:id                             → ECN详情
POST   /api/v1/bom-ecn/:id/approve                     → 批准ECN
POST   /api/v1/bom-ecn/:id/reject                      → 驳回ECN

# 产品级BOM (旧接口, 已废弃, 返回重定向提示)
GET    /api/v1/products/:id/bom                        → (废弃)
GET    /api/v1/products/:id/bom/versions               → (废弃)
POST   /api/v1/products/:id/bom/items                  → (废弃)
PUT    /api/v1/products/:id/bom/items/:itemId          → (废弃)
DELETE /api/v1/products/:id/bom/items/:itemId          → (废弃)
POST   /api/v1/products/:id/bom/release                → (废弃)
GET    /api/v1/products/:id/bom/compare                → (废弃)
```

#### Service 层核心逻辑
- **ProjectBOMService**: BOM 生命周期管理(draft→pending_review→published→frozen)、单个/批量添加行项(仅draft/rejected可编辑)、行项计数维护、物料关联
- **BOM ECN**: 草稿编辑模式、变更提交、审批流程
- 产品级 BOM (OldBOMHandler) 已废弃，全部重定向到项目级 BOM API

#### 完成度评估: **基本可用**

### 前端

#### 页面
- `BOM.tsx`: 产品级BOM管理页
- `BOMManagement.tsx`: 项目BOM管理列表页
- `BOMManagementDetail.tsx`: 项目BOM详细编辑页

#### 组件 (11个BOM专用组件)
- `EBOMControl.tsx`: 电子BOM编辑器
- `PBOMControl.tsx`: 结构BOM编辑器
- `MBOMControl.tsx`: 制造BOM编辑器
- `DynamicBOMTable.tsx`: 多类型BOM表格
- `BOMConfigPanel.tsx`: BOM配置面板
- `AddMaterialModal.tsx`: 物料选择弹窗
- `SupplierSelect.tsx`: 供应商选择下拉
- `BOMCategoryView.tsx`: 分类树视图
- `BOMItemMobileForm.tsx`: 移动端BOM表单
- `BOMEditableTable.tsx` / `EBOMEditableTable.tsx` / `PBOMEditableTable.tsx`: 可编辑表格

#### 核心功能
- 三种BOM类型支持 (EBOM/PBOM/MBOM)
- BOM行项增删改查
- 物料选择 (搜索+筛选)
- 供应商分配
- 成本计算 (单价 × 数量)
- 多级BOM结构 (父子层级)
- BOM版本管理与发布
- BOM对比
- 项目级BOM成本汇总
- 分类树视图
- 3D模型/STL查看器集成
- 移动端适配

#### 缺失功能
- **BOM导入**: 有 API 端点但前端实现深度未知
- **BOM差异可视化**: 对比功能较基础
- **替代料管理**: Entity 有 IsAlternative/AlternativeFor 字段但前端展示有限
- **BOM审批流程可视化**: 无审批进度展示
- **BOM变更历史时间线**: 无可视化
- **多级BOM展开/折叠**: 交互待优化
- **BOM Where-Used查询**: 无反向追溯 ("这个物料被哪些BOM引用")

#### 完成度评估: **基本可用**

### 综合评估 (重点审计)

- **当前状态**: BOM管理是系统中功能最丰富的模块，支持多类型BOM(EBOM/SBOM/OBOM/FWBOM)、完整审批流(draft→pending_review→published→frozen)、ECN草稿模式、导入导出、成本汇总、BOM对比
- **版本管理**: ✅ 支持版本号、BOM发布/冻结，有产品级BOM到项目级BOM的完整迁移
- **ECN联动**: ✅ 有独立的 BOM ECN 模块(草稿→提交→审批)，也有独立 ECN 模块可关联 BOM 行项
- **导入导出**: ⚠️ API存在(export/import/parse/template)，前端集成程度待验证
- **成本汇总**: ✅ 有 bom-cost-summary 端点和前端展示
- **与"产品化"的差距**: 需要 Where-Used 反向查询、更强的多级BOM可视化、替代料管理UI、BOM审批进度追踪、变更历史可视化、完整的导入验证和错误处理

---

## 三、物料管理 (Material / CMF / CMF Variant / Language Variant / SKU)

### 后端

#### Entity
- **MaterialCategory**: ID, Code, Name, ParentID, Path, Level, SortOrder
- **Material**: ID, Code, Name, CategoryID, Status(active/inactive/obsolete), Unit(pcs/kg/m/set), Description, Specs(JSONB), LeadTimeDays, MinOrderQty, SafetyStock, StandardCost, LastCost, Currency(CNY), CreatedBy, DeletedAt(软删除)

#### API 端点
```
GET    /api/v1/materials                              → 物料列表(分页/搜索/分类筛选)
POST   /api/v1/materials                              → 创建物料
GET    /api/v1/materials/:id                          → 物料详情
PUT    /api/v1/materials/:id                          → 更新物料
GET    /api/v1/material-categories                    → 分类列表(树形)

# CMF (Color/Material/Finish) 设计管理
GET    /api/v1/projects/:id/tasks/:taskId/cmf/appearance-parts → 外观件列表
GET    /api/v1/projects/:id/tasks/:taskId/cmf/designs → CMF设计列表
POST   /api/v1/projects/:id/tasks/:taskId/cmf/designs → 创建CMF设计
PUT    /api/v1/projects/:id/tasks/:taskId/cmf/designs/:designId → 更新CMF设计
DELETE /api/v1/projects/:id/tasks/:taskId/cmf/designs/:designId → 删除CMF设计
GET    /api/v1/projects/:id/cmf/designs               → 按项目查CMF设计
POST   /api/v1/cmf-designs/:designId/drawings         → 添加图纸
DELETE /api/v1/cmf-designs/:designId/drawings/:drawingId → 删除图纸

# CMF变体
GET    /api/v1/projects/:id/bom-items/:itemId/cmf-variants → CMF变体列表
POST   /api/v1/projects/:id/bom-items/:itemId/cmf-variants → 创建变体
PUT    /api/v1/projects/:id/cmf-variants/:variantId   → 更新变体
DELETE /api/v1/projects/:id/cmf-variants/:variantId   → 删除变体
GET    /api/v1/projects/:id/appearance-parts           → 外观件列表
GET    /api/v1/projects/:id/srm/items                  → SRM物料列表

# 语言变体
GET    /api/v1/projects/:id/bom-items/:itemId/lang-variants → 语言变体列表
POST   /api/v1/projects/:id/bom-items/:itemId/lang-variants → 创建语言变体
PUT    /api/v1/projects/:id/lang-variants/:variantId  → 更新语言变体
DELETE /api/v1/projects/:id/lang-variants/:variantId  → 删除语言变体
GET    /api/v1/projects/:id/multilang-parts            → 多语言件列表

# SKU管理
GET    /api/v1/projects/:id/skus                       → SKU列表
POST   /api/v1/projects/:id/skus                       → 创建SKU
PUT    /api/v1/projects/:id/skus/:skuId                → 更新SKU
DELETE /api/v1/projects/:id/skus/:skuId                → 删除SKU
GET    /api/v1/projects/:id/skus/:skuId/cmf            → SKU的CMF配置
PUT    /api/v1/projects/:id/skus/:skuId/cmf            → 批量保存CMF配置
GET    /api/v1/projects/:id/skus/:skuId/bom-items      → SKU的BOM行项
PUT    /api/v1/projects/:id/skus/:skuId/bom-items      → 批量保存BOM行项
GET    /api/v1/projects/:id/skus/:skuId/full-bom       → SKU完整BOM
```

#### Service 层核心逻辑
- 物料CRUD、分类树管理
- CMF设计与图纸关联
- CMF变体 (同一BOM项不同颜色/材质/表面处理)
- 语言变体 (同一BOM项不同语言版本)
- SKU管理 (CMF配置+BOM行项组合)

#### 完成度评估: **基本可用**

### 前端

#### 页面
- `Materials.tsx`: 物料库页面，树形分类+表格列表

#### 组件
- `CMFPanel.tsx`: CMF查看面板
- `CMFEditControl.tsx`: CMF编辑控件
- `CMFVariantEditor.tsx`: CMF变体编辑器

#### 核心功能
- 物料CRUD、搜索、分类树筛选
- 物料属性展示 (代码/名称/分类/单位/规格/成本/供应链参数)
- JSON规格查看器
- 移动端适配

#### 缺失功能
- **物料删除**: 无前端入口 (后端也无 DELETE 路由)
- **批量导入物料**: 无
- **物料变更历史**: 无
- **物料关联查询**: 无 (哪些BOM用了这个物料)
- **物料审批流程**: 无 (创建即生效)
- **CMF/SKU管理独立页面**: 当前嵌入在项目/BOM详情中，无独立管理入口

#### 完成度评估: **基本可用**

### 综合评估
- **当前状态**: 物料基础CRUD完整，CMF/语言变体/SKU管理有完整的后端API，是消费电子PLM的特色功能
- **与"产品化"的差距**: 缺物料导入、物料审批流程、物料变更追踪、Where-Used反查、批量操作、物料生命周期管理

---

## 四、ECN 变更管理 (ECN / BOM ECN)

### 后端

#### Entity
- **ECN**: ID, Code(ECN-YYYY-XXXX自动生成), Title, ProductID, ChangeType(design/material/process/document), Urgency(low/medium/high/critical), Status(draft/pending/approved/rejected/implemented/cancelled), Reason, Description, ImpactAnalysis, RequestedBy, ApprovedBy, ImplementedBy, FeishuApprovalCode, FeishuInstanceCode
- **ECNAffectedItem**: ID, ECNID, ItemType(bom_item/material/document/drawing), ItemID, BeforeValue(JSONB), AfterValue(JSONB), ChangeDescription
- **ECNApproval**: ID, ECNID, ApproverID, Sequence, Status(pending/approved/rejected), Decision, Comment, DecidedAt

#### API 端点 (20+ 路由)
```
GET    /api/v1/ecns                                    → ECN列表(分页/关键词/状态/类型/紧急度筛选)
POST   /api/v1/ecns                                    → 创建ECN
GET    /api/v1/ecns/stats                              → ECN统计(待审/执行中/月度创建关闭)
GET    /api/v1/ecns/my-pending                         → 我的待审ECN
GET    /api/v1/ecns/:id                                → ECN详情
PUT    /api/v1/ecns/:id                                → 更新ECN
POST   /api/v1/ecns/:id/submit                         → 提交审批
POST   /api/v1/ecns/:id/approve                        → 审批通过
POST   /api/v1/ecns/:id/reject                         → 审批驳回
POST   /api/v1/ecns/:id/implement                      → 标记已实施
GET    /api/v1/ecns/:id/affected-items                 → 受影响项列表
POST   /api/v1/ecns/:id/affected-items                 → 添加受影响项
PUT    /api/v1/ecns/:id/affected-items/:itemId         → 更新受影响项
DELETE /api/v1/ecns/:id/affected-items/:itemId         → 删除受影响项
GET    /api/v1/ecns/:id/approvals                      → 审批记录列表
POST   /api/v1/ecns/:id/approvers                      → 添加审批人
GET    /api/v1/ecns/:id/tasks                          → ECN任务列表
POST   /api/v1/ecns/:id/tasks                          → 创建ECN任务
PUT    /api/v1/ecns/:id/tasks/:taskId                  → 更新ECN任务
POST   /api/v1/ecns/:id/apply-bom-changes              → 应用BOM变更
GET    /api/v1/ecns/:id/history                        → 变更历史
```

#### Service 层核心逻辑
- ECN全生命周期: draft → pending → approved → implemented
- 多级顺序审批(Sequence排序，逐个审批，全部通过才能进入approved)
- 受影响项管理: 支持BOM行项/物料/文档/图纸，记录变更前后值(JSONB)
- ECN编号自动生成: PostgreSQL序列 (ECN-YYYY-XXXX)
- 飞书审批集成: 有字段预留(FeishuApprovalCode/FeishuInstanceCode)，Submit方法有TODO注释
- BOM变更应用: 有 apply-bom-changes 端点

#### 完成度评估: **基本可用**

### 前端

#### 页面
- `ECN/index.tsx`: ECN列表页，统计卡片(待审/执行中/月度)，表格/卡片视图切换，多维筛选
- `ECN/ECNDetail.tsx`: ECN详情页，审批流程、受影响项、执行任务、变更历史
- `ECN/ECNForm.tsx`: ECN创建/编辑表单

#### 核心功能
- ECN创建(变更类型/紧急度/原因/描述/影响分析)
- 统计仪表盘(待审批数/执行中/月度统计)
- 受影响项追踪
- 审批流程(含评论)
- 执行任务管理(分配人/截止日)
- 变更历史时间线
- 表格/卡片视图切换
- 多维筛选(状态/类型/紧急度/产品)

#### 缺失功能
- **ECN与BOM自动联动**: apply-bom-changes 端点存在，但前端触发方式不清晰
- **变更影响分析可视化**: 无图形化展示
- **ECN审批流程图**: 无可视化审批进度
- **ECN模板**: 无预设模板
- **ECN报表/趋势分析**: 基础统计，无趋势图

#### 完成度评估: **基本可用**

### 综合评估
- **当前状态**: ECN模块功能相当完整，覆盖了创建/审批/执行/实施全流程，有统计面板和历史追踪
- **与"产品化"的差距**: 飞书审批集成为TODO状态；缺ECN模板；缺与BOM的自动联动验证；缺变更影响的图形化分析

---

## 五、文档管理 (Document / Part Drawing)

### 后端

#### Entity
- **DocumentCategory**: ID, Code, Name, ParentID, SortOrder (自引用树形)
- **Document**: ID, Code(自动生成), Title, CategoryID, RelatedType(product/project/ecn/material), RelatedID, Status(draft/released/obsolete), Version, Description, FileName, FilePath, FileSize, MimeType, FeishuDocToken, FeishuDocURL, UploadedBy, ReleasedBy, ReleasedAt, DeletedAt
- **DocumentVersion**: ID, DocumentID, Version, FileName, FilePath, FileSize, ChangeSummary, CreatedBy
- **TaskAttachment**: ID, TaskID, FileName, FilePath, FileSize, MimeType, UploadedBy
- **OperationLog**: ID, UserID, UserName, Module, Action, TargetType, TargetID, TargetName, Before(JSONB), After(JSONB), IP, UserAgent
- **SystemConfig**: ID, Key, Value, ValueType, Module, Description, IsPublic
- **CodeRule**: ID, EntityType, Prefix, Separator, DateFormat, SeqLength, CurrentSeq, ResetCycle, Example

#### API 端点 (15+ 路由)
```
GET    /api/v1/documents                               → 文档列表(分页/搜索/分类筛选)
POST   /api/v1/documents                               → 上传文档(FormData)
GET    /api/v1/documents/:id                           → 文档详情
PUT    /api/v1/documents/:id                           → 更新元数据
DELETE /api/v1/documents/:id                           → 删除文档(软删除)
GET    /api/v1/documents/:id/download                  → 下载文档
POST   /api/v1/documents/:id/release                   → 发布文档
POST   /api/v1/documents/:id/obsolete                  → 废弃文档
GET    /api/v1/documents/:id/versions                  → 版本列表
POST   /api/v1/documents/:id/versions                  → 上传新版本
GET    /api/v1/documents/:id/versions/:versionId/download → 下载指定版本
GET    /api/v1/document-categories                     → 分类列表

# 零件图纸管理
GET    /api/v1/projects/:id/bom-items/:itemId/drawings → 零件图纸列表
POST   /api/v1/projects/:id/bom-items/:itemId/drawings → 上传图纸
DELETE /api/v1/projects/:id/bom-items/:itemId/drawings/:drawingId → 删除图纸
GET    /api/v1/projects/:id/bom-items/:itemId/drawings/:drawingId/download → 下载图纸
GET    /api/v1/projects/:id/boms/:bomId/drawings       → BOM图纸列表

# 文件上传
POST   /api/v1/upload                                  → 通用文件上传
GET    /api/v1/files/:fileId/3d                        → 3D模型查看
```

#### Service 层核心逻辑
- **DocumentService**: MinIO文件存储、版本管理(自动递增1.0→1.1→1.2)、发布/废弃工作流、分类管理、按关联对象查询
- 存储路径: `documents/{YYYY/MM/DD}/{uuid}{ext}`
- 支持飞书文档Token集成

#### 完成度评估: **基本可用**

### 前端

#### 页面
- `Documents.tsx`: 文档管理页

#### 核心功能
- 文档上传 (自动检测文件类型: PDF/Word/Excel/图片/ZIP)
- 分类筛选
- 版本管理 (上传新版本/查看历史)
- 搜索 (标题/关键词)
- 文件下载
- 文档发布/废弃工作流
- 文件类型图标区分

#### 缺失功能
- **在线预览**: 无 (仅下载)
- **文档权限控制**: 无细粒度权限
- **全文搜索**: 无 (仅标题搜索)
- **文档审批流程**: 无 (直接发布)
- **文档关联视图**: 无 (虽然后端支持 RelatedType/RelatedID)
- **文档标签**: 后端支持但前端展示有限

#### 完成度评估: **基本可用**

### 综合评估
- **当前状态**: 文档上传/版本/下载/分类基础完整，3D模型查看是亮点
- **与"产品化"的差距**: 缺在线预览(特别是PDF/图片)、全文搜索、文档审批、权限控制、关联展示

---

## 六、审批流程 (Approval / Approval Definition / Workflow)

### 后端

#### Entity
- **ApprovalRequest**: ID, ProjectID, TaskID, Title, Description, Type(task_review/definition), Status(pending/approved/rejected/cancelled), RequestedBy, FormData(JSONB), CurrentNode, FlowSnapshot(JSON), CreatedAt
- **ApprovalReviewer**: ID, ApprovalRequestID, UserID, NodeIndex, Sequence, Status(pending/approved/rejected), Decision, Comment, DecidedAt
- **ApprovalDefinition**: ID, Code, Name, Description, Icon, GroupName, FormSchema(JSON), FlowSchema(JSON), Visibility, AdminUserID, Status(draft/published), SortOrder, CreatedBy, DeletedAt
- **ApprovalGroup**: ID, Name, SortOrder
- **FlowSchema**: Nodes(数组), 每个Node含 Type(submit/approve/cc), Name, Config(ApproverType/ApproverIDs/MultiApprove等)

#### API 端点 (15+ 路由)
```
POST   /api/v1/approvals                               → 创建审批请求
GET    /api/v1/approvals                               → 审批列表
GET    /api/v1/approvals/:id                           → 审批详情
POST   /api/v1/approvals/:id/approve                   → 审批通过
POST   /api/v1/approvals/:id/reject                    → 审批驳回

GET    /api/v1/approval-definitions                    → 审批定义列表(分组)
POST   /api/v1/approval-definitions                    → 创建审批定义
GET    /api/v1/approval-definitions/:id                → 审批定义详情
PUT    /api/v1/approval-definitions/:id                → 更新审批定义
DELETE /api/v1/approval-definitions/:id                → 删除审批定义(仅draft)
POST   /api/v1/approval-definitions/:id/publish        → 发布审批定义
POST   /api/v1/approval-definitions/:id/unpublish      → 取消发布
POST   /api/v1/approval-definitions/:id/submit         → 提交审批实例

GET    /api/v1/approval-groups                         → 审批分组列表
POST   /api/v1/approval-groups                         → 创建分组
DELETE /api/v1/approval-groups/:id                     → 删除分组
```

#### Service 层核心逻辑
- **ApprovalService**: 多节点审批流(支持多个approve节点串行)、单节点内多人审批、单人驳回即全部驳回、审批完成自动启动依赖任务、飞书消息通知审批人/发起人
- **ApprovalDefinitionService**: 审批模板管理、表单Schema+流程Schema定义、发布/取消发布、支持指定人/自选人/提交人/角色等审批类型
- **与飞书集成**: 通过 FeishuClient 发送交互式卡片通知，非飞书原生审批流 (自建审批流+飞书通知)

#### 完成度评估: **基本可用**

### 前端

#### 页面
- `Approvals.tsx`: 审批中心(审批/发起 Tab切换)
- `ApprovalAdmin.tsx`: 审批管理后台(定义+分组管理)
- `ApprovalEditor.tsx`: 审批定义编辑器

#### 核心功能
- 审批中心: 查看待审批列表、审批/驳回(含评论)
- 发起审批: 浏览已发布定义、提交审批请求、动态表单渲染
- 管理后台: 审批定义CRUD、分组管理、发布/取消
- 审批定义编辑器: 定义审批步骤/审批人/条件

#### 缺失功能
- **审批流程可视化**: 无流程图展示
- **审批统计/报表**: 无
- **审批委托/转交**: 无
- **审批催办**: 无
- **条件分支**: FlowSchema设计支持但前端实现深度不明
- **飞书原生审批**: 当前是自建审批+飞书通知，非飞书审批中心集成

#### 完成度评估: **基本可用**

### 综合评估
- **当前状态**: 自建审批流程完整 (定义→发布→发起→多节点审批→通知)，飞书集成为消息通知级别
- **与飞书集成状态**: ⚠️ 当前是 **自建审批流 + 飞书消息卡片通知**，并非飞书审批中心原生集成。ECN模块的飞书审批字段已预留但标注TODO。SRM采样验证模块有较深的飞书审批集成(创建审批定义+提交实例+处理回调)
- **与"产品化"的差距**: 需审批流程可视化、审批委托/转交/催办、审批统计、移动端审批优化；如需深度飞书集成需改造为飞书审批中心方式

---

## 七、用户与权限 (User / Role / Auth)

### 后端

#### Entity
- **User**: ID, Username, Email, Name, AvatarURL, EmployeeNo, Mobile, Status(active/inactive/suspended), DepartmentID, FeishuOpenID, FeishuUserID, FeishuUnionID, LastLoginAt, CreatedAt, UpdatedAt, DeletedAt + Roles(多对多), Permissions(运行时加载)
- **Department**: ID, Name, ParentID, FeishuDeptID, FeishuOpenDeptID, Status, SortOrder
- **Role**: ID, Code, Name, Description, IsSystem, CreatedAt, UpdatedAt + Users, Permissions(多对多)
- **Permission**: ID, Code, Name, Description, Module, CreatedAt
- **RoleUser**: RoleID, UserID (关联表)
- **RolePermission**: RoleID, PermissionID (关联表)

#### API 端点
```
GET    /api/v1/auth/feishu/login                       → 飞书OAuth登录
GET    /api/v1/auth/feishu/callback                    → 飞书OAuth回调
POST   /api/v1/auth/refresh                            → 刷新Token
GET    /api/v1/auth/me                                 → 当前用户信息
POST   /api/v1/auth/logout                             → 退出登录

GET    /api/v1/users                                   → 用户列表
GET    /api/v1/users/search                            → 搜索用户
GET    /api/v1/users/:id                               → 用户详情

GET    /api/v1/roles                                   → 角色列表
POST   /api/v1/roles                                   → 创建角色
GET    /api/v1/roles/:id                               → 角色详情
PUT    /api/v1/roles/:id                               → 更新角色
DELETE /api/v1/roles/:id                               → 删除角色
GET    /api/v1/roles/:id/members                       → 角色成员列表
POST   /api/v1/roles/:id/members                       → 添加成员
DELETE /api/v1/roles/:id/members                       → 移除成员
GET    /api/v1/departments                             → 部门列表
GET    /api/v1/task-roles                              → 任务角色列表
GET    /api/v1/feishu/roles                            → 飞书角色列表

POST   /api/v1/admin/sync-contacts                     → 同步飞书通讯录

# Webhooks (无需认证)
POST   /api/v1/webhooks/feishu/approval                → 飞书审批回调
POST   /api/v1/webhooks/feishu/event                   → 飞书事件验证
```

#### Service 层核心逻辑
- **AuthService**: 飞书OAuth认证流程 (获取AppToken→换取UserToken→获取UserInfo→创建/更新用户→生成JWT)、JWT Token管理(Access+Refresh，Redis存储)、Token刷新、支持API Token管理员绕行+用户模拟(X-Impersonate-User)
- **ContactSyncService**: 飞书通讯录同步(部门+人员批量upsert)、处理唯一性约束冲突

#### 完成度评估: **基本可用**

### 前端

#### 页面
- `Login.tsx`: 飞书OAuth登录页
- `RoleManagement.tsx`: 角色管理页

#### 核心功能
- 飞书一键登录
- 角色CRUD
- 角色成员管理(添加/移除)
- 部门视图

#### 缺失功能
- **用户管理页面**: 无独立用户管理 (仅通过飞书同步)
- **权限管理界面**: 无 (角色-权限映射无前端管理)
- **权限校验中间件**: 后端有 RequirePermission/RequireRole 中间件但大部分路由未使用
- **数据权限**: 无 (无法限制"只看自己创建的项目")
- **操作审计日志查看**: 后端有 OperationLog 但无前端展示

#### 完成度评估: **骨架**

### 综合评估
- **当前状态**: 飞书登录+JWT认证完整，角色管理基本可用，但权限体系未完整落地
- **与"产品化"的差距**: 权限是最大短板 — 需要完善API级权限校验、数据级权限(数据隔离)、权限管理UI、操作审计日志查看；需用户管理界面(用于停用/编辑非飞书同步信息)

---

## 八、SRM 供应商管理 (完整 SRM 模块)

### 后端

#### Entity (13个实体，20个数据表)

**供应商管理:**
- **Supplier**: ID, Code, Name, ShortName, Category(structural/electronic/optical/packaging/manufacturer/other), Level(potential/qualified/preferred/strategic), Status(pending/active/suspended/blacklisted), 地址信息, BusinessScope, AnnualRevenue, EmployeeCount, FactoryArea, Certifications(JSONB), 银行信息, Tags(JSONB), TechCapability, Cooperation, CapacityLimit, QualityScore, DeliveryScore, PriceScore, OverallScore
- **SupplierContact**: ID, SupplierID, Name, Title, Phone, Email, Wechat, IsPrimary
- **SupplierMaterial**: ID, SupplierID, CategoryID, MaterialID, LeadTimeDays, MOQ, UnitPrice, Currency

**采购需求:**
- **PurchaseRequest**: ID, PRCode, Title, Type(sample/production), Priority(urgent/high/normal/low), Status(draft/pending/approved/sourcing/completed/cancelled), ProjectID, SRMProjectID, BOMID, Phase, RequiredDate, RequestedBy, ApprovedBy, Notes
- **PRItem**: ID, PRID, MaterialID, MaterialCode/Name/Specification, Category, SourceBOMType(EBOM/SBOM/ABOM/TOOLING), MaterialGroup(electronic/structural/assembly/tooling), ImageURL, ProcessType, ToolingCost, ToolingStatus, JigPhase, JigProgress, Quantity, Unit, Status(pending/quoting/sourcing/ordered/shipped/received/inspecting/passed/failed/inspected/completed), SupplierID, UnitPrice, TotalAmount, ExpectedDate, ActualDate, InspectionResult, Round, SortOrder

**采购订单:**
- **PurchaseOrder**: ID, POCode, SupplierID, PRID, SRMProjectID, Type, Status(draft/submitted/approved/sent/partial/received/completed/cancelled), Round, PrevPOID, Related8DID, TotalAmount, Currency, ExpectedDate, ActualDate, ShippingAddress, PaymentTerms
- **POItem**: ID, POID, PRItemID, BOMItemID, MaterialID, MaterialCode/Name/Specification, Quantity, Unit, UnitPrice, TotalAmount, ReceivedQty, Status(pending/shipped/partial/received)

**质量检验:**
- **Inspection**: ID, InspectionCode, POID, POItemID, SupplierID, MaterialID/Code/Name, Quantity, SampleQty, Status(pending/inspecting/completed), Result, OverallResult, InspectionItems(JSONB), ReportURL, Inspector, InspectedAt
- **InspectionItem**: ID, InspectionID, POItemID, MaterialName/Code, InspectedQty, QualifiedQty, DefectQty, DefectDesc, Result(passed/failed/conditional)
- **CorrectiveAction** (8D): ID, CACode, InspectionID, SupplierID, ProblemDesc, Severity(critical/major/minor), Status(open/responded/verified/closed), RootCause, CorrectiveAction, PreventiveAction, ResponseDeadline, RespondedAt, VerifiedAt, ClosedAt

**供应商评估:**
- **SupplierEvaluation**: ID, SupplierID, Period, EvalType(quarterly/monthly/annual), QualityScore/DeliveryScore/PriceScore/ServiceScore/TotalScore, 权重(0.30/0.25/0.25/0.20), Grade(A/B/C/D), TotalPOs, OnTimePOs, QualityPassed/Total, Status(draft/submitted/approved)

**对账结算:**
- **Settlement**: ID, SettlementCode, SupplierID, PeriodStart/End, Status(draft/confirmed/invoiced/paid), POAmount/ReceivedAmount/Deduction/FinalAmount, InvoiceNo/URL/Amount, ConfirmedByBuyer/Supplier
- **SettlementDispute**: ID, SettlementID, DisputeType(price_diff/quantity_diff/quality_deduction/other), AmountDiff, Status(open/resolved), Resolution

**打样管理:**
- **SamplingRequest**: ID, PRItemID, SupplierID, Round, SampleQty, Status(preparing/shipping/arrived/verifying/passed/failed), VerifyResult, RejectReason, ApprovalID(飞书审批实例)

**项目与库存:**
- **SRMProject**: ID, Code, Name, Type, Phase, Status(active/completed/cancelled), PLMProjectID/TaskID/BOMID, 进度计数(TotalItems/SourcingCount/OrderedCount/ReceivedCount/PassedCount/FailedCount), EstimatedDays, StartDate/TargetDate/ActualDate
- **InventoryRecord**: ID, MaterialName/Code/MPN, SupplierID, Quantity, SafetyStock, Unit, Warehouse, LastInDate
- **InventoryTransaction**: ID, InventoryID, Type(in/out/adjust), Quantity, ReferenceType(inspection/manual/adjust), ReferenceID, Operator
- **DelayRequest**: ID, Code, SRMProjectID, PRItemID, MaterialName, OriginalDays/RequestedDays, Reason, ReasonType, Status(pending/approved/rejected)
- **ActivityLog**: ID, EntityType, EntityID/Code, Action, FromStatus/ToStatus, Content, Attachments/Metadata(JSONB), OperatorID/Name
- **Equipment**: (最小实现, 仅ID和时间戳)
- **RFQ/RFQQuote**: (最小实现, 仅ID和时间戳)

#### API 端点 (80+ 路由)
```
# 供应商管理
GET    /api/v1/srm/suppliers                           → 供应商列表
POST   /api/v1/srm/suppliers                           → 创建供应商
GET    /api/v1/srm/suppliers/:id                       → 供应商详情(含评分)
PUT    /api/v1/srm/suppliers/:id                       → 更新供应商
DELETE /api/v1/srm/suppliers/:id                       → 删除供应商(软删)
GET    /api/v1/srm/suppliers/:id/contacts              → 联系人列表
POST   /api/v1/srm/suppliers/:id/contacts              → 添加联系人
DELETE /api/v1/srm/suppliers/:id/contacts/:contactId   → 删除联系人

# 采购需求 (PR)
GET    /api/v1/srm/purchase-requests                   → PR列表
POST   /api/v1/srm/purchase-requests                   → 创建PR
POST   /api/v1/srm/purchase-requests/from-bom          → 从BOM创建PR
GET    /api/v1/srm/purchase-requests/:id               → PR详情
PUT    /api/v1/srm/purchase-requests/:id               → 更新PR
POST   /api/v1/srm/purchase-requests/:id/approve       → 审批PR
PUT    /api/v1/srm/purchase-requests/:id/items/:itemId/assign-supplier → 分配供应商
POST   /api/v1/srm/purchase-requests/:id/generate-pos  → 从PR生成PO

# 采购订单 (PO)
GET    /api/v1/srm/purchase-orders                     → PO列表
GET    /api/v1/srm/purchase-orders/export              → 导出PO
POST   /api/v1/srm/purchase-orders                     → 创建PO
POST   /api/v1/srm/purchase-orders/from-bom            → 从BOM生成PO
GET    /api/v1/srm/purchase-orders/:id                 → PO详情
PUT    /api/v1/srm/purchase-orders/:id                 → 更新PO
POST   /api/v1/srm/purchase-orders/:id/submit          → 提交PO
POST   /api/v1/srm/purchase-orders/:id/approve         → 审批PO
POST   /api/v1/srm/purchase-orders/:id/items/:itemId/receive → 收货
DELETE /api/v1/srm/purchase-orders/:id                 → 删除PO(仅草稿)

# 来料检验
GET    /api/v1/srm/inspections                         → 检验列表
POST   /api/v1/srm/inspections                         → 创建检验
POST   /api/v1/srm/inspections/from-po                 → 从PO创建检验
GET    /api/v1/srm/inspections/:id                     → 检验详情
PUT    /api/v1/srm/inspections/:id                     → 更新检验
POST   /api/v1/srm/inspections/:id/complete            → 完成检验

# 库存管理
GET    /api/v1/srm/inventory                           → 库存列表
GET    /api/v1/srm/inventory/:id/transactions          → 流水记录
POST   /api/v1/srm/inventory/in                        → 入库
POST   /api/v1/srm/inventory/out                       → 出库
POST   /api/v1/srm/inventory/adjust                    → 库存调整

# 采购项目
GET    /api/v1/srm/projects                            → 项目列表
POST   /api/v1/srm/projects                            → 创建项目
GET    /api/v1/srm/projects/:id                        → 项目详情
PUT    /api/v1/srm/projects/:id                        → 更新项目
GET    /api/v1/srm/projects/:id/progress               → 进度概览
GET    /api/v1/srm/projects/:id/progress-by-group      → 按组进度
GET    /api/v1/srm/projects/:id/items-by-group         → 按组物料列表
GET    /api/v1/srm/projects/:id/activities             → 操作日志

# 对账结算
GET    /api/v1/srm/settlements                         → 结算列表
GET    /api/v1/srm/settlements/export                  → 导出结算
POST   /api/v1/srm/settlements                         → 创建结算
POST   /api/v1/srm/settlements/generate                → 自动生成结算
GET    /api/v1/srm/settlements/:id                     → 结算详情
PUT    /api/v1/srm/settlements/:id                     → 更新结算
DELETE /api/v1/srm/settlements/:id                     → 删除结算(仅草稿)
POST   /api/v1/srm/settlements/:id/confirm-buyer       → 采购方确认
POST   /api/v1/srm/settlements/:id/confirm-supplier    → 供应商确认
POST   /api/v1/srm/settlements/:id/disputes            → 添加差异
PUT    /api/v1/srm/settlements/:id/disputes/:disputeId → 更新差异

# 8D纠正措施
GET    /api/v1/srm/corrective-actions                  → 8D列表
POST   /api/v1/srm/corrective-actions                  → 创建8D
GET    /api/v1/srm/corrective-actions/:id              → 8D详情
PUT    /api/v1/srm/corrective-actions/:id              → 更新8D
POST   /api/v1/srm/corrective-actions/:id/respond      → 供应商回复
POST   /api/v1/srm/corrective-actions/:id/verify       → 验证
POST   /api/v1/srm/corrective-actions/:id/close        → 关闭

# 供应商评估
GET    /api/v1/srm/evaluations                         → 评估列表
POST   /api/v1/srm/evaluations                         → 创建评估
POST   /api/v1/srm/evaluations/auto-generate           → 自动生成评估
GET    /api/v1/srm/evaluations/supplier/:supplierId    → 供应商历史
GET    /api/v1/srm/evaluations/:id                     → 评估详情
PUT    /api/v1/srm/evaluations/:id                     → 更新评估
POST   /api/v1/srm/evaluations/:id/submit              → 提交评估
POST   /api/v1/srm/evaluations/:id/approve             → 审批评估

# RFQ询价
GET    /api/v1/srm/rfq                                 → 询价列表
POST   /api/v1/srm/rfq                                 → 创建询价
GET    /api/v1/srm/rfq/:id                             → 询价详情
POST   /api/v1/srm/rfq/:id/quotes                      → 添加报价
PUT    /api/v1/srm/rfq/:id/quotes/:quoteId             → 更新报价
POST   /api/v1/srm/rfq/:id/quotes/:quoteId/select      → 选择报价
POST   /api/v1/srm/rfq/:id/convert-to-po               → 转为PO
GET    /api/v1/srm/rfq/:id/comparison                  → 比价分析

# 打样管理
POST   /api/v1/srm/pr-items/:itemId/sampling           → 创建打样请求
GET    /api/v1/srm/pr-items/:itemId/sampling            → 打样记录
PUT    /api/v1/srm/sampling/:id/status                  → 更新打样状态
POST   /api/v1/srm/sampling/:id/request-verify          → 发起验证(飞书审批)
POST   /api/v1/srm/sampling/verify-callback             → 验证回调

# 延期审批
GET    /api/v1/srm/delay-requests                      → 延期列表
POST   /api/v1/srm/delay-requests                      → 创建延期
GET    /api/v1/srm/delay-requests/:id                  → 延期详情
POST   /api/v1/srm/delay-requests/:id/approve          → 审批通过
POST   /api/v1/srm/delay-requests/:id/reject           → 审批驳回

# 设备管理
GET    /api/v1/srm/equipment                           → 设备列表
POST   /api/v1/srm/equipment                           → 创建设备
GET    /api/v1/srm/equipment/:id                       → 设备详情
PUT    /api/v1/srm/equipment/:id                       → 更新设备
DELETE /api/v1/srm/equipment/:id                       → 删除设备

# 仪表盘
GET    /api/v1/srm/dashboard/sampling-progress          → 打样进度
```

#### Service 层核心逻辑 (完整业务链路分析)

**主流程: PLM BOM审批 → SRM全链路**
```
PLM项目BOM审批通过
    → SRMProjectService.CreateFromBOM() 自动创建采购项目
    → ProcurementService.AutoCreatePRFromBOM() 自动创建采购需求
    → 供应商分配 (AssignSupplierToItem)
    → PO生成 (GeneratePOsFromPR, 按供应商自动分组)
    → PO审批 (ApprovePO, 飞书通知)
    → 来料检验 (CreateInspectionFromPO)
    → 检验完成 → 合格品自动入库 (InventoryService.StockInFromInspection)
    → 不合格 → 8D纠正措施
    → 结算 (SettlementService.Generate, 双方确认, 对账差异处理)
```

**打样子流程:**
```
PRItem(pending) → 创建打样请求(round 1)
    → preparing → shipping → arrived
    → 发起飞书审批验证 (RequestVerify)
    → 飞书回调: 通过→PRItem(quoting) / 驳回→可重试(round 2)
```

**评估流程:**
```
TotalScore = Quality×0.30 + Delivery×0.25 + Price×0.25 + Service×0.20
Grade: A(≥90), B(≥75), C(≥60), D(<60)
审批通过后自动更新供应商评分
```

#### 完成度评估: **基本可用** (RFQ/Equipment为骨架级)

### 前端

#### 页面 (12+ 个)
- `SRMDashboard.tsx`: SRM总览(采购项目/统计数据)
- `Suppliers.tsx`: 供应商列表(分类/等级/状态筛选)
- `PurchaseRequests.tsx`: 采购需求管理
- `PurchaseOrders.tsx`: 采购订单管理
- `Inspections.tsx`: 来料检验管理
- `KanbanBoard.tsx`: 看板视图(PR/PO流程)
- `Projects.tsx`: SRM项目列表
- `ProjectDetail/`: SRM项目详情
- `Inventory.tsx`: 库存管理
- `Settlements/`: 对账结算
- `CorrectiveActions/`: 8D纠正措施
- `Evaluations/`: 供应商评估
- `Equipment/`: 设备管理

#### 核心功能
- 完整采购链路前端: 供应商→PR→PO→检验→入库→结算
- 看板视图(PR/PO流程可视化)
- 供应商360度画像(分类/等级/评分/认证/银行信息)
- 按物料组Tab展示(电子/结构/组装/模具)
- PO导出
- 结算导出
- 进度追踪(订购/收货/检验/通过)
- 延期审批

#### 缺失功能
- **RFQ询价**: 后端Entity最小实现，前端无独立页面
- **供应商门户**: 无供应商自助登录
- **比价分析可视化**: 后端有comparison端点，前端展示深度有限
- **采购报表**: 无采购趋势/支出分析图表
- **库存预警**: 有SafetyStock字段但无预警机制
- **批量操作**: 无批量审批/批量下单

#### 完成度评估: **基本可用**

### 综合评估 (重点审计)
- **当前状态**: SRM是系统中业务链路最完整的模块，从供应商管理到结算对账形成闭环，PLM→SRM自动联动是亮点
- **完整业务链路评估**:
  - ✅ 供应商管理: 完整 (360度画像/联系人/等级/评分)
  - ✅ 询价: 后端骨架(Entity/Handler/Service占位)，**需要完善**
  - ✅ 采购: 完整 (PR→PO, 自动从BOM生成, 按供应商分组)
  - ✅ 入库: 完整 (检验合格自动入库, 手动出入库/调整)
  - ✅ 结算: 完整 (自动生成/双方确认/差异处理/发票)
  - ⚠️ 打样: 完整 (多轮打样/飞书审批验证)
  - ⚠️ 8D: 后端基本完整，Service实现较简
- **与"产品化"的差距**: RFQ询价模块需完善(核心缺失)；需供应商门户(让供应商自助操作)；需采购报表/分析；库存预警机制；批量操作

---

## 九、Dashboard

### 后端
- 无独立 Dashboard 后端服务 (PLM侧)
- SRM 有 `DashboardService` 提供打样进度统计

#### API 端点
```
GET    /api/v1/srm/dashboard/sampling-progress          → SRM打样进度
GET    /api/v1/sse/events                              → SSE实时事件推送
```

### 前端

#### 页面
- `Dashboard.tsx`: 系统首页仪表盘

#### 核心功能
- 欢迎信息 (当前用户)
- 统计卡片: 活跃项目数 / 待办任务数 / 已完成项目数 / 物料总数
- 我的任务列表 (Top 10, 优先级标签, 截止日期)
- 参与项目列表 (Top 6, 进度条, 阶段标签)
- SSE 实时更新
- 骨架屏加载效果
- 移动端适配

#### 缺失功能
- **图表可视化**: 无饼图/柱图/折线图
- **自定义Dashboard**: 无
- **SRM Dashboard集成**: SRM有独立Dashboard但与PLM首页分离
- **KPI指标**: 无项目健康度/延期率/资源利用率
- **快捷操作**: 无快速创建项目/任务入口

#### 完成度评估: **基本可用**

### 综合评估
- **当前状态**: 基础仪表盘可用，有实时更新能力，但可视化和分析能力弱
- **与"产品化"的差距**: 需要添加图表(echarts/recharts)、自定义布局、深度KPI分析、PLM+SRM统一仪表盘

---

## 十、产品管理 (Product / Codename)

### 后端

#### Entity
- **ProductCategory**: ID, Code, Name, ParentID, Path, Level, SortOrder, Description (树形自引用)
- **Product**: ID, Code(自动生成PRD-NIMO-XXXX), Name, CategoryID, Status(draft/developing/active/discontinued), Description, Specs(JSONB), ThumbnailURL, CurrentBOMVersion, CreatedBy, UpdatedBy, ReleasedAt, DiscontinuedAt, DeletedAt
- **ProjectCodename**: ID, Codename, CodenameType(platform/product), Generation, Theme, Description, IsUsed, UsedByProjectID

#### API 端点
```
GET    /api/v1/products                                → 产品列表
POST   /api/v1/products                                → 创建产品
GET    /api/v1/products/:id                            → 产品详情
PUT    /api/v1/products/:id                            → 更新产品(active状态不可直接修改)
DELETE /api/v1/products/:id                            → 删除产品(active状态不可删)
POST   /api/v1/products/:id/release                    → 发布产品(需有已发布BOM)
GET    /api/v1/product-categories                      → 产品分类列表
GET    /api/v1/codenames                               → 代号列表
```

#### Service 层核心逻辑
- **ProductService**: CRUD + Redis缓存、发布需要已有 CurrentBOMVersion、Active状态产品禁止直接修改(需走ECN流程)、编码自动生成

#### 完成度评估: **基本可用**

### 前端

#### 页面
- `Products.tsx`: 产品管理页

#### 核心功能
- 产品列表(编码/名称/分类/版本/状态)
- CRUD操作(弹窗表单)
- 搜索/状态筛选
- 分页

#### 缺失功能
- **产品详情页**: 无独立详情页(仅弹窗)
- **产品关联视图**: 无 (关联BOM/项目/ECN/文档)
- **产品生命周期可视化**: 无
- **产品分类管理**: 后端支持但前端无管理入口
- **产品比较**: 无

#### 完成度评估: **基本可用**

### 综合评估
- **当前状态**: 产品CRUD完整，与BOM/ECN有业务规则联动 (发布需BOM，Active不可直接改)
- **与"产品化"的差距**: 需产品详情页(展示关联BOM/项目/ECN)、产品分类管理UI、产品生命周期可视化

---

## 十一、工艺路线 (Routing)

### 后端

#### API 端点
```
# 智能路由规则 (条件注册, h.Routing != nil 时才启用)
GET    /api/v1/routing-rules                           → 规则列表
POST   /api/v1/routing-rules                           → 创建规则
POST   /api/v1/routing-rules/test                      → 测试路由
GET    /api/v1/routing-rules/:id                       → 规则详情
PUT    /api/v1/routing-rules/:id                       → 更新规则
DELETE /api/v1/routing-rules/:id                       → 删除规则
GET    /api/v1/routing-logs                            → 路由日志

# 制造路线 (隶属于项目BOM)
GET    /api/v1/projects/:id/routes                     → 工艺路线列表
POST   /api/v1/projects/:id/boms/:bomId/routes         → 创建工艺路线
GET    /api/v1/projects/:id/routes/:routeId            → 工艺路线详情
PUT    /api/v1/projects/:id/routes/:routeId            → 更新工艺路线
POST   /api/v1/projects/:id/routes/:routeId/steps      → 创建工序
PUT    /api/v1/projects/:id/routes/:routeId/steps/:stepId → 更新工序
DELETE /api/v1/projects/:id/routes/:routeId/steps/:stepId → 删除工序
POST   /api/v1/projects/:id/routes/:routeId/steps/:stepId/materials → 添加工序物料
DELETE /api/v1/projects/:id/routes/:routeId/steps/:stepId/materials/:materialId → 删除工序物料
```

#### 完成度评估: **骨架**

### 前端
- 无独立工艺路线页面
- 路由规则前端展示不明

#### 完成度评估: **仅定义** (后端API存在但前端无对应页面)

### 综合评估
- **当前状态**: 后端有工艺路线和工序的完整API，但标记为条件注册(可能是实验性功能)，前端无对应页面
- **与"产品化"的差距**: 需要完整的工艺路线管理前端、工序可视化(流程图)、与BOM/产品的关联展示

---

## 十二、交付物管理 (Deliverable)

### 后端

#### Entity
- **PhaseDeliverable**: ID, PhaseID, Name, DeliverableType(document/bom/review), ResponsibleRole, IsRequired, Status(pending/submitted/approved), DocumentID, BOMID, SubmittedAt, SubmittedBy, ApprovedAt, ApprovedBy, SortOrder

#### API 端点
```
GET    /api/v1/projects/:id/deliverables               → 项目交付物列表
GET    /api/v1/projects/:id/phases/:phaseId/deliverables → 按阶段查交付物
PUT    /api/v1/projects/:id/deliverables/:deliverableId → 更新交付物
```

#### 完成度评估: **基本可用** (后端完整，与阶段/文档/BOM关联)

### 前端
- 无独立交付物管理页面
- 在任务表单/项目详情中被引用
- `api/deliverables.ts`: API定义存在

#### 完成度评估: **骨架** (API定义在前端，但无独立展示页面)

### 综合评估
- **当前状态**: 后端模型和API完整，支持按阶段管理交付物并关联文档/BOM，但前端缺独立展示
- **与"产品化"的差距**: 需要交付物清单页面、与阶段门审核的联动、完成率追踪面板

---

## 特别关注事项

### 1. BOM管理深度审计

| 审计项 | 状态 | 详情 |
|-------|------|------|
| 版本管理 | ✅ 完整 | 支持版本号、draft→published→frozen生命周期、BOM对比 |
| ECN联动 | ✅ 基本可用 | 独立BOM ECN模块(草稿/提交/审批)、独立ECN模块可关联BOM行项 |
| 导入导出 | ⚠️ API存在 | 有export/import/parse/template端点，前端集成深度待验证 |
| 成本汇总 | ✅ 完整 | 有bom-cost-summary端点和前端展示，支持单价×数量计算 |
| 多类型BOM | ✅ 完整 | EBOM/SBOM/OBOM/FWBOM + PBOM/MBOM转换端点 |
| Where-Used | ❌ 缺失 | 无反向查询(物料→引用的BOM) |
| 替代料管理 | ⚠️ 字段存在 | IsAlternative/AlternativeFor字段存在但前端管理有限 |

### 2. 项目管理深度审计

| 审计项 | 状态 | 详情 |
|-------|------|------|
| 阶段流转自动化 | ✅ 完整 | AutomationService自动检测阶段完成、依赖驱动任务启动 |
| 甘特图 | ❌ 缺失 | 无甘特图视图(仅列表) |
| 看板 | ❌ PLM缺失 | 项目管理无看板(SRM有独立看板) |
| 依赖管理 | ✅ 完整 | FS/SS/FF/SF四种依赖+LagDays |
| 模板驱动 | ✅ 完整 | 模板→项目转化含日期计算、角色分配、任务复制 |

### 3. SRM 完整业务链路审计

| 链路环节 | 状态 | 详情 |
|---------|------|------|
| 供应商管理 | ✅ 完整 | 360度画像/联系人/评分/等级 |
| 询价(RFQ) | ⚠️ 骨架 | Entity/Handler有占位，Service未实现 |
| 采购(PR→PO) | ✅ 完整 | 自动从BOM生成、按供应商分组、状态机流转 |
| 打样 | ✅ 完整 | 多轮打样、飞书审批验证、状态自动推进 |
| 检验 | ✅ 完整 | 从PO创建、逐项检验、合格品自动入库 |
| 入库 | ✅ 完整 | 自动入库+手动出入库+库存调整+流水记录 |
| 结算 | ✅ 完整 | 自动生成+双方确认+差异处理+发票管理 |
| 8D纠正 | ⚠️ 基本可用 | Entity/Handler完整，Service实现较简 |
| 评估 | ✅ 完整 | 加权评分+等级+历史+自动更新供应商评分 |

### 4. 审批与飞书集成状态

| 集成方式 | 模块 | 状态 |
|---------|------|------|
| 飞书OAuth登录 | Auth | ✅ 完整 |
| 飞书通讯录同步 | ContactSync | ✅ 完整 |
| 飞书消息通知 | Approval/Workflow/SRM | ✅ 完整 (交互式卡片) |
| 飞书任务同步 | Workflow | ✅ 完整 (创建任务+完成同步) |
| 飞书日历事件 | FeishuService | ✅ 完整 (ReviewMeeting) |
| 飞书原生审批 | SRM Sampling | ✅ 完整 (创建审批定义+提交实例+回调处理) |
| 飞书原生审批 | ECN | ⚠️ 预留 (字段存在，Submit方法有TODO) |
| 飞书原生审批 | PLM Approval | ❌ 未集成 (自建审批流+飞书通知) |

---

## 汇总表

| 模块 | 后端完成度 | 前端完成度 | 产品化差距 | 优先级建议 |
|------|-----------|-----------|-----------|-----------|
| **项目管理** | 基本可用 | 基本可用 | 缺甘特图/看板/资源视图/报表 | 🔴 高 |
| **BOM管理** | 基本可用 | 基本可用 | 缺Where-Used/替代料UI/导入验证/审批进度 | 🔴 高 |
| **物料管理** | 基本可用 | 基本可用 | 缺批量导入/变更历史/审批流/Where-Used | 🟡 中 |
| **ECN变更管理** | 基本可用 | 基本可用 | 缺飞书审批集成/ECN模板/影响分析可视化 | 🟡 中 |
| **文档管理** | 基本可用 | 基本可用 | 缺在线预览/全文搜索/审批/权限控制 | 🟡 中 |
| **审批流程** | 基本可用 | 基本可用 | 缺流程可视化/委托转交/催办/飞书深度集成 | 🟡 中 |
| **用户与权限** | 基本可用 | 骨架 | 缺权限管理UI/API权限校验/数据权限/审计日志 | 🔴 高 |
| **SRM供应商管理** | 基本可用 | 基本可用 | 缺RFQ完善/供应商门户/采购报表/库存预警 | 🟡 中 |
| **Dashboard** | 基本可用 | 基本可用 | 缺图表/KPI/自定义布局/统一PLM+SRM | 🟢 低 |
| **产品管理** | 基本可用 | 基本可用 | 缺产品详情页/关联视图/分类管理UI | 🟢 低 |
| **工艺路线** | 骨架 | 仅定义 | 需完整前端页面/工序可视化/BOM关联 | 🟢 低 |
| **交付物管理** | 基本可用 | 骨架 | 需独立展示页面/阶段门联动/完成率追踪 | 🟢 低 |

### 总体评价

**系统整体完成度: ~65-70%**

**核心优势:**
1. 后端架构清晰 (Handler→Service→Repository 三层分离)
2. 数据模型设计完整 (50+ 张表，关系合理)
3. SRM业务链路最完整 (供应商→采购→检验→入库→结算闭环)
4. 飞书集成深度较好 (OAuth/通讯录/消息/任务/日历/审批)
5. BOM管理功能丰富 (多类型/版本/ECN/导入导出/成本)

**最大短板 (产品化阻碍):**
1. **权限体系未落地**: 大部分API无细粒度权限校验，无数据权限，无权限管理UI
2. **可视化能力弱**: 无甘特图、无流程图、无图表分析、Dashboard基础
3. **导入/批量操作缺失**: 物料无批量导入，无批量审批/操作
4. **RFQ询价模块未实现**: SRM核心环节缺失
5. **前端若干模块缺独立页面**: 工艺路线、交付物、CMF/SKU管理

**建议优先处理顺序:**
1. 🔴 权限体系完善 (产品化必需)
2. 🔴 项目管理甘特图 (用户核心需求)
3. 🔴 BOM Where-Used + 导入增强
4. 🟡 RFQ询价模块完善
5. 🟡 飞书审批深度集成 (ECN/BOM审批)
6. 🟡 Dashboard图表化
7. 🟢 工艺路线前端
8. 🟢 交付物管理前端

---

*本报告基于代码文件实际内容分析生成，审计时间: 2026-02-24*
