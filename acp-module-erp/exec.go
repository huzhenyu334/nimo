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
	// Inventory (7)
	"receive_inventory":   cmdReceiveInventory,
	"issue_inventory":     cmdIssueInventory,
	"transfer_inventory":  cmdTransferInventory,
	"adjust_inventory":    cmdAdjustInventory,
	"scrap_inventory":     cmdScrapInventory,
	"reserve_inventory":   cmdReserveInventory,
	"unreserve_inventory": cmdUnreserveInventory,
	// Production (7)
	"run_mrp":                cmdRunMRP,
	"confirm_mrp_suggestion": cmdConfirmMRPSuggestion,
	"create_work_order":      cmdCreateWorkOrder,
	"release_work_order":     cmdReleaseWorkOrder,
	"issue_wo_materials":     cmdIssueWOMaterials,
	"report_wo_progress":     cmdReportWOProgress,
	"complete_work_order":    cmdCompleteWorkOrder,
	// Finance (5)
	"create_journal_entry":  cmdCreateJournalEntry,
	"post_journal_entry":    cmdPostJournalEntry,
	"reverse_journal_entry": cmdReverseJournalEntry,
	"close_period":          cmdClosePeriod,
	"generate_report":       cmdGenerateReport,
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

func cmdReceiveInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseID := getStr(input, "warehouse_id")
	qty := getFloat(input, "quantity")
	if materialID == "" {
		return nil, fmt.Errorf("material_id is required")
	}
	if warehouseID == "" {
		return nil, fmt.Errorf("warehouse_id is required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Create transaction
	var txn ErpInventoryTransaction
	code, err := autoCodeWithRetry(db, "erp_inventory_transactions", "IT", func(c string) error {
		txn = ErpInventoryTransaction{
			ID:            uuid.New().String(),
			Code:          c,
			Type:          "receive",
			MaterialID:    materialID,
			ToWarehouseID: warehouseID,
			ToLocationID:  getStr(input, "location_id"),
			Quantity:      qty,
			UnitCost:      getFloat(input, "unit_cost"),
			LotNumber:     getStr(input, "lot_number"),
			ReferenceType: getStr(input, "reference_type"),
			ReferenceID:   getStr(input, "reference_id"),
			Notes:         getStr(input, "notes"),
			CreatedBy:     getStr(input, "created_by"),
		}
		return db.Create(&txn).Error
	})
	if err != nil {
		return nil, fmt.Errorf("receive inventory failed: %w", err)
	}

	// Upsert inventory level
	var inv ErpInventory
	result := db.Where("material_id = ? AND warehouse_id = ?", materialID, warehouseID).First(&inv)
	if result.Error != nil {
		inv = ErpInventory{
			ID:          uuid.New().String(),
			MaterialID:  materialID,
			WarehouseID: warehouseID,
			LocationID:  getStr(input, "location_id"),
			LotNumber:   getStr(input, "lot_number"),
			Quantity:    qty,
			UnitCost:    getFloat(input, "unit_cost"),
			Status:      "available",
		}
		db.Create(&inv)
	} else {
		db.Model(&inv).Update("quantity", gorm.Expr("quantity + ?", qty))
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.received",
		fmt.Sprintf("入库: %s, 数量%.2f", code, qty),
		map[string]any{"transaction_id": txn.ID, "code": code, "material_id": materialID, "quantity": qty})

	return map[string]any{"transaction_id": txn.ID, "code": code}, nil
}

func cmdIssueInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseID := getStr(input, "warehouse_id")
	qty := getFloat(input, "quantity")
	if materialID == "" {
		return nil, fmt.Errorf("material_id is required")
	}
	if warehouseID == "" {
		return nil, fmt.Errorf("warehouse_id is required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Check available stock
	var inv ErpInventory
	if err := db.Where("material_id = ? AND warehouse_id = ?", materialID, warehouseID).First(&inv).Error; err != nil {
		return nil, fmt.Errorf("no inventory record for this material/warehouse")
	}
	available := inv.Quantity - inv.ReservedQty
	if available < qty {
		return nil, fmt.Errorf("insufficient stock: available %.2f, requested %.2f", available, qty)
	}

	// Create transaction
	var txn ErpInventoryTransaction
	code, err := autoCodeWithRetry(db, "erp_inventory_transactions", "IT", func(c string) error {
		txn = ErpInventoryTransaction{
			ID:              uuid.New().String(),
			Code:            c,
			Type:            "issue",
			MaterialID:      materialID,
			FromWarehouseID: warehouseID,
			FromLocationID:  getStr(input, "location_id"),
			Quantity:        qty,
			UnitCost:        inv.UnitCost,
			LotNumber:       getStr(input, "lot_number"),
			ReferenceType:   getStr(input, "reference_type"),
			ReferenceID:     getStr(input, "reference_id"),
			Notes:           getStr(input, "notes"),
			CreatedBy:       getStr(input, "created_by"),
		}
		return db.Create(&txn).Error
	})
	if err != nil {
		return nil, fmt.Errorf("issue inventory failed: %w", err)
	}

	// Decrease inventory
	db.Model(&inv).Update("quantity", gorm.Expr("quantity - ?", qty))

	emitEvent(adapter, runID, stepID, "erp.inventory.issued",
		fmt.Sprintf("出库: %s, 数量%.2f", code, qty),
		map[string]any{"transaction_id": txn.ID, "code": code, "material_id": materialID, "quantity": qty})

	return map[string]any{"transaction_id": txn.ID, "code": code}, nil
}

func cmdTransferInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	fromWarehouse := getStr(input, "from_warehouse_id")
	toWarehouse := getStr(input, "to_warehouse_id")
	qty := getFloat(input, "quantity")
	if materialID == "" || fromWarehouse == "" || toWarehouse == "" {
		return nil, fmt.Errorf("material_id, from_warehouse_id, to_warehouse_id are required")
	}
	if qty <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Check source stock
	var srcInv ErpInventory
	if err := db.Where("material_id = ? AND warehouse_id = ?", materialID, fromWarehouse).First(&srcInv).Error; err != nil {
		return nil, fmt.Errorf("no inventory at source warehouse")
	}
	available := srcInv.Quantity - srcInv.ReservedQty
	if available < qty {
		return nil, fmt.Errorf("insufficient stock: available %.2f, requested %.2f", available, qty)
	}

	// Create transaction
	var txn ErpInventoryTransaction
	code, err := autoCodeWithRetry(db, "erp_inventory_transactions", "IT", func(c string) error {
		txn = ErpInventoryTransaction{
			ID:              uuid.New().String(),
			Code:            c,
			Type:            "transfer",
			MaterialID:      materialID,
			FromWarehouseID: fromWarehouse,
			FromLocationID:  getStr(input, "from_location_id"),
			ToWarehouseID:   toWarehouse,
			ToLocationID:    getStr(input, "to_location_id"),
			Quantity:        qty,
			UnitCost:        srcInv.UnitCost,
			LotNumber:       getStr(input, "lot_number"),
			Notes:           getStr(input, "notes"),
			CreatedBy:       getStr(input, "created_by"),
		}
		return db.Create(&txn).Error
	})
	if err != nil {
		return nil, fmt.Errorf("transfer inventory failed: %w", err)
	}

	// Decrease source
	db.Model(&srcInv).Update("quantity", gorm.Expr("quantity - ?", qty))

	// Increase destination (upsert)
	var dstInv ErpInventory
	result := db.Where("material_id = ? AND warehouse_id = ?", materialID, toWarehouse).First(&dstInv)
	if result.Error != nil {
		dstInv = ErpInventory{
			ID:          uuid.New().String(),
			MaterialID:  materialID,
			WarehouseID: toWarehouse,
			LocationID:  getStr(input, "to_location_id"),
			Quantity:    qty,
			UnitCost:    srcInv.UnitCost,
			Status:      "available",
		}
		db.Create(&dstInv)
	} else {
		db.Model(&dstInv).Update("quantity", gorm.Expr("quantity + ?", qty))
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.transferred",
		fmt.Sprintf("调拨: %s, 数量%.2f", code, qty),
		map[string]any{"transaction_id": txn.ID, "code": code, "material_id": materialID, "quantity": qty})

	return map[string]any{"transaction_id": txn.ID, "code": code}, nil
}

func cmdAdjustInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
	materialID := getStr(input, "material_id")
	warehouseID := getStr(input, "warehouse_id")
	newQty := getFloat(input, "new_quantity")
	if materialID == "" || warehouseID == "" {
		return nil, fmt.Errorf("material_id and warehouse_id are required")
	}

	var inv ErpInventory
	result := db.Where("material_id = ? AND warehouse_id = ?", materialID, warehouseID).First(&inv)
	oldQty := float64(0)
	if result.Error == nil {
		oldQty = inv.Quantity
	}

	diff := newQty - oldQty
	txnType := "adjust_in"
	if diff < 0 {
		txnType = "adjust_out"
	}

	// Create transaction
	var txn ErpInventoryTransaction
	code, err := autoCodeWithRetry(db, "erp_inventory_transactions", "IT", func(c string) error {
		txn = ErpInventoryTransaction{
			ID:            uuid.New().String(),
			Code:          c,
			Type:          txnType,
			MaterialID:    materialID,
			ToWarehouseID: warehouseID,
			Quantity:      diff,
			Notes:         getStr(input, "reason"),
			CreatedBy:     getStr(input, "created_by"),
		}
		if diff < 0 {
			txn.FromWarehouseID = warehouseID
			txn.ToWarehouseID = ""
			txn.Quantity = -diff
		}
		return db.Create(&txn).Error
	})
	if err != nil {
		return nil, fmt.Errorf("adjust inventory failed: %w", err)
	}

	// Update or create inventory
	if result.Error != nil {
		inv = ErpInventory{
			ID:          uuid.New().String(),
			MaterialID:  materialID,
			WarehouseID: warehouseID,
			Quantity:    newQty,
			Status:      "available",
		}
		db.Create(&inv)
	} else {
		db.Model(&inv).Update("quantity", newQty)
	}

	emitEvent(adapter, runID, stepID, "erp.inventory.adjusted",
		fmt.Sprintf("库存调整: %s, %.2f -> %.2f", code, oldQty, newQty),
		map[string]any{"transaction_id": txn.ID, "code": code, "old_qty": oldQty, "new_qty": newQty})

	return map[string]any{"transaction_id": txn.ID, "code": code, "old_qty": oldQty, "new_qty": newQty}, nil
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

func cmdReserveInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
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
	available := inv.Quantity - inv.ReservedQty
	if available < qty {
		return nil, fmt.Errorf("insufficient available stock: %.2f, requested %.2f", available, qty)
	}

	db.Model(&inv).Update("reserved_qty", gorm.Expr("reserved_qty + ?", qty))

	emitEvent(adapter, runID, stepID, "erp.inventory.reserved",
		fmt.Sprintf("库存预留: 物料%s, 数量%.2f", materialID, qty),
		map[string]any{"material_id": materialID, "warehouse_id": warehouseID, "quantity": qty,
			"reference_type": getStr(input, "reference_type"), "reference_id": getStr(input, "reference_id")})

	return map[string]any{"material_id": materialID, "warehouse_id": warehouseID, "reserved_qty": qty}, nil
}

func cmdUnreserveInventory(db *gorm.DB, adapter sdk.EngineAdapter, runID, stepID string, input map[string]any) (any, error) {
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
	if inv.ReservedQty < qty {
		qty = inv.ReservedQty // unreserve only what's reserved
	}

	db.Model(&inv).Update("reserved_qty", gorm.Expr("reserved_qty - ?", qty))

	emitEvent(adapter, runID, stepID, "erp.inventory.unreserved",
		fmt.Sprintf("取消预留: 物料%s, 数量%.2f", materialID, qty),
		map[string]any{"material_id": materialID, "warehouse_id": warehouseID, "quantity": qty})

	return map[string]any{"material_id": materialID, "warehouse_id": warehouseID, "unreserved_qty": qty}, nil
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

	now := time.Now()
	db.Model(&wo).Updates(map[string]any{"status": "completed", "actual_end": &now})

	// Receive completed products into inventory
	if wo.CompletedQty > 0 {
		warehouseID := wo.WarehouseID
		var inv ErpInventory
		result := db.Where("material_id = ? AND warehouse_id = ?", wo.ProductID, warehouseID).First(&inv)
		if result.Error != nil {
			inv = ErpInventory{
				ID:          uuid.New().String(),
				MaterialID:  wo.ProductID,
				WarehouseID: warehouseID,
				Quantity:    wo.CompletedQty,
				Status:      "available",
			}
			db.Create(&inv)
		} else {
			db.Model(&inv).Update("quantity", gorm.Expr("quantity + ?", wo.CompletedQty))
		}

		// Create receive transaction
		autoCodeWithRetry(db, "erp_inventory_transactions", "IT", func(c string) error {
			txn := ErpInventoryTransaction{
				ID:            uuid.New().String(),
				Code:          c,
				Type:          "wo_receipt",
				MaterialID:    wo.ProductID,
				ToWarehouseID: warehouseID,
				Quantity:      wo.CompletedQty,
				ReferenceType: "work_order",
				ReferenceID:   wo.ID,
			}
			return db.Create(&txn).Error
		})
	}

	emitEvent(adapter, runID, stepID, "erp.work_order.completed",
		fmt.Sprintf("工单完工: %s, 完成%.0f", wo.Code, wo.CompletedQty),
		map[string]any{"work_order_id": wo.ID, "code": wo.Code, "completed_qty": wo.CompletedQty})

	return map[string]any{"work_order_id": wo.ID, "code": wo.Code, "completed_qty": wo.CompletedQty}, nil
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
