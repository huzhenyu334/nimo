package erp

import "time"

// ---------------------------------------------------------------------------
// GORM Models — ERP Module database tables
// All tables use "erp_" prefix to isolate from ACP core tables.
// ---------------------------------------------------------------------------

// ── Base Data ──

// ErpCustomer 客户主数据
type ErpCustomer struct {
	ID              string    `gorm:"primaryKey;size:100" json:"id"`
	Code            string    `gorm:"size:50;uniqueIndex" json:"code"`
	Name            string    `gorm:"size:200;not null" json:"name"`
	Type            string    `gorm:"size:30" json:"type"`
	TaxID           string    `gorm:"size:50" json:"tax_id"`
	ContactName     string    `gorm:"size:100" json:"contact_name"`
	ContactPhone    string    `gorm:"size:50" json:"contact_phone"`
	ContactEmail    string    `gorm:"size:200" json:"contact_email"`
	BillingAddress  string    `gorm:"type:text" json:"billing_address"`
	ShippingAddress string    `gorm:"type:text" json:"shipping_address"`
	PaymentTerms    string    `gorm:"size:50" json:"payment_terms"`
	CreditLimit     float64   `json:"credit_limit"`
	Currency        string    `gorm:"size:10;default:CNY" json:"currency"`
	Status          string    `gorm:"size:30;default:active" json:"status"`
	Tags            string    `gorm:"type:jsonb;default:'[]'" json:"tags"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ErpCustomer) TableName() string { return "erp_customers" }

// ErpWarehouse 仓库
type ErpWarehouse struct {
	ID        string    `gorm:"primaryKey;size:100" json:"id"`
	Code      string    `gorm:"size:50;uniqueIndex" json:"code"`
	Name      string    `gorm:"size:200;not null" json:"name"`
	Type      string    `gorm:"size:30" json:"type"`
	Address   string    `gorm:"type:text" json:"address"`
	ManagerID string    `gorm:"size:100" json:"manager_id"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	Status    string    `gorm:"size:30;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ErpWarehouse) TableName() string { return "erp_warehouses" }

// ErpLocation 仓库库位
type ErpLocation struct {
	ID          string    `gorm:"primaryKey;size:100" json:"id"`
	WarehouseID string    `gorm:"size:100;not null;index" json:"warehouse_id"`
	Code        string    `gorm:"size:50;uniqueIndex" json:"code"`
	Name        string    `gorm:"size:200" json:"name"`
	Type        string    `gorm:"size:30" json:"type"`
	Capacity    int       `json:"capacity"`
	Status      string    `gorm:"size:30;default:active" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ErpLocation) TableName() string { return "erp_locations" }

// ── Sales Management ──

// ErpQuotation 销售报价单
type ErpQuotation struct {
	ID           string     `gorm:"primaryKey;size:100" json:"id"`
	Code         string     `gorm:"size:50;uniqueIndex" json:"code"`
	CustomerID   string     `gorm:"size:100;not null;index" json:"customer_id"`
	ContactName  string     `gorm:"size:100" json:"contact_name"`
	Currency     string     `gorm:"size:10;default:CNY" json:"currency"`
	ExchangeRate float64    `gorm:"default:1" json:"exchange_rate"`
	Subtotal     float64    `json:"subtotal"`
	TaxAmount    float64    `json:"tax_amount"`
	Total        float64    `json:"total"`
	ValidUntil   *time.Time `json:"valid_until"`
	PaymentTerms string     `gorm:"size:50" json:"payment_terms"`
	Notes        string     `gorm:"type:text" json:"notes"`
	Status       string     `gorm:"size:30;default:draft" json:"status"`
	CreatedBy    string     `gorm:"size:100" json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (ErpQuotation) TableName() string { return "erp_quotations" }

// ErpQuotationItem 报价单行项
type ErpQuotationItem struct {
	ID          string    `gorm:"primaryKey;size:100" json:"id"`
	QuotationID string    `gorm:"size:100;not null;index" json:"quotation_id"`
	ProductID   string    `gorm:"size:100;index" json:"product_id"`
	SkuID       string    `gorm:"size:100" json:"sku_id"`
	Description string    `gorm:"type:text" json:"description"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	DiscountPct float64   `gorm:"default:0" json:"discount_pct"`
	TaxRate     float64   `gorm:"default:0" json:"tax_rate"`
	LineTotal   float64   `json:"line_total"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ErpQuotationItem) TableName() string { return "erp_quotation_items" }

// ErpSalesOrder 销售订单
type ErpSalesOrder struct {
	ID              string     `gorm:"primaryKey;size:100" json:"id"`
	Code            string     `gorm:"size:50;uniqueIndex" json:"code"`
	QuotationID     string     `gorm:"size:100;index" json:"quotation_id"`
	CustomerID      string     `gorm:"size:100;not null;index" json:"customer_id"`
	ShippingAddress string     `gorm:"type:text" json:"shipping_address"`
	Currency        string     `gorm:"size:10;default:CNY" json:"currency"`
	Subtotal        float64    `json:"subtotal"`
	TaxAmount       float64    `json:"tax_amount"`
	Total           float64    `json:"total"`
	PaymentTerms    string     `gorm:"size:50" json:"payment_terms"`
	ExpectedDate    *time.Time `json:"expected_date"`
	ShippingMethod  string     `gorm:"size:50" json:"shipping_method"`
	Priority        string     `gorm:"size:20;default:normal" json:"priority"`
	Status          string     `gorm:"size:30;default:draft" json:"status"`
	Notes           string     `gorm:"type:text" json:"notes"`
	CreatedBy       string     `gorm:"size:100" json:"created_by"`
	ConfirmedAt     *time.Time `json:"confirmed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (ErpSalesOrder) TableName() string { return "erp_sales_orders" }

// ErpSalesOrderItem 销售订单行项
type ErpSalesOrderItem struct {
	ID           string     `gorm:"primaryKey;size:100" json:"id"`
	OrderID      string     `gorm:"size:100;not null;index" json:"order_id"`
	ProductID    string     `gorm:"size:100;index" json:"product_id"`
	SkuID        string     `gorm:"size:100" json:"sku_id"`
	BomID        string     `gorm:"size:100" json:"bom_id"`
	Description  string     `gorm:"type:text" json:"description"`
	Quantity     float64    `json:"quantity"`
	DeliveredQty float64    `gorm:"default:0" json:"delivered_qty"`
	UnitPrice    float64    `json:"unit_price"`
	DiscountPct  float64    `gorm:"default:0" json:"discount_pct"`
	TaxRate      float64    `gorm:"default:0" json:"tax_rate"`
	LineTotal    float64    `json:"line_total"`
	ExpectedDate *time.Time `json:"expected_date"`
	Status       string     `gorm:"size:30;default:pending" json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (ErpSalesOrderItem) TableName() string { return "erp_sales_order_items" }

// ── Shipment & Delivery ──

// ErpShipment 发货单
type ErpShipment struct {
	ID              string     `gorm:"primaryKey;size:100" json:"id"`
	Code            string     `gorm:"size:50;uniqueIndex" json:"code"`
	OrderID         string     `gorm:"size:100;not null;index" json:"order_id"`
	WarehouseID     string     `gorm:"size:100;index" json:"warehouse_id"`
	ShippingAddress string     `gorm:"type:text" json:"shipping_address"`
	Carrier         string     `gorm:"size:100" json:"carrier"`
	TrackingNo      string     `gorm:"size:100" json:"tracking_no"`
	ShippedAt       *time.Time `json:"shipped_at"`
	DeliveredAt     *time.Time `json:"delivered_at"`
	Status          string     `gorm:"size:30;default:pending" json:"status"`
	Notes           string     `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (ErpShipment) TableName() string { return "erp_shipments" }

// ErpShipmentItem 发货单行项
type ErpShipmentItem struct {
	ID            string    `gorm:"primaryKey;size:100" json:"id"`
	ShipmentID    string    `gorm:"size:100;not null;index" json:"shipment_id"`
	OrderItemID   string    `gorm:"size:100;index" json:"order_item_id"`
	ProductID     string    `gorm:"size:100;index" json:"product_id"`
	Quantity      float64   `json:"quantity"`
	LotNumber     string    `gorm:"size:100" json:"lot_number"`
	SerialNumbers string    `gorm:"type:text" json:"serial_numbers"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (ErpShipmentItem) TableName() string { return "erp_shipment_items" }

// ErpReturn 退货单 (RMA)
type ErpReturn struct {
	ID          string    `gorm:"primaryKey;size:100" json:"id"`
	Code        string    `gorm:"size:50;uniqueIndex" json:"code"`
	OrderID     string    `gorm:"size:100;not null;index" json:"order_id"`
	CustomerID  string    `gorm:"size:100;index" json:"customer_id"`
	Reason      string    `gorm:"type:text" json:"reason"`
	Type        string    `gorm:"size:30" json:"type"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `gorm:"size:30;default:pending" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ErpReturn) TableName() string { return "erp_returns" }

// ── Inventory Management ──

// ErpInventory 库存余额（每个 material × warehouse × location × lot × status 一行）
//
// PRD v1: Sprint 1 升级 — 加入 SourceType/SourceID/Supplier* 用于追溯链；
// ReceivedAt 用于精确库龄计算；Notes 给操作员备注。
type ErpInventory struct {
	ID            string     `gorm:"primaryKey;size:100" json:"id"`
	MaterialID    string     `gorm:"size:100;not null;index:idx_inv_lookup,priority:1" json:"material_id"`
	WarehouseID   string     `gorm:"size:100;not null;index:idx_inv_lookup,priority:2" json:"warehouse_id"`
	LocationID    string     `gorm:"size:100;index:idx_inv_lookup,priority:3" json:"location_id"`
	LotNumber     string     `gorm:"size:100;index:idx_inv_lookup,priority:4" json:"lot_number"`
	Status        string     `gorm:"size:30;default:available;index:idx_inv_lookup,priority:5" json:"status"` // available / reserved / quality_hold / blocked / in_transit
	Quantity      float64    `gorm:"default:0" json:"quantity"`
	ReservedQty   float64    `gorm:"default:0" json:"reserved_qty"`
	UnitCost      float64    `gorm:"default:0" json:"unit_cost"`
	ReceivedAt    *time.Time `json:"received_at"`           // 这一批次首次入库时间（用于库龄）
	ExpiryDate    *time.Time `json:"expiry_date"`           // 失效日期
	SourceType    string     `gorm:"size:30" json:"source_type"`    // po_receipt / production_in / transfer / count_in / manual_in
	SourceID      string     `gorm:"size:100;index" json:"source_id"`
	SupplierID    string     `gorm:"size:100;index" json:"supplier_id"`
	SupplierLot   string     `gorm:"size:100" json:"supplier_lot"`
	Notes         string     `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (ErpInventory) TableName() string { return "erp_inventory" }

// ErpInventoryTransaction 库存事务流水
//
// PRD v1: Sprint 1 升级 — 加入 from/to status、operator 信息、reason、approval、
// total_amount、journal_entry_id；为后续 Sprint 2 财务联动准备。
type ErpInventoryTransaction struct {
	ID              string    `gorm:"primaryKey;size:100" json:"id"`
	Code            string    `gorm:"size:50;uniqueIndex" json:"code"`
	Type            string    `gorm:"size:30;not null;index" json:"type"`
	MaterialID      string    `gorm:"size:100;not null;index" json:"material_id"`
	FromWarehouseID string    `gorm:"size:100;index" json:"from_warehouse_id"`
	FromLocationID  string    `gorm:"size:100" json:"from_location_id"`
	FromStatus      string    `gorm:"size:30" json:"from_status"`
	ToWarehouseID   string    `gorm:"size:100;index" json:"to_warehouse_id"`
	ToLocationID    string    `gorm:"size:100" json:"to_location_id"`
	ToStatus        string    `gorm:"size:30" json:"to_status"`
	Quantity        float64   `json:"quantity"`
	UnitCost        float64   `json:"unit_cost"`
	TotalAmount     float64   `gorm:"default:0" json:"total_amount"`
	LotNumber       string    `gorm:"size:100;index" json:"lot_number"`
	SerialNumbers   string    `gorm:"type:jsonb;default:'[]'" json:"serial_numbers"` // []string
	ReferenceType   string    `gorm:"size:30;index" json:"reference_type"`            // po / work_order / sales_order / count / adjustment / scrap
	ReferenceID     string    `gorm:"size:100;index" json:"reference_id"`
	JournalEntryID  string    `gorm:"size:100;index" json:"journal_entry_id"`         // 关联 Sprint 2 生成的会计凭证
	Reason          string    `gorm:"type:text" json:"reason"`
	Notes           string    `gorm:"type:text" json:"notes"`
	OperatorID      string    `gorm:"size:100;index" json:"operator_id"`
	OperatorName    string    `gorm:"size:100" json:"operator_name"`
	CreatedBy       string    `gorm:"size:100" json:"created_by"` // legacy field, kept for backward compat
	ApprovalID      string    `gorm:"size:100" json:"approval_id"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ErpInventoryTransaction) TableName() string { return "erp_inventory_transactions" }

// ErpMaterialInventoryAttrs 物料的库存特有属性
//
// PRD v1: 物料主数据归 PLM (plm_materials) 拥有，ERP 不复制；但库存维度
// 需要的属性（ABC 分类、安全库存、跟踪模式、移动加权平均成本）放在这里。
// 第一次入库时由 ensureInventoryAttrs 自动 lazy create。
type ErpMaterialInventoryAttrs struct {
	MaterialID       string     `gorm:"primaryKey;size:100" json:"material_id"` // FK to plm_materials.id
	DefaultUnit      string     `gorm:"size:20;default:pcs" json:"default_unit"`
	ABCClass         string     `gorm:"size:1" json:"abc_class"` // A / B / C
	SafetyStock      float64    `gorm:"default:0" json:"safety_stock"`
	ReorderPoint     float64    `gorm:"default:0" json:"reorder_point"`
	MaxStock         float64    `gorm:"default:0" json:"max_stock"`
	LeadTimeDays     int        `gorm:"default:0" json:"lead_time_days"`
	TrackingMode     string     `gorm:"size:10;default:lot" json:"tracking_mode"` // none / lot / serial
	DefaultWarehouse string     `gorm:"size:100" json:"default_warehouse"`
	ShelfLifeDays    int        `gorm:"default:0" json:"shelf_life_days"`
	StandardCost     float64    `gorm:"default:0" json:"standard_cost"`
	CurrentAvgCost   float64    `gorm:"default:0" json:"current_avg_cost"` // 移动加权平均成本（Sprint 2 计算）
	LastReceivedAt   *time.Time `json:"last_received_at"`
	LastIssuedAt     *time.Time `json:"last_issued_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (ErpMaterialInventoryAttrs) TableName() string { return "erp_material_inventory_attrs" }

// ErpInventoryAuditLog 主数据变更审计 (PRD v1 Sprint 4 才用，Sprint 1 先建表)
type ErpInventoryAuditLog struct {
	ID           string    `gorm:"primaryKey;size:100" json:"id"`
	EntityType   string    `gorm:"size:30;index" json:"entity_type"` // warehouse / location / material_attrs / inventory
	EntityID     string    `gorm:"size:100;index" json:"entity_id"`
	Action       string    `gorm:"size:20" json:"action"` // create / update / delete / approve
	BeforeData   string    `gorm:"type:jsonb" json:"before_data"`
	AfterData    string    `gorm:"type:jsonb" json:"after_data"`
	OperatorID   string    `gorm:"size:100" json:"operator_id"`
	OperatorName string    `gorm:"size:100" json:"operator_name"`
	Reason       string    `gorm:"type:text" json:"reason"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

func (ErpInventoryAuditLog) TableName() string { return "erp_inventory_audit_log" }

// ── Sprint 3: Inventory Count & Reservation ──

// ErpInventoryCount 盘点单头表
//
// PRD v1: 支持 full (全盘) / cycle (循环盘) / spot (抽盘) 三种类型。
// 生命周期: draft → snapshot_taken → counting → review → posted
type ErpInventoryCount struct {
	ID           string     `gorm:"primaryKey;size:100" json:"id"`
	Code         string     `gorm:"size:50;uniqueIndex" json:"code"`
	Name         string     `gorm:"size:200" json:"name"`
	Type         string     `gorm:"size:20;index" json:"type"` // full / cycle / spot
	WarehouseID  string     `gorm:"size:100;index" json:"warehouse_id"`
	ABCFilter    string     `gorm:"size:20" json:"abc_filter"` // A / B / C / ABC
	ScheduledAt  *time.Time `json:"scheduled_at"`
	SnapshotAt   *time.Time `json:"snapshot_at"`
	PostedAt     *time.Time `json:"posted_at"`
	Status       string     `gorm:"size:30;default:draft;index" json:"status"` // draft / snapshot_taken / counting / review / posted / canceled
	TotalLines   int        `json:"total_lines"`
	CountedLines int        `json:"counted_lines"`
	Variance     float64    `json:"variance"`
	Notes        string     `gorm:"type:text" json:"notes"`
	CreatedBy    string     `gorm:"size:100" json:"created_by"`
	PostedBy     string     `gorm:"size:100" json:"posted_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (ErpInventoryCount) TableName() string { return "erp_inventory_counts" }

// ErpInventoryCountLine 盘点明细
type ErpInventoryCountLine struct {
	ID          string     `gorm:"primaryKey;size:100" json:"id"`
	CountID     string     `gorm:"size:100;not null;index" json:"count_id"`
	MaterialID  string     `gorm:"size:100;not null;index" json:"material_id"`
	WarehouseID string     `gorm:"size:100;index" json:"warehouse_id"`
	LocationID  string     `gorm:"size:100" json:"location_id"`
	LotNumber   string     `gorm:"size:100" json:"lot_number"`
	Status      string     `gorm:"size:30" json:"status"` // inventory status being counted
	SnapshotQty float64    `json:"snapshot_qty"`          // 盘点前系统数量（冻结值）
	CountedQty  float64    `gorm:"default:-1" json:"counted_qty"` // 实盘数量（-1 表示未盘）
	Variance    float64    `json:"variance"`                     // counted - snapshot
	UnitCost    float64    `json:"unit_cost"`
	VarianceAmt float64    `json:"variance_amt"`
	CountedBy   string     `gorm:"size:100" json:"counted_by"`
	CountedAt   *time.Time `json:"counted_at"`
	Notes       string     `gorm:"type:text" json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ErpInventoryCountLine) TableName() string { return "erp_inventory_count_lines" }

// ErpInventoryReservation 库存预留
//
// PRD v1: 销售订单确认或工单发布时预留库存；取消/发货后释放。
// 状态: active / released / consumed / expired
type ErpInventoryReservation struct {
	ID            string     `gorm:"primaryKey;size:100" json:"id"`
	Code          string     `gorm:"size:50;uniqueIndex" json:"code"`
	MaterialID    string     `gorm:"size:100;not null;index" json:"material_id"`
	WarehouseID   string     `gorm:"size:100;index" json:"warehouse_id"`
	LotNumber     string     `gorm:"size:100" json:"lot_number"`
	ReservedQty   float64    `json:"reserved_qty"`
	ConsumedQty   float64    `gorm:"default:0" json:"consumed_qty"`
	SourceType    string     `gorm:"size:30;index" json:"source_type"` // sales_order / work_order / manual
	SourceID      string     `gorm:"size:100;index" json:"source_id"`
	Priority      int        `gorm:"default:0" json:"priority"`
	Status        string     `gorm:"size:30;default:active;index" json:"status"`
	ExpiresAt     *time.Time `json:"expires_at"`
	ReservedBy    string     `gorm:"size:100" json:"reserved_by"`
	Notes         string     `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (ErpInventoryReservation) TableName() string { return "erp_inventory_reservations" }

// ErpSerialNumber 序列号追踪
//
// Sprint 4 扩展: 增加 work_order_id / shipment_id / return_id / notes
// 形成完整追溯链（生产工单 → 发货 → 客户 → 退货）
type ErpSerialNumber struct {
	ID             string     `gorm:"primaryKey;size:100" json:"id"`
	SerialNumber   string     `gorm:"size:100;uniqueIndex" json:"serial_number"`
	MaterialID     string     `gorm:"size:100;index" json:"material_id"`
	ProductID      string     `gorm:"size:100;index" json:"product_id"`
	Status         string     `gorm:"size:30;default:in_stock;index" json:"status"` // in_stock / shipped / delivered / returned / scrapped
	WarehouseID    string     `gorm:"size:100;index" json:"warehouse_id"`
	LotNumber      string     `gorm:"size:100;index" json:"lot_number"`
	WorkOrderID    string     `gorm:"size:100;index" json:"work_order_id"`
	ShipmentID     string     `gorm:"size:100;index" json:"shipment_id"`
	ReturnID       string     `gorm:"size:100;index" json:"return_id"`
	CustomerID     string     `gorm:"size:100;index" json:"customer_id"`
	ManufacturedAt *time.Time `json:"manufactured_at"`
	SoldTo         string     `gorm:"size:100" json:"sold_to"`
	SoldAt         *time.Time `json:"sold_at"`
	WarrantyUntil  *time.Time `json:"warranty_until"`
	Notes          string     `gorm:"type:text" json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (ErpSerialNumber) TableName() string { return "erp_serial_numbers" }

// ── MRP ──

// ErpMRPResult MRP 计算结果
type ErpMRPResult struct {
	ID               string     `gorm:"primaryKey;size:100" json:"id"`
	RunID            string     `gorm:"size:100;not null;index" json:"run_id"`
	MaterialID       string     `gorm:"size:100;not null;index" json:"material_id"`
	DemandSource     string     `gorm:"size:30" json:"demand_source"`
	DemandID         string     `gorm:"size:100" json:"demand_id"`
	GrossRequirement float64    `json:"gross_requirement"`
	OnHand           float64    `json:"on_hand"`
	OnOrder          float64    `json:"on_order"`
	NetRequirement   float64    `json:"net_requirement"`
	Action           string     `gorm:"size:30" json:"action"`
	SuggestedQty     float64    `json:"suggested_qty"`
	SuggestedDate    *time.Time `json:"suggested_date"`
	BomID            string     `gorm:"size:100" json:"bom_id"`
	BomLevel         int        `gorm:"default:0" json:"bom_level"`
	Status           string     `gorm:"size:30;default:pending" json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (ErpMRPResult) TableName() string { return "erp_mrp_results" }

// ── Production ──

// ErpWorkOrder 生产工单
type ErpWorkOrder struct {
	ID           string     `gorm:"primaryKey;size:100" json:"id"`
	Code         string     `gorm:"size:50;uniqueIndex" json:"code"`
	ProductID    string     `gorm:"size:100;not null;index" json:"product_id"`
	BomID        string     `gorm:"size:100;index" json:"bom_id"`
	OrderID      string     `gorm:"size:100;index" json:"order_id"`
	MrpResultID  string     `gorm:"size:100" json:"mrp_result_id"`
	PlannedQty   float64    `json:"planned_qty"`
	CompletedQty float64    `gorm:"default:0" json:"completed_qty"`
	ScrapQty     float64    `gorm:"default:0" json:"scrap_qty"`
	WarehouseID  string     `gorm:"size:100" json:"warehouse_id"`
	PlannedStart *time.Time `json:"planned_start"`
	PlannedEnd   *time.Time `json:"planned_end"`
	ActualStart  *time.Time `json:"actual_start"`
	ActualEnd    *time.Time `json:"actual_end"`
	Priority     string     `gorm:"size:20;default:normal" json:"priority"`
	Status       string     `gorm:"size:30;default:draft" json:"status"`
	Notes        string     `gorm:"type:text" json:"notes"`
	CreatedBy    string     `gorm:"size:100" json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (ErpWorkOrder) TableName() string { return "erp_work_orders" }

// ErpWOMaterialIssue 工单领料
type ErpWOMaterialIssue struct {
	ID          string     `gorm:"primaryKey;size:100" json:"id"`
	WorkOrderID string     `gorm:"size:100;not null;index" json:"work_order_id"`
	MaterialID  string     `gorm:"size:100;not null;index" json:"material_id"`
	BomItemID   string     `gorm:"size:100" json:"bom_item_id"`
	RequiredQty float64    `json:"required_qty"`
	IssuedQty   float64    `gorm:"default:0" json:"issued_qty"`
	WarehouseID string     `gorm:"size:100" json:"warehouse_id"`
	LotNumber   string     `gorm:"size:100" json:"lot_number"`
	IssuedAt    *time.Time `json:"issued_at"`
	IssuedBy    string     `gorm:"size:100" json:"issued_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ErpWOMaterialIssue) TableName() string { return "erp_wo_material_issues" }

// ErpWOReport 工单报工
type ErpWOReport struct {
	ID          string     `gorm:"primaryKey;size:100" json:"id"`
	WorkOrderID string     `gorm:"size:100;not null;index" json:"work_order_id"`
	Operation   string     `gorm:"size:100" json:"operation"`
	OperatorID  string     `gorm:"size:100" json:"operator_id"`
	GoodQty     float64    `json:"good_qty"`
	DefectQty   float64    `gorm:"default:0" json:"defect_qty"`
	ScrapQty    float64    `gorm:"default:0" json:"scrap_qty"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Notes       string     `gorm:"type:text" json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ErpWOReport) TableName() string { return "erp_wo_reports" }

// ── Finance ──

// ErpAccount 会计科目
type ErpAccount struct {
	ID       string    `gorm:"primaryKey;size:100" json:"id"`
	Code     string    `gorm:"size:50;uniqueIndex" json:"code"`
	Name     string    `gorm:"size:200;not null" json:"name"`
	Type     string    `gorm:"size:30;not null" json:"type"`
	ParentID string    `gorm:"size:100;index" json:"parent_id"`
	Level    int       `gorm:"default:1" json:"level"`
	IsLeaf   bool      `gorm:"default:true" json:"is_leaf"`
	Currency string    `gorm:"size:10;default:CNY" json:"currency"`
	Status   string    `gorm:"size:30;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ErpAccount) TableName() string { return "erp_accounts" }

// ErpJournalEntry 会计凭证
type ErpJournalEntry struct {
	ID          string     `gorm:"primaryKey;size:100" json:"id"`
	Code        string     `gorm:"size:50;uniqueIndex" json:"code"`
	Period      string     `gorm:"size:10" json:"period"`
	EntryDate   *time.Time `json:"entry_date"`
	SourceType  string     `gorm:"size:30" json:"source_type"`
	SourceID    string     `gorm:"size:100" json:"source_id"`
	Description string     `gorm:"type:text" json:"description"`
	TotalDebit  float64    `json:"total_debit"`
	TotalCredit float64    `json:"total_credit"`
	Status      string     `gorm:"size:30;default:draft" json:"status"`
	PostedBy    string     `gorm:"size:100" json:"posted_by"`
	PostedAt    *time.Time `json:"posted_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ErpJournalEntry) TableName() string { return "erp_journal_entries" }

// ErpPeriodLock 会计期间锁定表（Sprint 2 关期使用）
type ErpPeriodLock struct {
	Period    string    `gorm:"primaryKey;size:10" json:"period"` // YYYY-MM
	Status    string    `gorm:"size:20;default:open" json:"status"` // open / closed
	ClosedBy  string    `gorm:"size:100" json:"closed_by"`
	ClosedAt  *time.Time `json:"closed_at"`
	Notes     string    `gorm:"type:text" json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ErpPeriodLock) TableName() string { return "erp_period_locks" }

// ErpJournalLine 凭证行项
type ErpJournalLine struct {
	ID             string    `gorm:"primaryKey;size:100" json:"id"`
	EntryID        string    `gorm:"size:100;not null;index" json:"entry_id"`
	AccountID      string    `gorm:"size:100;not null;index" json:"account_id"`
	Debit          float64   `gorm:"default:0" json:"debit"`
	Credit         float64   `gorm:"default:0" json:"credit"`
	Currency       string    `gorm:"size:10;default:CNY" json:"currency"`
	OriginalAmount float64   `json:"original_amount"`
	Description    string    `gorm:"type:text" json:"description"`
	CustomerID     string    `gorm:"size:100" json:"customer_id"`
	SupplierID     string    `gorm:"size:100" json:"supplier_id"`
	DepartmentID   string    `gorm:"size:100" json:"department_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ErpJournalLine) TableName() string { return "erp_journal_lines" }

// ErpSalesInvoice 销售发票
type ErpSalesInvoice struct {
	ID             string     `gorm:"primaryKey;size:100" json:"id"`
	Code           string     `gorm:"size:50;uniqueIndex" json:"code"`
	OrderID        string     `gorm:"size:100;index" json:"order_id"`
	CustomerID     string     `gorm:"size:100;not null;index" json:"customer_id"`
	InvoiceDate    *time.Time `json:"invoice_date"`
	DueDate        *time.Time `json:"due_date"`
	Currency       string     `gorm:"size:10;default:CNY" json:"currency"`
	Subtotal       float64    `json:"subtotal"`
	TaxAmount      float64    `json:"tax_amount"`
	Total          float64    `json:"total"`
	PaidAmount     float64    `gorm:"default:0" json:"paid_amount"`
	Status         string     `gorm:"size:30;default:draft" json:"status"`
	JournalEntryID string     `gorm:"size:100" json:"journal_entry_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (ErpSalesInvoice) TableName() string { return "erp_sales_invoices" }

// ErpReceipt 收款单
type ErpReceipt struct {
	ID             string     `gorm:"primaryKey;size:100" json:"id"`
	Code           string     `gorm:"size:50;uniqueIndex" json:"code"`
	CustomerID     string     `gorm:"size:100;not null;index" json:"customer_id"`
	Amount         float64    `json:"amount"`
	Currency       string     `gorm:"size:10;default:CNY" json:"currency"`
	PaymentMethod  string     `gorm:"size:30" json:"payment_method"`
	BankAccount    string     `gorm:"size:100" json:"bank_account"`
	ReferenceNo    string     `gorm:"size:100" json:"reference_no"`
	ReceivedDate   *time.Time `json:"received_date"`
	Status         string     `gorm:"size:30;default:draft" json:"status"`
	JournalEntryID string     `gorm:"size:100" json:"journal_entry_id"`
	Notes          string     `gorm:"type:text" json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (ErpReceipt) TableName() string { return "erp_receipts" }

// ErpReceiptAllocation 收款核销
type ErpReceiptAllocation struct {
	ID        string    `gorm:"primaryKey;size:100" json:"id"`
	ReceiptID string    `gorm:"size:100;not null;index" json:"receipt_id"`
	InvoiceID string    `gorm:"size:100;not null;index" json:"invoice_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ErpReceiptAllocation) TableName() string { return "erp_receipt_allocations" }

// ── Quality Management ──

// ErpOQCInspection OQC 出货检验
type ErpOQCInspection struct {
	ID             string     `gorm:"primaryKey;size:100" json:"id"`
	Code           string     `gorm:"size:50;uniqueIndex" json:"code"`
	ShipmentID     string     `gorm:"size:100;not null;index" json:"shipment_id"`
	ProductID      string     `gorm:"size:100;index" json:"product_id"`
	LotNumber      string     `gorm:"size:100" json:"lot_number"`
	SampleSize     int        `json:"sample_size"`
	TotalInspected int        `json:"total_inspected"`
	PassCount      int        `json:"pass_count"`
	FailCount      int        `json:"fail_count"`
	Result         string     `gorm:"size:30" json:"result"`
	InspectorID    string     `gorm:"size:100" json:"inspector_id"`
	InspectedAt    *time.Time `json:"inspected_at"`
	Notes          string     `gorm:"type:text" json:"notes"`
	DefectDetails  string     `gorm:"type:text" json:"defect_details"`
	Status         string     `gorm:"size:30;default:pending" json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (ErpOQCInspection) TableName() string { return "erp_oqc_inspections" }

// ErpNCRReport 不合格品报告
type ErpNCRReport struct {
	ID          string     `gorm:"primaryKey;size:100" json:"id"`
	Code        string     `gorm:"size:50;uniqueIndex" json:"code"`
	Source      string     `gorm:"size:30" json:"source"`
	SourceID    string     `gorm:"size:100" json:"source_id"`
	ProductID   string     `gorm:"size:100;index" json:"product_id"`
	MaterialID  string     `gorm:"size:100;index" json:"material_id"`
	LotNumber   string     `gorm:"size:100" json:"lot_number"`
	DefectQty   float64    `json:"defect_qty"`
	DefectType  string     `gorm:"size:50" json:"defect_type"`
	Description string     `gorm:"type:text" json:"description"`
	Disposition string     `gorm:"size:30" json:"disposition"`
	Severity    string     `gorm:"size:20" json:"severity"`
	Status      string     `gorm:"size:30;default:open" json:"status"`
	OwnerID     string     `gorm:"size:100" json:"owner_id"`
	ClosedAt    *time.Time `json:"closed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ErpNCRReport) TableName() string { return "erp_ncr_reports" }

// ErpCAPA 纠正与预防措施
type ErpCAPA struct {
	ID           string     `gorm:"primaryKey;size:100" json:"id"`
	Code         string     `gorm:"size:50;uniqueIndex" json:"code"`
	Type         string     `gorm:"size:30" json:"type"`
	NcrID        string     `gorm:"size:100;index" json:"ncr_id"`
	Title        string     `gorm:"size:500;not null" json:"title"`
	RootCause    string     `gorm:"type:text" json:"root_cause"`
	ActionPlan   string     `gorm:"type:text" json:"action_plan"`
	OwnerID      string     `gorm:"size:100" json:"owner_id"`
	DueDate      *time.Time `json:"due_date"`
	Verification string     `gorm:"type:text" json:"verification"`
	Status       string     `gorm:"size:30;default:open" json:"status"`
	ClosedAt     *time.Time `json:"closed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (ErpCAPA) TableName() string { return "erp_capa" }

// ── Audit Log ──

// ErpAuditLog tracks field-level changes on ERP entities
type ErpAuditLog struct {
	ID         string    `gorm:"primaryKey;size:100" json:"id"`
	EntityType string    `gorm:"size:50;index" json:"entity_type"`
	EntityID   string    `gorm:"size:100;index" json:"entity_id"`
	Field      string    `gorm:"size:100" json:"field"`
	OldValue   string    `gorm:"type:text" json:"old_value"`
	NewValue   string    `gorm:"type:text" json:"new_value"`
	UserID     string    `gorm:"size:100" json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ErpAuditLog) TableName() string { return "erp_audit_logs" }
