# ERP系统产品需求文档 v2.0

**版本**: v2.0  
**日期**: 2026-02-05  
**项目**: nimo智能眼镜ERP系统  
**公司**: Bitfantasy 比特幻境

---

## 一、项目概述

### 1.1 背景

nimo智能眼镜需要一套ERP系统管理从采购到售后的全业务流程，实现供应链透明化和成本精准控制。

### 1.2 系统定位

ERP系统是nimo的**业务运营中心**，负责管理采购、库存、生产、销售、售后、财务等核心业务流程。从PLM系统获取BOM数据，执行实际业务操作。

### 1.3 核心目标

1. **端到端流程打通**：采购→制造→销售→售后
2. **供应链透明化**：实时掌握供应链状态
3. **成本精准控制**：物料、制造、物流、售后成本
4. **快速响应市场**：支持多SKU、多渠道

### 1.4 系统边界

**ERP系统负责：**
- 供应商管理
- 采购管理
- 库存管理
- 生产管理
- 销售管理
- 售后管理
- 财务管理

**PLM系统负责（ERP不处理）：**
- 产品数据管理
- BOM设计与版本管理
- 研发项目管理
- 设计文档管理
- 工程变更管理

**数据流向：**
```
PLM系统 ---(BOM/物料/产品数据)---> ERP系统
ERP系统 ---(成本/库存反馈)---> PLM系统
```

---

## 二、用户角色与权限

### 2.1 角色定义

| 角色代码 | 角色名称 | 说明 | 飞书角色映射 |
|---------|---------|------|-------------|
| ADMIN | 系统管理员 | 系统配置、权限管理 | 飞书超级管理员 |
| PROC_LEAD | 采购负责人 | 采购审批、供应商管理 | 采购部主管 |
| PROC | 采购员 | 日常采购操作 | 采购部 |
| WH_LEAD | 仓库负责人 | 库存管理、盘点审批 | 仓库主管 |
| WH | 仓库员 | 出入库操作 | 仓库 |
| MFG_LEAD | 生产负责人 | 生产计划、工单审批 | 生产部主管 |
| MFG | 生产员 | 生产执行、报工 | 生产部 |
| SALES_LEAD | 销售负责人 | 订单审核、渠道管理 | 销售部主管 |
| SALES | 销售员 | 订单处理 | 销售部 |
| SERVICE | 售后人员 | 售后服务 | 售后部 |
| FIN_LEAD | 财务负责人 | 财务审批、报表 | 财务部主管 |
| FIN | 财务员 | 日常财务操作 | 财务部 |

### 2.2 权限矩阵

| 功能模块 | ADMIN | PROC_LEAD | PROC | WH_LEAD | WH | MFG_LEAD | MFG | SALES_LEAD | SALES | FIN_LEAD |
|---------|-------|-----------|------|---------|-----|----------|-----|------------|-------|----------|
| 供应商-管理 | ✓ | ✓ | ✓ | - | - | - | - | - | - | - |
| 采购订单-创建 | ✓ | ✓ | ✓ | - | - | - | - | - | - | - |
| 采购订单-审批 | ✓ | ✓ | - | - | - | - | - | - | - | ✓ |
| 库存-入库 | ✓ | - | - | ✓ | ✓ | - | - | - | - | - |
| 库存-出库 | ✓ | - | - | ✓ | ✓ | - | - | - | - | - |
| 库存-盘点 | ✓ | - | - | ✓ | - | - | - | - | - | - |
| 工单-创建 | ✓ | - | - | - | - | ✓ | - | - | - | - |
| 工单-执行 | ✓ | - | - | - | - | ✓ | ✓ | - | - | - |
| 销售订单-创建 | ✓ | - | - | - | - | - | - | ✓ | ✓ | - |
| 销售订单-审批 | ✓ | - | - | - | - | - | - | ✓ | - | - |
| 财务-应付 | ✓ | - | - | - | - | - | - | - | - | ✓ |
| 财务-应收 | ✓ | - | - | - | - | - | - | - | - | ✓ |

---

## 三、功能模块详细设计

### 3.1 供应商管理模块

#### 3.1.1 功能清单

- 供应商档案管理
- 供应商资质管理
- 供应商评级评估
- 询价管理（RFQ）
- 合同管理
- 供应商绩效评估

#### 3.1.2 供应商属性

| 字段名 | 字段代码 | 类型 | 必填 | 说明 |
|-------|---------|------|-----|------|
| 供应商编码 | supplier_code | String(50) | 是 | 唯一标识 |
| 供应商名称 | name | String(200) | 是 | 公司全称 |
| 供应商类型 | type | Enum | 是 | 原材料/元器件/结构件/包装 |
| 联系人 | contact_name | String(100) | 是 | 主要联系人 |
| 联系电话 | phone | String(20) | 是 | - |
| 邮箱 | email | String(100) | 否 | - |
| 地址 | address | String(500) | 是 | 详细地址 |
| 付款条款 | payment_terms | String(100) | 否 | 账期等 |
| 评级 | rating | Enum | 否 | A/B/C/D |
| 状态 | status | Enum | 是 | ACTIVE/INACTIVE/BLACKLIST |

#### 3.1.3 供应商评分维度

| 维度 | 权重 | 计算方式 |
|-----|------|---------|
| 质量得分 | 40% | 100 - (不良率 × 1000) |
| 交期得分 | 30% | 准时交货率 × 100 |
| 价格得分 | 20% | 市场比价得分 |
| 服务得分 | 10% | 响应速度、配合度 |

**评级规则：**
- A级：综合得分 ≥ 90
- B级：综合得分 ≥ 75
- C级：综合得分 ≥ 60
- D级：综合得分 < 60

---

### 3.2 采购管理模块

#### 3.2.1 功能清单

- 采购需求管理（PR - Purchase Requisition）
- 采购订单管理（PO - Purchase Order）
- 到货管理
- 来料检验（IQC）
- 发票管理
- 付款管理

#### 3.2.2 采购流程

```
MRP计算/手动创建 → 采购需求PR → 审批 → 生成采购订单PO → 发送供应商
                                                              ↓
付款 ← 发票匹配 ← 入库 ← IQC质检 ← 收货 ← 供应商发货
```

#### 3.2.3 采购订单状态

| 状态 | 代码 | 说明 | 允许操作 |
|-----|------|------|---------|
| 草稿 | DRAFT | 新建未提交 | 编辑、删除、提交 |
| 待审批 | PENDING | 已提交待审批 | 审批、驳回 |
| 已审批 | APPROVED | 审批通过 | 发送、取消 |
| 已发送 | SENT | 已发送供应商 | 收货 |
| 部分到货 | PARTIAL | 部分收货 | 继续收货 |
| 全部到货 | RECEIVED | 全部收货 | 关闭 |
| 已关闭 | CLOSED | 订单完结 | - |
| 已取消 | CANCELLED | 订单取消 | - |

#### 3.2.4 审批规则

| 采购金额 | 审批流程 |
|---------|---------|
| < 5,000元 | 采购负责人 |
| 5,000 - 50,000元 | 采购负责人 → 财务负责人 |
| > 50,000元 | 采购负责人 → 财务负责人 → 总经理 |

---

### 3.3 库存管理模块

#### 3.3.1 功能清单

- 入库管理
  - 采购入库
  - 生产入库（成品）
  - 退货入库
- 出库管理
  - 生产领料
  - 销售出库
  - 报废出库
- 库存查询
- 库存盘点
- 安全库存预警
- 呆滞料管理

#### 3.3.2 库存类型

| 类型 | 代码 | 说明 |
|-----|------|------|
| 原材料 | RAW | 采购的原材料、元器件 |
| 半成品 | WIP | 生产过程中的半成品 |
| 成品 | FG | 完成生产的成品 |
| 备品备件 | SPARE | 售后维修用备件 |

#### 3.3.3 库位管理

```
仓库（Warehouse）
  └── 库区（Zone）
        └── 货架（Rack）
              └── 库位（Location）
```

#### 3.3.4 批次管理

- **批次号生成规则**：年月日+流水号，如 20260205001
- **序列号管理**：成品按序列号管理，支持单件追溯
- **先进先出（FIFO）**：出库时按批次入库时间排序

#### 3.3.5 库存预警

| 预警类型 | 触发条件 | 通知对象 |
|---------|---------|---------|
| 低库存预警 | 可用库存 < 安全库存 | 采购负责人 |
| 超储预警 | 库存 > 最大库存 × 120% | 仓库负责人 |
| 呆滞预警 | 超过90天无出库 | 仓库负责人、采购负责人 |
| 有效期预警 | 距过期 < 30天 | 仓库负责人 |

---

### 3.4 生产管理模块

#### 3.4.1 功能清单

- MRP物料需求计划
- 生产计划管理
- 工单管理
- 领料管理
- 生产报工
- 过程检验（IPQC）
- 最终检验（FQC）
- 成品入库

#### 3.4.2 生产流程

```
销售订单/生产计划 → MRP计算 → 生成采购需求 + 生产计划
                              ↓
                        创建生产工单
                              ↓
                        工单下达 → 领料
                              ↓
                        生产执行 → IPQC过程检验
                              ↓
                        生产完工 → FQC最终检验
                              ↓
                        成品入库
```

#### 3.4.3 工单状态

| 状态 | 代码 | 说明 |
|-----|------|------|
| 已创建 | CREATED | 工单已创建，未排产 |
| 已计划 | PLANNED | 已排产，确定生产日期 |
| 已下达 | RELEASED | 已下达到生产线 |
| 生产中 | IN_PROGRESS | 正在生产 |
| 已完成 | COMPLETED | 生产完成 |
| 已关闭 | CLOSED | 工单关闭 |

#### 3.4.4 MRP计算逻辑

```
毛需求 = 销售订单需求 + 安全库存需求
净需求 = 毛需求 - 现有库存 - 在途库存 - 在制数量
计划订单 = 净需求 / 批量取整
```

#### 3.4.5 质检类型

| 类型 | 代码 | 执行时机 | 说明 |
|-----|------|---------|------|
| 来料检验 | IQC | 采购入库前 | 检验供应商来料 |
| 过程检验 | IPQC | 生产过程中 | 抽检关键工序 |
| 最终检验 | FQC | 生产完成后 | 全检或抽检成品 |
| 出货检验 | OQC | 发货前 | 抽检出货产品 |

---

### 3.5 销售管理模块

#### 3.5.1 功能清单

- 客户管理
- 销售订单管理
- 发货管理
- 退换货管理
- 渠道管理
- 价格管理

#### 3.5.2 订单来源

| 渠道类型 | 说明 | 特点 |
|---------|------|------|
| 官网直销 | 公司官网订单 | 零售价，直接发货 |
| 电商平台 | 天猫、京东等 | 平台价格，API对接 |
| 线下代理 | 代理商批发 | 渠道价，账期结算 |
| 线下门店 | 眼镜门店零售 | 零售价，门店配送 |

#### 3.5.3 销售订单状态

| 状态 | 代码 | 说明 |
|-----|------|------|
| 待确认 | PENDING | 订单待确认 |
| 已确认 | CONFIRMED | 订单已确认，等待拣货 |
| 拣货中 | PICKING | 仓库拣货中 |
| 已发货 | SHIPPED | 已交付物流 |
| 已签收 | DELIVERED | 客户已签收 |
| 已完成 | COMPLETED | 订单完结 |
| 已取消 | CANCELLED | 订单取消 |

#### 3.5.4 价格管理

| 价格类型 | 说明 | 适用对象 |
|---------|------|---------|
| 零售价 | 官方指导价 | 终端消费者 |
| 渠道价 | 代理商价格 | 代理商 |
| 促销价 | 活动特价 | 指定活动 |
| 会员价 | 会员专享价 | 会员客户 |

---

### 3.6 售后管理模块

#### 3.6.1 功能清单

- 服务请求管理
- 维修工单管理
- 退换货处理
- 备件管理
- 质量追溯
- 故障分析

#### 3.6.2 服务类型

| 类型 | 代码 | 说明 | SLA |
|-----|------|------|-----|
| 维修 | REPAIR | 产品故障维修 | 7个工作日 |
| 退货 | RETURN | 产品退货退款 | 3个工作日 |
| 换货 | EXCHANGE | 产品换新 | 5个工作日 |
| 咨询 | INQUIRY | 使用咨询 | 24小时 |

#### 3.6.3 服务工单状态

| 状态 | 代码 | 说明 |
|-----|------|------|
| 已创建 | CREATED | 工单已创建 |
| 已分配 | ASSIGNED | 已分配处理人 |
| 处理中 | IN_PROGRESS | 正在处理 |
| 等备件 | WAITING_PARTS | 等待备件到货 |
| 已完成 | COMPLETED | 处理完成 |
| 已关闭 | CLOSED | 工单关闭 |

#### 3.6.4 质量追溯

通过产品序列号（SN）追溯：
- 生产批次
- 生产日期
- 使用的BOM版本
- 关键物料批次
- 供应商信息
- 生产工单
- 质检记录

---

### 3.7 财务管理模块

#### 3.7.1 功能清单

- 应付账款管理
- 应收账款管理
- 成本核算
- 利润分析
- 财务报表

#### 3.7.2 应付管理

**采购付款流程：**
```
采购订单 → 收货 → 供应商开票 → 发票核对 → 付款申请 → 审批 → 付款执行
```

**账期管理：**
- 月结30天
- 月结60天
- 款到发货
- 预付款

#### 3.7.3 应收管理

**销售收款流程：**
```
销售订单 → 发货 → 开票 → 收款确认 → 核销
```

#### 3.7.4 成本核算

| 成本类型 | 说明 | 计算方式 |
|---------|------|---------|
| 物料成本 | BOM物料成本 | 标准成本 × 用量 |
| 制造费用 | 生产制造费用 | 工时 × 费率 |
| 人工成本 | 直接人工 | 工时 × 人工费率 |
| 运费 | 物流运输费用 | 实际发生 |

---

## 四、数据模型设计

### 4.1 供应商表 suppliers

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| supplier_code | VARCHAR(50) | 是 | 供应商编码（唯一） |
| name | VARCHAR(200) | 是 | 供应商名称 |
| type | ENUM | 是 | 供应商类型 |
| contact_name | VARCHAR(100) | 是 | 联系人 |
| phone | VARCHAR(20) | 是 | 电话 |
| email | VARCHAR(100) | 否 | 邮箱 |
| address | VARCHAR(500) | 是 | 地址 |
| payment_terms | VARCHAR(100) | 否 | 付款条款 |
| rating | ENUM | 否 | 评级 A/B/C/D |
| status | ENUM | 是 | 状态 |
| created_at | TIMESTAMP | 是 | 创建时间 |
| updated_at | TIMESTAMP | 是 | 更新时间 |

### 4.2 采购订单表 purchase_orders

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| po_code | VARCHAR(50) | 是 | 订单编号（唯一） |
| supplier_id | UUID | 是 | 供应商ID |
| status | ENUM | 是 | 订单状态 |
| total_amount | DECIMAL(12,2) | 是 | 总金额 |
| currency | VARCHAR(10) | 是 | 币种 |
| order_date | DATE | 是 | 下单日期 |
| expected_date | DATE | 是 | 预计到货日期 |
| received_date | DATE | 否 | 实际到货日期 |
| created_by | VARCHAR(64) | 是 | 创建人（飞书ID） |
| approved_by | VARCHAR(64) | 否 | 审批人 |
| approved_at | TIMESTAMP | 否 | 审批时间 |
| created_at | TIMESTAMP | 是 | 创建时间 |

### 4.3 采购订单明细表 po_items

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| po_id | UUID | 是 | 关联采购订单 |
| material_id | UUID | 是 | 物料ID（来自PLM） |
| quantity | DECIMAL(12,4) | 是 | 采购数量 |
| unit | VARCHAR(20) | 是 | 单位 |
| unit_price | DECIMAL(12,4) | 是 | 单价 |
| amount | DECIMAL(12,2) | 是 | 金额 |
| received_qty | DECIMAL(12,4) | 否 | 已收货数量 |
| status | ENUM | 是 | 行状态 |

### 4.4 库存表 inventory

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| material_id | UUID | 是 | 物料ID |
| warehouse_id | UUID | 是 | 仓库ID |
| location | VARCHAR(50) | 否 | 库位 |
| batch_no | VARCHAR(50) | 否 | 批次号 |
| serial_no | VARCHAR(100) | 否 | 序列号（成品） |
| quantity | DECIMAL(12,4) | 是 | 库存数量 |
| reserved_qty | DECIMAL(12,4) | 是 | 预留数量 |
| available_qty | DECIMAL(12,4) | 是 | 可用数量 |
| unit_cost | DECIMAL(12,4) | 否 | 单位成本 |
| created_at | TIMESTAMP | 是 | 入库时间 |

### 4.5 库存交易表 inventory_transactions

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| material_id | UUID | 是 | 物料ID |
| warehouse_id | UUID | 是 | 仓库ID |
| transaction_type | ENUM | 是 | 交易类型（入库/出库/调整） |
| quantity | DECIMAL(12,4) | 是 | 数量（正为入，负为出） |
| batch_no | VARCHAR(50) | 否 | 批次号 |
| reference_type | VARCHAR(50) | 是 | 来源类型（PO/WO/SO等） |
| reference_id | UUID | 是 | 来源单据ID |
| created_by | VARCHAR(64) | 是 | 操作人 |
| created_at | TIMESTAMP | 是 | 交易时间 |

### 4.6 生产工单表 work_orders

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| wo_code | VARCHAR(50) | 是 | 工单编号（唯一） |
| product_id | UUID | 是 | 产品ID（来自PLM） |
| bom_id | UUID | 是 | BOM ID（来自PLM） |
| planned_qty | DECIMAL(12,4) | 是 | 计划数量 |
| completed_qty | DECIMAL(12,4) | 是 | 完成数量 |
| scrap_qty | DECIMAL(12,4) | 是 | 报废数量 |
| status | ENUM | 是 | 工单状态 |
| planned_start | DATE | 是 | 计划开始日期 |
| planned_end | DATE | 是 | 计划结束日期 |
| actual_start | DATE | 否 | 实际开始日期 |
| actual_end | DATE | 否 | 实际结束日期 |
| created_by | VARCHAR(64) | 是 | 创建人 |
| created_at | TIMESTAMP | 是 | 创建时间 |

### 4.7 销售订单表 sales_orders

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| so_code | VARCHAR(50) | 是 | 订单编号（唯一） |
| customer_id | UUID | 是 | 客户ID |
| channel | ENUM | 是 | 销售渠道 |
| status | ENUM | 是 | 订单状态 |
| total_amount | DECIMAL(12,2) | 是 | 总金额 |
| currency | VARCHAR(10) | 是 | 币种 |
| order_date | DATE | 是 | 下单日期 |
| shipping_date | DATE | 否 | 发货日期 |
| shipping_address | VARCHAR(500) | 是 | 收货地址 |
| created_at | TIMESTAMP | 是 | 创建时间 |

### 4.8 销售订单明细表 so_items

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| so_id | UUID | 是 | 关联销售订单 |
| product_id | UUID | 是 | 产品ID |
| quantity | DECIMAL(12,4) | 是 | 数量 |
| unit_price | DECIMAL(12,4) | 是 | 单价 |
| amount | DECIMAL(12,2) | 是 | 金额 |
| shipped_qty | DECIMAL(12,4) | 否 | 已发货数量 |
| status | ENUM | 是 | 行状态 |

### 4.9 服务工单表 service_orders

| 字段 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| id | UUID | 是 | 主键 |
| service_code | VARCHAR(50) | 是 | 服务单号（唯一） |
| customer_id | UUID | 是 | 客户ID |
| product_sn | VARCHAR(100) | 是 | 产品序列号 |
| service_type | ENUM | 是 | 服务类型 |
| status | ENUM | 是 | 工单状态 |
| description | TEXT | 是 | 问题描述 |
| solution | TEXT | 否 | 解决方案 |
| assignee_id | VARCHAR(64) | 否 | 处理人 |
| created_at | TIMESTAMP | 是 | 创建时间 |
| completed_at | TIMESTAMP | 否 | 完成时间 |

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
  "message": "参数校验失败",
  "data": null,
  "timestamp": 1706025600000
}
```

### 5.3 供应商接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/suppliers | 创建供应商 |
| GET | /api/v1/suppliers | 查询供应商列表 |
| GET | /api/v1/suppliers/{id} | 查询供应商详情 |
| PUT | /api/v1/suppliers/{id} | 更新供应商 |
| DELETE | /api/v1/suppliers/{id} | 删除供应商 |
| GET | /api/v1/suppliers/{id}/performance | 查询绩效评估 |

### 5.4 采购接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/purchase-orders | 创建采购订单 |
| GET | /api/v1/purchase-orders | 查询采购订单列表 |
| GET | /api/v1/purchase-orders/{id} | 查询采购订单详情 |
| PUT | /api/v1/purchase-orders/{id} | 更新采购订单 |
| POST | /api/v1/purchase-orders/{id}/submit | 提交审批 |
| POST | /api/v1/purchase-orders/{id}/approve | 审批通过 |
| POST | /api/v1/purchase-orders/{id}/reject | 审批驳回 |
| POST | /api/v1/purchase-orders/{id}/send | 发送供应商 |
| POST | /api/v1/purchase-orders/{id}/receive | 收货 |
| GET | /api/v1/purchase-requisitions | MRP采购建议 |

### 5.5 库存接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| GET | /api/v1/inventory | 库存查询 |
| GET | /api/v1/inventory/{material_id} | 物料库存详情 |
| POST | /api/v1/inventory/inbound | 入库 |
| POST | /api/v1/inventory/outbound | 出库 |
| POST | /api/v1/inventory/transfer | 库存调拨 |
| POST | /api/v1/inventory/adjust | 库存调整 |
| GET | /api/v1/inventory/alerts | 库存预警 |
| POST | /api/v1/inventory/stocktake | 创建盘点单 |

### 5.6 生产接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/work-orders | 创建工单 |
| GET | /api/v1/work-orders | 查询工单列表 |
| GET | /api/v1/work-orders/{id} | 查询工单详情 |
| PUT | /api/v1/work-orders/{id} | 更新工单 |
| POST | /api/v1/work-orders/{id}/release | 下达工单 |
| POST | /api/v1/work-orders/{id}/pick | 领料 |
| POST | /api/v1/work-orders/{id}/report | 生产报工 |
| POST | /api/v1/work-orders/{id}/complete | 工单完工 |
| GET | /api/v1/mrp/run | 执行MRP计算 |
| GET | /api/v1/mrp/result | 查询MRP结果 |

### 5.7 销售接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/sales-orders | 创建销售订单 |
| GET | /api/v1/sales-orders | 查询销售订单列表 |
| GET | /api/v1/sales-orders/{id} | 查询销售订单详情 |
| PUT | /api/v1/sales-orders/{id} | 更新销售订单 |
| POST | /api/v1/sales-orders/{id}/confirm | 确认订单 |
| POST | /api/v1/sales-orders/{id}/pick | 拣货 |
| POST | /api/v1/sales-orders/{id}/ship | 发货 |
| POST | /api/v1/sales-orders/{id}/cancel | 取消订单 |

### 5.8 售后接口

| 方法 | 路径 | 说明 |
|-----|------|------|
| POST | /api/v1/service-orders | 创建服务单 |
| GET | /api/v1/service-orders | 查询服务单列表 |
| GET | /api/v1/service-orders/{id} | 查询服务单详情 |
| PUT | /api/v1/service-orders/{id} | 更新服务单 |
| POST | /api/v1/service-orders/{id}/assign | 分配处理人 |
| POST | /api/v1/service-orders/{id}/complete | 完成处理 |
| GET | /api/v1/products/{sn}/trace | 质量追溯 |

---

## 六、飞书集成设计

### 6.1 单点登录

与PLM系统共用飞书OAuth认证，统一用户体系。

### 6.2 审批集成

| 审批场景 | 审批定义Code | 审批人规则 |
|---------|-------------|-----------|
| 采购订单审批 | po_approval | 按金额分级 |
| 付款申请审批 | payment_approval | 财务负责人 |
| 销售退换货审批 | return_approval | 售后经理 |
| 库存调整审批 | inv_adjust_approval | 仓库负责人 |

### 6.3 消息通知

| 事件 | 通知对象 | 通知方式 |
|-----|---------|---------|
| 采购到货 | 采购员、仓库员 | 飞书消息 |
| 库存预警 | 采购负责人 | 飞书消息 |
| 工单进度 | 生产负责人 | 飞书消息 |
| 发货完成 | 销售员、客户 | 飞书消息 |
| 服务工单分配 | 售后人员 | 飞书消息 |

---

## 七、PLM集成设计

### 7.1 BOM数据同步

ERP从PLM获取已发布的MBOM（制造BOM）作为生产和采购的依据。

**同步规则：**
- 只同步状态为RELEASED的MBOM
- BOM变更时自动触发同步
- 保留历史版本关联

### 7.2 物料主数据同步

物料基础信息从PLM同步到ERP，ERP维护供应商、价格、库存等业务属性。

**字段映射：**

| PLM字段 | ERP字段 | 同步方向 |
|--------|--------|---------|
| material_code | material_code | PLM→ERP |
| name | name | PLM→ERP |
| specification | specification | PLM→ERP |
| unit | unit | PLM→ERP |
| standard_cost | standard_cost | ERP→PLM |

### 7.3 产品数据同步

产品基础信息从PLM同步到ERP，ERP维护销售相关属性。

---

## 八、技术架构

### 8.1 技术选型

与PLM系统统一技术架构，降低运维复杂度，共享基础设施。

| 层级 | 技术 | 说明 |
|-----|------|------|
| 后端 | Go (Pure Go) | 与PLM统一，高性能，单二进制部署 |
| 数据库 | PostgreSQL 16 | 与PLM共享同一数据库实例 |
| ORM | GORM | 与PLM统一 |
| HTTP框架 | Gin | 与PLM统一 |
| 认证 | 飞书SSO + JWT | 与PLM共享认证体系 |
| 前端 | 内嵌SPA (单HTML) | 与PLM统一方案，嵌入二进制 |
| 部署 | 单二进制 + systemd | 简单可靠，与PLM同一服务器 |

### 8.2 系统架构

**单体架构，模块化设计：**

```
nimo-plm (单一Go二进制)
├── cmd/server/main.go          # 统一入口
├── internal/
│   ├── handler/                 # HTTP处理层
│   │   ├── plm_*.go            # PLM相关路由
│   │   └── erp_*.go            # ERP相关路由
│   ├── service/                 # 业务逻辑层
│   │   ├── plm_*.go            # PLM业务
│   │   ├── supplier_service.go  # 供应商管理
│   │   ├── procurement_service.go # 采购管理
│   │   ├── inventory_service.go # 库存管理
│   │   ├── manufacturing_service.go # 生产管理(含MRP)
│   │   ├── sales_service.go     # 销售管理
│   │   ├── service_order_service.go # 售后管理
│   │   └── finance_service.go   # 财务管理
│   ├── repository/              # 数据访问层
│   └── model/entity/           # 数据模型
│       ├── plm_*.go            # PLM实体
│       └── erp_*.go            # ERP实体
└── web/index.html              # PLM+ERP统一前端
```

**PLM与ERP共享的优势：**
- 用户认证统一（飞书SSO + JWT）
- 物料/BOM/产品数据零延迟共享（同一DB，无需同步）
- 部署运维极简（一个二进制、一个数据库）
- MRP计算直接读取PLM的BOM数据，无需API调用

---

## 九、非功能性需求

### 9.1 性能要求

| 指标 | 要求 |
|-----|------|
| API响应时间 | < 200ms（P95，Go性能优势） |
| MRP计算时间 | < 10秒（1000 SKU，Go并发计算） |
| 报表生成时间 | < 5秒 |
| 并发用户 | 支持200并发 |
| 系统可用性 | 99.9% |

### 9.2 安全要求

| 类别 | 要求 |
|-----|------|
| 认证 | 飞书OAuth 2.0 + JWT |
| 授权 | RBAC权限控制 |
| 传输 | 全站HTTPS（TLS 1.3） |
| 审计 | 全量操作日志 |
| 备份 | 每日自动备份，保留30天 |

### 9.3 兼容性要求

| 类别 | 要求 |
|-----|------|
| 浏览器 | Chrome 90+, Edge 90+, Safari 14+ |
| 分辨率 | 1920×1080（最佳），最低1366×768 |

---

## 十、实施计划

### 10.1 里程碑

| 阶段 | 时间 | 交付物 |
|-----|------|-------|
| 第一阶段 | 第1-4周 | 供应商管理、采购管理 |
| 第二阶段 | 第5-8周 | 库存管理、生产管理 |
| 第三阶段 | 第9-12周 | 销售管理、售后管理 |
| 第四阶段 | 第13-16周 | 财务管理、PLM集成 |
| 第五阶段 | 第17-20周 | 系统集成、优化上线 |

### 10.2 团队配置

| 角色 | 人数 | 职责 |
|-----|------|------|
| 项目负责人 | 1 (Claude COO) | 需求拆解、任务调度、质量把关 |
| 编程执行 | 1 (Claude Code) | Go后端 + 前端开发 |
| CEO评审 | 1 (泽斌) | 需求确认、验收 |

---

## 附录

### 附录A：错误码定义

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
| 30002 | PLM同步错误 |
| 50001 | 系统内部错误 |

### 附录B：状态枚举定义

**采购订单状态 PurchaseOrderStatus：**
DRAFT, PENDING, APPROVED, SENT, PARTIAL, RECEIVED, CLOSED, CANCELLED

**工单状态 WorkOrderStatus：**
CREATED, PLANNED, RELEASED, IN_PROGRESS, COMPLETED, CLOSED

**销售订单状态 SalesOrderStatus：**
PENDING, CONFIRMED, PICKING, SHIPPED, DELIVERED, COMPLETED, CANCELLED

**服务工单状态 ServiceOrderStatus：**
CREATED, ASSIGNED, IN_PROGRESS, WAITING_PARTS, COMPLETED, CLOSED

---

**文档版本历史：**

| 版本 | 日期 | 作者 | 变更说明 |
|-----|------|------|---------|
| v1.0 | 2026-01-30 | 企业效能专家团队 | 初稿 |
| v2.0 | 2026-02-05 | Claude Opus | 重构，完善到可编码级别 |
| v2.1 | 2026-02-07 | Claude COO | 统一技术架构为Pure Go单体，与PLM一致 |

---

**文档状态**：完成  
**下次评审日期**：2026-02-06
