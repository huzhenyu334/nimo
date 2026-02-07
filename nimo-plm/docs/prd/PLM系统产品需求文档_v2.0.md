# PLM系统产品需求文档 v2.0

**版本**: v2.0  
**日期**: 2026-02-05  
**项目**: nimo智能眼镜PLM系统  
**公司**: Bitfantasy 比特幻境

---

## 一、项目概述

### 1.1 项目背景

nimo是一家智能眼镜终端消费品牌，产品为带近视功能的AI智能眼镜。公司采用"平台+镜框"快速迭代模式：
- **镜腿平台**：包含MicroLED芯片、光机、波导镜片、电子部分、电池、铰链等核心组件
- **镜框迭代**：在平台基础上快速迭代不同外观的镜框
- **年产出能力**：一年可出几十甚至上百个不同SKU

### 1.2 系统定位

PLM系统是nimo的**产品数据中心**，负责管理从产品立项到量产全生命周期的所有产品数据。与ERP系统通过BOM数据进行集成，PLM输出BOM数据供ERP执行采购、生产等业务。

### 1.3 核心目标

1. **建立产品数据单一数据源**：所有产品、BOM、文档、测试数据集中管理
2. **规范研发流程**：EVT→DVT→PVT→MP四阶段门禁管理
3. **提升协作效率**：与飞书深度集成，任务/审批/通知无缝衔接
4. **支撑快速迭代**：支持平台复用、镜框快速派生、BOM版本管理

### 1.4 系统边界

**PLM系统负责：**
- 产品数据管理（PDM）
- BOM版本管理
- 项目任务管理（四阶段）
- 设计文档管理
- 工程变更管理（ECN）
- 测试数据管理

**PLM系统不负责（由ERP处理）：**
- 采购执行
- 库存管理
- 生产执行
- 销售订单
- 财务核算

---

## 二、用户角色与权限

### 2.1 角色定义

| 角色代码 | 角色名称 | 说明 | 飞书角色映射 |
|---------|---------|------|-------------|
| ADMIN | 系统管理员 | 系统配置、权限管理 | 飞书超级管理员 |
| PM | 项目经理 | 项目全局管理、阶段评审 | 项目管理部 |
| HW_LEAD | 硬件负责人 | 硬件设计审批、BOM审核 | 硬件研发部主管 |
| HW_ENG | 硬件工程师 | 硬件设计、BOM编辑 | 硬件研发部 |
| SW_LEAD | 软件负责人 | 软件设计审批 | 软件研发部主管 |
| SW_ENG | 软件工程师 | 软件开发、测试 | 软件研发部 |
| ID_ENG | 工业设计师 | 外观设计、CMF | 工业设计部 |
| ME_ENG | 结构工程师 | 结构设计 | 结构研发部 |
| OPT_ENG | 光学工程师 | 光学设计 | 光学研发部 |
| QA_ENG | 质量工程师 | 测试执行、质量分析 | 质量部 |
| PROC | 采购员 | 查看BOM、物料信息（只读） | 采购部 |
| VIEWER | 只读用户 | 查看产品信息 | 其他部门 |

### 2.2 权限矩阵

| 功能模块 | ADMIN | PM | HW_LEAD | HW_ENG | QA_ENG | PROC | VIEWER |
|---------|-------|-----|---------|--------|--------|------|--------|
| 产品-创建 | ✓ | ✓ | ✓ | - | - | - | - |
| 产品-编辑 | ✓ | ✓ | ✓ | ✓ | - | - | - |
| 产品-删除 | ✓ | - | - | - | - | - | - |
| 产品-查看 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| BOM-创建 | ✓ | ✓ | ✓ | ✓ | - | - | - |
| BOM-编辑 | ✓ | ✓ | ✓ | ✓ | - | - | - |
| BOM-审批 | ✓ | ✓ | ✓ | - | - | - | - |
| BOM-发布 | ✓ | ✓ | ✓ | - | - | - | - |
| BOM-查看 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| 项目-创建 | ✓ | ✓ | - | - | - | - | - |
| 项目-管理 | ✓ | ✓ | - | - | - | - | - |
| 任务-分配 | ✓ | ✓ | ✓ | - | - | - | - |
| 任务-执行 | ✓ | ✓ | ✓ | ✓ | ✓ | - | - |
| ECN-发起 | ✓ | ✓ | ✓ | ✓ | ✓ | - | - |
| ECN-审批 | ✓ | ✓ | ✓ | - | - | - | - |
| 文档-上传 | ✓ | ✓ | ✓ | ✓ | ✓ | - | - |
| 文档-下载 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| 系统-配置 | ✓ | - | - | - | - | - | - |

### 2.3 数据权限规则

1. **部门隔离**：默认用户只能查看本部门参与的项目
2. **项目成员**：被分配到项目的成员可查看该项目所有数据
3. **跨部门查看**：需项目经理授权
4. **敏感数据**：成本信息仅PM、ADMIN、采购可见

---

## 三、功能模块详细设计

### 3.1 产品数据管理（PDM）

#### 3.1.1 产品层级结构

```
产品系列（Series）
  └── 产品平台（Platform）- 镜腿平台
        └── 产品型号（Product/SKU）- 具体眼镜
              └── 产品版本（Revision）
```

**示例：**
```
nimo系列
  └── nimo G1平台（镜腿平台）
        ├── nimo G1-A001（圆框眼镜）
        │     ├── v1.0（EVT）
        │     ├── v1.1（DVT）
        │     └── v2.0（MP）
        └── nimo G1-A002（方框眼镜）
              └── v1.0
```

#### 3.1.2 产品基本属性

| 字段名 | 字段代码 | 类型 | 必填 | 说明 |
|-------|---------|------|-----|------|
| 产品编码 | sku | String(50) | 是 | 唯一标识，规则：品牌-平台-序号 |
| 产品名称 | name | String(200) | 是 | 产品中文名称 |
| 英文名称 | name_en | String(200) | 否 | 产品英文名称 |
| 产品系列 | series_id | UUID | 是 | 关联系列 |
| 产品平台 | platform_id | UUID | 否 | 关联平台 |
| 产品类型 | product_type | Enum | 是 | PLATFORM/SKU |
| 产品状态 | status | Enum | 是 | CONCEPT/EVT/DVT/PVT/MP/EOL |
| 当前版本 | current_version | String(20) | 是 | 当前生效版本号 |
| 上市日期 | launch_date | Date | 否 | 计划上市日期 |
| 停产日期 | eol_date | Date | 否 | 停产日期 |
| 目标成本 | target_cost | Decimal | 否 | 目标BOM成本（元） |
| 目标售价 | target_price | Decimal | 否 | 目标零售价（元） |
| 创建人 | created_by | String(64) | 是 | 飞书用户ID |
| 创建时间 | created_at | Timestamp | 是 | - |
| 更新时间 | updated_at | Timestamp | 是 | - |

#### 3.1.3 智能眼镜扩展属性

| 字段名 | 字段代码 | 类型 | 说明 |
|-------|---------|------|------|
| 镜框材质 | frame_material | Enum | TR90/钛合金/板材/混合 |
| 镜框颜色 | frame_color | String | 颜色描述 |
| 镜框尺寸 | frame_size | JSON | {width, height, bridge, temple} |
| 度数范围 | diopter_range | String | 支持的近视度数范围 |
| 显示类型 | display_type | Enum | 单色/彩色 |
| 分辨率 | resolution | String | 显示分辨率 |
| 视场角 | fov | Decimal | FOV度数 |
| 电池容量 | battery_mah | Integer | 电池容量mAh |
| 蓝牙版本 | bt_version | String | 蓝牙版本 |
| 重量 | weight_g | Decimal | 整机重量(g) |
| 防水等级 | ip_rating | String | 防水防尘等级 |

#### 3.1.4 产品状态流转

```
CONCEPT --立项评审通过--> EVT --EVT评审通过--> DVT --DVT评审通过--> PVT --PVT评审通过--> MP
                                                                                        |
                                                                                        v
                                                                                      EOL
```

| 当前状态 | 目标状态 | 触发条件 | 审批流程 |
|---------|---------|---------|---------|
| CONCEPT | EVT | 立项评审通过 | PM审批 |
| EVT | DVT | EVT评审通过 | PM + 技术负责人审批 |
| DVT | PVT | DVT评审通过 + 设计冻结 | PM + 技术负责人 + 质量负责人审批 |
| PVT | MP | PVT评审通过 + 量产准备完成 | PM + 生产负责人 + 质量负责人审批 |
| MP | EOL | 停产决策 | PM审批 |
| 任意状态 | CONCEPT | 项目终止 | ADMIN审批 |

#### 3.1.5 功能用例

**UC-PDM-001：创建产品**

- **参与者**：PM、HW_LEAD
- **前置条件**：用户已登录且有创建权限
- **主流程**：
  1. 用户点击"新建产品"
  2. 选择产品类型（平台/SKU）
  3. 如选择SKU，选择所属平台
  4. 填写产品基本信息
  5. 填写产品扩展属性
  6. 系统校验产品编码唯一性
  7. 系统创建产品，状态为CONCEPT
  8. 系统触发飞书通知相关人员

**UC-PDM-002：从平台派生产品**

- **参与者**：PM、HW_LEAD、HW_ENG
- **前置条件**：存在可用平台产品
- **主流程**：
  1. 用户选择平台产品
  2. 点击"派生新SKU"
  3. 系统自动复制平台的设计BOM
  4. 用户修改镜框相关物料
  5. 填写新SKU的差异属性
  6. 保存，创建新SKU产品
- **业务规则**：派生产品自动继承平台的镜腿BOM

---

### 3.2 BOM管理模块

#### 3.2.1 BOM类型定义

| BOM类型 | 代码 | 说明 | 使用阶段 |
|--------|------|------|---------|
| 设计BOM | EBOM | 工程设计阶段的BOM，包含所有设计物料 | EVT/DVT |
| 制造BOM | MBOM | 生产制造使用的BOM，包含实际采购物料 | PVT/MP |
| 成本BOM | CBOM | 财务成本核算用BOM，包含成本信息 | 全阶段 |
| 服务BOM | SBOM | 售后维修用BOM，包含可替换件 | MP/EOL |

#### 3.2.2 智能眼镜BOM结构

```
整机（Level 0）
├── 左镜腿组件（Level 1）
│   ├── 主板PCBA（Level 2）
│   │   ├── 主控芯片（Level 3）
│   │   ├── 蓝牙模块（Level 3）
│   │   ├── 电源管理IC（Level 3）
│   │   └── ...其他电子元器件
│   ├── 电池组件（Level 2）
│   ├── 光机模组（Level 2）
│   │   ├── MicroLED芯片（Level 3）
│   │   └── 光学镜片（Level 3）
│   ├── 镜腿结构件（Level 2）
│   └── 铰链组件（Level 2）
├── 右镜腿组件（Level 1）
│   ├── 副板PCBA（Level 2）
│   ├── 麦克风模组（Level 2）
│   ├── 电池组件（Level 2）
│   ├── 镜腿结构件（Level 2）
│   └── 铰链组件（Level 2）
├── 镜框组件（Level 1）
│   ├── 镜框主体（Level 2）
│   ├── 波导镜片（Level 2）
│   ├── 鼻托组件（Level 2）
│   └── 镜腿连接件（Level 2）
└── 包装组件（Level 1）
    ├── 彩盒（Level 2）
    ├── 说明书（Level 2）
    ├── 充电线（Level 2）
    ├── 眼镜盒（Level 2）
    └── 擦镜布（Level 2）
```

#### 3.2.3 BOM状态流转

```
DRAFT --提交审批--> REVIEWING --审批通过--> APPROVED --发布--> RELEASED --> OBSOLETE
                       |
                       v
                   审批驳回 --> DRAFT
```

**状态说明：**
- DRAFT：草稿，可自由编辑
- REVIEWING：审批中，不可编辑
- APPROVED：已审批，待发布
- RELEASED：已发布，不可修改，如需变更需创建新版本
- OBSOLETE：已废弃

#### 3.2.4 核心功能用例

**UC-BOM-001：创建BOM**

1. 用户选择产品
2. 点击"创建BOM"，选择BOM类型
3. 系统生成BOM版本号（自动递增）
4. 用户添加BOM物料明细
   - 选择已有物料或创建新物料
   - 设置用量、单位、位号
   - 设置父级物料（构建层级）
5. 系统自动计算BOM成本
6. 保存为DRAFT状态

**UC-BOM-002：BOM审批流程**

1. 工程师提交BOM审批
2. 系统触发飞书审批流程
3. HW_LEAD审核BOM结构、物料选型
4. PM审核BOM成本
5. 全部通过后，BOM状态变为APPROVED
6. 系统通知相关人员

**UC-BOM-003：BOM版本比较**

1. 用户选择产品
2. 选择两个BOM版本
3. 系统对比显示：
   - 新增物料（绿色标记）
   - 删除物料（红色标记）
   - 变更物料（黄色标记）
   - 成本变化汇总

**UC-BOM-004：从平台BOM派生**

1. 选择平台产品的已发布BOM
2. 点击"派生到SKU"
3. 选择目标SKU产品
4. 系统复制BOM结构
5. 用户修改镜框相关物料
6. 保存为新SKU的DRAFT BOM

---

### 3.3 项目任务管理模块

#### 3.3.1 四阶段任务模板

**EVT阶段（4-6周）- 工程验证测试**

| 任务编码 | 任务名称 | 负责部门 | 工期 | 前置任务 |
|---------|---------|---------|------|---------|
| EVT-001 | 硬件设计完成 | 硬件研发 | 14天 | - |
| EVT-001-01 | 电路原理图设计 | 硬件研发 | 7天 | - |
| EVT-001-02 | PCB布局设计 | 硬件研发 | 5天 | EVT-001-01 |
| EVT-001-03 | 结构3D设计 | 结构研发 | 10天 | - |
| EVT-001-04 | 光学设计 | 光学研发 | 8天 | - |
| EVT-001-05 | 散热设计验证 | 硬件研发 | 3天 | EVT-001-02, EVT-001-03 |
| EVT-002 | EVT样机制作 | 样机工程 | 10天 | EVT-001 |
| EVT-002-01 | PCB打样 | 样机工程 | 5天 | EVT-001-02 |
| EVT-002-02 | 结构件CNC加工 | 样机工程 | 7天 | EVT-001-03 |
| EVT-002-03 | 光学件采购 | 采购部 | 10天 | EVT-001-04 |
| EVT-002-04 | 样机组装 | 样机工程 | 3天 | EVT-002-01, EVT-002-02, EVT-002-03 |
| EVT-002-05 | 基本功能测试 | 测试部 | 2天 | EVT-002-04 |
| EVT-003 | EVT测试验证 | 测试部 | 7天 | EVT-002 |
| EVT-004 | EVT评审 | 项目管理 | 2天 | EVT-003 |

**DVT阶段（8-12周）- 设计验证测试**

主要任务：设计优化、DVT样机制作（50-100台）、全面测试验证、认证准备、DVT评审

**PVT阶段（6-8周）- 生产验证测试**

主要任务：生产准备、PVT试产（500-1000台）、量产验证、市场准备、PVT评审

**MP阶段（持续）- 批量生产**

主要任务：量产启动、持续改进、市场反馈

#### 3.3.2 甘特图功能

**显示要素：**
- 任务条：显示任务时间跨度
- 依赖线：显示任务依赖关系
- 里程碑：菱形标记
- 进度条：任务完成百分比
- 关键路径：红色高亮
- 今日线：当前日期标记
- 资源负载：人员工作负荷

**交互功能：**
- 拖拽调整任务时间
- 拖拽调整依赖关系
- 点击展开/折叠子任务
- 右键菜单操作
- 缩放时间刻度（日/周/月）
- 筛选（按阶段/负责人/状态）

#### 3.3.3 任务依赖类型

| 类型 | 代码 | 说明 |
|-----|------|------|
| 完成-开始 | FS | 前置任务完成后，后续任务才能开始 |
| 开始-开始 | SS | 前置任务开始后，后续任务可以开始 |
| 完成-完成 | FF | 前置任务完成后，后续任务才能完成 |
| 开始-完成 | SF | 前置任务开始后，后续任务才能完成 |

---

### 3.4 文档管理模块

#### 3.4.1 文档分类

| 分类代码 | 分类名称 | 说明 |
|---------|---------|------|
| DESIGN | 设计文档 | 原理图、PCB、结构图、光学设计 |
| SPEC | 规格书 | 产品规格书、物料规格书 |
| TEST | 测试文档 | 测试用例、测试报告 |
| QUALITY | 质量文档 | 检验标准、质量报告 |
| CERT | 认证文档 | 3C、FCC、CE等认证资料 |
| MFG | 生产文档 | 作业指导书、工艺文件 |
| USER | 用户文档 | 说明书、快速指南 |
| OTHER | 其他 | 会议纪要、培训资料等 |

#### 3.4.2 功能说明

1. **文档上传**：支持多文件上传，存储到MinIO
2. **版本管理**：同一文档支持多版本，保留历史
3. **文档审批**：重要文档通过飞书审批流程
4. **关联管理**：文档可关联到产品、项目、任务

---

### 3.5 工程变更管理（ECN）

#### 3.5.1 ECN类型

| 类型 | 说明 | 审批级别 |
|-----|------|---------|
| MINOR | 微小变更：不影响外观、性能、成本 | 部门负责人 |
| NORMAL | 普通变更：影响BOM但不影响外观 | 部门负责人 + PM |
| MAJOR | 重大变更：影响外观、性能或成本>5% | PM + 技术总监 |
| CRITICAL | 关键变更：涉及安全或认证 | PM + 技术总监 + CEO |

#### 3.5.2 ECN流程

```
DRAFT --提交--> SUBMITTED --分配评审人--> REVIEWING
                                          |
                    +---------------------+---------------------+
                    |                                           |
                    v                                           v
               REJECTED                                    APPROVED
                                                              |
                                                              v
                                                        IMPLEMENTED
                                                              |
                                                              v
                                                          CLOSED
```

---

### 3.6 测试管理模块

#### 3.6.1 测试类型

| 测试类型 | 代码 | 说明 | 执行阶段 |
|---------|-----|------|---------|
| 功能测试 | FUNC | 基本功能验证 | EVT起 |
| 电气测试 | ELEC | 电气性能测试 | EVT起 |
| 光学测试 | OPT | 光学性能测试 | EVT起 |
| 结构测试 | MECH | 结构强度测试 | EVT起 |
| 环境测试 | ENV | 温湿度、振动 | DVT起 |
| 可靠性测试 | REL | 寿命、老化 | DVT起 |
| 安规测试 | SAFETY | 安全认证测试 | DVT起 |
| 用户体验测试 | UX | 佩戴舒适度等 | DVT起 |

#### 3.6.2 功能说明

1. **测试计划管理**：制定测试计划，关联项目和产品
2. **测试用例库**：维护可复用的测试用例
3. **测试执行记录**：记录测试结果，支持PASS/FAIL/BLOCKED
4. **缺陷跟踪**：记录缺陷，关联ECN

---

## 四、数据模型设计

### 4.1 产品表 products

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| sku | VARCHAR(50) | 是 | 产品编码（唯一） |
| name | VARCHAR(200) | 是 | 产品名称 |
| name_en | VARCHAR(200) | 否 | 英文名称 |
| series_id | UUID | 是 | 关联系列 |
| platform_id | UUID | 否 | 关联平台 |
| product_type | ENUM | 是 | PLATFORM/SKU |
| status | ENUM | 是 | 产品状态 |
| current_version | VARCHAR(20) | 是 | 当前版本 |
| target_cost | DECIMAL(12,2) | 否 | 目标成本 |
| target_price | DECIMAL(12,2) | 否 | 目标售价 |
| attributes | JSONB | 否 | 扩展属性 |
| created_by | VARCHAR(64) | 是 | 创建人（飞书ID） |
| created_at | TIMESTAMP | 是 | 创建时间 |
| updated_at | TIMESTAMP | 是 | 更新时间 |

### 4.2 BOM版本表 bom_versions

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| product_id | UUID | 是 | 关联产品 |
| version | VARCHAR(20) | 是 | 版本号 |
| bom_type | ENUM | 是 | EBOM/MBOM/CBOM/SBOM |
| status | ENUM | 是 | BOM状态 |
| effective_date | DATE | 否 | 生效日期 |
| expiry_date | DATE | 否 | 失效日期 |
| total_cost | DECIMAL(12,2) | 否 | BOM总成本 |
| description | TEXT | 否 | 版本说明 |
| created_by | VARCHAR(64) | 是 | 创建人 |
| approved_by | VARCHAR(64) | 否 | 审批人 |
| approved_at | TIMESTAMP | 否 | 审批时间 |
| created_at | TIMESTAMP | 是 | 创建时间 |
| updated_at | TIMESTAMP | 是 | 更新时间 |

### 4.3 BOM物料明细表 bom_items

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| bom_id | UUID | 是 | 关联BOM版本 |
| parent_item_id | UUID | 否 | 父级物料（空为顶层） |
| level | INTEGER | 是 | BOM层级（0-n） |
| seq_no | INTEGER | 是 | 序号 |
| material_id | UUID | 是 | 关联物料主数据 |
| quantity | DECIMAL(12,4) | 是 | 用量 |
| unit | VARCHAR(20) | 是 | 单位 |
| unit_price | DECIMAL(12,4) | 否 | 单价 |
| amount | DECIMAL(12,2) | 否 | 金额 |
| reference | VARCHAR(100) | 否 | 位号/参考 |
| remark | TEXT | 否 | 备注 |
| is_key_part | BOOLEAN | 是 | 是否关键物料 |
| is_customized | BOOLEAN | 是 | 是否定制件 |

### 4.4 物料主数据表 materials

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| material_code | VARCHAR(50) | 是 | 物料编码（唯一） |
| name | VARCHAR(200) | 是 | 物料名称 |
| name_en | VARCHAR(200) | 否 | 英文名称 |
| category | ENUM | 是 | 电子/光学/结构/包装/辅料 |
| sub_category | VARCHAR(50) | 否 | 子分类 |
| specification | TEXT | 否 | 规格描述 |
| unit | VARCHAR(20) | 是 | 计量单位 |
| standard_cost | DECIMAL(12,4) | 否 | 标准成本 |
| lead_time_days | INTEGER | 否 | 标准采购周期 |
| moq | INTEGER | 否 | 最小起订量 |
| mpq | INTEGER | 否 | 最小包装量 |
| status | ENUM | 是 | ACTIVE/INACTIVE/OBSOLETE |
| created_at | TIMESTAMP | 是 | - |
| updated_at | TIMESTAMP | 是 | - |

### 4.5 项目表 projects

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| project_code | VARCHAR(50) | 是 | 项目编码 |
| name | VARCHAR(200) | 是 | 项目名称 |
| product_id | UUID | 是 | 关联产品 |
| current_phase | ENUM | 是 | 当前阶段 |
| status | ENUM | 是 | 项目状态 |
| pm_user_id | VARCHAR(64) | 是 | 项目经理（飞书ID） |
| planned_start_date | DATE | 是 | 计划开始 |
| planned_end_date | DATE | 是 | 计划结束 |
| actual_start_date | DATE | 否 | 实际开始 |
| actual_end_date | DATE | 否 | 实际结束 |
| progress | INTEGER | 是 | 整体进度(0-100) |
| description | TEXT | 否 | 项目描述 |
| created_at | TIMESTAMP | 是 | - |
| updated_at | TIMESTAMP | 是 | - |

### 4.6 任务表 tasks

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| project_id | UUID | 是 | 关联项目 |
| task_code | VARCHAR(50) | 是 | 任务编码 |
| name | VARCHAR(200) | 是 | 任务名称 |
| description | TEXT | 否 | 任务描述 |
| phase | ENUM | 是 | EVT/DVT/PVT/MP |
| parent_id | UUID | 否 | 父任务ID |
| seq_no | INTEGER | 是 | 序号 |
| assignee_id | VARCHAR(64) | 否 | 负责人（飞书ID） |
| assignee_dept | VARCHAR(100) | 否 | 负责部门 |
| status | ENUM | 是 | 任务状态 |
| priority | ENUM | 是 | LOW/MEDIUM/HIGH/URGENT |
| progress | INTEGER | 是 | 进度(0-100) |
| planned_start | DATE | 是 | 计划开始 |
| planned_end | DATE | 是 | 计划结束 |
| actual_start | DATE | 否 | 实际开始 |
| actual_end | DATE | 否 | 实际结束 |
| duration_days | INTEGER | 是 | 计划工期 |
| is_milestone | BOOLEAN | 是 | 是否里程碑 |
| feishu_task_id | VARCHAR(100) | 否 | 飞书任务ID |
| created_at | TIMESTAMP | 是 | - |
| updated_at | TIMESTAMP | 是 | - |

### 4.7 任务依赖关系表 task_dependencies

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | UUID | 主键 |
| task_id | UUID | 当前任务ID |
| depends_on_task_id | UUID | 依赖任务ID |
| dependency_type | ENUM | FS/SS/FF/SF |
| lag_days | INTEGER | 延迟天数 |

### 4.8 文档表 documents

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| doc_code | VARCHAR(50) | 是 | 文档编码 |
| title | VARCHAR(200) | 是 | 文档标题 |
| category | ENUM | 是 | 文档分类 |
| product_id | UUID | 否 | 关联产品 |
| project_id | UUID | 否 | 关联项目 |
| current_version | INTEGER | 是 | 当前版本号 |
| status | ENUM | 是 | 文档状态 |
| confidential_level | ENUM | 是 | 机密级别 |
| created_by | VARCHAR(64) | 是 | 创建人 |
| created_at | TIMESTAMP | 是 | - |
| updated_at | TIMESTAMP | 是 | - |

### 4.9 ECN表 engineering_changes

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| ecn_code | VARCHAR(50) | 是 | ECN编号 |
| title | VARCHAR(200) | 是 | 变更标题 |
| type | ENUM | 是 | MINOR/NORMAL/MAJOR/CRITICAL |
| product_id | UUID | 是 | 关联产品 |
| status | ENUM | 是 | ECN状态 |
| reason | TEXT | 是 | 变更原因 |
| description | TEXT | 是 | 变更描述 |
| impact_analysis | TEXT | 是 | 影响分析 |
| cost_impact | DECIMAL(12,2) | 否 | 成本影响 |
| schedule_impact | INTEGER | 否 | 进度影响（天） |
| requested_by | VARCHAR(64) | 是 | 发起人 |
| effective_date | DATE | 否 | 生效日期 |
| created_at | TIMESTAMP | 是 | - |
| updated_at | TIMESTAMP | 是 | - |

---

## 五、API接口设计

### 5.1 接口规范

- **协议**：HTTPS
- **格式**：JSON
- **认证**：Bearer Token（JWT）
- **版本**：URL路径 /api/v1/
- **编码**：UTF-8

### 5.2 响应格式

**成功响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "timestamp": 1706025600000
}
```

**错误响应：**
```json
{
  "code": 10001,
  "message": "Product SKU already exists",
  "data": null,
  "timestamp": 1706025600000
}
```

### 5.3 产品接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/products | 创建产品 |
| GET | /api/v1/products | 查询产品列表 |
| GET | /api/v1/products/{id} | 查询产品详情 |
| PUT | /api/v1/products/{id} | 更新产品 |
| DELETE | /api/v1/products/{id} | 删除产品 |
| POST | /api/v1/products/{id}/transition | 产品状态变更 |
| POST | /api/v1/products/{id}/derive | 从平台派生SKU |

### 5.4 BOM接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/products/{product_id}/boms | 创建BOM |
| GET | /api/v1/products/{product_id}/boms | 查询产品的BOM列表 |
| GET | /api/v1/boms/{id} | 查询BOM详情 |
| PUT | /api/v1/boms/{id} | 更新BOM |
| DELETE | /api/v1/boms/{id} | 删除BOM |
| POST | /api/v1/boms/{id}/submit-approval | 提交BOM审批 |
| POST | /api/v1/boms/{id}/approve | 审批通过 |
| POST | /api/v1/boms/{id}/reject | 审批驳回 |
| POST | /api/v1/boms/{id}/release | 发布BOM |
| GET | /api/v1/boms/compare | BOM版本比较 |
| POST | /api/v1/boms/{id}/derive | 从平台BOM派生 |

### 5.5 项目任务接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/projects | 创建项目 |
| GET | /api/v1/projects | 查询项目列表 |
| GET | /api/v1/projects/{id} | 查询项目详情 |
| PUT | /api/v1/projects/{id} | 更新项目 |
| GET | /api/v1/projects/{id}/tasks | 查询任务列表（甘特图数据） |
| POST | /api/v1/tasks | 创建任务 |
| PUT | /api/v1/tasks/{id} | 更新任务 |
| PATCH | /api/v1/tasks/{id}/progress | 更新任务进度 |
| POST | /api/v1/tasks/{id}/sync-to-feishu | 同步任务到飞书 |
| POST | /api/v1/projects/{id}/phase-review | 发起阶段评审 |

### 5.6 文档接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/documents | 上传文档 |
| GET | /api/v1/documents | 查询文档列表 |
| GET | /api/v1/documents/{id} | 查询文档详情 |
| GET | /api/v1/documents/{id}/download | 下载文档 |
| POST | /api/v1/documents/{id}/versions | 上传新版本 |

### 5.7 ECN接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/ecns | 创建ECN |
| GET | /api/v1/ecns | 查询ECN列表 |
| GET | /api/v1/ecns/{id} | 查询ECN详情 |
| PUT | /api/v1/ecns/{id} | 更新ECN |
| POST | /api/v1/ecns/{id}/submit | 提交ECN |
| POST | /api/v1/ecns/{id}/approve | 审批ECN |
| POST | /api/v1/ecns/{id}/implement | 实施ECN |
| POST | /api/v1/ecns/{id}/close | 关闭ECN |

### 5.8 飞书同步接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/feishu/sync/organization | 手动同步组织架构 |
| POST | /api/v1/feishu/webhook | 飞书Webhook回调 |

### 5.9 错误码定义

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 10001 | 参数校验失败 |
| 10002 | 资源不存在 |
| 10003 | 资源已存在 |
| 10004 | 操作不允许 |
| 20001 | 未认证 |
| 20002 | 权限不足 |
| 20003 | Token过期 |
| 30001 | 飞书API错误 |
| 50001 | 系统内部错误 |

---

## 六、飞书集成设计

### 6.1 单点登录（SSO）

**登录流程：**
1. 用户访问PLM系统
2. 系统检测无有效Session，重定向到飞书OAuth授权页
3. 用户在飞书登录（扫码/账号密码）
4. 飞书回调PLM，携带authorization_code
5. PLM用code换取access_token
6. PLM用access_token获取用户信息
7. PLM创建本地Session，颁发JWT
8. 用户进入系统

**飞书应用权限：**
- contact:user.base:readonly（用户基本信息）
- contact:user.employee_id:readonly（工号）
- contact:department:readonly（部门信息）
- task:task（任务管理）
- approval:approval（审批流程）

### 6.2 组织架构同步

**同步策略：**
- 全量同步：每天凌晨2点
- 增量同步：监听飞书事件（员工入职/离职/调岗）
- 缓存策略：用户信息缓存4小时

### 6.3 任务同步

**同步规则：**
- 创建项目时自动在飞书创建任务
- 任务分配时同步负责人到飞书任务
- 进度更新双向同步（飞书↔PLM）
- 任务完成状态双向同步

### 6.4 审批集成

| 审批场景 | 审批定义Code | 审批人规则 |
|---------|-------------|-----------|
| BOM审批 | bom_approval | 部门负责人 → PM |
| ECN审批 | ecn_approval | 根据ECN类型动态确定 |
| 阶段评审 | phase_review | PM + 技术负责人 |
| 文档发布 | doc_release | 部门负责人 |

### 6.5 消息通知

| 事件 | 通知对象 | 通知方式 |
|-----|---------|---------|
| 任务分配 | 任务负责人 | 飞书消息 |
| 任务到期提醒 | 任务负责人 | 飞书消息（提前1天） |
| 任务延期 | PM + 负责人 | 飞书消息 |
| 审批待办 | 审批人 | 飞书审批通知 |
| 审批完成 | 发起人 | 飞书消息 |
| 阶段转换 | 项目成员 | 飞书群消息 |

---

## 七、技术架构

### 7.1 技术选型

| 层级 | 技术 | 说明 |
|-----|------|------|
| 前端 | React 18 + TypeScript + Ant Design 5 | 企业级UI |
| 构建 | Vite 5 | 快速构建 |
| 状态管理 | Zustand + React Query | 轻量级 |
| 甘特图 | Frappe Gantt / DHTMLX Gantt | 专业甘特图组件 |
| 后端 | Go 1.21 + Gin | 高性能 |
| 集成服务 | Python 3.11 + FastAPI | 飞书SDK支持好 |
| 数据库 | PostgreSQL 16 | 主数据库 |
| 文档存储 | MongoDB 7 | BOM版本、文档 |
| 缓存 | Redis 7 | Session、缓存 |
| 文件存储 | MinIO | S3兼容对象存储 |
| 消息队列 | RabbitMQ | 异步任务 |
| 容器 | Docker + Kubernetes | 容器编排 |
| CI/CD | GitLab CI | 自动化部署 |

### 7.2 微服务划分

| 服务名 | 职责 | 技术栈 | 端口 |
|-------|------|-------|------|
| user-service | 用户认证、权限管理 | Go + Gin | 8001 |
| product-service | 产品数据管理 | Go + Gin | 8002 |
| bom-service | BOM管理 | Go + Gin | 8003 |
| project-service | 项目任务管理 | Go + Gin | 8004 |
| doc-service | 文档管理 | Go + Gin | 8005 |
| ecn-service | 变更管理 | Go + Gin | 8006 |
| test-service | 测试管理 | Go + Gin | 8007 |
| integration-service | 飞书集成 | Python + FastAPI | 8008 |

### 7.3 部署架构

**环境规划：**

| 环境 | 用途 | 配置 |
|-----|------|------|
| dev | 开发环境 | 单节点Docker Compose |
| test | 测试环境 | 单节点 |
| staging | 预发布环境 | 与生产同配置 |
| prod | 生产环境 | Kubernetes集群 |

**生产环境配置：**
- 微服务：每个服务2副本
- PostgreSQL：主从复制
- MongoDB：3节点副本集
- Redis：哨兵模式

---

## 八、非功能性需求

### 8.1 性能要求

| 指标 | 要求 |
|-----|------|
| 页面加载时间 | < 2秒（首屏） |
| API响应时间 | < 500ms（P95） |
| BOM查询响应 | < 1秒（1000级物料） |
| 甘特图渲染 | < 2秒（500任务） |
| 并发用户 | 支持100并发 |
| 系统可用性 | 99.9% |

### 8.2 安全要求

| 类别 | 要求 |
|-----|------|
| 认证 | 飞书OAuth 2.0 + JWT |
| 授权 | RBAC权限控制 |
| 传输 | 全站HTTPS（TLS 1.3） |
| 存储 | 敏感数据AES加密 |
| 审计 | 全量操作日志 |
| 备份 | 每日自动备份，保留30天 |

### 8.3 兼容性要求

| 类别 | 要求 |
|-----|------|
| 浏览器 | Chrome 90+, Edge 90+, Safari 14+ |
| 分辨率 | 1920×1080（最佳），最低1366×768 |
| 飞书版本 | 飞书6.0+ |

---

## 九、实施计划

### 9.1 里程碑

| 阶段 | 时间 | 交付物 |
|-----|------|-------|
| 第一阶段 | 第1-4周 | 用户体系 + 产品管理基础功能 |
| 第二阶段 | 第5-8周 | BOM管理 + 飞书集成 |
| 第三阶段 | 第9-12周 | 项目任务管理 + 甘特图 |
| 第四阶段 | 第13-16周 | 文档管理 + ECN + 测试管理 |
| 第五阶段 | 第17-20周 | 系统集成 + 性能优化 + 上线 |

### 9.2 团队配置

| 角色 | 人数 | 职责 |
|-----|------|------|
| 技术负责人 | 1 | 架构设计、技术决策 |
| 后端开发（Go） | 3 | Go微服务开发 |
| 前端开发 | 2 | React开发 |
| 测试工程师 | 1 | 测试用例、自动化测试 |
| DevOps | 1 | CI/CD、运维 |

---

## 附录

### 附录A：产品状态枚举

| 状态 | 代码 | 说明 |
|-----|------|------|
| 概念阶段 | CONCEPT | 产品立项前 |
| 工程验证 | EVT | Engineering Validation Test |
| 设计验证 | DVT | Design Validation Test |
| 生产验证 | PVT | Production Validation Test |
| 量产 | MP | Mass Production |
| 停产 | EOL | End of Life |

### 附录B：BOM状态枚举

| 状态 | 代码 | 说明 |
|-----|------|------|
| 草稿 | DRAFT | 可编辑 |
| 审批中 | REVIEWING | 不可编辑 |
| 已审批 | APPROVED | 待发布 |
| 已发布 | RELEASED | 生效中 |
| 已废弃 | OBSOLETE | 不再使用 |

### 附录C：任务状态枚举

| 状态 | 代码 | 说明 |
|-----|------|------|
| 未开始 | NOT_STARTED | 任务未开始 |
| 进行中 | IN_PROGRESS | 任务执行中 |
| 已完成 | COMPLETED | 任务完成 |
| 已阻塞 | BLOCKED | 任务被阻塞 |
| 已取消 | CANCELLED | 任务取消 |

### 附录D：ECN状态枚举

| 状态 | 代码 | 说明 |
|-----|------|------|
| 草稿 | DRAFT | 新建ECN |
| 已提交 | SUBMITTED | 已提交评审 |
| 评审中 | REVIEWING | 评审进行中 |
| 已批准 | APPROVED | 评审通过 |
| 已实施 | IMPLEMENTED | 变更已实施 |
| 已关闭 | CLOSED | ECN关闭 |
| 已拒绝 | REJECTED | 评审拒绝 |

---

**文档版本历史：**

| 版本 | 日期 | 作者 | 变更说明 |
|-----|------|------|---------|
| v1.0 | 2026-01-30 | 企业效能专家团队 | 初稿 |
| v2.0 | 2026-02-05 | Claude Opus | 重构，完善到可编码级别 |

---

**文档状态**：完成  
**下次评审日期**：2026-02-06
