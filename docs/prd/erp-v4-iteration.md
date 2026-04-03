# ERP v4 迭代 PRD — 审计驱动的系统性改进

> 版本: v4.0 | 日期: 2026-04-03
> 基于: v3 全模块审计结果（25 个页面、15 个 API 端点、17 个实体的完整检查）

---

## 一、审计发现总览

### v3 做对了什么

- **差异化页面设计**：14 个独特 view 页面（管道/360 视图/操作面板/搜索台/控制台/追溯等），不再千篇一律
- **业务场景导向**：每个页面围绕"用户在做什么"设计，不是围绕"数据有什么字段"
- **数据完整**：17 个实体共 55 条测试数据，15 个自定义 API 端点全部可用

### v3 的 5 类系统性问题

| 类别 | 问题 | 影响范围 | 优先级 |
|------|------|---------|--------|
| **A. 前后端不对齐** | 前端调用不存在的 API / 不使用已有 API | Dashboard, MRP, AR, Quality | P0 |
| **B. 交互有形无实** | 看板不能拖拽、操作按钮无后端支持 | Pipeline, Schedule, CAPA | P0 |
| **C. 业务流程断裂** | 关键工作流缺少环节（退款、对账、自动 NCR） | 销售/财务/质量全链路 | P1 |
| **D. 性能与数据量** | 大量 page_size=500 无分页、客户端重复计算 | 所有 list 页面 | P1 |
| **E. 功能完成度** | 现金流量表桩代码、打印导出缺失、模板硬编码 | 报表/凭证/发货 | P2 |

---

## 二、P0 修复 — 系统能正常工作

### 2.1 前后端 API 对齐

| 前端页面 | 调用的 API | 状态 | 修复方案 |
|---------|-----------|------|---------|
| ErpDashboard | `/api/m/erp/chatter` | **404** | 后端新增 chatter API：查询最近操作日志 |
| ErpDashboard | `/api/m/erp/sales-trend` | **404** | 后端新增 sales-trend API：按月汇总销售额 |
| ErpMRPConsole | `POST /api/m/erp/mrp-run` | **404** | 后端路由注册：映射到 `run_mrp` 命令 |
| ErpARWorkspace | 客户端计算账龄 | 重复 | 改用已有 `/ar-aging` 端点 |
| ErpQualityConsole | 客户端计算趋势 | 重复 | 改用已有 `/oqc-dashboard` + `/quality-trend` 端点 |
| ErpCustomer360 | 客户端计算对账 | 重复 | 改用已有 `/customer-statement/:id` 端点 |

**需新增的后端 API（routes.go）：**

```go
// 最近活动日志（Dashboard 用）
GET /api/m/erp/activity-log?limit=10

// 销售趋势（Dashboard 用）
GET /api/m/erp/sales-trend?months=6

// MRP 执行（MRPConsole 用）
POST /api/m/erp/mrp-run
```

### 2.2 看板拖拽交互

三个看板页面（SalesPipeline, ProductionSchedule, CAPABoard）都是"视觉看板" — 只能看不能拖。需要实现真正的拖拽状态变更。

**技术方案**：使用 `@dnd-kit/core` + `@dnd-kit/sortable`（React DnD 库，ACP 项目已有类似依赖）

**拖拽行为**：
- 拖动卡片到另一列 → 调用 `executeCommand(module, 'update_xxx', { id, status: newStatus })` 
- 拖拽时卡片半透明 + 占位符 + 目标列高亮
- 拖拽后乐观更新（先更新 UI，后台同步）
- 失败回滚 + toast 提示

**状态变更校验**：不是任意状态都能拖拽，需校验有效转换：
- 销售订单：draft→confirmed→producing→ready→shipped→delivered（单向）
- 工单：draft→released→in_progress→completed（单向）
- CAPA：open→in_progress→pending_verification→closed（单向）

### 2.3 操作按钮后端支持

以下前端操作按钮目前无对应后端命令或命令名不匹配：

| 页面 | 按钮 | 需要的命令 | 状态 |
|------|------|-----------|------|
| ErpShippingCenter | 创建发货单 | `create_shipment` | ✅ 已有 |
| ErpShippingCenter | 开始拣货 | `update_shipment_status` | ❌ 需新增 |
| ErpQuoteBoard | 转为订单 | `confirm_quotation` | ✅ 已有 |
| ErpQuoteBoard | 延期 | `extend_quotation` | ❌ 需新增 |
| ErpARWorkspace | 发送催款 | `send_reminder` | ❌ 需新增（或用通知服务） |
| ErpQualityConsole | 创建 OQC | `create_oqc` | ✅ 已有 |

**需新增的命令（exec.go）：**

```go
// 更新发货单状态（支持 picking/packed/shipped/delivered）
"update_shipment_status": cmdUpdateShipmentStatus

// 延长报价有效期
"extend_quotation": cmdExtendQuotation

// 发送催款通知（记录催款操作，可选飞书通知）
"send_payment_reminder": cmdSendPaymentReminder
```

---

## 三、P1 改进 — 业务流程完整性

### 3.1 销售全链路补全

**当前断裂点**：
- 报价 → 订单的转换只有命令，无 UI 确认流程
- 订单确认后不自动触发 MRP
- 发票开具后无自动生成凭证
- 退货无退款/信用票关联

**改进**：

**a) 订单确认联动 MRP**
```
用户点击 [确认订单] 
→ 弹出确认对话框："确认后将自动计算物料需求（MRP），是否继续？"
→ 确认 → 调 confirm_order 命令 → 后端自动运行 MRP → 返回 MRP 建议数
→ Toast: "订单已确认，MRP 生成 12 条采购建议"
→ 显示 [查看 MRP 结果] 链接
```

**b) 发票自动生成凭证**
```
发票状态变为 "issued" 时
→ 后端自动创建 Journal Entry:
   借: 应收账款 (1122)  金额: 发票总额
   贷: 主营业务收入 (6001)  金额: 不含税金额
   贷: 应交税费 (2221)  金额: 税额
→ 发票详情页显示关联凭证链接
```

**c) 退货退款流程**
```
退货状态变为 "completed" + type="refund" 时
→ 后端自动创建：
   1. 红字发票（或信用票）
   2. 冲销凭证
   3. 库存入库事务（退回的产品）
→ 退货详情页显示关联的发票和凭证
```

### 3.2 库存事务自动化

**当前问题**：入库/出库是手动操作，没有跟 PO/WO/SO 自动联动

**改进**：

```
采购到货 → 自动创建入库事务 + 更新库存
工单领料 → 自动创建出库事务 + 扣减库存
发货确认 → 自动创建出库事务 + 扣减库存
退货入库 → 自动创建入库事务 + 增加库存
```

在 `confirm_shipment`、`issue_wo_materials`、`complete_work_order` 等命令中增加库存事务联动逻辑。

### 3.3 质量自动触发

**当前问题**：OQC 检验失败后需要手动创建 NCR

**改进**：
```
OQC result = "fail" 或 "conditional" 时
→ 自动创建 NCR（source=oqc, source_id=oqc_id）
→ OQC 详情页显示关联 NCR 链接
→ 质量控制台待处理列表自动出现新 NCR
```

### 3.4 财务对账自动匹配

**当前问题**：收款录入后需要手动核销到发票

**改进**：
```
录入收款时
→ 如果指定了 customer_id
→ 自动匹配该客户最早的未付发票
→ 按 FIFO 原则自动核销
→ 更新发票 paid_amount 和 status
→ 如果收款 > 当前发票余额，自动分摊到下一张发票
```

---

## 四、P1 改进 — 性能优化

### 4.1 分页查询

当前多个页面用 `page_size: 500` 一次性加载所有数据。需改为按需加载：

| 页面 | 当前 | 改进 |
|------|------|------|
| ErpSalesPipeline | 加载全部订单 (500) | 分页加载，支持 "加载更多" |
| ErpARWorkspace | 加载全部发票 (500) | 使用 `/ar-aging` 后端聚合 |
| ErpStockSearch | 加载全部库存 | 搜索时才查询（改为服务端搜索） |
| ErpCustomer360 | 加载客户全部订单 | 分页 + 按年份折叠 |

### 4.2 后端聚合替代前端计算

以下计算应该在后端完成，前端直接用结果：

| 计算 | 当前 | 改进 |
|------|------|------|
| 账龄分析 | 前端遍历发票 | 用 `/ar-aging` |
| 质量趋势 | 前端遍历 OQC | 用 `/quality-trend` |
| 客户对账 | 前端拼接发票+收款 | 用 `/customer-statement/:id` |
| 销售管道汇总 | 前端分组计算 | 新增 `/pipeline-summary` |
| 库存预警 | 前端比较库存 | 新增 `/stock-alerts` |

**需新增的后端聚合 API：**

```go
// 销售管道汇总（每个阶段的数量和金额）
GET /api/m/erp/pipeline-summary

// 库存预警列表（低于安全库存的物料）
GET /api/m/erp/stock-alerts
```

---

## 五、P2 改进 — 功能完善

### 5.1 报表导出

**ReportCenter 增加导出功能**：
- [导出 PDF] — 使用 `html2canvas` + `jsPDF` 生成
- [导出 Excel] — 使用 `xlsx` 库导出表格数据
- [打印] — 使用 `@media print` CSS + `window.print()`

### 5.2 现金流量表

ReportCenter 的现金流量表目前是桩代码。需要实现：

```
经营活动:
  销售收款              ¥X,XXX,XXX  (sum of receipts)
  采购付款             (¥X,XXX,XXX) (sum of AP payments)
  ──────────────────────────────
  经营活动现金流净额     ¥X,XXX,XXX

投资活动:
  购置设备             (¥X,XXX,XXX)
  ──────────────────────────────
  投资活动现金流净额    (¥X,XXX,XXX)

现金净增加额            ¥X,XXX,XXX
期初现金余额            ¥X,XXX,XXX
期末现金余额            ¥X,XXX,XXX
```

需新增后端 API：
```go
GET /api/m/erp/cash-flow?period=2026-04
```

### 5.3 打印模板

**关键单据打印**：
- 销售订单确认书
- 发货单/装箱单
- 销售发票
- 对账单

每个单据详情页增加 [打印] 按钮，打开打印预览页面（A4 格式、公司抬头、表格、签章区）。

使用 CSS `@media print` 实现，打印页面独立样式。

### 5.4 凭证模板配置化

当前 JournalEntry 的常用分录模板硬编码在组件里。改为后端配置：

```go
// 新增实体: erp_journal_templates
type ErpJournalTemplate struct {
    ID          string `gorm:"primaryKey"`
    Name        string // "销售收入"
    Description string
    Lines       string // JSON: [{account_code, debit_formula, credit_formula, description}]
    IsDefault   bool
    SortOrder   int
}
```

凭证录入页从后端加载模板列表，点击模板自动填充分录行。

### 5.5 审计日志

每个实体的详情页增加 "变更记录" 区块：

```
字段        旧值          新值          操作人    时间
status     draft        confirmed     张三    4/2 15:00
total      ¥1,003,550   ¥1,134,012   张三    4/2 14:45
```

需要：
- 后端：在 generic handler 的 update 方法中记录字段变更
- 前端：详情页底部 Chatter 区域增加变更日志 Tab
- 新表：`erp_audit_logs (entity, entity_id, field, old_value, new_value, user_id, created_at)`

---

## 六、UI 微调清单

| 页面 | 改进 | 原因 |
|------|------|------|
| ErpSalesPipeline | 管道汇总条添加动画（数字 count-up） | 增强数据变化感知 |
| ErpCustomer360 | 信用健康度条添加阈值线标记 | 直观显示安全/危险线 |
| ErpStockSearch | 搜索结果卡片添加"最后变动"时间 | 判断数据时效 |
| ErpProductionSchedule | 甘特图条添加今日线（红色竖线） | 直观看进度 |
| ErpJournalEntry | 科目下拉添加最近使用排序 | 减少搜索时间 |
| ErpReportCenter | 数字添加千分位分隔符 | 财务数据可读性 |
| ErpQualityConsole | SPC 图添加 UCL/LCL 控制限标注线 | 真正的 SPC |
| ErpCAPABoard | 卡片添加逾期天数红色角标 | 紧急度一目了然 |
| 所有详情页 | Chatter 添加 @提及功能 | 协作沟通 |
| 所有页面 | 空状态添加引导操作按钮 | 新用户引导 |

---

## 七、实施优先级

### Sprint 1 — P0 修复（让系统能正常工作）

1. 后端新增 3 个缺失 API（activity-log, sales-trend, mrp-run POST）
2. 后端新增 3 个缺失命令（update_shipment_status, extend_quotation, send_payment_reminder）
3. 前端 6 个页面改用已有后端 API（AR/Quality/Customer360）
4. 看板拖拽实现（SalesPipeline, ProductionSchedule, CAPABoard）

### Sprint 2 — P1 业务流程（让流程跑通）

1. 订单确认联动 MRP
2. 发票自动生成凭证
3. 库存事务自动联动（发货/领料/完工）
4. OQC 失败自动创建 NCR
5. 收款自动核销发票

### Sprint 3 — P1 性能 + P2 功能

1. 分页查询优化
2. 后端聚合 API（pipeline-summary, stock-alerts）
3. 现金流量表实现
4. 报表导出（PDF/Excel）
5. 打印模板
6. 审计日志

### Sprint 4 — UI 打磨

1. 看板拖拽动效
2. SPC 控制限
3. 凭证模板配置化
4. 空状态引导
5. UI 微调清单
