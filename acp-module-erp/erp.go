package erp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/bitfantasy/acp-module-sdk"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ERPModule implements the Module interface for enterprise resource planning.
type ERPModule struct {
	Deps    *sdk.ExecutorDeps
	BaseDir string // absolute path to the module directory (for serving static assets)
}

var _ sdk.Module = (*ERPModule)(nil)

// ---------------------------------------------------------------------------
// Executor interface
// ---------------------------------------------------------------------------

func (m *ERPModule) Type() string { return "erp" }

func (m *ERPModule) ValidateConfig(input map[string]any) error { return nil }

func (m *ERPModule) Commands() []sdk.CommandDef { return erpCommands() }

// ExecuteStep dispatches to command handler, passing adapter for event logging.
func (m *ERPModule) ExecuteStep(ctx context.Context, adapter sdk.EngineAdapter, rc *sdk.RunContext) error {
	handler, ok := commandHandlers[rc.Command]
	if !ok {
		errMsg := fmt.Sprintf("unknown erp command: %s", rc.Command)
		adapter.UpdateStepStatus(rc.RunID, rc.StepID, "failed", "", &errMsg)
		return fmt.Errorf("%s", errMsg)
	}
	// Pass adapter + runID + stepID for event logging
	result, err := handler(m.Deps.DB, adapter, rc.RunID, rc.StepID, rc.Input)
	if err != nil {
		errStr := err.Error()
		adapter.UpdateStepStatus(rc.RunID, rc.StepID, "failed", "", &errStr)
		return err
	}
	output, _ := json.Marshal(result)
	adapter.UpdateStepStatus(rc.RunID, rc.StepID, "completed", string(output), nil)
	return nil
}

// Meta returns visual editor metadata.
func (m *ERPModule) Meta() sdk.ExecutorMeta {
	return sdk.ExecutorMeta{
		Type:        "erp",
		Category:    "module",
		Label:       "企业资源计划",
		Icon:        "BankOutlined",
		Color:       "#1677ff",
		BgColor:     "#111a2e",
		Description: "销售·库存·生产·财务一体化管理",
	}
}

// ---------------------------------------------------------------------------
// Module interface
// ---------------------------------------------------------------------------

func (m *ERPModule) ModuleMeta() sdk.ModuleMeta {
	return sdk.ModuleMeta{
		Type:        "erp",
		Label:       "企业资源计划",
		Icon:        "BankOutlined",
		Color:       "#1677ff",
		BgColor:     "#111a2e",
		Description: "销售·库存·生产·财务一体化管理",
		Version:     "1.0.0",
	}
}

func (m *ERPModule) Entities() []sdk.EntityDef {
	return []sdk.EntityDef{
		// 基础数据
		customerEntity(),
		warehouseEntity(),
		locationEntity(),
		// 销售管理
		quotationEntity(),
		quotationItemEntity(),
		salesOrderEntity(),
		salesOrderItemEntity(),
		shipmentEntity(),
		shipmentItemEntity(),
		returnEntity(),
		// 库存管理
		inventoryEntity(),
		inventoryTransactionEntity(),
		serialNumberEntity(),
		// 生产管理
		mrpResultEntity(),
		workOrderEntity(),
		woMaterialIssueEntity(),
		woReportEntity(),
		// 财务管理
		accountEntity(),
		journalEntryEntity(),
		journalLineEntity(),
		salesInvoiceEntity(),
		receiptEntity(),
		receiptAllocationEntity(),
		// 质量管理
		oqcInspectionEntity(),
		ncrReportEntity(),
		capaEntity(),
	}
}

func (m *ERPModule) Components() []sdk.ComponentDef {
	return nil
}

func (m *ERPModule) Views() []sdk.ViewDef {
	return []sdk.ViewDef{
		// Dashboard
		{Name: "dashboard", Label: "工作台", Type: "custom", Component: "ErpDashboard", Icon: "DashboardOutlined"},
		// Sales
		{Name: "sales_pipeline", Label: "销售管道", Type: "custom", Component: "ErpSalesPipeline", Icon: "FunnelPlotOutlined"},
		{Name: "quote_board", Label: "报价跟进", Type: "custom", Component: "ErpQuoteBoard", Icon: "FileTextOutlined"},
		{Name: "shipping_center", Label: "发货中心", Type: "custom", Component: "ErpShippingCenter", Icon: "SendOutlined"},
		{Name: "ar_workspace", Label: "收款工作台", Type: "custom", Component: "ErpARWorkspace", Icon: "DollarOutlined"},
		// Inventory
		{Name: "stock_search", Label: "库存查询", Type: "custom", Component: "ErpStockSearch", Icon: "SearchOutlined"},
		{Name: "stock_ops", Label: "出入库操作", Type: "custom", Component: "ErpStockOps", Icon: "SwapOutlined"},
		{Name: "serial_trace", Label: "序列号追溯", Type: "custom", Component: "ErpSerialTrace", Icon: "BarcodeOutlined"},
		// Production
		{Name: "mrp_console", Label: "MRP控制台", Type: "custom", Component: "ErpMRPConsole", Icon: "CalculatorOutlined"},
		{Name: "production_schedule", Label: "生产调度", Type: "custom", Component: "ErpProductionSchedule", Icon: "ScheduleOutlined"},
		// Finance
		{Name: "journal_entry", Label: "凭证录入", Type: "custom", Component: "ErpJournalEntry", Icon: "EditOutlined"},
		{Name: "report_center", Label: "报表中心", Type: "custom", Component: "ErpReportCenter", Icon: "BarChartOutlined"},
		// Quality
		{Name: "quality_console", Label: "质量控制台", Type: "custom", Component: "ErpQualityConsole", Icon: "SafetyCertificateOutlined"},
		{Name: "capa_board", Label: "CAPA跟踪", Type: "custom", Component: "ErpCAPABoard", Icon: "ToolOutlined"},
	}
}

func (m *ERPModule) Nav() []sdk.NavEntry {
	return []sdk.NavEntry{
		// 工作台
		{Key: "/m/erp/view/dashboard", Label: "工作台", Icon: "DashboardOutlined", View: "dashboard"},
		// 销售
		{Key: "erp-sales", Label: "销售", Icon: "ShoppingCartOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/view/sales_pipeline", Label: "销售管道", View: "sales_pipeline"},
			{Key: "/m/erp/customers", Label: "客户", Entity: "customers"},
			{Key: "/m/erp/view/quote_board", Label: "报价跟进", View: "quote_board"},
			{Key: "/m/erp/view/shipping_center", Label: "发货中心", View: "shipping_center"},
			{Key: "/m/erp/view/ar_workspace", Label: "收款工作台", View: "ar_workspace"},
			{Key: "/m/erp/returns", Label: "退货", Entity: "returns"},
		}},
		// 库存
		{Key: "erp-inventory", Label: "库存", Icon: "DatabaseOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/view/stock_search", Label: "库存查询", View: "stock_search"},
			{Key: "/m/erp/view/stock_ops", Label: "出入库操作", View: "stock_ops"},
			{Key: "/m/erp/view/serial_trace", Label: "序列号追溯", View: "serial_trace"},
		}},
		// 生产
		{Key: "erp-production", Label: "生产", Icon: "ExperimentOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/view/mrp_console", Label: "MRP控制台", View: "mrp_console"},
			{Key: "/m/erp/view/production_schedule", Label: "生产调度", View: "production_schedule"},
			{Key: "/m/erp/work_orders", Label: "工单详情", Entity: "work_orders"},
		}},
		// 财务
		{Key: "erp-finance", Label: "财务", Icon: "AccountBookOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/view/journal_entry", Label: "凭证录入", View: "journal_entry"},
			{Key: "/m/erp/view/report_center", Label: "报表中心", View: "report_center"},
			{Key: "/m/erp/sales_invoices", Label: "销售发票", Entity: "sales_invoices"},
			{Key: "/m/erp/receipts", Label: "收款记录", Entity: "receipts"},
		}},
		// 质量
		{Key: "erp-quality", Label: "质量", Icon: "SafetyCertificateOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/view/quality_console", Label: "质量控制台", View: "quality_console"},
			{Key: "/m/erp/view/capa_board", Label: "CAPA跟踪", View: "capa_board"},
		}},
		// 设置
		{Key: "erp-settings", Label: "设置", Icon: "SettingOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/warehouses", Label: "仓库", Entity: "warehouses"},
			{Key: "/m/erp/accounts", Label: "科目表", Entity: "accounts"},
		}},
	}
}

func (m *ERPModule) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 基础数据
		&ErpCustomer{}, &ErpWarehouse{}, &ErpLocation{},
		// 销售管理
		&ErpQuotation{}, &ErpQuotationItem{},
		&ErpSalesOrder{}, &ErpSalesOrderItem{},
		&ErpShipment{}, &ErpShipmentItem{},
		&ErpReturn{},
		// 库存管理
		&ErpInventory{}, &ErpInventoryTransaction{}, &ErpSerialNumber{},
		// 生产管理
		&ErpMRPResult{}, &ErpWorkOrder{},
		&ErpWOMaterialIssue{}, &ErpWOReport{},
		// 财务管理
		&ErpAccount{}, &ErpJournalEntry{}, &ErpJournalLine{},
		&ErpSalesInvoice{}, &ErpReceipt{}, &ErpReceiptAllocation{},
		// 质量管理
		&ErpOQCInspection{}, &ErpNCRReport{}, &ErpCAPA{},
		// 审计日志
		&ErpAuditLog{},
	)
}

// Routes registers custom routes
func (m *ERPModule) Routes(rg *gin.RouterGroup) {
	registerRoutes(m, rg)
}
