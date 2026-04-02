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
		{
			Name:      "dashboard",
			Label:     "工作台",
			Type:      "custom",
			Component: "ErpDashboard",
			Icon:      "DashboardOutlined",
		},
		// Order Board
		{
			Name:         "order_board",
			Label:        "订单看板",
			Type:         "kanban",
			Entity:       "sales_orders",
			Icon:         "ShoppingCartOutlined",
			KanbanField:  "status",
			KanbanStages: []string{"draft", "confirmed", "producing", "ready", "shipped", "delivered"},
		},
		// Work Order Board
		{
			Name:         "wo_board",
			Label:        "工单看板",
			Type:         "kanban",
			Entity:       "work_orders",
			Icon:         "ToolOutlined",
			KanbanField:  "status",
			KanbanStages: []string{"draft", "released", "in_progress", "completed"},
		},
		// Inventory Dashboard
		{
			Name:      "inventory_dashboard",
			Label:     "库存总览",
			Type:      "custom",
			Component: "ErpInventoryDashboard",
			Icon:      "InboxOutlined",
		},
		// MRP Console
		{
			Name:      "mrp_console",
			Label:     "MRP控制台",
			Type:      "custom",
			Component: "ErpMRPConsole",
			Icon:      "CalculatorOutlined",
		},
		// Finance Dashboard
		{
			Name:      "finance_dashboard",
			Label:     "财务仪表盘",
			Type:      "custom",
			Component: "ErpFinanceDashboard",
			Icon:      "BankOutlined",
		},
	}
}

func (m *ERPModule) Nav() []sdk.NavEntry {
	return []sdk.NavEntry{
		// 工作台
		{Key: "/m/erp/view/dashboard", Label: "工作台", Icon: "DashboardOutlined", View: "dashboard"},
		// 销售
		{Key: "erp-sales", Label: "销售", Icon: "ShoppingCartOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/customers", Label: "客户", Entity: "customers"},
			{Key: "/m/erp/quotations", Label: "报价", Entity: "quotations"},
			{Key: "/m/erp/sales_orders", Label: "销售订单", Entity: "sales_orders"},
			{Key: "/m/erp/shipments", Label: "发货", Entity: "shipments"},
			{Key: "/m/erp/returns", Label: "退货", Entity: "returns"},
		}},
		// 库存
		{Key: "erp-inventory", Label: "库存", Icon: "DatabaseOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/view/inventory_dashboard", Label: "库存总览", View: "inventory_dashboard"},
			{Key: "/m/erp/inventory_transactions", Label: "出入库", Entity: "inventory_transactions"},
			{Key: "/m/erp/serial_numbers", Label: "序列号", Entity: "serial_numbers"},
		}},
		// 生产
		{Key: "erp-production", Label: "生产", Icon: "ExperimentOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/view/mrp_console", Label: "MRP控制台", View: "mrp_console"},
			{Key: "/m/erp/work_orders", Label: "生产工单", Entity: "work_orders"},
		}},
		// 财务
		{Key: "erp-finance", Label: "财务", Icon: "AccountBookOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/view/finance_dashboard", Label: "财务仪表盘", View: "finance_dashboard"},
			{Key: "/m/erp/accounts", Label: "科目", Entity: "accounts"},
			{Key: "/m/erp/journal_entries", Label: "凭证", Entity: "journal_entries"},
			{Key: "/m/erp/sales_invoices", Label: "销售发票", Entity: "sales_invoices"},
			{Key: "/m/erp/receipts", Label: "收款", Entity: "receipts"},
		}},
		// 质量
		{Key: "erp-quality", Label: "质量", Icon: "SafetyCertificateOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/oqc_inspections", Label: "OQC检验", Entity: "oqc_inspections"},
			{Key: "/m/erp/ncr_reports", Label: "NCR", Entity: "ncr_reports"},
			{Key: "/m/erp/capas", Label: "CAPA", Entity: "capas"},
		}},
		// 设置
		{Key: "erp-settings", Label: "设置", Icon: "SettingOutlined", Children: []sdk.NavEntry{
			{Key: "/m/erp/warehouses", Label: "仓库", Entity: "warehouses"},
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
	)
}

// Routes registers custom routes
func (m *ERPModule) Routes(rg *gin.RouterGroup) {
	registerRoutes(m, rg)
}
