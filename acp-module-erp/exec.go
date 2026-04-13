package erp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	sdk "github.com/bitfantasy/acp-module-sdk"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Command handler registry
// ---------------------------------------------------------------------------

type commandHandler func(db *gorm.DB, adapter sdk.EngineAdapter, runID string, stepID string, input map[string]any) (any, error)

var commandHandlers = map[string]commandHandler{
	// Sales (9)
	"create_quotation":     cmdCreateQuotation,
	"confirm_quotation":    cmdConfirmQuotation,
	"create_sales_order":   cmdCreateSalesOrder,
	"confirm_order":        cmdConfirmOrder,
	"create_shipment":      cmdCreateShipment,
	"confirm_shipment":     cmdConfirmShipment,
	"create_sales_invoice": cmdCreateSalesInvoice,
	"record_receipt":       cmdRecordReceipt,
	"create_return":        cmdCreateReturn,
	// Inventory (8)
	"receive_inventory":            cmdReceiveInventory,
	"issue_inventory":              cmdIssueInventory,
	"transfer_inventory":           cmdTransferInventory,
	"adjust_inventory":             cmdAdjustInventory,
	"scrap_inventory":              cmdScrapInventory,
	"reserve_inventory":            cmdReserveInventory,
	"unreserve_inventory":          cmdUnreserveInventory,
	"bootstrap_default_warehouses":  cmdBootstrapDefaultWarehouses,
	"bootstrap_chart_of_accounts":   cmdBootstrapChartOfAccounts,
	// Sprint 3: Count & Quality hold
	"create_inventory_count": cmdCreateInventoryCount,
	"submit_count_result":    cmdSubmitCountResult,
	"post_inventory_count":   cmdPostInventoryCount,
	"quality_hold":           cmdQualityHold,
	"quality_release":        cmdQualityRelease,
	// Sprint 4: Serial numbers
	"generate_serial_numbers": cmdGenerateSerialNumbers,
	"update_serial_status":    cmdUpdateSerialStatus,
	// Production (7)
	"run_mrp":                cmdRunMRP,
	"confirm_mrp_suggestion": cmdConfirmMRPSuggestion,
	"create_work_order":      cmdCreateWorkOrder,
	"release_work_order":     cmdReleaseWorkOrder,
	"issue_wo_materials":     cmdIssueWOMaterials,
	"report_wo_progress":     cmdReportWOProgress,
	"complete_work_order":    cmdCompleteWorkOrder,
	// Finance (6)
	"create_journal_entry":    cmdCreateJournalEntry,
	"post_journal_entry":      cmdPostJournalEntry,
	"reverse_journal_entry":   cmdReverseJournalEntry,
	"close_period":            cmdClosePeriod,
	"generate_report":         cmdGenerateReport,
	"post_ap_from_settlement": cmdPostAPFromSettlement,
	// Quality (6)
	"create_oqc":      cmdCreateOQC,
	"complete_oqc":    cmdCompleteOQC,
	"create_ncr":      cmdCreateNCR,
	"disposition_ncr":  cmdDispositionNCR,
	"create_capa":     cmdCreateCAPA,
	"close_capa":      cmdCloseCAPA,
	// Additional (3)
	"update_shipment_status": cmdUpdateShipmentStatus,
	"extend_quotation":       cmdExtendQuotation,
	"send_payment_reminder":  cmdSendPaymentReminder,
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// emitEvent writes a structured event via the engine adapter.
func emitEvent(adapter sdk.EngineAdapter, runID, stepID, eventType, summary string, payload map[string]any) {
	if adapter != nil {
		adapter.WriteEvent(runID, stepID, 0, eventType, "info", summary, payload, 0)
	}
}

// getStr extracts a string from input map.
func getStr(input map[string]any, key string) string {
	if v, ok := input[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// getInt extracts an int from input map.
func getInt(input map[string]any, key string, def int) int {
	if v, ok := input[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case json.Number:
			i, _ := n.Int64()
			return int(i)
		case string:
			if i, err := strconv.Atoi(n); err == nil {
				return i
			}
		}
	}
	return def
}

// getFloat extracts a float from input map.
func getFloat(input map[string]any, key string) float64 {
	if v, ok := input[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		case json.Number:
			f, _ := n.Float64()
			return f
		case string:
			if f, err := strconv.ParseFloat(n, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

// getStrSlice extracts a string slice.
func getStrSlice(input map[string]any, key string) []string {
	v, ok := input[key]
	if !ok {
		return nil
	}
	switch arr := v.(type) {
	case []string:
		return arr
	case []any:
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			out = append(out, fmt.Sprintf("%v", item))
		}
		return out
	}
	return nil
}

// parseDate parses YYYY-MM-DD.
func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

// autoCode generates a code like "SO-0001" using MAX of existing numeric codes.
func autoCode(db *gorm.DB, table, prefix string) string {
	col := "code"
	regexPattern := fmt.Sprintf("^%s-[0-9]+$", prefix)
	var maxCode string
	db.Table(table).Select(fmt.Sprintf("MAX(%s)", col)).
		Where(fmt.Sprintf("%s ~ ?", col), regexPattern).
		Row().Scan(&maxCode)
	next := 1
	if maxCode != "" {
		parts := strings.Split(maxCode, "-")
		if len(parts) >= 2 {
			if n, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				next = n + 1
			}
		}
	}
	return fmt.Sprintf("%s-%04d", prefix, next)
}

// autoCodeWithRetry generates a unique code and retries on duplicate key errors.
func autoCodeWithRetry(db *gorm.DB, table, prefix string, createFn func(code string) error) (string, error) {
	for attempt := 0; attempt < 15; attempt++ {
		code := autoCode(db, table, prefix)
		if err := createFn(code); err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "duplicate key") || strings.Contains(errMsg, "23505") {
				base := 50 * (1 << attempt)
				if base > 2000 {
					base = 2000
				}
				time.Sleep(time.Duration(base+rand.Intn(base)) * time.Millisecond)
				continue
			}
			return "", err
		}
		return code, nil
	}
	return "", fmt.Errorf("failed to generate unique %s code after 15 retries", prefix)
}

// getMapSlice extracts an array of objects from input.
func getMapSlice(input map[string]any, key string) []map[string]any {
	v, ok := input[key]
	if !ok {
		return nil
	}
	switch arr := v.(type) {
	case []map[string]any:
		return arr
	case []any:
		out := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
}

// findAccountByCode looks up an account by its code and returns the ID.
func findAccountByCode(db *gorm.DB, code string) string {
	var account ErpAccount
	if err := db.Where("code = ?", code).First(&account).Error; err != nil {
		return ""
	}
	return account.ID
}

// ===========================================================================
// Sales commands (9)
// ===========================================================================

func cmdCreateQuotation(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	customerID := getStr(input, "customer_id")
	if customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	var quot ErpQuotation
	code, err := autoCodeWithRetry(db, "erp_quotations", "QT", func(c string) error {
		quot = ErpQuotation{
			ID:           uuid.New().String(),
			Code:         c,
			CustomerID:   customerID,
			ContactName:  getStr(input, "contact_name"),
			Currency:     getStr(input, "currency"),
			PaymentTerms: getStr(input, "payment_terms"),
			ValidUntil:   parseDate(getStr(input, "valid_until")),
			Notes:        getStr(input, "notes"),
			Status:       "draft",
			CreatedBy:    getStr(input, "created_by"),
		}
		if quot.Currency == "" {
			quot.Currency = "CNY"
		}
		quot.ExchangeRate = getFloat(input, "exchange_rate")
		if quot.ExchangeRate == 0 {
			quot.ExchangeRate = 1
		}
		return db.Create(&quot).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create quotation failed: %w", err)
	}

	// Create line items
	items := getMapSlice(input, "items")
	var subtotal, taxTotal float64
	for i, item := range items {
		qty := getFloat(item, "quantity")
		price := getFloat(item, "unit_price")
		discount := getFloat(item, "discount_pct")
		taxRate := getFloat(item, "tax_rate")
		lineTotal := qty * price * (1 - discount/100)
		lineTax := lineTotal * taxRate / 100
		subtotal += lineTotal
		taxTotal += lineTax

		li := ErpQuotationItem{
			ID:          uuid.New().String(),
			QuotationID: quot.ID,
			ProductID:   getStr(item, "product_id"),
			SkuID:       getStr(item, "sku_id"),
			Description: getStr(item, "description"),
			Quantity:    qty,
			UnitPrice:   price,
			DiscountPct: discount,
			TaxRate:     taxRate,
			LineTotal:   lineTotal,
			SortOrder:   i + 1,
		}
		db.Create(&li)
	}

	// Update totals
	db.Model(&quot).Updates(map[string]any{
		"subtotal":   subtotal,
		"tax_amount": taxTotal,
		"total":      subtotal + taxTotal,
	})

	emitEvent(adapter, runID, stepID, "erp.quotation.created",
		fmt.Sprintf("报价单创建: %s", code),
		map[string]any{"quotation_id": quot.ID, "code": code})

	return map[string]any{"quotation_id": quot.ID, "code": code}, nil
}

func cmdConfirmQuotation(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	quotID := getStr(input, "quotation_id")
	if quotID == "" {
		return nil, fmt.Errorf("quotation_id is required")
	}
	var quot ErpQuotation
	if err := db.First(&quot, "id = ?", quotID).Error; err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	if quot.Status != "draft" {
		return nil, fmt.Errorf("quotation not in draft status, current: %s", quot.Status)
	}

	db.Model(&quot).Update("status", "confirmed")

	emitEvent(adapter, runID, stepID, "erp.quotation.confirmed",
		fmt.Sprintf("报价单确认: %s", quot.Code),
		map[string]any{"quotation_id": quot.ID, "code": quot.Code})

	return map[string]any{"quotation_id": quot.ID, "code": quot.Code}, nil
}

func cmdCreateSalesOrder(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	customerID := getStr(input, "customer_id")
	if customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	var order ErpSalesOrder
	code, err := autoCodeWithRetry(db, "erp_sales_orders", "SO", func(c string) error {
		order = ErpSalesOrder{
			ID:              uuid.New().String(),
			Code:            c,
			QuotationID:     getStr(input, "quotation_id"),
			CustomerID:      customerID,
			ShippingAddress: getStr(input, "shipping_address"),
			Currency:        getStr(input, "currency"),
			PaymentTerms:    getStr(input, "payment_terms"),
			ExpectedDate:    parseDate(getStr(input, "expected_date")),
			ShippingMethod:  getStr(input, "shipping_method"),
			Priority:        getStr(input, "priority"),
			Notes:           getStr(input, "notes"),
			Status:          "draft",
			CreatedBy:       getStr(input, "created_by"),
		}
		if order.Currency == "" {
			order.Currency = "CNY"
		}
		if order.Priority == "" {
			order.Priority = "normal"
		}
		return db.Create(&order).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create sales order failed: %w", err)
	}

	// Create line items
	items := getMapSlice(input, "items")
	var subtotal, taxTotal float64
	for _, item := range items {
		qty := getFloat(item, "quantity")
		price := getFloat(item, "unit_price")
		discount := getFloat(item, "discount_pct")
		taxRate := getFloat(item, "tax_rate")
		lineTotal := qty * price * (1 - discount/100)
		lineTax := lineTotal * taxRate / 100
		subtotal += lineTotal
		taxTotal += lineTax

		li := ErpSalesOrderItem{
			ID:           uuid.New().String(),
			OrderID:      order.ID,
			ProductID:    getStr(item, "product_id"),
			SkuID:        getStr(item, "sku_id"),
			BomID:        getStr(item, "bom_id"),
			Description:  getStr(item, "description"),
			Quantity:     qty,
			UnitPrice:    price,
			DiscountPct:  discount,
			TaxRate:      taxRate,
			LineTotal:    lineTotal,
			ExpectedDate: parseDate(getStr(item, "expected_date")),
			Status:       "pending",
		}
		db.Create(&li)
	}

	// Update totals
	db.Model(&order).Updates(map[string]any{
		"subtotal":   subtotal,
		"tax_amount": taxTotal,
		"total":      subtotal + taxTotal,
	})

	emitEvent(adapter, runID, stepID, "erp.sales_order.created",
		fmt.Sprintf("销售订单创建: %s", code),
		map[string]any{"order_id": order.ID, "code": code})

	return map[string]any{"order_id": order.ID, "code": code}, nil
}

func cmdConfirmOrder(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	orderID := getStr(input, "order_id")
	if orderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}
	var order ErpSalesOrder
	if err := db.First(&order, "id = ?", orderID).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != "draft" {
		return nil, fmt.Errorf("order not in draft status, current: %s", order.Status)
	}

	now := time.Now()
	db.Model(&order).Updates(map[string]any{"status": "confirmed", "confirmed_at": &now})

	// Auto-run MRP for this order's items
	mrpResult, _ := cmdRunMRP(db, adapter, runID, stepID, map[string]any{"demand_source": "so"})
	mrpSuggestions := 0
	if m, ok := mrpResult.(map[string]any); ok {
		if sc, ok := m["suggestion_count"]; ok {
			if n, ok := sc.(int); ok {
				mrpSuggestions = n
			}
		}
	}

	emitEvent(adapter, runID, stepID, "erp.sales_order.confirmed",
		fmt.Sprintf("销售订单确认: %s, MRP生成%d条建议", order.Code, mrpSuggestions),
		map[string]any{"order_id": order.ID, "code": order.Code, "mrp_suggestions": mrpSuggestions})

	return map[string]any{"order_id": order.ID, "code": order.Code, "mrp_suggestions": mrpSuggestions}, nil
}

func cmdCreateShipment(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	orderID := getStr(input, "order_id")
	if orderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}
	var order ErpSalesOrder
	if err := db.First(&order, "id = ?", orderID).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	var shipment ErpShipment
	code, err := autoCodeWithRetry(db, "erp_shipments", "SH", func(c string) error {
		shipment = ErpShipment{
			ID:              uuid.New().String(),
			Code:            c,
			OrderID:         orderID,
			WarehouseID:     getStr(input, "warehouse_id"),
			ShippingAddress: getStr(input, "shipping_address"),
			Carrier:         getStr(input, "carrier"),
			TrackingNo:      getStr(input, "tracking_no"),
			Notes:           getStr(input, "notes"),
			Status:          "pending",
		}
		if shipment.ShippingAddress == "" {
			shipment.ShippingAddress = order.ShippingAddress
		}
		return db.Create(&shipment).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create shipment failed: %w", err)
	}

	// Create shipment items
	items := getMapSlice(input, "items")
	for _, item := range items {
		si := ErpShipmentItem{
			ID:            uuid.New().String(),
			ShipmentID:    shipment.ID,
			OrderItemID:   getStr(item, "order_item_id"),
			ProductID:     getStr(item, "product_id"),
			Quantity:      getFloat(item, "quantity"),
			LotNumber:     getStr(item, "lot_number"),
			SerialNumbers: getStr(item, "serial_numbers"),
		}
		db.Create(&si)
	}

	emitEvent(adapter, runID, stepID, "erp.shipment.created",
		fmt.Sprintf("发货单创建: %s", code),
		map[string]any{"shipment_id": shipment.ID, "code": code, "order_id": orderID})

	return map[string]any{"shipment_id": shipment.ID, "code": code}, nil
}

func cmdConfirmShipment(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	shipmentID := getStr(input, "shipment_id")
	if shipmentID == "" {
		return nil, fmt.Errorf("shipment_id is required")
	}
	var shipment ErpShipment
	if err := db.First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return nil, fmt.Errorf("shipment not found")
	}
	if shipment.Status != "pending" {
		return nil, fmt.Errorf("shipment not in pending status, current: %s", shipment.Status)
	}

	now := time.Now()
	db.Model(&shipment).Updates(map[string]any{"status": "shipped", "shipped_at": &now})

	// Update delivered qty on order items
	var items []ErpShipmentItem
	db.Where("shipment_id = ?", shipmentID).Find(&items)
	for _, item := range items {
		if item.OrderItemID != "" {
			db.Model(&ErpSalesOrderItem{}).Where("id = ?", item.OrderItemID).
				Update("delivered_qty", gorm.Expr("delivered_qty + ?", item.Quantity))
		}
	}

	// Auto-create inventory issue transaction for each shipment item
	for _, item := range items {
		autoCodeWithRetry(db, "erp_inventory_transactions", "IT", func(c string) error {
			txn := ErpInventoryTransaction{
				ID: uuid.New().String(), Code: c,
				Type: "issue", MaterialID: item.ProductID,
				FromWarehouseID: shipment.WarehouseID,
				Quantity:        item.Quantity,
				ReferenceType:   "so", ReferenceID: shipment.OrderID,
				Notes: fmt.Sprintf("发货出库 %s", shipment.Code),
			}
			return db.Create(&txn).Error
		})
		// Update inventory
		db.Exec("UPDATE erp_inventory SET quantity = quantity - ? WHERE material_id = ? AND warehouse_id = ?",
			item.Quantity, item.ProductID, shipment.WarehouseID)
	}

	// Check if all order items are fully delivered → update order status
	var order ErpSalesOrder
	if db.First(&order, "id = ?", shipment.OrderID).Error == nil {
		var pendingCount int64
		db.Model(&ErpSalesOrderItem{}).
			Where("order_id = ? AND delivered_qty < quantity", order.ID).
			Count(&pendingCount)
		if pendingCount == 0 {
			db.Model(&order).Update("status", "shipped")
		}
	}

	emitEvent(adapter, runID, stepID, "erp.shipment.confirmed",
		fmt.Sprintf("发货确认: %s", shipment.Code),
		map[string]any{"shipment_id": shipment.ID, "code": shipment.Code})

	return map[string]any{"shipment_id": shipment.ID, "code": shipment.Code}, nil
}

func cmdCreateSalesInvoice(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	orderID := getStr(input, "order_id")
	customerID := getStr(input, "customer_id")
	if customerID == "" && orderID != "" {
		var order ErpSalesOrder
		if db.First(&order, "id = ?", orderID).Error == nil {
			customerID = order.CustomerID
		}
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	var invoice ErpSalesInvoice
	code, err := autoCodeWithRetry(db, "erp_sales_invoices", "INV", func(c string) error {
		invoice = ErpSalesInvoice{
			ID:          uuid.New().String(),
			Code:        c,
			OrderID:     orderID,
			CustomerID:  customerID,
			InvoiceDate: parseDate(getStr(input, "invoice_date")),
			DueDate:     parseDate(getStr(input, "due_date")),
			Currency:    getStr(input, "currency"),
			Subtotal:    getFloat(input, "subtotal"),
			TaxAmount:   getFloat(input, "tax_amount"),
			Total:       getFloat(input, "total"),
			Status:      "draft",
		}
		if invoice.Currency == "" {
			invoice.Currency = "CNY"
		}
		// Default totals from order
		if invoice.Total == 0 && orderID != "" {
			var order ErpSalesOrder
			if db.First(&order, "id = ?", orderID).Error == nil {
				invoice.Subtotal = order.Subtotal
				invoice.TaxAmount = order.TaxAmount
				invoice.Total = order.Total
			}
		}
		return db.Create(&invoice).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create sales invoice failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.sales_invoice.created",
		fmt.Sprintf("销售发票创建: %s", code),
		map[string]any{"invoice_id": invoice.ID, "code": code})

	// Auto-generate journal entry
	entryDate := time.Now()
	period := entryDate.Format("2006-01")

	var entry ErpJournalEntry
	autoCodeWithRetry(db, "erp_journal_entries", "JE", func(c string) error {
		entry = ErpJournalEntry{
			ID:          uuid.New().String(),
			Code:        c,
			Period:      period,
			EntryDate:   &entryDate,
			SourceType:  "sales_invoice",
			SourceID:    invoice.ID,
			Description: fmt.Sprintf("销售收入-%s", invoice.Code),
			TotalDebit:  invoice.Total,
			TotalCredit: invoice.Total,
			Status:      "posted",
		}
		return db.Create(&entry).Error
	})

	// Create journal lines
	// Debit: 应收账款 (1122)
	db.Create(&ErpJournalLine{
		ID: uuid.New().String(), EntryID: entry.ID,
		AccountID:  findAccountByCode(db, "1122"),
		Debit:      invoice.Total, Credit: 0,
		Description: "应收账款",
		CustomerID:  invoice.CustomerID,
	})
	// Credit: 主营业务收入 (6001)
	db.Create(&ErpJournalLine{
		ID: uuid.New().String(), EntryID: entry.ID,
		AccountID:  findAccountByCode(db, "6001"),
		Debit:      0, Credit: invoice.Subtotal,
		Description: "主营业务收入",
	})
	// Credit: 应交税费 (2221)
	if invoice.TaxAmount > 0 {
		db.Create(&ErpJournalLine{
			ID: uuid.New().String(), EntryID: entry.ID,
			AccountID:  findAccountByCode(db, "2221"),
			Debit:      0, Credit: invoice.TaxAmount,
			Description: "销项税额",
		})
	}

	return map[string]any{"invoice_id": invoice.ID, "code": code, "journal_entry_id": entry.ID}, nil
}

func cmdRecordReceipt(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	customerID := getStr(input, "customer_id")
	if customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}
	amount := getFloat(input, "amount")
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	var receipt ErpReceipt
	code, err := autoCodeWithRetry(db, "erp_receipts", "REC", func(c string) error {
		receipt = ErpReceipt{
			ID:            uuid.New().String(),
			Code:          c,
			CustomerID:    customerID,
			Amount:        amount,
			Currency:      getStr(input, "currency"),
			PaymentMethod: getStr(input, "payment_method"),
			BankAccount:   getStr(input, "bank_account"),
			ReferenceNo:   getStr(input, "reference_no"),
			ReceivedDate:  parseDate(getStr(input, "received_date")),
			Notes:         getStr(input, "notes"),
			Status:        "draft",
		}
		if receipt.Currency == "" {
			receipt.Currency = "CNY"
		}
		return db.Create(&receipt).Error
	})
	if err != nil {
		return nil, fmt.Errorf("record receipt failed: %w", err)
	}

	// Auto-allocate to invoices if specified
	allocations := getMapSlice(input, "allocations")
	for _, alloc := range allocations {
		invoiceID := getStr(alloc, "invoice_id")
		allocAmt := getFloat(alloc, "amount")
		if invoiceID == "" || allocAmt <= 0 {
			continue
		}
		ra := ErpReceiptAllocation{
			ID:        uuid.New().String(),
			ReceiptID: receipt.ID,
			InvoiceID: invoiceID,
			Amount:    allocAmt,
		}
		db.Create(&ra)

		// Update invoice paid amount
		db.Model(&ErpSalesInvoice{}).Where("id = ?", invoiceID).
			Update("paid_amount", gorm.Expr("paid_amount + ?", allocAmt))

		// Check if fully paid
		var inv ErpSalesInvoice
		if db.First(&inv, "id = ?", invoiceID).Error == nil {
			if inv.PaidAmount >= inv.Total {
				db.Model(&inv).Update("status", "paid")
			}
		}
	}

	// FIFO auto-allocate to unpaid invoices
	if len(allocations) == 0 {
		remaining := receipt.Amount

		var invoices []ErpSalesInvoice
		db.Where("customer_id = ? AND status IN ?", receipt.CustomerID, []string{"issued", "partially_paid"}).
			Order("due_date ASC").Find(&invoices)

		for _, inv := range invoices {
			if remaining <= 0 {
				break
			}

			balance := inv.Total - inv.PaidAmount
			if balance <= 0 {
				continue
			}

			allocAmount := balance
			if allocAmount > remaining {
				allocAmount = remaining
			}

			// Create allocation
			db.Create(&ErpReceiptAllocation{
				ID:        uuid.New().String(),
				ReceiptID: receipt.ID,
				InvoiceID: inv.ID,
				Amount:    allocAmount,
			})

			// Update invoice paid amount
			newPaid := inv.PaidAmount + allocAmount
			newStatus := "partially_paid"
			if newPaid >= inv.Total {
				newStatus = "paid"
			}
			db.Model(&inv).Updates(map[string]any{
				"paid_amount": newPaid,
				"status":      newStatus,
			})

			remaining -= allocAmount
		}
	}

	emitEvent(adapter, runID, stepID, "erp.receipt.recorded",
		fmt.Sprintf("收款记录: %s, %.2f", code, amount),
		map[string]any{"receipt_id": receipt.ID, "code": code, "amount": amount})

	return map[string]any{"receipt_id": receipt.ID, "code": code}, nil
}

func cmdCreateReturn(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	orderID := getStr(input, "order_id")
	if orderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}

	customerID := getStr(input, "customer_id")
	if customerID == "" {
		var order ErpSalesOrder
		if db.First(&order, "id = ?", orderID).Error == nil {
			customerID = order.CustomerID
		}
	}

	var ret ErpReturn
	code, err := autoCodeWithRetry(db, "erp_returns", "RMA", func(c string) error {
		ret = ErpReturn{
			ID:          uuid.New().String(),
			Code:        c,
			OrderID:     orderID,
			CustomerID:  customerID,
			Reason:      getStr(input, "reason"),
			Type:        getStr(input, "type"),
			TotalAmount: getFloat(input, "total_amount"),
			Status:      "pending",
		}
		if ret.Type == "" {
			ret.Type = "return"
		}
		return db.Create(&ret).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create return failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.return.created",
		fmt.Sprintf("退货单创建: %s", code),
		map[string]any{"return_id": ret.ID, "code": code, "order_id": orderID})

	return map[string]any{"return_id": ret.ID, "code": code}, nil
}

// ===========================================================================
// Inventory commands (7)
// ===========================================================================

// cmdReceiveInventory — PRD v1 Sprint 1 重写
//
// 收货入库。一笔收货 = 一行 ErpInventory（按 material × warehouse × location ×
// lot × status 唯一），同一组合存在则累加 quantity，不存在则新建。同时写一条
// ErpInventoryTransaction 流水。Sprint 2 会在这里加移动加权平均 + 凭证生成。
//
// 接受 warehouse_id 也接受 warehouse code（如 "WH-FG-01"）。lazy create
// ErpMaterialInventoryAttrs。
func cmdReceiveInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseRef := getStr(input, "warehouse_id")
	qty := getFloat(input, "quantity")
	if materialID == "" {
		return nil, fmt.Errorf("material_id is required")
	}
	if warehouseRef == "" {
		return nil, fmt.Errorf("warehouse_id is required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	warehouseID := resolveWarehouseID(db, warehouseRef)
	locationID := getStr(input, "location_id")
	lotNumber := getStr(input, "lot_number")
	unitCost := getFloat(input, "unit_cost")
	now := time.Now()

	status := getStr(input, "status")
	if status == "" {
		status = "available"
	}

	// Lazy-create material inventory attrs
	ensureInventoryAttrs(db, materialID, getStr(input, "unit"), warehouseID)

	var txn ErpInventoryTransaction
	var inv ErpInventory
	var journalEntryID string
	var newAvgCost float64

	// PRD v1 Sprint 2: period lock check
	period := now.Format("2006-01")
	if isPeriodLocked(db, period) {
		return nil, fmt.Errorf("accounting period %s is closed; cannot post inventory transactions", period)
	}

	// Determine the txn type for journaling — production_in if from work
	// order, else po_receipt for purchase receipts.
	txnType := "po_receipt"
	if rt := getStr(input, "reference_type"); rt == "work_order" {
		txnType = "production_in"
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// Upsert inventory by composite key
		query := tx.Where(
			"material_id = ? AND warehouse_id = ? AND COALESCE(location_id, '') = ? AND COALESCE(lot_number, '') = ? AND status = ?",
			materialID, warehouseID, locationID, lotNumber, status,
		)
		if err := query.First(&inv).Error; err != nil {
			inv = ErpInventory{
				ID:          uuid.New().String(),
				MaterialID:  materialID,
				WarehouseID: warehouseID,
				LocationID:  locationID,
				LotNumber:   lotNumber,
				Status:      status,
				Quantity:    qty,
				UnitCost:    unitCost,
				ReceivedAt:  &now,
				SourceType:  getStr(input, "reference_type"),
				SourceID:    getStr(input, "reference_id"),
				SupplierID:  getStr(input, "supplier_id"),
				SupplierLot: getStr(input, "supplier_lot"),
				Notes:       getStr(input, "notes"),
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if expiry := parseDate(getStr(input, "expiry_date")); expiry != nil {
				inv.ExpiryDate = expiry
			}
			if err := tx.Create(&inv).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&inv).Updates(map[string]any{
				"quantity":   gorm.Expr("quantity + ?", qty),
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
		}

		// PRD v1 Sprint 2: 移动加权平均成本
		avg, err := updateMovingAverageCost(tx, materialID, warehouseID, qty, unitCost)
		if err != nil {
			return fmt.Errorf("update avg cost: %w", err)
		}
		newAvgCost = avg

		code, err := autoCodeWithRetry(tx, "erp_inventory_transactions", "IT", func(c string) error {
			txn = ErpInventoryTransaction{
				ID:             uuid.New().String(),
				Code:           c,
				Type:           txnType,
				MaterialID:     materialID,
				ToWarehouseID:  warehouseID,
				ToLocationID:   locationID,
				ToStatus:       status,
				Quantity:       qty,
				UnitCost:       unitCost,
				TotalAmount:    qty * unitCost,
				LotNumber:      lotNumber,
				ReferenceType:  getStr(input, "reference_type"),
				ReferenceID:    getStr(input, "reference_id"),
				Reason:         getStr(input, "reason"),
				Notes:          getStr(input, "notes"),
				OperatorID:     getStr(input, "operator_id"),
				OperatorName:   getStr(input, "operator_name"),
				CreatedBy:      getStr(input, "created_by"),
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			return tx.Create(&txn).Error
		})
		if err != nil {
			return err
		}
		_ = code

		// PRD v1 Sprint 2: 自动生成会计凭证
		jeID, err := postInventoryJournal(tx, &txn, getStr(input, "supplier_id"))
		if err != nil {
			return fmt.Errorf("post inventory journal: %w", err)
		}
		if jeID != "" {
			journalEntryID = jeID
			tx.Model(&txn).Update("journal_entry_id", jeID)
		}

		tx.Model(&ErpMaterialInventoryAttrs{}).
			Where("material_id = ?", materialID).
			Updates(map[string]any{"last_received_at": &now, "updated_at": now})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("receive inventory failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.received",
		fmt.Sprintf("入库: %s, 物料 %s, 数量 %.2f, 仓库 %s, 成本 %.4f", txn.Code, materialID, qty, warehouseID, newAvgCost),
		map[string]any{
			"transaction_id":   txn.ID,
			"code":             txn.Code,
			"material_id":      materialID,
			"warehouse_id":     warehouseID,
			"lot_number":       lotNumber,
			"quantity":         qty,
			"unit_cost":        unitCost,
			"avg_cost":         newAvgCost,
			"total_amount":     qty * unitCost,
			"journal_entry_id": journalEntryID,
		})

	return map[string]any{
		"transaction_id":   txn.ID,
		"code":             txn.Code,
		"inventory_id":     inv.ID,
		"material_id":      materialID,
		"warehouse_id":     warehouseID,
		"quantity":         qty,
		"avg_cost":         newAvgCost,
		"journal_entry_id": journalEntryID,
	}, nil
}

// cmdIssueInventory — PRD v1 Sprint 1 重写
//
// 出库（生产领料 / 销售出库 / 手工出库）。FIFO 选批次（按 received_at 升序），
// 不够时跨多个批次扣减。每个批次扣完写一行流水。
func cmdIssueInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseRef := getStr(input, "warehouse_id")
	qty := getFloat(input, "quantity")
	if materialID == "" {
		return nil, fmt.Errorf("material_id is required")
	}
	if warehouseRef == "" {
		return nil, fmt.Errorf("warehouse_id is required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	warehouseID := resolveWarehouseID(db, warehouseRef)
	now := time.Now()
	txnType := getStr(input, "type")
	if txnType == "" {
		txnType = "production_issue"
	}

	period := now.Format("2006-01")
	if isPeriodLocked(db, period) {
		return nil, fmt.Errorf("accounting period %s is closed; cannot post inventory transactions", period)
	}

	var totalIssued float64
	var totalCost float64
	var txnIDs []string
	var firstTxnCode string
	var journalEntryIDs []string

	err := db.Transaction(func(tx *gorm.DB) error {
		// FIFO: pick batches ordered by received_at; only available status counts
		var batches []ErpInventory
		if err := tx.Where(
			"material_id = ? AND warehouse_id = ? AND status = ? AND quantity - reserved_qty > 0",
			materialID, warehouseID, "available",
		).Order("received_at ASC NULLS LAST, created_at ASC").Find(&batches).Error; err != nil {
			return err
		}

		var totalAvailable float64
		for _, b := range batches {
			totalAvailable += b.Quantity - b.ReservedQty
		}
		if totalAvailable < qty {
			return fmt.Errorf("insufficient stock for material %s in warehouse %s: available %.2f, requested %.2f",
				materialID, warehouseID, totalAvailable, qty)
		}

		remaining := qty
		for i := range batches {
			if remaining <= 0 {
				break
			}
			b := &batches[i]
			batchAvail := b.Quantity - b.ReservedQty
			take := batchAvail
			if take > remaining {
				take = remaining
			}
			if take <= 0 {
				continue
			}
			if err := tx.Model(b).Updates(map[string]any{
				"quantity":   gorm.Expr("quantity - ?", take),
				"updated_at": now,
			}).Error; err != nil {
				return err
			}

			var txn ErpInventoryTransaction
			code, err := autoCodeWithRetry(tx, "erp_inventory_transactions", "IT", func(c string) error {
				txn = ErpInventoryTransaction{
					ID:               uuid.New().String(),
					Code:             c,
					Type:             txnType,
					MaterialID:       materialID,
					FromWarehouseID:  warehouseID,
					FromLocationID:   b.LocationID,
					FromStatus:       b.Status,
					Quantity:         take,
					UnitCost:         b.UnitCost,
					TotalAmount:      take * b.UnitCost,
					LotNumber:        b.LotNumber,
					ReferenceType:    getStr(input, "reference_type"),
					ReferenceID:      getStr(input, "reference_id"),
					Reason:           getStr(input, "reason"),
					Notes:            getStr(input, "notes"),
					OperatorID:       getStr(input, "operator_id"),
					OperatorName:     getStr(input, "operator_name"),
					CreatedBy:        getStr(input, "created_by"),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				return tx.Create(&txn).Error
			})
			if err != nil {
				return err
			}

			// PRD v1 Sprint 2: 出库自动生成会计凭证（每个批次一笔）
			jeID, err := postInventoryJournal(tx, &txn, b.SupplierID)
			if err != nil {
				return fmt.Errorf("post inventory journal: %w", err)
			}
			if jeID != "" {
				tx.Model(&txn).Update("journal_entry_id", jeID)
				journalEntryIDs = append(journalEntryIDs, jeID)
			}

			txnIDs = append(txnIDs, txn.ID)
			if firstTxnCode == "" {
				firstTxnCode = code
			}
			totalIssued += take
			totalCost += take * b.UnitCost
			remaining -= take
		}

		tx.Model(&ErpMaterialInventoryAttrs{}).
			Where("material_id = ?", materialID).
			Updates(map[string]any{"last_issued_at": &now, "updated_at": now})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("issue inventory failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.issued",
		fmt.Sprintf("出库: %s, 物料 %s, 数量 %.2f, 总成本 %.2f", firstTxnCode, materialID, totalIssued, totalCost),
		map[string]any{
			"transaction_ids":   txnIDs,
			"first_code":        firstTxnCode,
			"material_id":       materialID,
			"warehouse_id":      warehouseID,
			"quantity":          totalIssued,
			"total_cost":        totalCost,
			"batches_used":      len(txnIDs),
			"journal_entry_ids": journalEntryIDs,
		})

	return map[string]any{
		"transaction_ids":   txnIDs,
		"code":              firstTxnCode,
		"material_id":       materialID,
		"warehouse_id":      warehouseID,
		"quantity":          totalIssued,
		"total_cost":        totalCost,
		"journal_entry_ids": journalEntryIDs,
	}, nil
}

// cmdTransferInventory — PRD v1 Sprint 1 重写
//
// 仓间 / 库位间 / 状态间转储。从源批次扣减，目标位置 upsert。
// 支持状态变更（如 available → quality_hold 用 from_status/to_status 描述）。
func cmdTransferInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	fromWarehouseRef := getStr(input, "from_warehouse_id")
	toWarehouseRef := getStr(input, "to_warehouse_id")
	qty := getFloat(input, "quantity")
	if materialID == "" || fromWarehouseRef == "" || toWarehouseRef == "" {
		return nil, fmt.Errorf("material_id, from_warehouse_id, to_warehouse_id are required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	fromWarehouseID := resolveWarehouseID(db, fromWarehouseRef)
	toWarehouseID := resolveWarehouseID(db, toWarehouseRef)
	fromLocation := getStr(input, "from_location_id")
	toLocation := getStr(input, "to_location_id")
	lot := getStr(input, "lot_number")
	fromStatus := getStr(input, "from_status")
	if fromStatus == "" {
		fromStatus = "available"
	}
	toStatus := getStr(input, "to_status")
	if toStatus == "" {
		toStatus = "available"
	}
	now := time.Now()

	var txn ErpInventoryTransaction

	err := db.Transaction(func(tx *gorm.DB) error {
		query := tx.Where(
			"material_id = ? AND warehouse_id = ? AND COALESCE(location_id,'') = ? AND COALESCE(lot_number,'') = ? AND status = ?",
			materialID, fromWarehouseID, fromLocation, lot, fromStatus,
		)
		var srcInv ErpInventory
		if err := query.First(&srcInv).Error; err != nil {
			return fmt.Errorf("no inventory at source (material=%s warehouse=%s lot=%s status=%s)", materialID, fromWarehouseID, lot, fromStatus)
		}
		available := srcInv.Quantity - srcInv.ReservedQty
		if available < qty {
			return fmt.Errorf("insufficient stock: available %.2f, requested %.2f", available, qty)
		}

		// Decrease source
		if err := tx.Model(&srcInv).Updates(map[string]any{
			"quantity":   gorm.Expr("quantity - ?", qty),
			"updated_at": now,
		}).Error; err != nil {
			return err
		}

		// Upsert destination
		var dstInv ErpInventory
		dstQuery := tx.Where(
			"material_id = ? AND warehouse_id = ? AND COALESCE(location_id,'') = ? AND COALESCE(lot_number,'') = ? AND status = ?",
			materialID, toWarehouseID, toLocation, lot, toStatus,
		)
		if err := dstQuery.First(&dstInv).Error; err != nil {
			dstInv = ErpInventory{
				ID:          uuid.New().String(),
				MaterialID:  materialID,
				WarehouseID: toWarehouseID,
				LocationID:  toLocation,
				LotNumber:   lot,
				Status:      toStatus,
				Quantity:    qty,
				UnitCost:    srcInv.UnitCost,
				ReceivedAt:  srcInv.ReceivedAt,
				ExpiryDate:  srcInv.ExpiryDate,
				SourceType:  "transfer",
				SourceID:    srcInv.ID,
				SupplierID:  srcInv.SupplierID,
				SupplierLot: srcInv.SupplierLot,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := tx.Create(&dstInv).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&dstInv).Updates(map[string]any{
				"quantity":   gorm.Expr("quantity + ?", qty),
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
		}

		_, err := autoCodeWithRetry(tx, "erp_inventory_transactions", "IT", func(c string) error {
			txn = ErpInventoryTransaction{
				ID:              uuid.New().String(),
				Code:            c,
				Type:            "transfer_out",
				MaterialID:      materialID,
				FromWarehouseID: fromWarehouseID,
				FromLocationID:  fromLocation,
				FromStatus:      fromStatus,
				ToWarehouseID:   toWarehouseID,
				ToLocationID:    toLocation,
				ToStatus:        toStatus,
				Quantity:        qty,
				UnitCost:        srcInv.UnitCost,
				TotalAmount:     qty * srcInv.UnitCost,
				LotNumber:       lot,
				ReferenceType:   getStr(input, "reference_type"),
				ReferenceID:     getStr(input, "reference_id"),
				Reason:          getStr(input, "reason"),
				Notes:           getStr(input, "notes"),
				OperatorID:      getStr(input, "operator_id"),
				OperatorName:    getStr(input, "operator_name"),
				CreatedBy:       getStr(input, "created_by"),
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			return tx.Create(&txn).Error
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("transfer inventory failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.transferred",
		fmt.Sprintf("调拨: %s, %s → %s, 数量 %.2f", txn.Code, fromWarehouseID, toWarehouseID, qty),
		map[string]any{
			"transaction_id":    txn.ID,
			"code":              txn.Code,
			"material_id":       materialID,
			"from_warehouse_id": fromWarehouseID,
			"to_warehouse_id":   toWarehouseID,
			"quantity":          qty,
		})

	return map[string]any{
		"transaction_id":    txn.ID,
		"code":              txn.Code,
		"material_id":       materialID,
		"from_warehouse_id": fromWarehouseID,
		"to_warehouse_id":   toWarehouseID,
		"quantity":          qty,
	}, nil
}

// cmdAdjustInventory — PRD v1 Sprint 1 重写
//
// 库存调整。把指定批次的数量改成 new_quantity，差值生成 adjustment_in/out 流水。
// 大额调整（金额 ≥ 500）应该走审批流，目前通过 input.approval_id 标记已审批。
func cmdAdjustInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseRef := getStr(input, "warehouse_id")
	newQty := getFloat(input, "new_quantity")
	if materialID == "" || warehouseRef == "" {
		return nil, fmt.Errorf("material_id and warehouse_id are required")
	}

	warehouseID := resolveWarehouseID(db, warehouseRef)
	locationID := getStr(input, "location_id")
	lot := getStr(input, "lot_number")
	status := getStr(input, "status")
	if status == "" {
		status = "available"
	}
	now := time.Now()
	approvalID := getStr(input, "approval_id")

	var txn ErpInventoryTransaction
	var oldQty float64

	err := db.Transaction(func(tx *gorm.DB) error {
		var inv ErpInventory
		query := tx.Where(
			"material_id = ? AND warehouse_id = ? AND COALESCE(location_id,'') = ? AND COALESCE(lot_number,'') = ? AND status = ?",
			materialID, warehouseID, locationID, lot, status,
		)
		exists := query.First(&inv).Error == nil
		if exists {
			oldQty = inv.Quantity
		}
		diff := newQty - oldQty
		if diff == 0 {
			return nil
		}

		// Approval gate: large adjustments need approval
		threshold := 500.0
		// If we have a unit_cost we can compute amount; otherwise treat by quantity diff alone
		amount := diff
		if exists {
			amount = diff * inv.UnitCost
		}
		if amount < 0 {
			amount = -amount
		}
		if amount >= threshold && approvalID == "" {
			return fmt.Errorf("adjustment of %.2f requires approval (set approval_id)", amount)
		}

		txnType := "adjustment_in"
		if diff < 0 {
			txnType = "adjustment_out"
		}

		if exists {
			if err := tx.Model(&inv).Updates(map[string]any{
				"quantity":   newQty,
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
		} else {
			inv = ErpInventory{
				ID:          uuid.New().String(),
				MaterialID:  materialID,
				WarehouseID: warehouseID,
				LocationID:  locationID,
				LotNumber:   lot,
				Status:      status,
				Quantity:    newQty,
				ReceivedAt:  &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := tx.Create(&inv).Error; err != nil {
				return err
			}
		}

		absDiff := diff
		if absDiff < 0 {
			absDiff = -absDiff
		}

		_, err := autoCodeWithRetry(tx, "erp_inventory_transactions", "IT", func(c string) error {
			txn = ErpInventoryTransaction{
				ID:              uuid.New().String(),
				Code:            c,
				Type:            txnType,
				MaterialID:      materialID,
				ToWarehouseID:   warehouseID,
				ToLocationID:    locationID,
				ToStatus:        status,
				FromWarehouseID: warehouseID,
				FromStatus:      status,
				Quantity:        absDiff,
				UnitCost:        inv.UnitCost,
				TotalAmount:     absDiff * inv.UnitCost,
				LotNumber:       lot,
				ReferenceType:   "adjustment",
				ReferenceID:     getStr(input, "reference_id"),
				Reason:          getStr(input, "reason"),
				Notes:           getStr(input, "notes"),
				ApprovalID:      approvalID,
				OperatorID:      getStr(input, "operator_id"),
				OperatorName:    getStr(input, "operator_name"),
				CreatedBy:       getStr(input, "created_by"),
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			return tx.Create(&txn).Error
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("adjust inventory failed: %w", err)
	}
	if txn.ID == "" {
		// no diff
		return map[string]any{"old_qty": oldQty, "new_qty": newQty, "diff": 0.0, "skipped": true}, nil
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.adjusted",
		fmt.Sprintf("库存调整: %s, %.2f → %.2f, 物料 %s", txn.Code, oldQty, newQty, materialID),
		map[string]any{
			"transaction_id": txn.ID,
			"code":           txn.Code,
			"material_id":    materialID,
			"old_qty":        oldQty,
			"new_qty":        newQty,
			"diff":           newQty - oldQty,
			"approval_id":    approvalID,
		})

	return map[string]any{
		"transaction_id": txn.ID,
		"code":           txn.Code,
		"material_id":    materialID,
		"old_qty":        oldQty,
		"new_qty":        newQty,
		"diff":           newQty - oldQty,
	}, nil
}

func cmdScrapInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseID := getStr(input, "warehouse_id")
	qty := getFloat(input, "quantity")
	if materialID == "" || warehouseID == "" {
		return nil, fmt.Errorf("material_id and warehouse_id are required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	var inv ErpInventory
	if err := db.Where("material_id = ? AND warehouse_id = ?", materialID, warehouseID).First(&inv).Error; err != nil {
		return nil, fmt.Errorf("no inventory record for this material/warehouse")
	}
	if inv.Quantity < qty {
		return nil, fmt.Errorf("insufficient stock: available %.2f, requested %.2f", inv.Quantity, qty)
	}

	// Create transaction
	var txn ErpInventoryTransaction
	code, err := autoCodeWithRetry(db, "erp_inventory_transactions", "IT", func(c string) error {
		txn = ErpInventoryTransaction{
			ID:              uuid.New().String(),
			Code:            c,
			Type:            "scrap",
			MaterialID:      materialID,
			FromWarehouseID: warehouseID,
			Quantity:        qty,
			UnitCost:        inv.UnitCost,
			Notes:           getStr(input, "reason"),
			CreatedBy:       getStr(input, "created_by"),
		}
		return db.Create(&txn).Error
	})
	if err != nil {
		return nil, fmt.Errorf("scrap inventory failed: %w", err)
	}

	// Decrease inventory
	db.Model(&inv).Update("quantity", gorm.Expr("quantity - ?", qty))

	emitEvent(adapter, runID, stepID, "erp.inventory.scrapped",
		fmt.Sprintf("报废: %s, 数量%.2f", code, qty),
		map[string]any{"transaction_id": txn.ID, "code": code, "material_id": materialID, "quantity": qty})

	return map[string]any{"transaction_id": txn.ID, "code": code}, nil
}

// cmdReserveInventory — PRD v1 Sprint 3 重写
//
// 为订单/工单预留库存：
// 1) 校验 available - reserved_qty >= qty
// 2) 创建 ErpInventoryReservation 记录（source_type, source_id, priority, expires_at）
// 3) 增加 ErpInventory.reserved_qty
// 同一 source_id 再次预留会追加（而非覆盖）。
func cmdReserveInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseRef := getStr(input, "warehouse_id")
	qty := getFloat(input, "quantity")
	if materialID == "" || warehouseRef == "" {
		return nil, fmt.Errorf("material_id and warehouse_id are required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	warehouseID := resolveWarehouseID(db, warehouseRef)
	lot := getStr(input, "lot_number")
	sourceType := getStr(input, "source_type")
	sourceID := getStr(input, "source_id")
	if sourceID == "" {
		// Backward compat with old OrderItemID key
		sourceID = getStr(input, "order_item_id")
		if sourceID != "" && sourceType == "" {
			sourceType = "sales_order"
		}
	}

	var reservation ErpInventoryReservation
	err := db.Transaction(func(tx *gorm.DB) error {
		// Pick the target inventory row (same warehouse, available status)
		var inv ErpInventory
		q := tx.Where("material_id = ? AND warehouse_id = ? AND status = ?", materialID, warehouseID, "available")
		if lot != "" {
			q = q.Where("COALESCE(lot_number,'') = ?", lot)
		}
		if err := q.Order("received_at ASC NULLS LAST").First(&inv).Error; err != nil {
			return fmt.Errorf("no available inventory for this material/warehouse/lot")
		}
		available := inv.Quantity - inv.ReservedQty
		if available < qty {
			return fmt.Errorf("insufficient available stock: %.2f, requested %.2f", available, qty)
		}
		if err := tx.Model(&inv).Update("reserved_qty", gorm.Expr("reserved_qty + ?", qty)).Error; err != nil {
			return err
		}

		now := time.Now()
		_, cerr := autoCodeWithRetry(tx, "erp_inventory_reservations", "RV", func(c string) error {
			reservation = ErpInventoryReservation{
				ID:          uuid.New().String(),
				Code:        c,
				MaterialID:  materialID,
				WarehouseID: warehouseID,
				LotNumber:   inv.LotNumber,
				ReservedQty: qty,
				SourceType:  sourceType,
				SourceID:    sourceID,
				Priority:    getInt(input, "priority", 0),
				Status:      "active",
				ReservedBy:  getStr(input, "operator_id"),
				Notes:       getStr(input, "notes"),
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			return tx.Create(&reservation).Error
		})
		return cerr
	})
	if err != nil {
		return nil, fmt.Errorf("reserve inventory failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.reserved",
		fmt.Sprintf("库存预留: 物料 %s, 数量 %.2f, 来源 %s/%s", materialID, qty, sourceType, sourceID),
		map[string]any{"reservation_id": reservation.ID, "material_id": materialID, "warehouse_id": warehouseID, "quantity": qty, "source_type": sourceType, "source_id": sourceID})

	return map[string]any{
		"reservation_id": reservation.ID,
		"code":           reservation.Code,
		"material_id":    materialID,
		"warehouse_id":   warehouseID,
		"quantity":       qty,
		"status":         "active",
	}, nil
}

// cmdUnreserveInventory — PRD v1 Sprint 3 重写
//
// 释放预留：接受 reservation_id 或 source_id/source_type。
// 找到 active 记录，标记 released，减少对应 ErpInventory.reserved_qty。
func cmdUnreserveInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	reservationID := getStr(input, "reservation_id")
	sourceID := getStr(input, "source_id")
	if sourceID == "" {
		sourceID = getStr(input, "order_item_id")
	}

	var totalReleased float64
	var count int
	err := db.Transaction(func(tx *gorm.DB) error {
		var reservations []ErpInventoryReservation
		q := tx.Where("status = ?", "active")
		if reservationID != "" {
			q = q.Where("id = ?", reservationID)
		} else if sourceID != "" {
			q = q.Where("source_id = ?", sourceID)
		} else {
			return fmt.Errorf("reservation_id or source_id is required")
		}
		if err := q.Find(&reservations).Error; err != nil {
			return err
		}
		if len(reservations) == 0 {
			return fmt.Errorf("no active reservation found")
		}
		now := time.Now()
		for _, r := range reservations {
			remaining := r.ReservedQty - r.ConsumedQty
			if remaining <= 0 {
				continue
			}
			if err := tx.Model(&r).Updates(map[string]any{
				"status":     "released",
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
			// Decrement inventory reserved_qty
			if err := tx.Model(&ErpInventory{}).
				Where("material_id = ? AND warehouse_id = ? AND COALESCE(lot_number,'') = ? AND status = ?",
					r.MaterialID, r.WarehouseID, r.LotNumber, "available").
				Update("reserved_qty", gorm.Expr("GREATEST(reserved_qty - ?, 0)", remaining)).Error; err != nil {
				return err
			}
			totalReleased += remaining
			count++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unreserve inventory failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.unreserved",
		fmt.Sprintf("取消预留: %d 笔, 释放数量 %.2f", count, totalReleased),
		map[string]any{"count": count, "total_released": totalReleased})

	return map[string]any{
		"released_count":    count,
		"total_released_qty": totalReleased,
	}, nil
}

// ===========================================================================
// Production commands (7)
// ===========================================================================

func cmdRunMRP(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	mrpRunID := uuid.New().String()

	// Get all confirmed sales orders
	var orders []ErpSalesOrder
	db.Where("status IN ?", []string{"confirmed", "producing"}).Find(&orders)

	suggestionCount := 0
	for _, order := range orders {
		var items []ErpSalesOrderItem
		db.Where("order_id = ? AND status = ?", order.ID, "pending").Find(&items)

		for _, item := range items {
			if item.BomID == "" {
				continue
			}
			// Get BOM items from PLM
			var bomItems []struct {
				MaterialID string  `gorm:"column:material_id"`
				Quantity   float64 `gorm:"column:quantity"`
			}
			db.Table("plm_bom_items").Where("bom_id = ?", item.BomID).Find(&bomItems)

			for _, bi := range bomItems {
				gross := bi.Quantity * item.Quantity

				// Get on-hand
				var onHand float64
				db.Table("erp_inventory").Where("material_id = ?", bi.MaterialID).
					Select("COALESCE(SUM(quantity - reserved_qty), 0)").Row().Scan(&onHand)

				// Get on-order (from SRM)
				var onOrder float64
				db.Table("srm_purchase_order_items poi").
					Joins("JOIN srm_purchase_orders po ON po.id = poi.purchase_order_id").
					Where("poi.material_id = ? AND po.status IN ?", bi.MaterialID, []string{"approved", "ordered"}).
					Select("COALESCE(SUM(poi.quantity - poi.received_qty), 0)").Row().Scan(&onOrder)

				net := gross - onHand - onOrder
				if net <= 0 {
					continue
				}

				// Check if material has BOM (semi-finished) -> produce, otherwise -> purchase
				action := "purchase"
				var bomCount int64
				db.Table("plm_boms").Where("product_id = ? AND status = ?", bi.MaterialID, "released").Count(&bomCount)
				if bomCount > 0 {
					action = "produce"
				}

				mrpResult := ErpMRPResult{
					ID:               uuid.New().String(),
					RunID:            mrpRunID,
					MaterialID:       bi.MaterialID,
					DemandSource:     "so",
					DemandID:         order.ID,
					GrossRequirement: gross,
					OnHand:           onHand,
					OnOrder:          onOrder,
					NetRequirement:   net,
					Action:           action,
					SuggestedQty:     net,
					BomID:            item.BomID,
					Status:           "suggested",
				}
				db.Create(&mrpResult)
				suggestionCount++
			}
		}
	}

	emitEvent(adapter, runID, stepID, "erp.mrp.completed",
		fmt.Sprintf("MRP运算完成: %d条建议", suggestionCount),
		map[string]any{"run_id": mrpRunID, "suggestion_count": suggestionCount})

	return map[string]any{"run_id": mrpRunID, "suggestion_count": suggestionCount}, nil
}

func cmdConfirmMRPSuggestion(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	resultID := getStr(input, "result_id")
	if resultID == "" {
		return nil, fmt.Errorf("result_id is required")
	}
	var mrpResult ErpMRPResult
	if err := db.First(&mrpResult, "id = ?", resultID).Error; err != nil {
		return nil, fmt.Errorf("MRP result not found")
	}
	if mrpResult.Status != "suggested" {
		return nil, fmt.Errorf("MRP result not in suggested status, current: %s", mrpResult.Status)
	}

	db.Model(&mrpResult).Update("status", "confirmed")

	emitEvent(adapter, runID, stepID, "erp.mrp.suggestion_confirmed",
		fmt.Sprintf("MRP建议确认: %s, 动作=%s, 数量=%.2f", mrpResult.MaterialID, mrpResult.Action, mrpResult.SuggestedQty),
		map[string]any{"result_id": mrpResult.ID, "action": mrpResult.Action, "quantity": mrpResult.SuggestedQty})

	return map[string]any{"result_id": mrpResult.ID, "action": mrpResult.Action}, nil
}

func cmdCreateWorkOrder(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	productID := getStr(input, "product_id")
	if productID == "" {
		return nil, fmt.Errorf("product_id is required")
	}
	plannedQty := getFloat(input, "planned_qty")
	if plannedQty <= 0 {
		return nil, fmt.Errorf("planned_qty must be positive")
	}

	var wo ErpWorkOrder
	code, err := autoCodeWithRetry(db, "erp_work_orders", "WO", func(c string) error {
		wo = ErpWorkOrder{
			ID:           uuid.New().String(),
			Code:         c,
			ProductID:    productID,
			BomID:        getStr(input, "bom_id"),
			OrderID:      getStr(input, "order_id"),
			MrpResultID:  getStr(input, "mrp_result_id"),
			PlannedQty:   plannedQty,
			WarehouseID:  getStr(input, "warehouse_id"),
			PlannedStart: parseDate(getStr(input, "planned_start")),
			PlannedEnd:   parseDate(getStr(input, "planned_end")),
			Priority:     getStr(input, "priority"),
			Notes:        getStr(input, "notes"),
			Status:       "draft",
			CreatedBy:    getStr(input, "created_by"),
		}
		if wo.Priority == "" {
			wo.Priority = "normal"
		}
		return db.Create(&wo).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create work order failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.work_order.created",
		fmt.Sprintf("生产工单创建: %s, 数量%.0f", code, plannedQty),
		map[string]any{"work_order_id": wo.ID, "code": code, "product_id": productID})

	return map[string]any{"work_order_id": wo.ID, "code": code}, nil
}

func cmdReleaseWorkOrder(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	woID := getStr(input, "work_order_id")
	if woID == "" {
		return nil, fmt.Errorf("work_order_id is required")
	}
	var wo ErpWorkOrder
	if err := db.First(&wo, "id = ?", woID).Error; err != nil {
		return nil, fmt.Errorf("work order not found")
	}
	if wo.Status != "draft" {
		return nil, fmt.Errorf("work order not in draft status, current: %s", wo.Status)
	}

	now := time.Now()
	db.Model(&wo).Updates(map[string]any{"status": "released", "actual_start": &now})

	// Auto-create material issue records from BOM
	if wo.BomID != "" {
		var bomItems []struct {
			ID         string  `gorm:"column:id"`
			MaterialID string  `gorm:"column:material_id"`
			Quantity   float64 `gorm:"column:quantity"`
		}
		db.Table("plm_bom_items").Where("bom_id = ?", wo.BomID).Find(&bomItems)
		for _, bi := range bomItems {
			issue := ErpWOMaterialIssue{
				ID:          uuid.New().String(),
				WorkOrderID: wo.ID,
				MaterialID:  bi.MaterialID,
				BomItemID:   bi.ID,
				RequiredQty: bi.Quantity * wo.PlannedQty,
				WarehouseID: wo.WarehouseID,
			}
			db.Create(&issue)
		}
	}

	emitEvent(adapter, runID, stepID, "erp.work_order.released",
		fmt.Sprintf("工单下达: %s", wo.Code),
		map[string]any{"work_order_id": wo.ID, "code": wo.Code})

	return map[string]any{"work_order_id": wo.ID, "code": wo.Code}, nil
}

func cmdIssueWOMaterials(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	woID := getStr(input, "work_order_id")
	if woID == "" {
		return nil, fmt.Errorf("work_order_id is required")
	}
	var wo ErpWorkOrder
	if err := db.First(&wo, "id = ?", woID).Error; err != nil {
		return nil, fmt.Errorf("work order not found")
	}
	if wo.Status != "released" && wo.Status != "in_progress" {
		return nil, fmt.Errorf("work order not in released/in_progress status, current: %s", wo.Status)
	}

	// Issue materials
	issues := getMapSlice(input, "issues")
	issuedCount := 0
	for _, iss := range issues {
		issueID := getStr(iss, "issue_id")
		qty := getFloat(iss, "quantity")
		if issueID == "" || qty <= 0 {
			continue
		}

		var mi ErpWOMaterialIssue
		if db.First(&mi, "id = ?", issueID).Error != nil {
			continue
		}

		// Deduct from inventory
		warehouseID := mi.WarehouseID
		if warehouseID == "" {
			warehouseID = wo.WarehouseID
		}

		var inv ErpInventory
		if db.Where("material_id = ? AND warehouse_id = ?", mi.MaterialID, warehouseID).First(&inv).Error == nil {
			available := inv.Quantity - inv.ReservedQty
			if available >= qty {
				db.Model(&inv).Update("quantity", gorm.Expr("quantity - ?", qty))
			} else {
				continue // skip if insufficient
			}
		} else {
			continue
		}

		now := time.Now()
		lotNumber := getStr(iss, "lot_number")
		db.Model(&mi).Updates(map[string]any{
			"issued_qty": gorm.Expr("issued_qty + ?", qty),
			"issued_at":  &now,
			"issued_by":  getStr(input, "issued_by"),
			"lot_number": lotNumber,
		})
		issuedCount++
	}

	// Update work order status to in_progress
	if wo.Status == "released" {
		db.Model(&wo).Update("status", "in_progress")
	}

	emitEvent(adapter, runID, stepID, "erp.wo.materials_issued",
		fmt.Sprintf("工单领料: %s, %d项", wo.Code, issuedCount),
		map[string]any{"work_order_id": wo.ID, "code": wo.Code, "issued_count": issuedCount})

	return map[string]any{"work_order_id": wo.ID, "issued_count": issuedCount}, nil
}

func cmdReportWOProgress(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	woID := getStr(input, "work_order_id")
	if woID == "" {
		return nil, fmt.Errorf("work_order_id is required")
	}
	var wo ErpWorkOrder
	if err := db.First(&wo, "id = ?", woID).Error; err != nil {
		return nil, fmt.Errorf("work order not found")
	}
	if wo.Status != "released" && wo.Status != "in_progress" {
		return nil, fmt.Errorf("work order not in active status, current: %s", wo.Status)
	}

	goodQty := getFloat(input, "good_qty")
	defectQty := getFloat(input, "defect_qty")
	scrapQty := getFloat(input, "scrap_qty")

	report := ErpWOReport{
		ID:          uuid.New().String(),
		WorkOrderID: woID,
		Operation:   getStr(input, "operation"),
		OperatorID:  getStr(input, "operator_id"),
		GoodQty:     goodQty,
		DefectQty:   defectQty,
		ScrapQty:    scrapQty,
		StartTime:   parseDate(getStr(input, "start_time")),
		EndTime:     parseDate(getStr(input, "end_time")),
		Notes:       getStr(input, "notes"),
	}
	db.Create(&report)

	// Update work order quantities
	db.Model(&wo).Updates(map[string]any{
		"completed_qty": gorm.Expr("completed_qty + ?", goodQty),
		"scrap_qty":     gorm.Expr("scrap_qty + ?", scrapQty),
		"status":        "in_progress",
	})

	emitEvent(adapter, runID, stepID, "erp.wo.progress_reported",
		fmt.Sprintf("工单报工: %s, 良品%.0f, 废品%.0f", wo.Code, goodQty, scrapQty),
		map[string]any{"work_order_id": wo.ID, "report_id": report.ID, "good_qty": goodQty})

	return map[string]any{"work_order_id": wo.ID, "report_id": report.ID}, nil
}

// cmdCompleteWorkOrder — PRD v1 Sprint 1 修复
//
// 完工工单：从 input 读 completed_qty 持久化到 WO 记录。
// 旧实现忘了从 input 读，直接用 wo.CompletedQty (永远是 0)，导致下游
// {{steps.mp_complete_wo.output.completed_qty}} 渲染为 0，mp_stock_in 失败。
//
// 不在这里做自动入库 — 由调用方流程显式调用 receive_inventory（与 PRD
// "工单完工 → 序列号生成 → 入库" 的多步分离一致，避免双重记账）。
func cmdCompleteWorkOrder(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	woID := getStr(input, "work_order_id")
	if woID == "" {
		return nil, fmt.Errorf("work_order_id is required")
	}
	var wo ErpWorkOrder
	if err := db.First(&wo, "id = ?", woID).Error; err != nil {
		return nil, fmt.Errorf("work order not found")
	}
	if wo.Status != "in_progress" && wo.Status != "released" {
		return nil, fmt.Errorf("work order not in active status, current: %s", wo.Status)
	}

	completedQty := getFloat(input, "completed_qty")
	if completedQty <= 0 {
		// Backward compat: if caller didn't pass it, use the WO's planned qty
		completedQty = wo.PlannedQty
	}
	if completedQty <= 0 {
		return nil, fmt.Errorf("completed_qty must be positive (input or wo.planned_qty)")
	}

	scrapQty := getFloat(input, "scrap_qty")
	now := time.Now()

	if err := db.Model(&wo).Updates(map[string]any{
		"status":        "completed",
		"completed_qty": completedQty,
		"scrap_qty":     scrapQty,
		"actual_end":    &now,
	}).Error; err != nil {
		return nil, fmt.Errorf("update work order failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.work_order.completed",
		fmt.Sprintf("工单完工: %s, 完成%.0f", wo.Code, completedQty),
		map[string]any{
			"work_order_id": wo.ID,
			"code":          wo.Code,
			"completed_qty": completedQty,
			"scrap_qty":     scrapQty,
		})

	return map[string]any{
		"work_order_id": wo.ID,
		"code":          wo.Code,
		"completed_qty": completedQty,
		"scrap_qty":     scrapQty,
	}, nil
}

// ===========================================================================
// Finance commands (5)
// ===========================================================================

func cmdCreateJournalEntry(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	entryDate := parseDate(getStr(input, "entry_date"))
	description := getStr(input, "description")
	lines := getMapSlice(input, "lines")

	if len(lines) == 0 {
		return nil, fmt.Errorf("lines required")
	}

	var totalDebit, totalCredit float64
	for _, l := range lines {
		totalDebit += getFloat(l, "debit")
		totalCredit += getFloat(l, "credit")
	}

	// Validate debit = credit
	if fmt.Sprintf("%.2f", totalDebit) != fmt.Sprintf("%.2f", totalCredit) {
		return nil, fmt.Errorf("借贷不平衡: 借方%.2f != 贷方%.2f", totalDebit, totalCredit)
	}

	var entry ErpJournalEntry
	code, err := autoCodeWithRetry(db, "erp_journal_entries", "JE", func(c string) error {
		period := ""
		if entryDate != nil {
			period = entryDate.Format("2006-01")
		}
		entry = ErpJournalEntry{
			ID:          uuid.New().String(),
			Code:        c,
			Period:      period,
			EntryDate:   entryDate,
			Description: description,
			TotalDebit:  totalDebit,
			TotalCredit: totalCredit,
			SourceType:  getStr(input, "source_type"),
			SourceID:    getStr(input, "source_id"),
			Status:      "draft",
		}
		return db.Create(&entry).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create journal entry failed: %w", err)
	}

	for _, l := range lines {
		line := ErpJournalLine{
			ID:          uuid.New().String(),
			EntryID:     entry.ID,
			AccountID:   getStr(l, "account_id"),
			Debit:       getFloat(l, "debit"),
			Credit:      getFloat(l, "credit"),
			Currency:    getStr(l, "currency"),
			Description: getStr(l, "description"),
			CustomerID:  getStr(l, "customer_id"),
			SupplierID:  getStr(l, "supplier_id"),
		}
		if line.Currency == "" {
			line.Currency = "CNY"
		}
		db.Create(&line)
	}

	emitEvent(adapter, runID, stepID, "erp.journal_entry.created",
		fmt.Sprintf("凭证创建: %s, 借方%.2f", code, totalDebit),
		map[string]any{"entry_id": entry.ID, "code": code})

	return map[string]any{"entry_id": entry.ID, "code": code}, nil
}

func cmdPostJournalEntry(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	entryID := getStr(input, "entry_id")
	if entryID == "" {
		return nil, fmt.Errorf("entry_id is required")
	}
	var entry ErpJournalEntry
	if err := db.First(&entry, "id = ?", entryID).Error; err != nil {
		return nil, fmt.Errorf("journal entry not found")
	}
	if entry.Status != "draft" {
		return nil, fmt.Errorf("journal entry not in draft status, current: %s", entry.Status)
	}

	now := time.Now()
	db.Model(&entry).Updates(map[string]any{
		"status":    "posted",
		"posted_by": getStr(input, "posted_by"),
		"posted_at": &now,
	})

	emitEvent(adapter, runID, stepID, "erp.journal_entry.posted",
		fmt.Sprintf("凭证过账: %s", entry.Code),
		map[string]any{"entry_id": entry.ID, "code": entry.Code})

	return map[string]any{"entry_id": entry.ID, "code": entry.Code}, nil
}

func cmdReverseJournalEntry(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	entryID := getStr(input, "entry_id")
	if entryID == "" {
		return nil, fmt.Errorf("entry_id is required")
	}
	var entry ErpJournalEntry
	if err := db.First(&entry, "id = ?", entryID).Error; err != nil {
		return nil, fmt.Errorf("journal entry not found")
	}
	if entry.Status != "posted" {
		return nil, fmt.Errorf("only posted entries can be reversed, current: %s", entry.Status)
	}

	// Get original lines
	var origLines []ErpJournalLine
	db.Where("entry_id = ?", entryID).Find(&origLines)

	// Create reversal entry
	var reversal ErpJournalEntry
	revCode, err := autoCodeWithRetry(db, "erp_journal_entries", "JE", func(c string) error {
		now := time.Now()
		reversal = ErpJournalEntry{
			ID:          uuid.New().String(),
			Code:        c,
			Period:      entry.Period,
			EntryDate:   &now,
			Description: fmt.Sprintf("冲销: %s - %s", entry.Code, entry.Description),
			TotalDebit:  entry.TotalCredit,
			TotalCredit: entry.TotalDebit,
			SourceType:  "reversal",
			SourceID:    entry.ID,
			Status:      "posted",
			PostedBy:    getStr(input, "posted_by"),
			PostedAt:    &now,
		}
		return db.Create(&reversal).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create reversal entry failed: %w", err)
	}

	// Create reversed lines (swap debit/credit)
	for _, ol := range origLines {
		rl := ErpJournalLine{
			ID:          uuid.New().String(),
			EntryID:     reversal.ID,
			AccountID:   ol.AccountID,
			Debit:       ol.Credit,
			Credit:      ol.Debit,
			Currency:    ol.Currency,
			Description: fmt.Sprintf("冲销: %s", ol.Description),
			CustomerID:  ol.CustomerID,
			SupplierID:  ol.SupplierID,
		}
		db.Create(&rl)
	}

	// Mark original as reversed
	db.Model(&entry).Update("status", "reversed")

	emitEvent(adapter, runID, stepID, "erp.journal_entry.reversed",
		fmt.Sprintf("凭证冲销: %s -> %s", entry.Code, revCode),
		map[string]any{"original_id": entry.ID, "reversal_id": reversal.ID, "reversal_code": revCode})

	return map[string]any{"original_id": entry.ID, "reversal_id": reversal.ID, "reversal_code": revCode}, nil
}

func cmdClosePeriod(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	period := getStr(input, "period")
	if period == "" {
		return nil, fmt.Errorf("period is required (format: YYYY-MM)")
	}

	// Check for unposted entries in this period
	var draftCount int64
	db.Model(&ErpJournalEntry{}).Where("period = ? AND status = ?", period, "draft").Count(&draftCount)
	if draftCount > 0 {
		return nil, fmt.Errorf("期间 %s 还有 %d 张未过账凭证", period, draftCount)
	}

	// Calculate period totals
	var totalDebit, totalCredit float64
	db.Model(&ErpJournalEntry{}).Where("period = ? AND status = ?", period, "posted").
		Select("COALESCE(SUM(total_debit), 0), COALESCE(SUM(total_credit), 0)").
		Row().Scan(&totalDebit, &totalCredit)

	emitEvent(adapter, runID, stepID, "erp.period.closed",
		fmt.Sprintf("期间关闭: %s, 借方合计%.2f, 贷方合计%.2f", period, totalDebit, totalCredit),
		map[string]any{"period": period, "total_debit": totalDebit, "total_credit": totalCredit, "draft_count": 0})

	return map[string]any{"period": period, "total_debit": totalDebit, "total_credit": totalCredit}, nil
}

func cmdGenerateReport(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	reportType := getStr(input, "report_type")
	period := getStr(input, "period")

	switch reportType {
	case "trial_balance":
		// Trial balance: sum debit/credit per account for the period
		type balanceRow struct {
			AccountID string  `json:"account_id"`
			Debit     float64 `json:"debit"`
			Credit    float64 `json:"credit"`
		}
		var rows []balanceRow
		query := db.Table("erp_journal_lines jl").
			Joins("JOIN erp_journal_entries je ON je.id = jl.entry_id").
			Where("je.status = ?", "posted").
			Select("jl.account_id, COALESCE(SUM(jl.debit), 0) as debit, COALESCE(SUM(jl.credit), 0) as credit").
			Group("jl.account_id")
		if period != "" {
			query = query.Where("je.period = ?", period)
		}
		query.Find(&rows)

		emitEvent(adapter, runID, stepID, "erp.report.generated",
			fmt.Sprintf("试算平衡表生成: %s, %d个科目", period, len(rows)),
			map[string]any{"report_type": reportType, "period": period, "account_count": len(rows)})

		return map[string]any{"report_type": reportType, "period": period, "rows": rows}, nil

	case "income_statement":
		// Sum revenue and expense accounts
		var revenue, expense float64
		db.Table("erp_journal_lines jl").
			Joins("JOIN erp_journal_entries je ON je.id = jl.entry_id").
			Joins("JOIN erp_accounts a ON a.id = jl.account_id").
			Where("je.status = ? AND je.period = ? AND a.type = ?", "posted", period, "revenue").
			Select("COALESCE(SUM(jl.credit - jl.debit), 0)").Row().Scan(&revenue)
		db.Table("erp_journal_lines jl").
			Joins("JOIN erp_journal_entries je ON je.id = jl.entry_id").
			Joins("JOIN erp_accounts a ON a.id = jl.account_id").
			Where("je.status = ? AND je.period = ? AND a.type = ?", "posted", period, "expense").
			Select("COALESCE(SUM(jl.debit - jl.credit), 0)").Row().Scan(&expense)

		profit := revenue - expense

		emitEvent(adapter, runID, stepID, "erp.report.generated",
			fmt.Sprintf("利润表生成: %s, 收入%.2f, 支出%.2f, 利润%.2f", period, revenue, expense, profit),
			map[string]any{"report_type": reportType, "period": period, "revenue": revenue, "expense": expense, "profit": profit})

		return map[string]any{"report_type": reportType, "period": period, "revenue": revenue, "expense": expense, "profit": profit}, nil

	default:
		return nil, fmt.Errorf("unsupported report type: %s (supported: trial_balance, income_statement)", reportType)
	}
}

// ===========================================================================
// Quality commands (6)
// ===========================================================================

func cmdCreateOQC(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	shipmentID := getStr(input, "shipment_id")
	if shipmentID == "" {
		return nil, fmt.Errorf("shipment_id is required")
	}

	var oqc ErpOQCInspection
	code, err := autoCodeWithRetry(db, "erp_oqc_inspections", "OQC", func(c string) error {
		oqc = ErpOQCInspection{
			ID:          uuid.New().String(),
			Code:        c,
			ShipmentID:  shipmentID,
			ProductID:   getStr(input, "product_id"),
			LotNumber:   getStr(input, "lot_number"),
			SampleSize:  getInt(input, "sample_size", 0),
			InspectorID: getStr(input, "inspector_id"),
			Notes:       getStr(input, "notes"),
			Status:      "pending",
		}
		return db.Create(&oqc).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create OQC inspection failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.oqc.created",
		fmt.Sprintf("OQC检验创建: %s", code),
		map[string]any{"oqc_id": oqc.ID, "code": code, "shipment_id": shipmentID})

	return map[string]any{"oqc_id": oqc.ID, "code": code}, nil
}

func cmdCompleteOQC(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	oqcID := getStr(input, "oqc_id")
	if oqcID == "" {
		return nil, fmt.Errorf("oqc_id is required")
	}
	var oqc ErpOQCInspection
	if err := db.First(&oqc, "id = ?", oqcID).Error; err != nil {
		return nil, fmt.Errorf("OQC inspection not found")
	}
	if oqc.Status != "pending" {
		return nil, fmt.Errorf("OQC not in pending status, current: %s", oqc.Status)
	}

	result := getStr(input, "result")
	if result != "pass" && result != "fail" && result != "conditional" {
		return nil, fmt.Errorf("result must be 'pass', 'fail', or 'conditional'")
	}

	now := time.Now()
	db.Model(&oqc).Updates(map[string]any{
		"total_inspected": getInt(input, "total_inspected", 0),
		"pass_count":      getInt(input, "pass_count", 0),
		"fail_count":      getInt(input, "fail_count", 0),
		"result":          result,
		"defect_details":  getStr(input, "defect_details"),
		"inspected_at":    &now,
		"status":          "completed",
	})

	emitEvent(adapter, runID, stepID, "erp.oqc.completed",
		fmt.Sprintf("OQC检验完成: %s, 结果=%s", oqc.Code, result),
		map[string]any{"oqc_id": oqc.ID, "code": oqc.Code, "result": result})

	retVal := map[string]any{"oqc_id": oqc.ID, "code": oqc.Code, "result": result}

	// Auto-create NCR on fail or conditional
	if result == "fail" || result == "conditional" {
		var ncr ErpNCRReport
		autoCodeWithRetry(db, "erp_ncr_reports", "NCR", func(c string) error {
			ncr = ErpNCRReport{
				ID:          uuid.New().String(),
				Code:        c,
				Source:      "oqc",
				SourceID:    oqcID,
				ProductID:   oqc.ProductID,
				LotNumber:   oqc.LotNumber,
				DefectQty:   float64(oqc.FailCount),
				DefectType:  "OQC检验不合格",
				Description: fmt.Sprintf("OQC检验 %s 结果: %s, 不合格 %d/%d", oqc.Code, result, oqc.FailCount, oqc.TotalInspected),
				Severity:    "major",
				Status:      "open",
			}
			return db.Create(&ncr).Error
		})

		emitEvent(adapter, runID, stepID, "erp.ncr.auto_created",
			fmt.Sprintf("自动创建NCR: %s (来源: OQC %s)", ncr.Code, oqc.Code),
			map[string]any{"ncr_id": ncr.ID, "oqc_id": oqcID})

		retVal["ncr_id"] = ncr.ID
		retVal["ncr_code"] = ncr.Code
	}

	return retVal, nil
}

func cmdCreateNCR(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	description := getStr(input, "description")
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	var ncr ErpNCRReport
	code, err := autoCodeWithRetry(db, "erp_ncr_reports", "NCR", func(c string) error {
		ncr = ErpNCRReport{
			ID:          uuid.New().String(),
			Code:        c,
			Source:      getStr(input, "source"),
			SourceID:    getStr(input, "source_id"),
			ProductID:   getStr(input, "product_id"),
			MaterialID:  getStr(input, "material_id"),
			LotNumber:   getStr(input, "lot_number"),
			DefectQty:   getFloat(input, "defect_qty"),
			DefectType:  getStr(input, "defect_type"),
			Description: description,
			Severity:    getStr(input, "severity"),
			OwnerID:     getStr(input, "owner_id"),
			Status:      "open",
		}
		if ncr.Severity == "" {
			ncr.Severity = "medium"
		}
		return db.Create(&ncr).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create NCR failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.ncr.created",
		fmt.Sprintf("NCR创建: %s, %s", code, ncr.DefectType),
		map[string]any{"ncr_id": ncr.ID, "code": code})

	return map[string]any{"ncr_id": ncr.ID, "code": code}, nil
}

func cmdDispositionNCR(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	ncrID := getStr(input, "ncr_id")
	if ncrID == "" {
		return nil, fmt.Errorf("ncr_id is required")
	}
	var ncr ErpNCRReport
	if err := db.First(&ncr, "id = ?", ncrID).Error; err != nil {
		return nil, fmt.Errorf("NCR not found")
	}
	if ncr.Status != "open" {
		return nil, fmt.Errorf("NCR not in open status, current: %s", ncr.Status)
	}

	disposition := getStr(input, "disposition")
	if disposition == "" {
		return nil, fmt.Errorf("disposition is required (e.g. rework, scrap, use_as_is, return_to_supplier)")
	}

	now := time.Now()
	db.Model(&ncr).Updates(map[string]any{
		"disposition": disposition,
		"status":      "dispositioned",
		"closed_at":   &now,
	})

	emitEvent(adapter, runID, stepID, "erp.ncr.dispositioned",
		fmt.Sprintf("NCR处置: %s, 处置=%s", ncr.Code, disposition),
		map[string]any{"ncr_id": ncr.ID, "code": ncr.Code, "disposition": disposition})

	return map[string]any{"ncr_id": ncr.ID, "code": ncr.Code, "disposition": disposition}, nil
}

func cmdCreateCAPA(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	title := getStr(input, "title")
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	var capa ErpCAPA
	code, err := autoCodeWithRetry(db, "erp_capa", "CAPA", func(c string) error {
		capa = ErpCAPA{
			ID:         uuid.New().String(),
			Code:       c,
			Type:       getStr(input, "type"),
			NcrID:      getStr(input, "ncr_id"),
			Title:      title,
			RootCause:  getStr(input, "root_cause"),
			ActionPlan: getStr(input, "action_plan"),
			OwnerID:    getStr(input, "owner_id"),
			DueDate:    parseDate(getStr(input, "due_date")),
			Status:     "open",
		}
		if capa.Type == "" {
			capa.Type = "corrective"
		}
		return db.Create(&capa).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create CAPA failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.capa.created",
		fmt.Sprintf("CAPA创建: %s, %s", code, title),
		map[string]any{"capa_id": capa.ID, "code": code})

	return map[string]any{"capa_id": capa.ID, "code": code}, nil
}

func cmdCloseCAPA(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	capaID := getStr(input, "capa_id")
	if capaID == "" {
		return nil, fmt.Errorf("capa_id is required")
	}
	var capa ErpCAPA
	if err := db.First(&capa, "id = ?", capaID).Error; err != nil {
		return nil, fmt.Errorf("CAPA not found")
	}
	if capa.Status != "open" {
		return nil, fmt.Errorf("CAPA not in open status, current: %s", capa.Status)
	}

	verification := getStr(input, "verification")
	now := time.Now()
	db.Model(&capa).Updates(map[string]any{
		"verification": verification,
		"root_cause":   getStr(input, "root_cause"),
		"status":       "closed",
		"closed_at":    &now,
	})

	emitEvent(adapter, runID, stepID, "erp.capa.closed",
		fmt.Sprintf("CAPA关闭: %s", capa.Code),
		map[string]any{"capa_id": capa.ID, "code": capa.Code})

	return map[string]any{"capa_id": capa.ID, "code": capa.Code}, nil
}

// ---------------------------------------------------------------------------
// Additional Commands
// ---------------------------------------------------------------------------

// cmdUpdateShipmentStatus — 更新发货单状态
func cmdUpdateShipmentStatus(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	shipmentID := getStr(input, "shipment_id")
	newStatus := getStr(input, "status")
	if shipmentID == "" || newStatus == "" {
		return nil, fmt.Errorf("shipment_id and status are required")
	}
	var shipment ErpShipment
	if err := db.First(&shipment, "id = ?", shipmentID).Error; err != nil {
		return nil, fmt.Errorf("shipment not found")
	}
	updates := map[string]any{"status": newStatus}
	now := time.Now()
	if newStatus == "shipped" {
		updates["shipped_at"] = &now
	}
	if newStatus == "delivered" {
		updates["delivered_at"] = &now
	}
	db.Model(&shipment).Updates(updates)
	emitEvent(adapter, runID, stepID, "erp.shipment.status_changed",
		fmt.Sprintf("发货单状态变更: %s → %s", shipment.Code, newStatus),
		map[string]any{"shipment_id": shipmentID, "status": newStatus})
	return map[string]any{"shipment_id": shipmentID, "status": newStatus}, nil
}

// cmdExtendQuotation — 延长报价单有效期
func cmdExtendQuotation(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	quotationID := getStr(input, "quotation_id")
	days := getInt(input, "days", 7)
	if quotationID == "" {
		return nil, fmt.Errorf("quotation_id is required")
	}
	var q ErpQuotation
	if err := db.First(&q, "id = ?", quotationID).Error; err != nil {
		return nil, fmt.Errorf("quotation not found")
	}
	newDate := time.Now().AddDate(0, 0, days)
	db.Model(&q).Update("valid_until", &newDate)
	return map[string]any{"quotation_id": quotationID, "valid_until": newDate.Format("2006-01-02")}, nil
}

// cmdSendPaymentReminder — 发送催款通知
func cmdSendPaymentReminder(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	invoiceID := getStr(input, "invoice_id")
	if invoiceID == "" {
		return nil, fmt.Errorf("invoice_id is required")
	}
	var inv ErpSalesInvoice
	if err := db.First(&inv, "id = ?", invoiceID).Error; err != nil {
		return nil, fmt.Errorf("invoice not found")
	}
	emitEvent(adapter, runID, stepID, "erp.payment.reminder_sent",
		fmt.Sprintf("催款通知已发送: %s ¥%.0f", inv.Code, inv.Total),
		map[string]any{"invoice_id": invoiceID, "amount": inv.Total, "customer_id": inv.CustomerID})
	return map[string]any{"invoice_id": invoiceID, "reminder_sent": true}, nil
}

// ---------------------------------------------------------------------------
// SRM → ERP AP posting
// ---------------------------------------------------------------------------
//
// cmdPostAPFromSettlement closes the industry-standard split:
//   - SRM owns the supplier-facing "结算单" (collaboration document): amounts,
//     invoice reference, buyer/supplier confirmation, disputes.
//   - ERP owns the book of record: accounts payable journal entries.
//
// This command is called from the flow YAML after the SRM side confirms a
// settlement (e.g. after srm:confirm_settlement_buyer). It:
//   1. Reads the srm_settlements row for the given settlement_id (raw SQL
//      since SRM types live in a separate Go module).
//   2. Lazily bootstraps the minimal chart-of-accounts rows needed (2202
//      应付账款, 1002 银行存款) when the accounts table is empty.
//   3. Creates a posted ErpJournalEntry with source_type="srm_settlement"
//      so finance can trace every AP entry back to its settlement.
//   4. Writes two ErpJournalLine rows: DR 应付账款, CR 银行存款.
//   5. Flips the SRM settlement status to "posted" via raw UPDATE (keeping
//      the SRM row authoritative for the collaboration side).

// srmSettlementLite is a minimal shape mirroring srm_settlements — enough
// for AP posting. We intentionally do not import the SRM package so the
// ERP module keeps a clean dependency boundary.
type srmSettlementLite struct {
	ID             string
	SettlementCode string
	SupplierID     string
	FinalAmount    float64
	InvoiceNo      string
	Status         string
	PeriodStart    *time.Time
	PeriodEnd      *time.Time
}

func (srmSettlementLite) TableName() string { return "srm_settlements" }

func cmdPostAPFromSettlement(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	settlementID := getStr(input, "settlement_id")
	if settlementID == "" {
		return nil, fmt.Errorf("settlement_id is required")
	}

	var stl srmSettlementLite
	if err := db.Table("srm_settlements").
		Select("id, settlement_code, supplier_id, final_amount, invoice_no, status, period_start, period_end").
		Where("id = ?", settlementID).
		Scan(&stl).Error; err != nil || stl.ID == "" {
		return nil, fmt.Errorf("settlement not found: %s", settlementID)
	}
	if stl.FinalAmount <= 0 {
		return nil, fmt.Errorf("settlement final_amount must be positive, got %.2f", stl.FinalAmount)
	}

	// Bootstrap AP + Cash accounts if they don't exist.
	apAccount := ensureAccount(db, "2202", "应付账款", "liability")
	cashAccount := ensureAccount(db, "1002", "银行存款", "asset")

	now := time.Now()
	period := ""
	if stl.PeriodEnd != nil {
		period = stl.PeriodEnd.Format("2006-01")
	} else {
		period = now.Format("2006-01")
	}

	var entry ErpJournalEntry
	entryCode, err := autoCodeWithRetry(db, "erp_journal_entries", "JE", func(c string) error {
		entry = ErpJournalEntry{
			ID:          uuid.New().String(),
			Code:        c,
			Period:      period,
			EntryDate:   &now,
			SourceType:  "srm_settlement",
			SourceID:    stl.ID,
			Description: fmt.Sprintf("供应商结算 %s (发票 %s)", stl.SettlementCode, stl.InvoiceNo),
			TotalDebit:  stl.FinalAmount,
			TotalCredit: stl.FinalAmount,
			Status:      "posted",
			PostedBy:    getStr(input, "posted_by"),
			PostedAt:    &now,
		}
		return db.Create(&entry).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create AP journal entry failed: %w", err)
	}

	// DR 应付账款
	if err := db.Create(&ErpJournalLine{
		ID:             uuid.New().String(),
		EntryID:        entry.ID,
		AccountID:      apAccount.ID,
		Debit:          stl.FinalAmount,
		Credit:         0,
		Currency:       "CNY",
		OriginalAmount: stl.FinalAmount,
		Description:    fmt.Sprintf("%s 应付账款冲销", stl.SettlementCode),
		SupplierID:     stl.SupplierID,
	}).Error; err != nil {
		return nil, fmt.Errorf("create DR line failed: %w", err)
	}
	// CR 银行存款
	if err := db.Create(&ErpJournalLine{
		ID:             uuid.New().String(),
		EntryID:        entry.ID,
		AccountID:      cashAccount.ID,
		Debit:          0,
		Credit:         stl.FinalAmount,
		Currency:       "CNY",
		OriginalAmount: stl.FinalAmount,
		Description:    fmt.Sprintf("%s 银行付款", stl.SettlementCode),
		SupplierID:     stl.SupplierID,
	}).Error; err != nil {
		return nil, fmt.Errorf("create CR line failed: %w", err)
	}

	// Flip SRM settlement status to "posted" so the collaboration row
	// reflects that finance has already booked it.
	db.Table("srm_settlements").Where("id = ?", stl.ID).Updates(map[string]any{
		"status":     "paid",
		"updated_at": now,
	})

	emitEvent(adapter, runID, stepID, "erp.ap.settlement_posted",
		fmt.Sprintf("AP 过账: %s → 凭证 %s, 金额 ¥%.2f", stl.SettlementCode, entryCode, stl.FinalAmount),
		map[string]any{
			"journal_entry_id": entry.ID,
			"journal_code":     entryCode,
			"settlement_id":    stl.ID,
			"supplier_id":      stl.SupplierID,
			"amount":           stl.FinalAmount,
		})

	return map[string]any{
		"journal_entry_id": entry.ID,
		"journal_code":     entryCode,
		"settlement_id":    stl.ID,
		"amount":           stl.FinalAmount,
	}, nil
}

// ensureAccount fetches the account with the given code, creating it with
// the supplied name/type if missing. Used for lazy bootstrapping of the
// minimal chart of accounts needed by AP posting.
func ensureAccount(db *gorm.DB, code, name, accountType string) ErpAccount {
	var acc ErpAccount
	if err := db.Where("code = ?", code).First(&acc).Error; err == nil {
		return acc
	}
	acc = ErpAccount{
		ID:       uuid.New().String(),
		Code:     code,
		Name:     name,
		Type:     accountType,
		Level:    1,
		IsLeaf:   true,
		Currency: "CNY",
		Status:   "active",
	}
	db.Create(&acc)
	return acc
}

// ---------------------------------------------------------------------------
// PRD v1 Sprint 1: Inventory bootstrap + lazy-create helpers
// ---------------------------------------------------------------------------

// defaultWarehouseSpec lists the 6 warehouses every nimo install needs:
// raw materials, WIP, finished goods, spare parts, quality hold (virtual)
// and scrap (virtual). The hold/scrap "warehouses" are conceptual buckets
// used for status transitions; they don't represent physical sites.
type defaultWarehouseSpec struct {
	Code, Name, Type string
}

var defaultWarehouses = []defaultWarehouseSpec{
	{Code: "WH-RAW-01", Name: "原料库", Type: "raw"},
	{Code: "WH-WIP-01", Name: "在制品库", Type: "wip"},
	{Code: "WH-FG-01", Name: "成品库", Type: "finished"},
	{Code: "WH-SPARE-01", Name: "备件库", Type: "spare"},
	{Code: "WH-QH-01", Name: "质量冻结库", Type: "quality_hold"},
	{Code: "WH-SCRAP-01", Name: "报废暂存库", Type: "scrap"},
}

// cmdBootstrapDefaultWarehouses creates the 6 standard warehouses if any
// of them is missing. Idempotent — safe to run multiple times.
//
// Used by Sprint 1 installation hook so the NPI flow's mp_stock_in step
// can resolve {{vars.finished_goods_warehouse}} to a real warehouse_id
// instead of the empty string.
func cmdBootstrapDefaultWarehouses(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	created := 0
	existing := 0
	codeToID := make(map[string]string, len(defaultWarehouses))
	now := time.Now()
	for _, spec := range defaultWarehouses {
		var wh ErpWarehouse
		if err := db.Where("code = ?", spec.Code).First(&wh).Error; err == nil {
			codeToID[spec.Code] = wh.ID
			existing++
			continue
		}
		wh = ErpWarehouse{
			ID:        uuid.New().String(),
			Code:      spec.Code,
			Name:      spec.Name,
			Type:      spec.Type,
			IsDefault: spec.Code == "WH-RAW-01",
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&wh).Error; err != nil {
			return nil, fmt.Errorf("create warehouse %s failed: %w", spec.Code, err)
		}
		codeToID[spec.Code] = wh.ID
		created++
	}

	emitEvent(adapter, runID, stepID, "erp.warehouse.bootstrapped",
		fmt.Sprintf("仓库初始化: 新建 %d, 已存在 %d", created, existing),
		map[string]any{"created": created, "existing": existing, "code_to_id": codeToID})

	return map[string]any{
		"created":      created,
		"existing":     existing,
		"warehouses":   defaultWarehouses,
		"code_to_id":   codeToID,
	}, nil
}

// resolveWarehouseID resolves a warehouse identifier to its UUID.
// Accepts either a UUID-like string (returns as-is) or a code like "WH-FG-01"
// (looked up in erp_warehouses). Used by inventory commands so YAMLs can
// reference warehouses by stable code instead of UUIDs.
func resolveWarehouseID(db *gorm.DB, ref string) string {
	if ref == "" {
		return ""
	}
	// Treat any reference that looks like our default code (starts with WH-)
	// as a code to look up; otherwise assume caller already passed an ID.
	if strings.HasPrefix(ref, "WH-") {
		var wh ErpWarehouse
		if err := db.Where("code = ?", ref).First(&wh).Error; err == nil {
			return wh.ID
		}
	}
	return ref
}

// ensureInventoryAttrs lazy-creates an ErpMaterialInventoryAttrs row when a
// material is touched by inventory operations for the first time. The PRD
// model: material master lives in plm_materials; ERP only owns the inventory
// projection (ABC, safety stock, tracking mode, current avg cost). On first
// touch we create a default-valued row so subsequent updates have somewhere
// to write. Tracking mode defaults to "lot" since most nimo electronics are
// lot-controlled by regulation.
func ensureInventoryAttrs(db *gorm.DB, materialID, defaultUnit, defaultWarehouseID string) ErpMaterialInventoryAttrs {
	var attrs ErpMaterialInventoryAttrs
	if err := db.Where("material_id = ?", materialID).First(&attrs).Error; err == nil {
		return attrs
	}
	if defaultUnit == "" {
		defaultUnit = "pcs"
	}
	attrs = ErpMaterialInventoryAttrs{
		MaterialID:       materialID,
		DefaultUnit:      defaultUnit,
		ABCClass:         "C",
		TrackingMode:     "lot",
		DefaultWarehouse: defaultWarehouseID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	db.Create(&attrs)
	return attrs
}

// ---------------------------------------------------------------------------
// PRD v1 Sprint 2: 移动加权平均成本 + 默认科目表 + 自动凭证生成
// ---------------------------------------------------------------------------

// updateMovingAverageCost recalculates the moving weighted average cost for
// this material across the warehouse.
//
// Called AFTER the receipt row has already been written to erp_inventory
// (the current quantity already includes the new batch). Formula reduces to:
//   avg = SUM(quantity × unit_cost) / SUM(quantity)
// across all available-status rows for the material in this warehouse.
//
// Stores the result on ErpMaterialInventoryAttrs.CurrentAvgCost.
func updateMovingAverageCost(tx *gorm.DB, materialID, warehouseID string, newQty, newCost float64) (float64, error) {
	if newQty <= 0 {
		return 0, nil
	}
	type aggResult struct {
		TotalQty   float64
		TotalValue float64
	}
	var agg aggResult
	if err := tx.Table("erp_inventory").
		Select("COALESCE(SUM(quantity), 0) AS total_qty, COALESCE(SUM(quantity * unit_cost), 0) AS total_value").
		Where("material_id = ? AND warehouse_id = ? AND status = ?", materialID, warehouseID, "available").
		Scan(&agg).Error; err != nil {
		return 0, err
	}
	avg := newCost
	if agg.TotalQty > 0 {
		avg = agg.TotalValue / agg.TotalQty
	}
	if err := tx.Model(&ErpMaterialInventoryAttrs{}).
		Where("material_id = ?", materialID).
		Updates(map[string]any{
			"current_avg_cost": avg,
			"updated_at":       time.Now(),
		}).Error; err != nil {
		return 0, err
	}
	return avg, nil
}

// chartOfAccountsSpec is the minimal Chinese GAAP chart of accounts the ERP
// inventory module needs for posting. Sprint 2 bootstraps these on first use.
type accountSpec struct {
	Code, Name, Type string
}

var defaultChartOfAccounts = []accountSpec{
	{Code: "1002", Name: "银行存款", Type: "asset"},
	{Code: "1403", Name: "原材料", Type: "asset"},
	{Code: "1405", Name: "库存商品", Type: "asset"},
	{Code: "1411", Name: "在制品", Type: "asset"},
	{Code: "2202", Name: "应付账款", Type: "liability"},
	{Code: "5301", Name: "营业外收入", Type: "income"},
	{Code: "5302", Name: "营业外支出", Type: "expense"},
	{Code: "6401", Name: "主营业务成本", Type: "expense"},
	{Code: "6603", Name: "财务费用", Type: "expense"},
}

// cmdBootstrapChartOfAccounts creates the default 9 accounts if missing.
// Idempotent.
func cmdBootstrapChartOfAccounts(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	created := 0
	existing := 0
	for _, spec := range defaultChartOfAccounts {
		var acc ErpAccount
		if err := db.Where("code = ?", spec.Code).First(&acc).Error; err == nil {
			existing++
			continue
		}
		ensureAccount(db, spec.Code, spec.Name, spec.Type)
		created++
	}
	emitEvent(adapter, runID, stepID, "erp.account.bootstrapped",
		fmt.Sprintf("科目表初始化: 新建 %d, 已存在 %d", created, existing),
		map[string]any{"created": created, "existing": existing})
	return map[string]any{"created": created, "existing": existing}, nil
}

// inventoryAccountForMaterial returns the asset account code that should
// hold this material's value. Default rule:
//   - finished products (warehouse type "finished") → 1405 库存商品
//   - WIP                                            → 1411 在制品
//   - everything else                                → 1403 原材料
// PRD v2 may override per material category later.
func inventoryAccountForMaterial(db *gorm.DB, warehouseID string) string {
	var wh ErpWarehouse
	if err := db.First(&wh, "id = ?", warehouseID).Error; err != nil {
		return "1403"
	}
	switch wh.Type {
	case "finished":
		return "1405"
	case "wip":
		return "1411"
	default:
		return "1403"
	}
}

// postInventoryJournal generates a balanced ErpJournalEntry + 2 lines for an
// inventory transaction. Returns the new entry ID. Should be called from
// within the same DB transaction as the inventory mutation so that "失败回滚"
// gives both directions atomicity.
//
// Mapping:
//   po_receipt        DR inventoryAcct  / CR 2202 应付账款
//   production_in     DR 1405 库存商品  / CR 1411 在制品
//   production_issue  DR 1411 在制品    / CR 1403 原材料
//   sales_issue       DR 6401 主营业务成本 / CR 1405 库存商品
//   adjustment_in     DR inventoryAcct  / CR 5301 营业外收入
//   adjustment_out    DR 5302 营业外支出 / CR inventoryAcct
//   scrap             DR 5302 营业外支出 / CR inventoryAcct
//   transfer_out      no posting (value-neutral)
func postInventoryJournal(tx *gorm.DB, txn *ErpInventoryTransaction, supplierID string) (string, error) {
	if txn.TotalAmount <= 0 {
		return "", nil
	}
	var drCode, crCode string
	switch txn.Type {
	case "po_receipt":
		drCode = inventoryAccountForMaterial(tx, txn.ToWarehouseID)
		crCode = "2202"
	case "production_in":
		drCode = "1405"
		crCode = "1411"
	case "production_issue":
		drCode = "1411"
		crCode = "1403"
	case "sales_issue":
		drCode = "6401"
		crCode = "1405"
	case "adjustment_in":
		drCode = inventoryAccountForMaterial(tx, txn.ToWarehouseID)
		crCode = "5301"
	case "adjustment_out", "scrap":
		drCode = "5302"
		crCode = inventoryAccountForMaterial(tx, txn.FromWarehouseID)
	default:
		return "", nil
	}

	dr := ensureAccount(tx, drCode, accountNameByCode(drCode), accountTypeByCode(drCode))
	cr := ensureAccount(tx, crCode, accountNameByCode(crCode), accountTypeByCode(crCode))

	now := time.Now()
	period := now.Format("2006-01")
	var entry ErpJournalEntry
	code, err := autoCodeWithRetry(tx, "erp_journal_entries", "JE", func(c string) error {
		entry = ErpJournalEntry{
			ID:          uuid.New().String(),
			Code:        c,
			Period:      period,
			EntryDate:   &now,
			SourceType:  "inventory_txn",
			SourceID:    txn.ID,
			Description: fmt.Sprintf("库存事务 %s (%s)", txn.Code, txn.Type),
			TotalDebit:  txn.TotalAmount,
			TotalCredit: txn.TotalAmount,
			Status:      "posted",
			PostedAt:    &now,
			PostedBy:    txn.OperatorID,
		}
		return tx.Create(&entry).Error
	})
	if err != nil {
		return "", fmt.Errorf("create journal entry failed: %w", err)
	}
	_ = code

	lines := []ErpJournalLine{
		{
			ID:             uuid.New().String(),
			EntryID:        entry.ID,
			AccountID:      dr.ID,
			Debit:          txn.TotalAmount,
			Credit:         0,
			Currency:       "CNY",
			OriginalAmount: txn.TotalAmount,
			Description:    fmt.Sprintf("%s DR", txn.Code),
			SupplierID:     supplierID,
		},
		{
			ID:             uuid.New().String(),
			EntryID:        entry.ID,
			AccountID:      cr.ID,
			Debit:          0,
			Credit:         txn.TotalAmount,
			Currency:       "CNY",
			OriginalAmount: txn.TotalAmount,
			Description:    fmt.Sprintf("%s CR", txn.Code),
			SupplierID:     supplierID,
		},
	}
	for i := range lines {
		if err := tx.Create(&lines[i]).Error; err != nil {
			return "", fmt.Errorf("create journal line failed: %w", err)
		}
	}
	return entry.ID, nil
}

// accountNameByCode / accountTypeByCode are helpers used by postInventoryJournal
// when ensureAccount is called for an account that may not yet exist (the
// chart of accounts is bootstrapped lazily on first reference).
func accountNameByCode(code string) string {
	for _, a := range defaultChartOfAccounts {
		if a.Code == code {
			return a.Name
		}
	}
	return code
}

func accountTypeByCode(code string) string {
	for _, a := range defaultChartOfAccounts {
		if a.Code == code {
			return a.Type
		}
	}
	return "asset"
}

// isPeriodLocked checks whether the current month period is closed via
// erp_period_locks. If a row exists for the given period code with
// status="closed", inventory mutations should be rejected.
//
// Sprint 2 implementation: simple lookup, no historical lock support yet.
func isPeriodLocked(tx *gorm.DB, period string) bool {
	type periodLock struct {
		Period string
		Status string
	}
	var pl periodLock
	err := tx.Table("erp_period_locks").Where("period = ?", period).Scan(&pl).Error
	if err != nil || pl.Period == "" {
		return false
	}
	return pl.Status == "closed"
}

// ===========================================================================
// PRD v1 Sprint 3: 盘点 + 不合格品状态机
// ===========================================================================

// cmdCreateInventoryCount creates a count sheet and freezes a snapshot of
// current inventory rows matching the filter. All matched rows become count
// lines with snapshot_qty = current quantity, counted_qty = -1 (not counted).
//
// type=full:  all materials in the warehouse
// type=cycle: filter by abc_class (from erp_material_inventory_attrs)
// type=spot:  only material_ids list
func cmdCreateInventoryCount(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	countType := getStr(input, "type")
	if countType == "" {
		countType = "full"
	}
	warehouseRef := getStr(input, "warehouse_id")
	warehouseID := resolveWarehouseID(db, warehouseRef)

	name := getStr(input, "name")
	if name == "" {
		name = fmt.Sprintf("盘点-%s-%s", countType, time.Now().Format("20060102"))
	}

	var scheduledPtr *time.Time
	if s := getStr(input, "scheduled_at"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			scheduledPtr = &t
		}
	}

	var count ErpInventoryCount
	var totalLines int

	err := db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		_, cerr := autoCodeWithRetry(tx, "erp_inventory_counts", "IC", func(c string) error {
			count = ErpInventoryCount{
				ID:          uuid.New().String(),
				Code:        c,
				Name:        name,
				Type:        countType,
				WarehouseID: warehouseID,
				ABCFilter:   getStr(input, "abc_filter"),
				ScheduledAt: scheduledPtr,
				SnapshotAt:  &now,
				Status:      "snapshot_taken",
				Notes:       getStr(input, "notes"),
				CreatedBy:   getStr(input, "operator_id"),
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			return tx.Create(&count).Error
		})
		if cerr != nil {
			return cerr
		}

		// Build filter query for inventory rows to snapshot
		query := tx.Model(&ErpInventory{}).Where("quantity > 0")
		if warehouseID != "" {
			query = query.Where("warehouse_id = ?", warehouseID)
		}

		// type-specific filters
		switch countType {
		case "cycle":
			abcFilter := getStr(input, "abc_filter")
			if abcFilter != "" && abcFilter != "ABC" {
				classes := strings.Split(abcFilter, "")
				query = query.Where("material_id IN (?)",
					tx.Table("erp_material_inventory_attrs").Select("material_id").Where("abc_class IN ?", classes))
			}
		case "spot":
			if ids, ok := input["material_ids"].([]any); ok && len(ids) > 0 {
				idList := make([]string, 0, len(ids))
				for _, v := range ids {
					if s, ok := v.(string); ok && s != "" {
						idList = append(idList, s)
					}
				}
				if len(idList) > 0 {
					query = query.Where("material_id IN ?", idList)
				}
			}
		}

		var invs []ErpInventory
		if err := query.Find(&invs).Error; err != nil {
			return err
		}

		// Create one count line per inventory row
		for _, inv := range invs {
			line := ErpInventoryCountLine{
				ID:          uuid.New().String(),
				CountID:     count.ID,
				MaterialID:  inv.MaterialID,
				WarehouseID: inv.WarehouseID,
				LocationID:  inv.LocationID,
				LotNumber:   inv.LotNumber,
				Status:      inv.Status,
				SnapshotQty: inv.Quantity,
				CountedQty:  -1, // not counted yet
				UnitCost:    inv.UnitCost,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := tx.Create(&line).Error; err != nil {
				return err
			}
			totalLines++
		}

		count.TotalLines = totalLines
		return tx.Model(&count).Update("total_lines", totalLines).Error
	})
	if err != nil {
		return nil, fmt.Errorf("create inventory count failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.count.created",
		fmt.Sprintf("盘点单 %s 创建: 类型 %s, %d 行", count.Code, countType, totalLines),
		map[string]any{"count_id": count.ID, "code": count.Code, "total_lines": totalLines})

	return map[string]any{
		"count_id":    count.ID,
		"code":        count.Code,
		"status":      "snapshot_taken",
		"total_lines": totalLines,
	}, nil
}

// cmdSubmitCountResult records counted quantities for selected lines and
// updates variance calculations. Status moves to "counting" then "review"
// once all lines have a counted value.
func cmdSubmitCountResult(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	countID := getStr(input, "count_id")
	if countID == "" {
		return nil, fmt.Errorf("count_id is required")
	}

	var count ErpInventoryCount
	if err := db.First(&count, "id = ?", countID).Error; err != nil {
		return nil, fmt.Errorf("count not found: %w", err)
	}
	if count.Status == "posted" || count.Status == "canceled" {
		return nil, fmt.Errorf("count %s is already %s", count.Code, count.Status)
	}

	linesInput, _ := input["lines"].([]any)
	if len(linesInput) == 0 {
		return nil, fmt.Errorf("lines is required")
	}
	operatorID := getStr(input, "operator_id")

	var totalVariance float64
	var countedCount int

	err := db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, raw := range linesInput {
			m, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			lineID, _ := m["line_id"].(string)
			if lineID == "" {
				continue
			}
			countedQty := getFloat(m, "counted_qty")
			notes, _ := m["notes"].(string)

			var line ErpInventoryCountLine
			if err := tx.First(&line, "id = ? AND count_id = ?", lineID, countID).Error; err != nil {
				return fmt.Errorf("line %s not found in count %s", lineID, countID)
			}
			variance := countedQty - line.SnapshotQty
			varianceAmt := variance * line.UnitCost
			updates := map[string]any{
				"counted_qty":  countedQty,
				"variance":     variance,
				"variance_amt": varianceAmt,
				"counted_by":   operatorID,
				"counted_at":   now,
				"notes":        notes,
				"updated_at":   now,
			}
			if err := tx.Model(&line).Updates(updates).Error; err != nil {
				return err
			}
		}

		// Recount completed lines
		tx.Model(&ErpInventoryCountLine{}).
			Where("count_id = ? AND counted_qty >= 0", countID).
			Count(&[]int64{0}[0])

		var cc int64
		tx.Model(&ErpInventoryCountLine{}).
			Where("count_id = ? AND counted_qty >= 0", countID).
			Count(&cc)
		countedCount = int(cc)

		// Sum variance amount
		type varAgg struct {
			TotalVariance float64
		}
		var va varAgg
		tx.Table("erp_inventory_count_lines").
			Select("COALESCE(SUM(variance_amt), 0) AS total_variance").
			Where("count_id = ? AND counted_qty >= 0", countID).
			Scan(&va)
		totalVariance = va.TotalVariance

		newStatus := "counting"
		if countedCount >= count.TotalLines && count.TotalLines > 0 {
			newStatus = "review"
		}
		return tx.Model(&count).Updates(map[string]any{
			"status":        newStatus,
			"counted_lines": countedCount,
			"variance":      totalVariance,
			"updated_at":    now,
		}).Error
	})
	if err != nil {
		return nil, fmt.Errorf("submit count result failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.count.counted",
		fmt.Sprintf("盘点 %s 进度: %d/%d, 差异金额 %.2f", count.Code, countedCount, count.TotalLines, totalVariance),
		map[string]any{"count_id": countID, "counted_lines": countedCount, "total_lines": count.TotalLines, "variance": totalVariance})

	return map[string]any{
		"count_id":      countID,
		"counted_lines": countedCount,
		"total_lines":   count.TotalLines,
		"variance":      totalVariance,
		"status":        map[bool]string{true: "review", false: "counting"}[countedCount >= count.TotalLines && count.TotalLines > 0],
	}, nil
}

// cmdPostInventoryCount posts the count sheet: creates adjustment_in/out
// transactions for every line with variance != 0, updates erp_inventory to
// match counted_qty, and generates one summary journal entry for the net
// variance.
func cmdPostInventoryCount(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	countID := getStr(input, "count_id")
	if countID == "" {
		return nil, fmt.Errorf("count_id is required")
	}

	var count ErpInventoryCount
	if err := db.First(&count, "id = ?", countID).Error; err != nil {
		return nil, fmt.Errorf("count not found: %w", err)
	}
	if count.Status == "posted" {
		return nil, fmt.Errorf("count already posted")
	}
	operatorID := getStr(input, "operator_id")
	txnCount := 0
	var journalEntryID string
	var totalVariance float64

	err := db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		period := now.Format("2006-01")
		if isPeriodLocked(tx, period) {
			return fmt.Errorf("period %s is locked", period)
		}

		var lines []ErpInventoryCountLine
		if err := tx.Where("count_id = ? AND counted_qty >= 0", countID).Find(&lines).Error; err != nil {
			return err
		}

		for _, line := range lines {
			if line.Variance == 0 {
				continue
			}
			absDiff := line.Variance
			if absDiff < 0 {
				absDiff = -absDiff
			}
			txnType := "adjustment_in"
			if line.Variance < 0 {
				txnType = "adjustment_out"
			}

			// Update inventory row (create if missing)
			var inv ErpInventory
			q := tx.Where(
				"material_id = ? AND warehouse_id = ? AND COALESCE(location_id,'') = ? AND COALESCE(lot_number,'') = ? AND status = ?",
				line.MaterialID, line.WarehouseID, line.LocationID, line.LotNumber, line.Status,
			)
			if q.First(&inv).Error == nil {
				if err := tx.Model(&inv).Update("quantity", line.CountedQty).Error; err != nil {
					return err
				}
			} else if line.CountedQty > 0 {
				inv = ErpInventory{
					ID:          uuid.New().String(),
					MaterialID:  line.MaterialID,
					WarehouseID: line.WarehouseID,
					LocationID:  line.LocationID,
					LotNumber:   line.LotNumber,
					Status:      line.Status,
					Quantity:    line.CountedQty,
					UnitCost:    line.UnitCost,
					ReceivedAt:  &now,
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				if err := tx.Create(&inv).Error; err != nil {
					return err
				}
			}

			var txn ErpInventoryTransaction
			_, cerr := autoCodeWithRetry(tx, "erp_inventory_transactions", "IT", func(c string) error {
				txn = ErpInventoryTransaction{
					ID:              uuid.New().String(),
					Code:            c,
					Type:            txnType,
					MaterialID:      line.MaterialID,
					ToWarehouseID:   line.WarehouseID,
					ToLocationID:    line.LocationID,
					ToStatus:        line.Status,
					FromWarehouseID: line.WarehouseID,
					FromStatus:      line.Status,
					Quantity:        absDiff,
					UnitCost:        line.UnitCost,
					TotalAmount:     absDiff * line.UnitCost,
					LotNumber:       line.LotNumber,
					ReferenceType:   "count",
					ReferenceID:     countID,
					Reason:          fmt.Sprintf("盘点 %s", count.Code),
					OperatorID:      operatorID,
					CreatedAt:       now,
					UpdatedAt:       now,
				}
				return tx.Create(&txn).Error
			})
			if cerr != nil {
				return cerr
			}

			// Post individual journal per variance txn
			if je, jerr := postInventoryJournal(tx, &txn, ""); jerr == nil && je != "" {
				tx.Model(&ErpInventoryTransaction{}).Where("id = ?", txn.ID).Update("journal_entry_id", je)
				if journalEntryID == "" {
					journalEntryID = je
				}
			}
			totalVariance += line.Variance * line.UnitCost
			txnCount++
		}

		return tx.Model(&count).Updates(map[string]any{
			"status":    "posted",
			"posted_at": now,
			"posted_by": operatorID,
			"variance":  totalVariance,
			"updated_at": now,
		}).Error
	})
	if err != nil {
		return nil, fmt.Errorf("post count failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.count.posted",
		fmt.Sprintf("盘点 %s 过账: %d 笔调整, 净差异 %.2f", count.Code, txnCount, totalVariance),
		map[string]any{"count_id": countID, "txn_count": txnCount, "total_variance": totalVariance, "journal_entry_id": journalEntryID})

	return map[string]any{
		"count_id":         countID,
		"txn_count":        txnCount,
		"journal_entry_id": journalEntryID,
		"total_variance":   totalVariance,
		"status":           "posted",
	}, nil
}

// cmdQualityHold freezes quantity from status=available to status=quality_hold
// for a given material/warehouse/lot. Generates a status-transfer transaction.
func cmdQualityHold(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseID := resolveWarehouseID(db, getStr(input, "warehouse_id"))
	lot := getStr(input, "lot_number")
	qty := getFloat(input, "quantity")
	reason := getStr(input, "reason")
	if materialID == "" || warehouseID == "" {
		return nil, fmt.Errorf("material_id and warehouse_id are required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	var txn ErpInventoryTransaction
	err := db.Transaction(func(tx *gorm.DB) error {
		var src ErpInventory
		q := tx.Where("material_id = ? AND warehouse_id = ? AND status = ?", materialID, warehouseID, "available")
		if lot != "" {
			q = q.Where("COALESCE(lot_number,'') = ?", lot)
		}
		if err := q.First(&src).Error; err != nil {
			return fmt.Errorf("no available inventory for material/warehouse/lot")
		}
		if src.Quantity-src.ReservedQty < qty {
			return fmt.Errorf("insufficient available qty: have %.2f, need %.2f", src.Quantity-src.ReservedQty, qty)
		}

		now := time.Now()
		// Decrement source
		if err := tx.Model(&src).Update("quantity", gorm.Expr("quantity - ?", qty)).Error; err != nil {
			return err
		}

		// Upsert destination (quality_hold row)
		var dst ErpInventory
		dq := tx.Where(
			"material_id = ? AND warehouse_id = ? AND COALESCE(location_id,'') = ? AND COALESCE(lot_number,'') = ? AND status = ?",
			materialID, warehouseID, src.LocationID, src.LotNumber, "quality_hold",
		)
		if dq.First(&dst).Error == nil {
			tx.Model(&dst).Updates(map[string]any{
				"quantity":   gorm.Expr("quantity + ?", qty),
				"updated_at": now,
			})
		} else {
			dst = ErpInventory{
				ID:          uuid.New().String(),
				MaterialID:  materialID,
				WarehouseID: warehouseID,
				LocationID:  src.LocationID,
				LotNumber:   src.LotNumber,
				Status:      "quality_hold",
				Quantity:    qty,
				UnitCost:    src.UnitCost,
				ReceivedAt:  src.ReceivedAt,
				SourceType:  "quality_hold",
				SourceID:    getStr(input, "ncr_id"),
				Notes:       reason,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := tx.Create(&dst).Error; err != nil {
				return err
			}
		}

		_, cerr := autoCodeWithRetry(tx, "erp_inventory_transactions", "IT", func(c string) error {
			txn = ErpInventoryTransaction{
				ID:              uuid.New().String(),
				Code:            c,
				Type:            "status_change",
				MaterialID:      materialID,
				FromWarehouseID: warehouseID,
				FromStatus:      "available",
				ToWarehouseID:   warehouseID,
				ToStatus:        "quality_hold",
				Quantity:        qty,
				UnitCost:        src.UnitCost,
				TotalAmount:     qty * src.UnitCost,
				LotNumber:       src.LotNumber,
				ReferenceType:   "ncr",
				ReferenceID:     getStr(input, "ncr_id"),
				Reason:          reason,
				OperatorID:      getStr(input, "operator_id"),
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			return tx.Create(&txn).Error
		})
		return cerr
	})
	if err != nil {
		return nil, fmt.Errorf("quality hold failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.quality_hold",
		fmt.Sprintf("质量冻结: 物料 %s, 数量 %.2f", materialID, qty),
		map[string]any{"transaction_id": txn.ID, "material_id": materialID, "quantity": qty, "reason": reason})

	return map[string]any{
		"transaction_id": txn.ID,
		"code":           txn.Code,
		"from_status":    "available",
		"to_status":      "quality_hold",
		"quantity":       qty,
	}, nil
}

// cmdQualityRelease moves quality_hold qty back to available, or scraps it.
func cmdQualityRelease(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseID := resolveWarehouseID(db, getStr(input, "warehouse_id"))
	lot := getStr(input, "lot_number")
	qty := getFloat(input, "quantity")
	releaseTo := getStr(input, "release_to")
	if releaseTo == "" {
		releaseTo = "available"
	}
	reason := getStr(input, "reason")
	if materialID == "" || warehouseID == "" {
		return nil, fmt.Errorf("material_id and warehouse_id are required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	var txn ErpInventoryTransaction
	err := db.Transaction(func(tx *gorm.DB) error {
		var src ErpInventory
		q := tx.Where("material_id = ? AND warehouse_id = ? AND status = ?", materialID, warehouseID, "quality_hold")
		if lot != "" {
			q = q.Where("COALESCE(lot_number,'') = ?", lot)
		}
		if err := q.First(&src).Error; err != nil {
			return fmt.Errorf("no quality_hold inventory found")
		}
		if src.Quantity < qty {
			return fmt.Errorf("insufficient quality_hold qty: have %.2f, need %.2f", src.Quantity, qty)
		}
		now := time.Now()
		// Decrement source
		if err := tx.Model(&src).Update("quantity", gorm.Expr("quantity - ?", qty)).Error; err != nil {
			return err
		}

		txnType := "status_change"
		if releaseTo == "scrap" {
			txnType = "scrap"
			// Don't create destination row for scrap
		} else {
			// Upsert destination available row
			var dst ErpInventory
			dq := tx.Where(
				"material_id = ? AND warehouse_id = ? AND COALESCE(location_id,'') = ? AND COALESCE(lot_number,'') = ? AND status = ?",
				materialID, warehouseID, src.LocationID, src.LotNumber, "available",
			)
			if dq.First(&dst).Error == nil {
				tx.Model(&dst).Updates(map[string]any{
					"quantity":   gorm.Expr("quantity + ?", qty),
					"updated_at": now,
				})
			} else {
				dst = ErpInventory{
					ID:          uuid.New().String(),
					MaterialID:  materialID,
					WarehouseID: warehouseID,
					LocationID:  src.LocationID,
					LotNumber:   src.LotNumber,
					Status:      "available",
					Quantity:    qty,
					UnitCost:    src.UnitCost,
					ReceivedAt:  src.ReceivedAt,
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				if err := tx.Create(&dst).Error; err != nil {
					return err
				}
			}
		}

		_, cerr := autoCodeWithRetry(tx, "erp_inventory_transactions", "IT", func(c string) error {
			txn = ErpInventoryTransaction{
				ID:              uuid.New().String(),
				Code:            c,
				Type:            txnType,
				MaterialID:      materialID,
				FromWarehouseID: warehouseID,
				FromStatus:      "quality_hold",
				ToWarehouseID:   warehouseID,
				ToStatus:        releaseTo,
				Quantity:        qty,
				UnitCost:        src.UnitCost,
				TotalAmount:     qty * src.UnitCost,
				LotNumber:       src.LotNumber,
				ReferenceType:   "ncr",
				Reason:          reason,
				OperatorID:      getStr(input, "operator_id"),
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			return tx.Create(&txn).Error
		})
		if cerr != nil {
			return cerr
		}

		// If scrap, also post scrap journal
		if releaseTo == "scrap" {
			if je, jerr := postInventoryJournal(tx, &txn, ""); jerr == nil && je != "" {
				tx.Model(&ErpInventoryTransaction{}).Where("id = ?", txn.ID).Update("journal_entry_id", je)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("quality release failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.quality_released",
		fmt.Sprintf("质量放行: %s → %s, 数量 %.2f", "quality_hold", releaseTo, qty),
		map[string]any{"transaction_id": txn.ID, "material_id": materialID, "quantity": qty, "release_to": releaseTo})

	return map[string]any{
		"transaction_id": txn.ID,
		"code":           txn.Code,
		"from_status":    "quality_hold",
		"to_status":      releaseTo,
		"quantity":       qty,
	}, nil
}

// ===========================================================================
// PRD v1 Sprint 4: 序列号管理
// ===========================================================================

// cmdGenerateSerialNumbers batch-generates serial numbers for a work order or
// receipt batch. SN format: PREFIX-YYYYMMDD-NNNNNN
func cmdGenerateSerialNumbers(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	qty := getInt(input, "quantity", 0)
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if qty > 5000 {
		return nil, fmt.Errorf("single batch limit is 5000")
	}
	prefix := getStr(input, "prefix")
	if prefix == "" {
		prefix = "SN"
	}
	warehouseID := resolveWarehouseID(db, getStr(input, "warehouse_id"))
	workOrderID := getStr(input, "work_order_id")
	productID := getStr(input, "product_id")
	materialID := getStr(input, "material_id")
	lot := getStr(input, "lot_number")

	// Get current max suffix for today
	today := time.Now().Format("20060102")
	baseStr := fmt.Sprintf("%s-%s-", prefix, today)
	var maxCount int64
	db.Model(&ErpSerialNumber{}).Where("serial_number LIKE ?", baseStr+"%").Count(&maxCount)
	startSeq := int(maxCount) + 1

	now := time.Now()
	serials := make([]string, 0, qty)
	records := make([]ErpSerialNumber, 0, qty)
	for i := 0; i < qty; i++ {
		sn := fmt.Sprintf("%s%06d", baseStr, startSeq+i)
		serials = append(serials, sn)
		records = append(records, ErpSerialNumber{
			ID:             uuid.New().String(),
			SerialNumber:   sn,
			MaterialID:     materialID,
			ProductID:      productID,
			Status:         "in_stock",
			WarehouseID:    warehouseID,
			LotNumber:      lot,
			WorkOrderID:    workOrderID,
			ManufacturedAt: &now,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	// Batch insert
	if err := db.CreateInBatches(&records, 200).Error; err != nil {
		return nil, fmt.Errorf("create serial numbers failed: %w", err)
	}

	emitEvent(adapter, runID, stepID, "erp.serial.generated",
		fmt.Sprintf("序列号生成: %d 个 (%s ~ %s)", qty, serials[0], serials[len(serials)-1]),
		map[string]any{"count": qty, "first_serial": serials[0], "last_serial": serials[len(serials)-1], "work_order_id": workOrderID})

	// Only return first/last for large batches to keep output small
	returnList := serials
	if len(returnList) > 100 {
		returnList = append([]string{}, serials[:10]...)
		returnList = append(returnList, "...")
		returnList = append(returnList, serials[len(serials)-10:]...)
	}

	return map[string]any{
		"count":          qty,
		"serial_numbers": returnList,
		"first_serial":   serials[0],
		"last_serial":    serials[len(serials)-1],
	}, nil
}

// cmdUpdateSerialStatus updates status for a list of serial numbers.
// Used when shipping, delivering, returning, or scrapping.
func cmdUpdateSerialStatus(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	status := getStr(input, "status")
	if status == "" {
		return nil, fmt.Errorf("status is required")
	}
	snList, _ := input["serial_numbers"].([]any)
	if len(snList) == 0 {
		return nil, fmt.Errorf("serial_numbers is required")
	}
	serials := make([]string, 0, len(snList))
	for _, v := range snList {
		if s, ok := v.(string); ok && s != "" {
			serials = append(serials, s)
		}
	}
	if len(serials) == 0 {
		return nil, fmt.Errorf("no valid serial numbers")
	}

	now := time.Now()
	updates := map[string]any{
		"status":     status,
		"updated_at": now,
	}
	if shipID := getStr(input, "shipment_id"); shipID != "" {
		updates["shipment_id"] = shipID
	}
	if returnID := getStr(input, "return_id"); returnID != "" {
		updates["return_id"] = returnID
	}
	if custID := getStr(input, "customer_id"); custID != "" {
		updates["customer_id"] = custID
		updates["sold_to"] = custID
		updates["sold_at"] = now
	}
	if soldTo := getStr(input, "sold_to"); soldTo != "" {
		updates["sold_to"] = soldTo
	}
	if reason := getStr(input, "reason"); reason != "" {
		updates["notes"] = reason
	}

	result := db.Model(&ErpSerialNumber{}).Where("serial_number IN ?", serials).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("update serial status failed: %w", result.Error)
	}

	emitEvent(adapter, runID, stepID, "erp.serial.updated",
		fmt.Sprintf("序列号状态更新: %d 个 → %s", int(result.RowsAffected), status),
		map[string]any{"updated": int(result.RowsAffected), "status": status})

	return map[string]any{
		"updated": int(result.RowsAffected),
		"status":  status,
	}, nil
}
