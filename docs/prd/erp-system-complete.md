# nimo ERP 系统完整 PRD

> 版本: v1.0 | 日期: 2026-04-02 | 状态: Draft
> 对标: SAP S/4HANA, Oracle NetSuite, Odoo 17, 金蝶云星空
> 定位: 消费电子硬件企业 ERP，深度集成 ACP/PLM/SRM

---

## 一、第一性原理：ERP 到底解决什么问题

一个企业的经营活动归根结底是四件事的闭环：

```
客户要什么 → 我们能做什么 → 需要买什么 → 花了多少钱赚了多少钱
  (销售)        (生产)          (采购)         (财务)
```

ERP 不是"功能的堆砌"，而是这个闭环的数字化。每一条数据都应该能追溯到这个闭环中的位置。如果一个功能不在这个闭环里，就不需要它。

### 核心数据流

```
                    ┌──────────────────────────────────────────────┐
                    │                   财务总账                     │
                    │   (所有业务单据最终归集为会计凭证)                │
                    └──────┬──────────────┬───────────────┬────────┘
                           │              │               │
                    ┌──────┴──────┐ ┌─────┴─────┐ ┌──────┴──────┐
                    │   应收账款   │ │  成本核算  │ │   应付账款   │
                    └──────┬──────┘ └─────┬─────┘ └──────┬──────┘
                           │              │               │
  ┌──────────┐      ┌──────┴──────┐ ┌─────┴─────┐ ┌──────┴──────┐
  │ 客户下单  │─────→│  销售订单   │ │  生产工单  │ │  采购订单   │←── SRM
  └──────────┘      └──────┬──────┘ └─────┬─────┘ └──────┬──────┘
                           │              │               │
                           │        ┌─────┴─────┐        │
                           └───────→│  库存管理  │←───────┘
                                    └─────┬─────┘
                                          │
                                    ┌─────┴─────┐
                                    │  MRP 引擎  │←── PLM (BOM)
                                    └───────────┘
```

**MRP 是核心引擎**：销售订单说"要 1000 副眼镜"，MRP 用 PLM 的 BOM 展开成物料需求，减去库存现有量，生成采购建议（→SRM）和生产工单。这就是 ERP 的灵魂。

### 与现有系统的边界

| 系统 | 职责 | 数据流向 ERP |
|------|------|-------------|
| **PLM** | 产品定义、BOM 结构、ECN 变更 | BOM → MRP 展开用 |
| **SRM** | 供应商管理、采购执行、来料检验 | PO 执行结果 → 库存入库 → 应付 |
| **ERP** | 销售、库存、生产、MRP、财务 | 消费 PLM/SRM 数据，产出财务结果 |

**铁律**：PLM 管"怎么做"，SRM 管"找谁买"，ERP 管"做多少、卖多少、赚多少"。功能不重复。

---

## 二、系统架构

### 技术架构

作为 ACP 模块实现，跟 PLM/SRM 完全一致的架构：

```
ACP Platform
├── acp-module-plm    (已有)
├── acp-module-srm    (已有)
└── acp-module-erp    (本 PRD)
    ├── entities.go       # 实体定义
    ├── commands.go       # 命令定义
    ├── exec.go           # 业务逻辑
    ├── routes.go         # 自定义 API
    ├── models.go         # GORM 数据模型
    └── mrp.go            # MRP 引擎
```

### 模块总览

```
┌─────────────────────────────────────────────────────┐
│                    ERP Module                        │
│                                                     │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐       │
│  │  销售管理  │  │  库存管理  │  │  生产管理  │       │
│  │           │  │           │  │           │       │
│  │ 报价→订单  │  │ 出入库    │  │ MRP→工单   │       │
│  │ →发货→发票 │  │ 批次/序列  │  │ →报工→完工 │       │
│  │ →收款     │  │ 库位管理   │  │           │       │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘       │
│        │              │              │              │
│  ┌─────┴──────────────┴──────────────┴─────┐       │
│  │              财务管理                      │       │
│  │  总账 · 应收 · 应付 · 成本 · 报表          │       │
│  └─────────────────────────────────────────┘       │
│                                                     │
│  ┌───────────┐  ┌───────────┐                       │
│  │  质量管理  │  │  基础数据  │                       │
│  │ OQC·NCR   │  │ 客户·仓库  │                       │
│  │ CAPA      │  │ 币种·税率  │                       │
│  └───────────┘  └───────────┘                       │
└─────────────────────────────────────────────────────┘
```

---

## 三、数据模型

### 3.1 基础数据

#### 客户 (erp_customers)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | 客户编码，自动生成 C-0001 |
| name | VARCHAR(200) | 客户名称 |
| short_name | VARCHAR(50) | 简称 |
| type | ENUM | enterprise / individual / distributor / oem |
| tax_id | VARCHAR(50) | 税号/统一社会信用代码 |
| contact_name | VARCHAR(100) | 联系人 |
| contact_phone | VARCHAR(30) | 联系电话 |
| contact_email | VARCHAR(100) | 邮箱 |
| billing_address | TEXT | 开票地址 |
| shipping_address | TEXT | 收货地址（JSON数组，支持多地址） |
| payment_terms | VARCHAR(50) | 付款条件：NET30 / NET60 / COD / prepaid |
| credit_limit | DECIMAL(15,2) | 信用额度 |
| currency | VARCHAR(3) | 默认币种 CNY/USD |
| status | ENUM | active / inactive / blocked |
| tags | JSONB | 标签 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

#### 仓库 (erp_warehouses)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(20) UK | 仓库编码 W-001 |
| name | VARCHAR(100) | 仓库名称 |
| type | ENUM | raw_material / semi_finished / finished_goods / returns |
| address | TEXT | 地址 |
| manager_id | VARCHAR(100) | 仓管员 |
| is_default | BOOL | 是否默认仓库 |
| status | ENUM | active / inactive |

#### 库位 (erp_locations)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| warehouse_id | VARCHAR(100) FK | 所属仓库 |
| code | VARCHAR(30) UK | 库位编码 A-01-01 (区-排-位) |
| name | VARCHAR(100) | 库位名称 |
| type | ENUM | storage / picking / staging / quarantine |
| capacity | INT | 容量上限 |
| status | ENUM | active / inactive |

---

### 3.2 销售管理

#### 销售报价 (erp_quotations)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | QT-20260402-001 |
| customer_id | VARCHAR(100) FK | 客户 |
| contact_name | VARCHAR(100) | 客户联系人 |
| currency | VARCHAR(3) | 币种 |
| exchange_rate | DECIMAL(10,6) | 汇率 |
| subtotal | DECIMAL(15,2) | 小计 |
| tax_amount | DECIMAL(15,2) | 税额 |
| total | DECIMAL(15,2) | 总计 |
| valid_until | DATE | 有效期 |
| payment_terms | VARCHAR(50) | 付款条件 |
| notes | TEXT | 备注 |
| status | ENUM | draft / sent / accepted / rejected / expired |
| created_by | VARCHAR(100) | |
| created_at | TIMESTAMP | |

#### 报价明细 (erp_quotation_items)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| quotation_id | VARCHAR(100) FK | |
| product_id | VARCHAR(100) | PLM 产品 ID |
| sku_id | VARCHAR(100) | PLM SKU ID（可选） |
| description | VARCHAR(500) | 描述（可覆盖产品名） |
| quantity | DECIMAL(15,4) | 数量 |
| unit_price | DECIMAL(15,4) | 单价 |
| discount_pct | DECIMAL(5,2) | 折扣% |
| tax_rate | DECIMAL(5,2) | 税率% |
| line_total | DECIMAL(15,2) | 行合计 |
| sort_order | INT | 排序 |

#### 销售订单 (erp_sales_orders)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | SO-20260402-001 |
| quotation_id | VARCHAR(100) | 来源报价（可选） |
| customer_id | VARCHAR(100) FK | 客户 |
| shipping_address | TEXT | 收货地址 |
| currency | VARCHAR(3) | |
| subtotal | DECIMAL(15,2) | |
| tax_amount | DECIMAL(15,2) | |
| total | DECIMAL(15,2) | |
| payment_terms | VARCHAR(50) | |
| expected_date | DATE | 期望交货日 |
| shipping_method | VARCHAR(50) | 物流方式 |
| priority | ENUM | normal / urgent / critical |
| status | ENUM | draft / confirmed / producing / ready / shipped / delivered / closed / cancelled |
| notes | TEXT | |
| created_by | VARCHAR(100) | |
| confirmed_at | TIMESTAMP | 确认时间 |
| created_at | TIMESTAMP | |

#### 销售订单明细 (erp_sales_order_items)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| order_id | VARCHAR(100) FK | |
| product_id | VARCHAR(100) | PLM 产品 |
| sku_id | VARCHAR(100) | PLM SKU |
| bom_id | VARCHAR(100) | PLM BOM（生产用哪个 BOM） |
| description | VARCHAR(500) | |
| quantity | DECIMAL(15,4) | 订单数量 |
| delivered_qty | DECIMAL(15,4) | 已发货数量 |
| unit_price | DECIMAL(15,4) | |
| discount_pct | DECIMAL(5,2) | |
| tax_rate | DECIMAL(5,2) | |
| line_total | DECIMAL(15,2) | |
| expected_date | DATE | 行级交期 |
| status | ENUM | pending / allocated / producing / ready / shipped / delivered |

#### 发货单 (erp_shipments)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | SH-20260402-001 |
| order_id | VARCHAR(100) FK | 销售订单 |
| warehouse_id | VARCHAR(100) FK | 出库仓库 |
| shipping_address | TEXT | 收货地址 |
| carrier | VARCHAR(100) | 承运商 |
| tracking_no | VARCHAR(100) | 物流单号 |
| shipped_at | TIMESTAMP | 发货时间 |
| delivered_at | TIMESTAMP | 签收时间 |
| status | ENUM | draft / picking / packed / shipped / delivered / returned |
| notes | TEXT | |

#### 发货明细 (erp_shipment_items)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| shipment_id | VARCHAR(100) FK | |
| order_item_id | VARCHAR(100) FK | 销售订单行 |
| product_id | VARCHAR(100) | |
| quantity | DECIMAL(15,4) | 发货数量 |
| lot_number | VARCHAR(50) | 批次号 |
| serial_numbers | JSONB | 序列号列表 |

#### 退货单 (erp_returns)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | RMA-20260402-001 |
| order_id | VARCHAR(100) FK | 原始销售订单 |
| customer_id | VARCHAR(100) FK | |
| reason | VARCHAR(500) | 退货原因 |
| type | ENUM | refund / exchange / repair |
| total_amount | DECIMAL(15,2) | 退款金额 |
| status | ENUM | requested / approved / received / inspected / completed / rejected |
| created_at | TIMESTAMP | |

---

### 3.3 库存管理

#### 库存记录 (erp_inventory)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| material_id | VARCHAR(100) | PLM 物料 ID |
| warehouse_id | VARCHAR(100) FK | 仓库 |
| location_id | VARCHAR(100) FK | 库位 |
| lot_number | VARCHAR(50) | 批次号 |
| quantity | DECIMAL(15,4) | 现有数量 |
| reserved_qty | DECIMAL(15,4) | 预留数量（已分配给订单） |
| available_qty | DECIMAL(15,4) GENERATED | = quantity - reserved_qty |
| unit_cost | DECIMAL(15,4) | 单位成本 |
| status | ENUM | available / quarantine / damaged |
| expiry_date | DATE | 有效期（可选） |
| updated_at | TIMESTAMP | |

> 物料主数据复用 PLM 的 `plm_materials` 表，不重复创建。

#### 库存事务 (erp_inventory_transactions)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | IT-20260402-001 |
| type | ENUM | receive / issue / transfer / adjust / scrap / return |
| material_id | VARCHAR(100) | 物料 |
| from_warehouse_id | VARCHAR(100) | 源仓库（出库/调拨） |
| from_location_id | VARCHAR(100) | 源库位 |
| to_warehouse_id | VARCHAR(100) | 目标仓库（入库/调拨） |
| to_location_id | VARCHAR(100) | 目标库位 |
| quantity | DECIMAL(15,4) | 数量 |
| unit_cost | DECIMAL(15,4) | 单位成本 |
| lot_number | VARCHAR(50) | 批次号 |
| reference_type | VARCHAR(30) | po / so / wo / adjust / scrap |
| reference_id | VARCHAR(100) | 关联单据 ID |
| notes | TEXT | |
| created_by | VARCHAR(100) | |
| created_at | TIMESTAMP | |

#### 序列号追踪 (erp_serial_numbers)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| serial_number | VARCHAR(100) UK | 唯一序列号 |
| material_id | VARCHAR(100) | 物料/成品 |
| product_id | VARCHAR(100) | PLM 产品 |
| status | ENUM | in_stock / sold / returned / scrapped / in_repair |
| warehouse_id | VARCHAR(100) | 当前仓库 |
| lot_number | VARCHAR(50) | 所属批次 |
| manufactured_at | DATE | 生产日期 |
| sold_to | VARCHAR(100) | 销售客户 |
| sold_at | DATE | 销售日期 |
| warranty_until | DATE | 保修截止 |

> 智能眼镜是高值消费电子，每一副都需要序列号追踪（售后、保修、召回）。

---

### 3.4 生产管理

#### MRP 运算结果 (erp_mrp_results)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| run_id | VARCHAR(100) | MRP 运算批次 ID |
| material_id | VARCHAR(100) | 物料 |
| demand_source | VARCHAR(30) | so / wo / forecast |
| demand_id | VARCHAR(100) | 需求来源单据 |
| gross_requirement | DECIMAL(15,4) | 毛需求 |
| on_hand | DECIMAL(15,4) | 现有库存 |
| on_order | DECIMAL(15,4) | 在途（SRM 未到货 PO） |
| net_requirement | DECIMAL(15,4) | 净需求 = gross - on_hand - on_order |
| action | ENUM | purchase / produce / none |
| suggested_qty | DECIMAL(15,4) | 建议数量 |
| suggested_date | DATE | 建议日期 |
| bom_id | VARCHAR(100) | 展开用的 BOM |
| bom_level | INT | BOM 展开层级 |
| status | ENUM | suggested / confirmed / executed |
| created_at | TIMESTAMP | |

#### 生产工单 (erp_work_orders)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | WO-20260402-001 |
| product_id | VARCHAR(100) | PLM 产品 |
| bom_id | VARCHAR(100) | PLM BOM |
| order_id | VARCHAR(100) | 来源销售订单（可选） |
| mrp_result_id | VARCHAR(100) | 来源 MRP 建议（可选） |
| planned_qty | DECIMAL(15,4) | 计划数量 |
| completed_qty | DECIMAL(15,4) | 已完工数量 |
| scrap_qty | DECIMAL(15,4) | 报废数量 |
| warehouse_id | VARCHAR(100) | 成品入库仓 |
| planned_start | DATE | 计划开始 |
| planned_end | DATE | 计划完成 |
| actual_start | TIMESTAMP | 实际开始 |
| actual_end | TIMESTAMP | 实际完成 |
| priority | ENUM | low / normal / high / urgent |
| status | ENUM | draft / released / in_progress / completed / closed / cancelled |
| notes | TEXT | |
| created_by | VARCHAR(100) | |
| created_at | TIMESTAMP | |

#### 工单物料领用 (erp_wo_material_issues)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| work_order_id | VARCHAR(100) FK | |
| material_id | VARCHAR(100) | 物料 |
| bom_item_id | VARCHAR(100) | PLM BOM 行 |
| required_qty | DECIMAL(15,4) | BOM 要求数量 |
| issued_qty | DECIMAL(15,4) | 实际领用数量 |
| warehouse_id | VARCHAR(100) | 领料仓库 |
| lot_number | VARCHAR(50) | 批次 |
| issued_at | TIMESTAMP | 领料时间 |
| issued_by | VARCHAR(100) | 领料人 |

#### 工单报工 (erp_wo_reports)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| work_order_id | VARCHAR(100) FK | |
| operation | VARCHAR(100) | 工序名称 |
| operator_id | VARCHAR(100) | 操作员 |
| good_qty | DECIMAL(15,4) | 良品数量 |
| defect_qty | DECIMAL(15,4) | 不良数量 |
| scrap_qty | DECIMAL(15,4) | 报废数量 |
| start_time | TIMESTAMP | 开始时间 |
| end_time | TIMESTAMP | 结束时间 |
| notes | TEXT | |

---

### 3.5 财务管理

#### 会计科目 (erp_accounts)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(20) UK | 科目编码 1001 / 1122 / 2202 |
| name | VARCHAR(100) | 科目名称 |
| type | ENUM | asset / liability / equity / revenue / expense |
| parent_id | VARCHAR(100) | 父科目 |
| level | INT | 层级 |
| is_leaf | BOOL | 是否末级科目 |
| currency | VARCHAR(3) | 核算币种（空=本位币） |
| status | ENUM | active / inactive |

> 预置中国企业会计准则科目表（一级+二级），用户可扩展。

#### 会计凭证 (erp_journal_entries)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | JE-202604-001 |
| period | VARCHAR(7) | 会计期间 2026-04 |
| entry_date | DATE | 记账日期 |
| source_type | VARCHAR(30) | sales_invoice / purchase_invoice / receipt / payment / manual |
| source_id | VARCHAR(100) | 来源单据 ID |
| description | TEXT | 摘要 |
| total_debit | DECIMAL(15,2) | 借方合计 |
| total_credit | DECIMAL(15,2) | 贷方合计 |
| status | ENUM | draft / posted / reversed |
| posted_by | VARCHAR(100) | 过账人 |
| posted_at | TIMESTAMP | 过账时间 |
| created_by | VARCHAR(100) | |
| created_at | TIMESTAMP | |

#### 凭证分录 (erp_journal_lines)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| entry_id | VARCHAR(100) FK | 凭证 |
| account_id | VARCHAR(100) FK | 科目 |
| debit | DECIMAL(15,2) | 借方金额 |
| credit | DECIMAL(15,2) | 贷方金额 |
| currency | VARCHAR(3) | 原币币种 |
| original_amount | DECIMAL(15,2) | 原币金额 |
| description | VARCHAR(500) | 行摘要 |
| customer_id | VARCHAR(100) | 辅助核算-客户 |
| supplier_id | VARCHAR(100) | 辅助核算-供应商 |
| department_id | VARCHAR(100) | 辅助核算-部门 |

#### 销售发票 (erp_sales_invoices)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | INV-20260402-001 |
| order_id | VARCHAR(100) FK | 销售订单 |
| customer_id | VARCHAR(100) FK | |
| invoice_date | DATE | 开票日期 |
| due_date | DATE | 到期日 |
| currency | VARCHAR(3) | |
| subtotal | DECIMAL(15,2) | |
| tax_amount | DECIMAL(15,2) | |
| total | DECIMAL(15,2) | |
| paid_amount | DECIMAL(15,2) | 已收金额 |
| balance | DECIMAL(15,2) GENERATED | = total - paid_amount |
| status | ENUM | draft / issued / partially_paid / paid / overdue / cancelled |
| journal_entry_id | VARCHAR(100) | 关联凭证 |
| created_at | TIMESTAMP | |

#### 收款记录 (erp_receipts)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | REC-20260402-001 |
| customer_id | VARCHAR(100) FK | |
| amount | DECIMAL(15,2) | 收款金额 |
| currency | VARCHAR(3) | |
| payment_method | ENUM | bank_transfer / check / cash / online |
| bank_account | VARCHAR(50) | 收款银行账号 |
| reference_no | VARCHAR(100) | 银行流水号 |
| received_date | DATE | 收款日期 |
| status | ENUM | draft / confirmed / reconciled |
| journal_entry_id | VARCHAR(100) | 关联凭证 |
| notes | TEXT | |

#### 收款核销 (erp_receipt_allocations)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| receipt_id | VARCHAR(100) FK | 收款 |
| invoice_id | VARCHAR(100) FK | 发票 |
| amount | DECIMAL(15,2) | 核销金额 |

---

### 3.6 质量管理（轻量）

> IQC（来料检验）已在 SRM 模块，这里补充 OQC 和 CAPA。

#### 出货检验 (erp_oqc_inspections)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | OQC-20260402-001 |
| shipment_id | VARCHAR(100) FK | 关联发货单 |
| product_id | VARCHAR(100) | |
| lot_number | VARCHAR(50) | 批次 |
| sample_size | INT | 抽样数 |
| total_inspected | INT | 检验总数 |
| pass_count | INT | 合格数 |
| fail_count | INT | 不合格数 |
| result | ENUM | pass / fail / conditional |
| inspector_id | VARCHAR(100) | 检验员 |
| inspected_at | TIMESTAMP | |
| notes | TEXT | |
| defect_details | JSONB | 不良明细 [{type, count, description}] |
| status | ENUM | pending / completed |

#### 不合格品报告 (erp_ncr_reports)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | NCR-20260402-001 |
| source | ENUM | iqc / ipqc / oqc / customer_return / internal |
| source_id | VARCHAR(100) | 来源检验单/退货单 ID |
| product_id | VARCHAR(100) | |
| material_id | VARCHAR(100) | |
| lot_number | VARCHAR(50) | |
| defect_qty | DECIMAL(15,4) | 不良数量 |
| defect_type | VARCHAR(100) | 不良类型 |
| description | TEXT | 不良描述 |
| disposition | ENUM | use_as_is / rework / scrap / return_to_supplier |
| severity | ENUM | minor / major / critical |
| status | ENUM | open / reviewing / dispositioned / closed |
| owner_id | VARCHAR(100) | 责任人 |
| created_at | TIMESTAMP | |
| closed_at | TIMESTAMP | |

#### CAPA (erp_capa)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | VARCHAR(100) PK | |
| code | VARCHAR(30) UK | CAPA-20260402-001 |
| type | ENUM | corrective / preventive |
| ncr_id | VARCHAR(100) | 关联 NCR |
| title | VARCHAR(200) | 标题 |
| root_cause | TEXT | 根因分析 |
| action_plan | TEXT | 纠正/预防措施 |
| owner_id | VARCHAR(100) | 责任人 |
| due_date | DATE | 截止日期 |
| verification | TEXT | 有效性验证 |
| status | ENUM | open / in_progress / pending_verification / closed |
| created_at | TIMESTAMP | |
| closed_at | TIMESTAMP | |

---

## 四、核心业务流程

### 4.1 Order-to-Cash（订单到收款）

```
报价 → 销售订单 → [MRP] → 生产工单 → 领料出库
                                        ↓
                              完工入库 → OQC检验
                                        ↓
                              发货出库 → 物流跟踪
                                        ↓
                              销售发票 → 收款核销 → 凭证
```

**状态流转 — 销售订单：**
```
draft → confirmed → producing → ready → shipped → delivered → closed
                                                              ↗
                                              cancelled ←────┘
```

### 4.2 MRP 物料需求计划

MRP 是 ERP 的计算引擎，连接销售、库存、采购、生产：

```
输入:
  1. 需求来源：销售订单 + 预测（可选）
  2. BOM：从 PLM 读取已发布的 BOM
  3. 库存：现有量 + 安全库存
  4. 在途：SRM 的未到货 PO

计算:
  对每个需求，展开 BOM 到最底层：
    毛需求 = 订单数量 × BOM 用量
    净需求 = 毛需求 - 现有库存 - 在途数量
    if 净需求 > 0:
      if 物料有BOM（半成品）→ 生成生产工单建议
      else（原材料）→ 生成采购申请建议 → SRM

输出:
  - 采购建议列表（→ SRM 自动创建 PR）
  - 生产工单建议列表
  - 缺料预警
```

**MRP 与 PLM/SRM 的数据交互：**

| 数据 | 来源 | 用途 |
|------|------|------|
| BOM 结构 | PLM `plm_boms` + `plm_bom_items` | 展开物料需求 |
| 物料主数据 | PLM `plm_materials` | 采购提前期、安全库存 |
| 在途 PO | SRM `srm_purchase_orders` | 扣减净需求 |
| 采购建议 | MRP 输出 | 自动创建 SRM 的 PR |

### 4.3 Procure-to-Pay（采购到付款）

```
[MRP 采购建议] → SRM 采购申请 → SRM 采购订单 → SRM 到货检验
                                                     ↓
                                         ERP 库存入库 (receive)
                                                     ↓
                                         SRM 供应商发票
                                                     ↓
                                         ERP 应付凭证 → 付款 → 凭证
```

> 采购执行在 SRM，ERP 只负责库存入库和应付账款。

### 4.4 生产执行

```
生产工单 (released)
    ↓
物料领用 (BOM展开，从库存扣减)
    ↓
生产报工 (每道工序报良品/不良品)
    ↓
完工入库 (成品入库，序列号生成)
    ↓
OQC 出货检验
```

### 4.5 财务自动凭证

所有业务单据自动生成会计凭证，不需要手工做账：

| 业务事件 | 借方 | 贷方 |
|---------|------|------|
| 销售发货 | 应收账款 | 主营业务收入 + 应交税费 |
| 收款 | 银行存款 | 应收账款 |
| 采购入库 | 原材料 | 应付账款 |
| 付款 | 应付账款 | 银行存款 |
| 生产领料 | 生产成本-直接材料 | 原材料 |
| 完工入库 | 库存商品 | 生产成本 |
| 销售出库（成本结转） | 主营业务成本 | 库存商品 |

---

## 五、命令定义（ACP Module Commands）

### 5.1 销售管理

| 命令 | 说明 | 关键 Input | 关键 Output |
|------|------|-----------|------------|
| create_quotation | 创建报价 | customer_id, items[] | quotation_id, code |
| confirm_quotation | 报价转订单 | quotation_id | order_id |
| create_sales_order | 创建销售订单 | customer_id, items[] | order_id, code |
| confirm_order | 确认订单（触发MRP） | order_id | mrp_run_id |
| create_shipment | 创建发货单 | order_id, items[] | shipment_id |
| confirm_shipment | 确认发货 | shipment_id, tracking_no | |
| create_sales_invoice | 创建发票 | order_id | invoice_id |
| record_receipt | 记录收款 | customer_id, amount, invoices[] | receipt_id |
| create_return | 创建退货 | order_id, items[], reason | return_id |

### 5.2 库存管理

| 命令 | 说明 | 关键 Input | 关键 Output |
|------|------|-----------|------------|
| receive_inventory | 入库 | material_id, warehouse_id, qty, lot | transaction_id |
| issue_inventory | 出库 | material_id, warehouse_id, qty | transaction_id |
| transfer_inventory | 调拨 | material_id, from, to, qty | transaction_id |
| adjust_inventory | 盘点调整 | material_id, warehouse_id, actual_qty | transaction_id, diff |
| scrap_inventory | 报废 | material_id, qty, reason | transaction_id |
| reserve_inventory | 预留（销售订单分配） | order_item_id, material_id, qty | |
| unreserve_inventory | 取消预留 | order_item_id | |

### 5.3 生产管理

| 命令 | 说明 | 关键 Input | 关键 Output |
|------|------|-----------|------------|
| run_mrp | 执行 MRP 运算 | demand_source | run_id, suggestions[] |
| confirm_mrp_suggestion | 确认 MRP 建议 | suggestion_id | wo_id 或 pr_id |
| create_work_order | 创建生产工单 | product_id, bom_id, qty | wo_id, code |
| release_work_order | 下达工单 | wo_id | |
| issue_wo_materials | 工单领料 | wo_id | transactions[] |
| report_wo_progress | 工序报工 | wo_id, operation, good_qty, defect_qty | |
| complete_work_order | 完工入库 | wo_id, completed_qty, serial_numbers[] | transaction_id |

### 5.4 财务管理

| 命令 | 说明 | 关键 Input | 关键 Output |
|------|------|-----------|------------|
| create_journal_entry | 创建凭证 | lines[] | entry_id |
| post_journal_entry | 过账 | entry_id | |
| reverse_journal_entry | 红冲 | entry_id | new_entry_id |
| close_period | 期末结转 | period | |
| generate_report | 生成报表 | report_type, period | report_data |

### 5.5 质量管理

| 命令 | 说明 | 关键 Input | 关键 Output |
|------|------|-----------|------------|
| create_oqc | 创建出货检验 | shipment_id | oqc_id |
| complete_oqc | 完成检验 | oqc_id, result, defects[] | |
| create_ncr | 创建不合格品报告 | source, defect_type, qty | ncr_id |
| disposition_ncr | NCR 处置 | ncr_id, disposition | |
| create_capa | 创建 CAPA | ncr_id, type, action_plan | capa_id |
| close_capa | 关闭 CAPA | capa_id, verification | |

---

## 六、自定义 API 路由

挂载在 `/api/m/erp/`：

```
# 销售
GET    /order-summary/:id              # 订单汇总（含发货、发票、收款状态）
GET    /customer-statement/:id         # 客户对账单
GET    /price-list                     # 价格表

# 库存
GET    /stock-summary                  # 库存汇总（按物料/仓库）
GET    /stock-movements/:material_id   # 物料出入库流水
GET    /stock-aging                    # 库龄分析
GET    /serial-trace/:serial           # 序列号追溯

# 生产
POST   /mrp-run                        # 执行 MRP
GET    /mrp-results/:run_id            # MRP 结果
GET    /production-schedule            # 生产排程看板
GET    /wo-dashboard                   # 工单仪表盘

# 财务
GET    /trial-balance                  # 试算平衡表
GET    /income-statement               # 利润表
GET    /balance-sheet                  # 资产负债表
GET    /cash-flow                      # 现金流量表
GET    /ar-aging                       # 应收账龄分析
GET    /ap-aging                       # 应付账龄分析
GET    /cost-analysis/:product_id      # 产品成本分析

# 质量
GET    /oqc-dashboard                  # OQC 仪表盘
GET    /quality-trend                  # 质量趋势
```

---

## 七、UI/UX 设计规范

### 7.1 设计原则

**对标**: Linear（交互流畅度）、Figma（信息密度）、Notion（内容组织）、SAP Fiori（企业级导航）

1. **信息密度优先** — ERP 用户需要同时看到大量数据，不为了美观牺牲效率
2. **操作就近原则** — 所有操作在当前上下文完成，不跳页
3. **实时反馈** — 每个操作立即看到结果，用乐观更新
4. **键盘优先** — 财务和仓储人员大量录入场景，全键盘可操作
5. **暗色主题** — 延续 ACP 的暗色设计语言

### 7.2 页面布局

```
┌─────────────────────────────────────────────────────┐
│  顶栏: 模块切换(PLM/SRM/ERP) + 全局搜索 + 通知     │
├────────┬────────────────────────────────────────────┤
│ 侧边栏  │  主内容区                                  │
│        │                                            │
│ 销售    │  ┌──────────────────────────────────────┐ │
│  报价   │  │  列表/看板/甘特 切换                   │ │
│  订单   │  │                                      │ │
│  发货   │  │  ┌────────────────┬─────────────┐   │ │
│  发票   │  │  │ 主列表          │ 详情面板     │   │ │
│  收款   │  │  │ (可筛选/排序)   │ (右侧滑出)  │   │ │
│        │  │  │               │             │   │ │
│ 库存    │  │  │               │             │   │ │
│  总览   │  │  │               │             │   │ │
│  出入库  │  │  └────────────────┴─────────────┘   │ │
│  盘点   │  │                                      │ │
│        │  │  底部: 汇总统计行                       │ │
│ 生产    │  └──────────────────────────────────────┘ │
│  MRP   │                                            │
│  工单   │                                            │
│        │                                            │
│ 财务    │                                            │
│  凭证   │                                            │
│  报表   │                                            │
│        │                                            │
│ 质量    │                                            │
│  OQC   │                                            │
│  NCR   │                                            │
└────────┴────────────────────────────────────────────┘
```

### 7.3 关键交互模式

**主从列表（Master-Detail）**
- 左侧列表，点击行在右侧滑出详情面板（不跳页）
- 详情面板可展开为全屏
- 列表支持多选 + 批量操作
- 列宽可拖拽调整，列顺序可拖拽排序

**快速创建**
- `Cmd+K` 全局命令面板（类 Linear/Raycast）
- 输入 "新建订单"、"创建工单" 直接进入表单
- 表单内 Tab 键快速切换字段

**数据联动预览**
- 鼠标悬停客户名 → 弹出卡片显示信用额度、最近订单、应收余额
- 悬停物料名 → 显示库存量、在途量、安全库存预警
- 悬停订单号 → 显示订单进度、发货状态、收款状态

**看板视图**
- 销售订单按状态分列（confirmed → producing → ready → shipped）
- 生产工单按状态分列
- 拖拽卡片切换状态

**仪表盘**
- 销售仪表盘: 本月销售额、订单数、待发货、逾期应收
- 库存仪表盘: 库存金额、周转天数、低库存预警、呆滞物料
- 生产仪表盘: 在制工单、产能利用率、良率趋势
- 财务仪表盘: 现金余额、应收/应付、利润趋势

### 7.4 颜色系统

延续 ACP 暗色主题，业务状态使用统一色彩：

| 语义 | 颜色 | 用途 |
|------|------|------|
| 信息/进行中 | `#1677ff` | 活跃状态、进行中 |
| 成功/完成 | `#52c41a` | 已完成、合格、已收款 |
| 警告/待处理 | `#faad14` | 待审批、即将到期 |
| 危险/逾期 | `#ff4d4f` | 逾期、不合格、欠款 |
| 中性/草稿 | `#8c8c8c` | 草稿、已取消 |
| 紫色/特殊 | `#722ed1` | MRP 建议、预测需求 |

---

## 八、PLM/SRM 集成点清单

### 读取 PLM 数据

| ERP 场景 | PLM 数据 | 调用方式 |
|---------|---------|---------|
| 销售订单选择产品 | `plm_products` + `plm_product_skus` | 模块间 entity 引用 |
| MRP 展开 BOM | `plm_boms` + `plm_bom_items` | 自定义路由跨模块查询 |
| 生产领料 | `plm_materials`（物料主数据） | entity 引用 |
| 成本核算 | `plm_bom_items`（标准用量） | 跨模块查询 |

### 与 SRM 交互

| ERP 场景 | SRM 交互 | 方向 |
|---------|---------|------|
| MRP 采购建议 | 自动创建 `srm_purchase_requests` | ERP → SRM |
| 采购到货入库 | 读取 `srm_purchase_orders` 到货信息 | SRM → ERP |
| 应付账款 | 读取 SRM 供应商发票 | SRM → ERP |
| 在途物料 | 查询 SRM 未到货 PO 数量 | ERP ← SRM |
| IQC 不合格 | SRM IQC → ERP NCR | SRM → ERP |

---

## 九、实施计划

### Phase 1 — 库存 + 基础数据（4 周）

最先做库存，因为它是所有其他模块的基础。

- 客户、仓库、库位基础数据
- 库存出入库（手动）
- 库存查询、流水
- 序列号管理
- 与 PLM 物料打通

### Phase 2 — 销售管理（3 周）

有了库存就可以做销售。

- 报价 → 销售订单 → 发货 → 发票 → 收款
- 订单看板视图
- 库存预留/释放
- 与 PLM 产品/SKU 打通

### Phase 3 — 生产 + MRP（4 周）

最复杂的模块，依赖库存和 PLM BOM。

- MRP 引擎（BOM 展开、净需求计算）
- 生产工单（创建、领料、报工、完工入库）
- MRP → SRM 自动创建采购申请
- 生产排程视图

### Phase 4 — 财务（4 周）

所有业务数据最终汇入财务。

- 科目表预置
- 业务单据自动生成凭证
- 应收/应付管理
- 期末结转
- 三大报表（资产负债表、利润表、现金流量表）

### Phase 5 — 质量 + 优化（2 周）

- OQC 出货检验
- NCR 不合格品报告
- CAPA 纠正预防
- 仪表盘完善
- 性能优化

---

## 十、核心指标

系统上线后应能回答以下问题：

| 问题 | 数据来源 |
|------|---------|
| 这个月卖了多少钱？ | 销售订单汇总 |
| 哪些订单还没发货？ | 销售订单 status=confirmed/producing |
| 仓库里还有多少原材料？ | 库存汇总 by 仓库/物料 |
| 这副眼镜的成本是多少？ | BOM 展开 × 物料成本 |
| 客户欠我们多少钱？ | 应收账龄分析 |
| 我们欠供应商多少钱？ | 应付账龄分析 |
| 下个月需要采购什么？ | MRP 运算结果 |
| 生产线良率怎么样？ | 工单报工汇总 |
| 这副眼镜经过了什么检验？ | 序列号追溯 |
| 公司赚了多少钱？ | 利润表 |
