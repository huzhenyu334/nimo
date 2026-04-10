package erp

import sdk "github.com/bitfantasy/acp-module-sdk"

// ---------------------------------------------------------------------------
// erpCommands returns all command definitions for the ERP module.
// ---------------------------------------------------------------------------

func erpCommands() []sdk.CommandDef {
	return []sdk.CommandDef{
		// ── 销售管理 (9) ─────────────────────────────────────────────
		{Name: "create_quotation", Label: "创建报价单", Description: "为客户创建销售报价单", InputType: CreateQuotationInput{}, OutputType: QuotationOutput{}},
		{Name: "confirm_quotation", Label: "确认报价单", Description: "确认报价单，自动生成销售订单", InputType: ConfirmQuotationInput{}, OutputType: ConfirmQuotationOutput{}},
		{Name: "create_sales_order", Label: "创建销售订单", Description: "直接创建销售订单（不经过报价）", InputType: CreateSalesOrderInput{}, OutputType: SalesOrderOutput{}},
		{Name: "confirm_order", Label: "确认订单", Description: "确认销售订单（触发MRP计算）", InputType: ConfirmOrderInput{}, OutputType: ConfirmOrderOutput{}},
		{Name: "create_shipment", Label: "创建发货单", Description: "为销售订单创建发货单", InputType: CreateShipmentInput{}, OutputType: ShipmentOutput{}},
		{Name: "confirm_shipment", Label: "确认发货", Description: "确认发货并录入物流信息", InputType: ConfirmShipmentInput{}, OutputType: ShipmentOutput{}},
		{Name: "create_sales_invoice", Label: "创建销售发票", Description: "根据销售订单创建发票", InputType: CreateSalesInvoiceInput{}, OutputType: SalesInvoiceOutput{}},
		{Name: "record_receipt", Label: "录入收款", Description: "录入客户付款记录", InputType: RecordReceiptInput{}, OutputType: ReceiptOutput{}},
		{Name: "create_return", Label: "创建退货单", Description: "创建销售退货单（退款/换货/返修）", InputType: CreateReturnInput{}, OutputType: ReturnOutput{}},

		// ── 库存管理 (7) ─────────────────────────────────────────────
		{Name: "receive_inventory", Label: "入库", Description: "物料入库操作", InputType: ReceiveInventoryInput{}, OutputType: InventoryTransactionOutput{}},
		{Name: "issue_inventory", Label: "出库", Description: "物料出库操作", InputType: IssueInventoryInput{}, OutputType: InventoryTransactionOutput{}},
		{Name: "transfer_inventory", Label: "调拨", Description: "仓库间物料调拨", InputType: TransferInventoryInput{}, OutputType: InventoryTransactionOutput{}},
		{Name: "adjust_inventory", Label: "库存调整", Description: "盘点后库存数量调整", InputType: AdjustInventoryInput{}, OutputType: InventoryTransactionOutput{}},
		{Name: "scrap_inventory", Label: "报废", Description: "物料报废出库", InputType: ScrapInventoryInput{}, OutputType: InventoryTransactionOutput{}},
		{Name: "reserve_inventory", Label: "库存预留", Description: "为订单行预留库存", InputType: ReserveInventoryInput{}, OutputType: ReservationOutput{}},
		{Name: "unreserve_inventory", Label: "取消预留", Description: "取消订单行的库存预留", InputType: UnreserveInventoryInput{}, OutputType: ReservationOutput{}},

		// ── 生产管理 (7) ─────────────────────────────────────────────
		{Name: "run_mrp", Label: "运行MRP", Description: "运行物料需求计划计算", InputType: RunMRPInput{}, OutputType: RunMRPOutput{}},
		{Name: "confirm_mrp_suggestion", Label: "确认MRP建议", Description: "确认MRP建议，生成工单或采购申请", InputType: ConfirmMRPSuggestionInput{}, OutputType: ConfirmMRPSuggestionOutput{}},
		{Name: "create_work_order", Label: "创建工单", Description: "手动创建生产工单", InputType: CreateWorkOrderInput{}, OutputType: WorkOrderOutput{}},
		{Name: "release_work_order", Label: "下达工单", Description: "下达工单到生产车间", InputType: ReleaseWorkOrderInput{}, OutputType: WorkOrderOutput{}},
		{Name: "issue_wo_materials", Label: "工单领料", Description: "按BOM自动领料到工单", InputType: IssueWOMaterialsInput{}, OutputType: IssueWOMaterialsOutput{}},
		{Name: "report_wo_progress", Label: "报工", Description: "报告工单工序生产进度", InputType: ReportWOProgressInput{}, OutputType: WOProgressOutput{}},
		{Name: "complete_work_order", Label: "完工入库", Description: "工单完工并成品入库", InputType: CompleteWorkOrderInput{}, OutputType: WorkOrderOutput{}},

		// ── 财务管理 (6) ─────────────────────────────────────────────
		{Name: "create_journal_entry", Label: "创建凭证", Description: "创建会计记账凭证", InputType: CreateJournalEntryInput{}, OutputType: JournalEntryOutput{}},
		{Name: "post_journal_entry", Label: "过账凭证", Description: "凭证过账（不可逆）", InputType: PostJournalEntryInput{}, OutputType: JournalEntryOutput{}},
		{Name: "reverse_journal_entry", Label: "冲销凭证", Description: "生成红字冲销凭证", InputType: ReverseJournalEntryInput{}, OutputType: ReverseJournalEntryOutput{}},
		{Name: "close_period", Label: "关闭会计期间", Description: "关闭指定会计期间", InputType: ClosePeriodInput{}, OutputType: ClosePeriodOutput{}},
		{Name: "generate_report", Label: "生成财务报表", Description: "生成试算平衡表/利润表/资产负债表", InputType: GenerateReportInput{}, OutputType: GenerateReportOutput{}},
		{Name: "post_ap_from_settlement", Label: "结算转AP凭证", Description: "从 SRM 结算单创建应付账款会计凭证(DR 应付账款 / CR 银行存款)", InputType: PostAPFromSettlementInput{}, OutputType: PostAPFromSettlementOutput{}},

		// ── 质量管理 (6) ─────────────────────────────────────────────
		{Name: "create_oqc", Label: "创建OQC检验", Description: "创建出货质量检验单", InputType: CreateOQCInput{}, OutputType: OQCOutput{}},
		{Name: "complete_oqc", Label: "完成OQC检验", Description: "录入OQC检验结果", InputType: CompleteOQCInput{}, OutputType: OQCOutput{}},
		{Name: "create_ncr", Label: "创建NCR", Description: "创建不合格品报告", InputType: CreateNCRInput{}, OutputType: NCROutput{}},
		{Name: "disposition_ncr", Label: "NCR处置", Description: "对不合格品做出处置决定", InputType: DispositionNCRInput{}, OutputType: NCROutput{}},
		{Name: "create_capa", Label: "创建CAPA", Description: "创建纠正/预防措施", InputType: CreateCAPAInput{}, OutputType: CAPAOutput{}},
		{Name: "close_capa", Label: "关闭CAPA", Description: "验证并关闭CAPA", InputType: CloseCAPAInput{}, OutputType: CAPAOutput{}},

		// ── 补充命令 (3) ─────────────────────────────────────────────
		{Name: "update_shipment_status", Label: "更新发货状态", Description: "更新发货单状态（shipped/delivered等）", InputType: UpdateShipmentStatusInput{}, OutputType: UpdateShipmentStatusOutput{}},
		{Name: "extend_quotation", Label: "延长报价有效期", Description: "延长报价单有效期指定天数", InputType: ExtendQuotationInput{}, OutputType: ExtendQuotationOutput{}},
		{Name: "send_payment_reminder", Label: "发送催款通知", Description: "对逾期发票发送催款通知", InputType: SendPaymentReminderInput{}, OutputType: SendPaymentReminderOutput{}},
	}
}

// ===========================================================================
// Input / Output types
// ===========================================================================

// ---------------------------------------------------------------------------
// 通用输出
// ---------------------------------------------------------------------------

type DeleteOutput struct {
	ID      string `json:"id" desc:"被删除记录的ID"`
	Deleted bool   `json:"deleted" desc:"是否删除成功"`
}

// ---------------------------------------------------------------------------
// 销售管理 — 报价单
// ---------------------------------------------------------------------------

type QuotationItemInput struct {
	ProductID   string  `json:"product_id" desc:"产品ID"`
	Quantity    float64 `json:"quantity" desc:"数量"`
	UnitPrice   float64 `json:"unit_price" desc:"单价"`
	DiscountPct float64 `json:"discount_pct,omitempty" desc:"折扣百分比(0-100)"`
	TaxRate     float64 `json:"tax_rate,omitempty" desc:"税率(0-100)"`
}

type CreateQuotationInput struct {
	CustomerID string               `json:"customer_id" desc:"客户ID"`
	Items      []QuotationItemInput `json:"items" desc:"报价明细行"`
}

type QuotationOutput struct {
	QuotationID string  `json:"quotation_id" desc:"报价单ID"`
	Code        string  `json:"code" desc:"报价单编号"`
	CustomerID  string  `json:"customer_id" desc:"客户ID"`
	Status      string  `json:"status" desc:"状态: draft/confirmed/cancelled"`
	TotalAmount float64 `json:"total_amount" desc:"总金额"`
	URL         string  `json:"url" desc:"报价单链接"`
}

type ConfirmQuotationInput struct {
	QuotationID string `json:"quotation_id" desc:"报价单ID"`
}

type ConfirmQuotationOutput struct {
	QuotationID string `json:"quotation_id" desc:"报价单ID"`
	OrderID     string `json:"order_id" desc:"自动生成的销售订单ID"`
	OrderCode   string `json:"order_code" desc:"销售订单编号"`
}

// ---------------------------------------------------------------------------
// 销售管理 — 销售订单
// ---------------------------------------------------------------------------

type SalesOrderItemInput struct {
	ProductID   string  `json:"product_id" desc:"产品ID"`
	Quantity    float64 `json:"quantity" desc:"数量"`
	UnitPrice   float64 `json:"unit_price" desc:"单价"`
	DiscountPct float64 `json:"discount_pct,omitempty" desc:"折扣百分比(0-100)"`
	TaxRate     float64 `json:"tax_rate,omitempty" desc:"税率(0-100)"`
}

type CreateSalesOrderInput struct {
	CustomerID      string                `json:"customer_id" desc:"客户ID"`
	Items           []SalesOrderItemInput `json:"items" desc:"订单明细行"`
	ExpectedDate    string                `json:"expected_date,omitempty" desc:"期望交货日期 YYYY-MM-DD"`
	ShippingAddress string                `json:"shipping_address,omitempty" desc:"收货地址"`
	PaymentTerms    string                `json:"payment_terms,omitempty" desc:"付款条件" enum:"prepaid,net30,net60,cod"`
}

type SalesOrderOutput struct {
	OrderID     string  `json:"order_id" desc:"订单ID"`
	Code        string  `json:"code" desc:"订单编号"`
	CustomerID  string  `json:"customer_id" desc:"客户ID"`
	Status      string  `json:"status" desc:"状态: draft/confirmed/in_production/shipped/completed/cancelled"`
	TotalAmount float64 `json:"total_amount" desc:"总金额"`
	URL         string  `json:"url" desc:"订单链接"`
}

type ConfirmOrderInput struct {
	OrderID string `json:"order_id" desc:"销售订单ID"`
}

type ConfirmOrderOutput struct {
	OrderID string `json:"order_id" desc:"销售订单ID"`
	Status  string `json:"status" desc:"确认后状态"`
}

// ---------------------------------------------------------------------------
// 销售管理 — 发货
// ---------------------------------------------------------------------------

type ShipmentItemInput struct {
	OrderItemID   string   `json:"order_item_id" desc:"订单行ID"`
	Quantity      float64  `json:"quantity" desc:"发货数量"`
	LotNumber     string   `json:"lot_number,omitempty" desc:"批次号"`
	SerialNumbers []string `json:"serial_numbers,omitempty" desc:"序列号列表"`
}

type CreateShipmentInput struct {
	OrderID string              `json:"order_id" desc:"销售订单ID"`
	Items   []ShipmentItemInput `json:"items" desc:"发货明细行"`
}

type ShipmentOutput struct {
	ShipmentID string `json:"shipment_id" desc:"发货单ID"`
	Code       string `json:"code" desc:"发货单编号"`
	OrderID    string `json:"order_id" desc:"关联订单ID"`
	Status     string `json:"status" desc:"状态: draft/confirmed/shipped/delivered"`
	Carrier    string `json:"carrier" desc:"物流商"`
	TrackingNo string `json:"tracking_no" desc:"物流单号"`
	URL        string `json:"url" desc:"发货单链接"`
}

type ConfirmShipmentInput struct {
	ShipmentID string `json:"shipment_id" desc:"发货单ID"`
	Carrier    string `json:"carrier" desc:"物流商"`
	TrackingNo string `json:"tracking_no" desc:"物流单号"`
}

// ---------------------------------------------------------------------------
// 销售管理 — 发票
// ---------------------------------------------------------------------------

type CreateSalesInvoiceInput struct {
	OrderID string `json:"order_id" desc:"销售订单ID"`
}

type SalesInvoiceOutput struct {
	InvoiceID   string  `json:"invoice_id" desc:"发票ID"`
	Code        string  `json:"code" desc:"发票编号"`
	OrderID     string  `json:"order_id" desc:"关联订单ID"`
	Status      string  `json:"status" desc:"状态: draft/posted/paid/cancelled"`
	TotalAmount float64 `json:"total_amount" desc:"发票金额"`
	URL         string  `json:"url" desc:"发票链接"`
}

// ---------------------------------------------------------------------------
// 销售管理 — 收款
// ---------------------------------------------------------------------------

type ReceiptInvoiceInput struct {
	InvoiceID string  `json:"invoice_id" desc:"发票ID"`
	Amount    float64 `json:"amount" desc:"本次核销金额"`
}

type RecordReceiptInput struct {
	CustomerID    string                `json:"customer_id" desc:"客户ID"`
	Amount        float64               `json:"amount" desc:"收款金额"`
	PaymentMethod string                `json:"payment_method" desc:"付款方式" enum:"bank_transfer,cash,check,online"`
	Invoices      []ReceiptInvoiceInput `json:"invoices,omitempty" desc:"核销发票列表"`
}

type ReceiptOutput struct {
	ReceiptID     string  `json:"receipt_id" desc:"收款记录ID"`
	Code          string  `json:"code" desc:"收款编号"`
	CustomerID    string  `json:"customer_id" desc:"客户ID"`
	Amount        float64 `json:"amount" desc:"收款金额"`
	PaymentMethod string  `json:"payment_method" desc:"付款方式"`
	Status        string  `json:"status" desc:"状态: posted"`
}

// ---------------------------------------------------------------------------
// 销售管理 — 退货
// ---------------------------------------------------------------------------

type ReturnItemInput struct {
	OrderItemID string  `json:"order_item_id" desc:"订单行ID"`
	Quantity    float64 `json:"quantity" desc:"退货数量"`
	Reason      string  `json:"reason" desc:"退货原因"`
}

type CreateReturnInput struct {
	OrderID string            `json:"order_id" desc:"销售订单ID"`
	Items   []ReturnItemInput `json:"items" desc:"退货明细行"`
	Type    string            `json:"type" desc:"退货类型" enum:"refund,exchange,repair"`
}

type ReturnOutput struct {
	ReturnID string `json:"return_id" desc:"退货单ID"`
	Code     string `json:"code" desc:"退货单编号"`
	OrderID  string `json:"order_id" desc:"关联订单ID"`
	Type     string `json:"type" desc:"退货类型: refund/exchange/repair"`
	Status   string `json:"status" desc:"状态: draft/approved/received/completed"`
	URL      string `json:"url" desc:"退货单链接"`
}

// ---------------------------------------------------------------------------
// 库存管理 — 入库
// ---------------------------------------------------------------------------

type ReceiveInventoryInput struct {
	MaterialID    string  `json:"material_id" desc:"物料ID"`
	WarehouseID   string  `json:"warehouse_id" desc:"仓库ID"`
	LocationID    string  `json:"location_id,omitempty" desc:"库位ID"`
	Quantity      float64 `json:"quantity" desc:"入库数量"`
	LotNumber     string  `json:"lot_number,omitempty" desc:"批次号"`
	UnitCost      float64 `json:"unit_cost,omitempty" desc:"单位成本"`
	ReferenceType string  `json:"reference_type,omitempty" desc:"来源类型" enum:"purchase_order,work_order,return,manual"`
	ReferenceID   string  `json:"reference_id,omitempty" desc:"来源单据ID"`
}

type InventoryTransactionOutput struct {
	TransactionID string  `json:"transaction_id" desc:"库存事务ID"`
	MaterialID    string  `json:"material_id" desc:"物料ID"`
	WarehouseID   string  `json:"warehouse_id" desc:"仓库ID"`
	Type          string  `json:"type" desc:"事务类型: receive/issue/transfer/adjust/scrap"`
	Quantity      float64 `json:"quantity" desc:"事务数量"`
	OnHandAfter   float64 `json:"on_hand_after" desc:"事务后库存数量"`
}

// ---------------------------------------------------------------------------
// 库存管理 — 出库
// ---------------------------------------------------------------------------

type IssueInventoryInput struct {
	MaterialID    string  `json:"material_id" desc:"物料ID"`
	WarehouseID   string  `json:"warehouse_id" desc:"仓库ID"`
	Quantity      float64 `json:"quantity" desc:"出库数量"`
	LotNumber     string  `json:"lot_number,omitempty" desc:"批次号"`
	ReferenceType string  `json:"reference_type,omitempty" desc:"来源类型" enum:"sales_order,work_order,manual"`
	ReferenceID   string  `json:"reference_id,omitempty" desc:"来源单据ID"`
}

// ---------------------------------------------------------------------------
// 库存管理 — 调拨
// ---------------------------------------------------------------------------

type TransferInventoryInput struct {
	MaterialID      string  `json:"material_id" desc:"物料ID"`
	FromWarehouseID string  `json:"from_warehouse_id" desc:"源仓库ID"`
	ToWarehouseID   string  `json:"to_warehouse_id" desc:"目标仓库ID"`
	Quantity        float64 `json:"quantity" desc:"调拨数量"`
	LotNumber       string  `json:"lot_number,omitempty" desc:"批次号"`
}

// ---------------------------------------------------------------------------
// 库存管理 — 调整
// ---------------------------------------------------------------------------

type AdjustInventoryInput struct {
	MaterialID  string  `json:"material_id" desc:"物料ID"`
	WarehouseID string  `json:"warehouse_id" desc:"仓库ID"`
	ActualQty   float64 `json:"actual_qty" desc:"实际盘点数量"`
	Reason      string  `json:"reason" desc:"调整原因"`
}

// ---------------------------------------------------------------------------
// 库存管理 — 报废
// ---------------------------------------------------------------------------

type ScrapInventoryInput struct {
	MaterialID  string  `json:"material_id" desc:"物料ID"`
	WarehouseID string  `json:"warehouse_id" desc:"仓库ID"`
	Quantity    float64 `json:"quantity" desc:"报废数量"`
	Reason      string  `json:"reason" desc:"报废原因"`
}

// ---------------------------------------------------------------------------
// 库存管理 — 预留
// ---------------------------------------------------------------------------

type ReserveInventoryInput struct {
	OrderItemID string  `json:"order_item_id" desc:"订单行ID"`
	MaterialID  string  `json:"material_id" desc:"物料ID"`
	WarehouseID string  `json:"warehouse_id" desc:"仓库ID"`
	Quantity    float64 `json:"quantity" desc:"预留数量"`
}

type ReservationOutput struct {
	ReservationID string  `json:"reservation_id" desc:"预留ID"`
	OrderItemID   string  `json:"order_item_id" desc:"订单行ID"`
	MaterialID    string  `json:"material_id" desc:"物料ID"`
	WarehouseID   string  `json:"warehouse_id" desc:"仓库ID"`
	Quantity      float64 `json:"quantity" desc:"预留数量"`
	Status        string  `json:"status" desc:"状态: reserved/released"`
}

type UnreserveInventoryInput struct {
	OrderItemID string `json:"order_item_id" desc:"订单行ID"`
}

// ---------------------------------------------------------------------------
// 生产管理 — MRP
// ---------------------------------------------------------------------------

type RunMRPInput struct {
	DemandSource string `json:"demand_source" desc:"需求来源" enum:"so,forecast"`
}

type RunMRPOutput struct {
	RunID           string `json:"run_id" desc:"MRP运行ID"`
	SuggestionCount int    `json:"suggestion_count" desc:"生成建议数量"`
	Status          string `json:"status" desc:"运行状态: completed"`
}

type ConfirmMRPSuggestionInput struct {
	SuggestionID string `json:"suggestion_id" desc:"MRP建议ID"`
}

type ConfirmMRPSuggestionOutput struct {
	SuggestionID      string `json:"suggestion_id" desc:"MRP建议ID"`
	ResultType        string `json:"result_type" desc:"结果类型: work_order/purchase_request"`
	WorkOrderID       string `json:"work_order_id,omitempty" desc:"生成的工单ID（生产建议时）"`
	PurchaseRequestID string `json:"purchase_request_id,omitempty" desc:"生成的采购申请ID（采购建议时）"`
}

// ---------------------------------------------------------------------------
// 生产管理 — 工单
// ---------------------------------------------------------------------------

type CreateWorkOrderInput struct {
	ProductID    string  `json:"product_id" desc:"产品ID"`
	BOMID        string  `json:"bom_id" desc:"BOM ID"`
	PlannedQty   float64 `json:"planned_qty" desc:"计划生产数量"`
	PlannedStart string  `json:"planned_start" desc:"计划开始日期 YYYY-MM-DD"`
	PlannedEnd   string  `json:"planned_end" desc:"计划结束日期 YYYY-MM-DD"`
	WarehouseID  string  `json:"warehouse_id" desc:"入库仓库ID"`
}

type WorkOrderOutput struct {
	WorkOrderID  string  `json:"work_order_id" desc:"工单ID"`
	Code         string  `json:"code" desc:"工单编号"`
	ProductID    string  `json:"product_id" desc:"产品ID"`
	Status       string  `json:"status" desc:"状态: draft/released/in_progress/completed/cancelled"`
	PlannedQty   float64 `json:"planned_qty" desc:"计划数量"`
	CompletedQty float64 `json:"completed_qty" desc:"完成数量"`
	URL          string  `json:"url" desc:"工单链接"`
}

type ReleaseWorkOrderInput struct {
	WorkOrderID string `json:"work_order_id" desc:"工单ID"`
}

// ---------------------------------------------------------------------------
// 生产管理 — 工单领料
// ---------------------------------------------------------------------------

type IssueWOMaterialsInput struct {
	WorkOrderID string `json:"work_order_id" desc:"工单ID"`
}

type WOMaterialTransaction struct {
	MaterialID    string  `json:"material_id" desc:"物料ID"`
	Quantity      float64 `json:"quantity" desc:"领料数量"`
	TransactionID string  `json:"transaction_id" desc:"库存事务ID"`
}

type IssueWOMaterialsOutput struct {
	WorkOrderID  string                  `json:"work_order_id" desc:"工单ID"`
	Transactions []WOMaterialTransaction `json:"transactions" desc:"领料事务列表"`
}

// ---------------------------------------------------------------------------
// 生产管理 — 报工
// ---------------------------------------------------------------------------

type ReportWOProgressInput struct {
	WorkOrderID string  `json:"work_order_id" desc:"工单ID"`
	Operation   string  `json:"operation" desc:"工序名称"`
	GoodQty     float64 `json:"good_qty" desc:"良品数量"`
	DefectQty   float64 `json:"defect_qty,omitempty" desc:"缺陷品数量"`
	ScrapQty    float64 `json:"scrap_qty,omitempty" desc:"报废数量"`
}

type WOProgressOutput struct {
	WorkOrderID string  `json:"work_order_id" desc:"工单ID"`
	Operation   string  `json:"operation" desc:"工序名称"`
	GoodQty     float64 `json:"good_qty" desc:"良品数量"`
	DefectQty   float64 `json:"defect_qty" desc:"缺陷品数量"`
	ScrapQty    float64 `json:"scrap_qty" desc:"报废数量"`
	YieldRate   float64 `json:"yield_rate" desc:"良率百分比"`
}

// ---------------------------------------------------------------------------
// 生产管理 — 完工入库
// ---------------------------------------------------------------------------

type CompleteWorkOrderInput struct {
	WorkOrderID   string   `json:"work_order_id" desc:"工单ID"`
	CompletedQty  float64  `json:"completed_qty" desc:"完工数量"`
	SerialNumbers []string `json:"serial_numbers,omitempty" desc:"序列号列表"`
}

// ---------------------------------------------------------------------------
// 财务管理 — 凭证
// ---------------------------------------------------------------------------

type JournalLineInput struct {
	AccountID   string  `json:"account_id" desc:"科目ID"`
	Debit       float64 `json:"debit,omitempty" desc:"借方金额"`
	Credit      float64 `json:"credit,omitempty" desc:"贷方金额"`
	Description string  `json:"description,omitempty" desc:"摘要"`
}

type CreateJournalEntryInput struct {
	EntryDate   string             `json:"entry_date" desc:"记账日期 YYYY-MM-DD"`
	Description string             `json:"description" desc:"凭证摘要"`
	Lines       []JournalLineInput `json:"lines" desc:"凭证分录行"`
}

type JournalEntryOutput struct {
	EntryID     string  `json:"entry_id" desc:"凭证ID"`
	Code        string  `json:"code" desc:"凭证编号"`
	EntryDate   string  `json:"entry_date" desc:"记账日期"`
	Status      string  `json:"status" desc:"状态: draft/posted/reversed"`
	TotalDebit  float64 `json:"total_debit" desc:"借方合计"`
	TotalCredit float64 `json:"total_credit" desc:"贷方合计"`
	URL         string  `json:"url" desc:"凭证链接"`
}

type PostJournalEntryInput struct {
	EntryID string `json:"entry_id" desc:"凭证ID"`
}

type ReverseJournalEntryInput struct {
	EntryID string `json:"entry_id" desc:"凭证ID"`
}

type ReverseJournalEntryOutput struct {
	OriginalEntryID string `json:"original_entry_id" desc:"原凭证ID"`
	NewEntryID      string `json:"new_entry_id" desc:"冲销凭证ID"`
	NewEntryCode    string `json:"new_entry_code" desc:"冲销凭证编号"`
}

type PostAPFromSettlementInput struct {
	SettlementID string `json:"settlement_id" desc:"SRM 结算单ID"`
	PostedBy     string `json:"posted_by,omitempty" desc:"过账人"`
}

type PostAPFromSettlementOutput struct {
	JournalEntryID string  `json:"journal_entry_id" desc:"生成的会计凭证ID"`
	JournalCode    string  `json:"journal_code" desc:"凭证编号"`
	SettlementID   string  `json:"settlement_id" desc:"源 SRM 结算单ID"`
	Amount         float64 `json:"amount" desc:"过账金额"`
}

// ---------------------------------------------------------------------------
// 财务管理 — 期间
// ---------------------------------------------------------------------------

type ClosePeriodInput struct {
	Period string `json:"period" desc:"会计期间 (如 2026-04)"`
}

type ClosePeriodOutput struct {
	Period string `json:"period" desc:"会计期间"`
	Status string `json:"status" desc:"状态: closed"`
}

// ---------------------------------------------------------------------------
// 财务管理 — 报表
// ---------------------------------------------------------------------------

type GenerateReportInput struct {
	ReportType string `json:"report_type" desc:"报表类型" enum:"trial_balance,income_statement,balance_sheet"`
	Period     string `json:"period" desc:"报表期间 (如 2026-04)"`
}

type GenerateReportOutput struct {
	ReportID   string `json:"report_id" desc:"报表ID"`
	ReportType string `json:"report_type" desc:"报表类型"`
	Period     string `json:"period" desc:"报表期间"`
	URL        string `json:"url" desc:"报表链接"`
}

// ---------------------------------------------------------------------------
// 质量管理 — OQC
// ---------------------------------------------------------------------------

type CreateOQCInput struct {
	ShipmentID string `json:"shipment_id" desc:"发货单ID"`
	ProductID  string `json:"product_id" desc:"产品ID"`
	SampleSize int    `json:"sample_size" desc:"抽样数量"`
}

type OQCOutput struct {
	OQCID      string `json:"oqc_id" desc:"OQC检验单ID"`
	Code       string `json:"code" desc:"检验单编号"`
	ShipmentID string `json:"shipment_id" desc:"关联发货单ID"`
	ProductID  string `json:"product_id" desc:"产品ID"`
	SampleSize int    `json:"sample_size" desc:"抽样数量"`
	Result     string `json:"result" desc:"检验结果: pending/pass/fail/conditional"`
	PassCount  int    `json:"pass_count" desc:"合格数"`
	FailCount  int    `json:"fail_count" desc:"不合格数"`
	URL        string `json:"url" desc:"检验单链接"`
}

type CompleteOQCInput struct {
	OQCID         string `json:"oqc_id" desc:"OQC检验单ID"`
	Result        string `json:"result" desc:"检验结果" enum:"pass,fail,conditional"`
	PassCount     int    `json:"pass_count" desc:"合格数"`
	FailCount     int    `json:"fail_count" desc:"不合格数"`
	DefectDetails string `json:"defect_details,omitempty" desc:"缺陷详情"`
}

// ---------------------------------------------------------------------------
// 质量管理 — NCR
// ---------------------------------------------------------------------------

type CreateNCRInput struct {
	Source      string  `json:"source" desc:"来源" enum:"iqc,oqc,customer_return,internal"`
	SourceID    string  `json:"source_id,omitempty" desc:"来源单据ID"`
	ProductID   string  `json:"product_id,omitempty" desc:"产品ID"`
	MaterialID  string  `json:"material_id,omitempty" desc:"物料ID"`
	DefectType  string  `json:"defect_type" desc:"缺陷类型"`
	DefectQty   float64 `json:"defect_qty" desc:"缺陷数量"`
	Description string  `json:"description" desc:"缺陷描述"`
	Severity    string  `json:"severity" desc:"严重程度" enum:"critical,major,minor"`
}

type NCROutput struct {
	NCRID       string  `json:"ncr_id" desc:"NCR ID"`
	Code        string  `json:"code" desc:"NCR编号"`
	Source      string  `json:"source" desc:"来源"`
	DefectType  string  `json:"defect_type" desc:"缺陷类型"`
	DefectQty   float64 `json:"defect_qty" desc:"缺陷数量"`
	Severity    string  `json:"severity" desc:"严重程度"`
	Status      string  `json:"status" desc:"状态: open/dispositioned/closed"`
	Disposition string  `json:"disposition" desc:"处置方式"`
	URL         string  `json:"url" desc:"NCR链接"`
}

type DispositionNCRInput struct {
	NCRID       string `json:"ncr_id" desc:"NCR ID"`
	Disposition string `json:"disposition" desc:"处置方式" enum:"use_as_is,rework,scrap,return_to_supplier"`
}

// ---------------------------------------------------------------------------
// 质量管理 — CAPA
// ---------------------------------------------------------------------------

type CreateCAPAInput struct {
	NCRID      string `json:"ncr_id" desc:"关联NCR ID"`
	Type       string `json:"type" desc:"类型" enum:"corrective,preventive"`
	Title      string `json:"title" desc:"标题"`
	RootCause  string `json:"root_cause" desc:"根本原因"`
	ActionPlan string `json:"action_plan" desc:"行动计划"`
	DueDate    string `json:"due_date" desc:"截止日期 YYYY-MM-DD"`
}

type CAPAOutput struct {
	CAPAID       string `json:"capa_id" desc:"CAPA ID"`
	Code         string `json:"code" desc:"CAPA编号"`
	NCRID        string `json:"ncr_id" desc:"关联NCR ID"`
	Type         string `json:"type" desc:"类型: corrective/preventive"`
	Status       string `json:"status" desc:"状态: open/in_progress/closed"`
	Title        string `json:"title" desc:"标题"`
	DueDate      string `json:"due_date" desc:"截止日期"`
	Verification string `json:"verification" desc:"验证结果"`
	URL          string `json:"url" desc:"CAPA链接"`
}

type CloseCAPAInput struct {
	CAPAID       string `json:"capa_id" desc:"CAPA ID"`
	Verification string `json:"verification" desc:"验证结果说明"`
}

// ---------------------------------------------------------------------------
// 补充命令 — 发货状态更新
// ---------------------------------------------------------------------------

type UpdateShipmentStatusInput struct {
	ShipmentID string `json:"shipment_id" desc:"发货单ID"`
	Status     string `json:"status" desc:"新状态" enum:"shipped,delivered,cancelled"`
}

type UpdateShipmentStatusOutput struct {
	ShipmentID string `json:"shipment_id" desc:"发货单ID"`
	Status     string `json:"status" desc:"更新后状态"`
}

// ---------------------------------------------------------------------------
// 补充命令 — 延长报价有效期
// ---------------------------------------------------------------------------

type ExtendQuotationInput struct {
	QuotationID string `json:"quotation_id" desc:"报价单ID"`
	Days        int    `json:"days,omitempty" desc:"延长天数（默认7天）"`
}

type ExtendQuotationOutput struct {
	QuotationID string `json:"quotation_id" desc:"报价单ID"`
	ValidUntil  string `json:"valid_until" desc:"新的有效截止日期 YYYY-MM-DD"`
}

// ---------------------------------------------------------------------------
// 补充命令 — 发送催款通知
// ---------------------------------------------------------------------------

type SendPaymentReminderInput struct {
	InvoiceID string `json:"invoice_id" desc:"发票ID"`
}

type SendPaymentReminderOutput struct {
	InvoiceID    string `json:"invoice_id" desc:"发票ID"`
	ReminderSent bool   `json:"reminder_sent" desc:"是否发送成功"`
}
