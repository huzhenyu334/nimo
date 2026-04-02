package erp

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// Custom API Routes — beyond generic CRUD
// Mounted under /api/m/erp/ by Module.Routes()
// ---------------------------------------------------------------------------

func registerRoutes(m *ERPModule, rg *gin.RouterGroup) {
	db := m.Deps.DB

	// ─── Sales ───

	// GET /order-summary/:id — 订单汇总（含发货、发票、收款状态）
	rg.GET("/order-summary/:id", func(c *gin.Context) {
		orderID := c.Param("id")
		var order ErpSalesOrder
		if err := db.First(&order, "id = ?", orderID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		var items []ErpSalesOrderItem
		db.Where("order_id = ?", orderID).Find(&items)
		var shipments []ErpShipment
		db.Where("order_id = ?", orderID).Find(&shipments)
		var invoices []ErpSalesInvoice
		db.Where("order_id = ?", orderID).Find(&invoices)

		var totalPaid float64
		for _, inv := range invoices {
			totalPaid += inv.PaidAmount
		}

		c.JSON(http.StatusOK, gin.H{
			"order":     order,
			"items":     items,
			"shipments": shipments,
			"invoices":  invoices,
			"summary": gin.H{
				"total":       order.Total,
				"shipped":     len(shipments),
				"invoiced":    len(invoices),
				"paid":        totalPaid,
				"outstanding": order.Total - totalPaid,
			},
		})
	})

	// GET /customer-statement/:id — 客户对账单
	rg.GET("/customer-statement/:id", func(c *gin.Context) {
		customerID := c.Param("id")
		var invoices []ErpSalesInvoice
		db.Where("customer_id = ?", customerID).Order("invoice_date DESC").Find(&invoices)
		var receipts []ErpReceipt
		db.Where("customer_id = ?", customerID).Order("received_date DESC").Find(&receipts)

		var totalInvoiced, totalPaid, balance float64
		for _, inv := range invoices {
			totalInvoiced += inv.Total
			totalPaid += inv.PaidAmount
		}
		balance = totalInvoiced - totalPaid

		c.JSON(http.StatusOK, gin.H{
			"customer_id": customerID,
			"invoices":    invoices,
			"receipts":    receipts,
			"summary": gin.H{
				"total_invoiced": totalInvoiced,
				"total_paid":     totalPaid,
				"balance":        balance,
			},
		})
	})

	// ─── Inventory ───

	// GET /stock-summary — 库存汇总（按物料/仓库）
	rg.GET("/stock-summary", func(c *gin.Context) {
		warehouseID := c.Query("warehouse_id")
		materialID := c.Query("material_id")

		query := db.Table("erp_inventory").
			Select("material_id, warehouse_id, SUM(quantity) as total_qty, SUM(reserved_qty) as total_reserved, SUM(quantity - reserved_qty) as available_qty").
			Group("material_id, warehouse_id")

		if warehouseID != "" {
			query = query.Where("warehouse_id = ?", warehouseID)
		}
		if materialID != "" {
			query = query.Where("material_id = ?", materialID)
		}

		var results []struct {
			MaterialID    string  `json:"material_id"`
			WarehouseID   string  `json:"warehouse_id"`
			TotalQty      float64 `json:"total_qty"`
			TotalReserved float64 `json:"total_reserved"`
			AvailableQty  float64 `json:"available_qty"`
		}
		query.Find(&results)
		c.JSON(http.StatusOK, gin.H{"items": results})
	})

	// GET /stock-movements/:material_id — 物料出入库流水
	rg.GET("/stock-movements/:material_id", func(c *gin.Context) {
		materialID := c.Param("material_id")
		var txns []ErpInventoryTransaction
		db.Where("material_id = ?", materialID).Order("created_at DESC").Limit(100).Find(&txns)
		c.JSON(http.StatusOK, gin.H{"items": txns})
	})

	// GET /stock-aging — 库龄分析
	rg.GET("/stock-aging", func(c *gin.Context) {
		var results []struct {
			MaterialID string  `json:"material_id"`
			Quantity   float64 `json:"quantity"`
			AgeDays    int     `json:"age_days"`
		}
		db.Raw(`
			SELECT material_id, quantity,
				EXTRACT(DAY FROM NOW() - updated_at)::int AS age_days
			FROM erp_inventory
			WHERE quantity > 0
			ORDER BY age_days DESC
		`).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results})
	})

	// GET /serial-trace/:serial — 序列号追溯
	rg.GET("/serial-trace/:serial", func(c *gin.Context) {
		serial := c.Param("serial")
		var sn ErpSerialNumber
		if err := db.First(&sn, "serial_number = ?", serial).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "serial number not found"})
			return
		}
		// Get related transactions
		var txns []ErpInventoryTransaction
		db.Where("lot_number = ? AND material_id = ?", sn.LotNumber, sn.MaterialID).
			Order("created_at").Find(&txns)
		c.JSON(http.StatusOK, gin.H{"serial": sn, "transactions": txns})
	})

	// ─── Production ───

	// GET /mrp-results/:run_id — MRP 运算结果
	rg.GET("/mrp-results/:run_id", func(c *gin.Context) {
		runID := c.Param("run_id")
		var results []ErpMRPResult
		db.Where("run_id = ?", runID).Find(&results)
		c.JSON(http.StatusOK, gin.H{"items": results, "total": len(results)})
	})

	// GET /production-schedule — 生产排程
	rg.GET("/production-schedule", func(c *gin.Context) {
		var orders []ErpWorkOrder
		db.Where("status IN ?", []string{"released", "in_progress"}).
			Order("priority DESC, planned_start ASC").Find(&orders)
		c.JSON(http.StatusOK, gin.H{"items": orders})
	})

	// GET /wo-dashboard — 工单仪表盘
	rg.GET("/wo-dashboard", func(c *gin.Context) {
		type StatusCount struct {
			Status string `json:"status"`
			Count  int64  `json:"count"`
		}
		var counts []StatusCount
		db.Table("erp_work_orders").
			Select("status, COUNT(*) as count").
			Group("status").Scan(&counts)

		var recentOrders []ErpWorkOrder
		db.Order("created_at DESC").Limit(10).Find(&recentOrders)

		c.JSON(http.StatusOK, gin.H{
			"status_counts": counts,
			"recent_orders": recentOrders,
		})
	})

	// ─── Finance ───

	// GET /trial-balance — 试算平衡表
	rg.GET("/trial-balance", func(c *gin.Context) {
		period := c.Query("period")
		if period == "" {
			period = time.Now().Format("2006-01")
		}
		var results []struct {
			AccountID   string  `json:"account_id"`
			AccountCode string  `json:"account_code"`
			AccountName string  `json:"account_name"`
			AccountType string  `json:"account_type"`
			Debit       float64 `json:"debit"`
			Credit      float64 `json:"credit"`
			Balance     float64 `json:"balance"`
		}
		db.Raw(`
			SELECT a.id as account_id, a.code as account_code, a.name as account_name, a.type as account_type,
				COALESCE(SUM(jl.debit), 0) as debit,
				COALESCE(SUM(jl.credit), 0) as credit,
				COALESCE(SUM(jl.debit), 0) - COALESCE(SUM(jl.credit), 0) as balance
			FROM erp_accounts a
			LEFT JOIN erp_journal_lines jl ON jl.account_id = a.id
			LEFT JOIN erp_journal_entries je ON je.id = jl.entry_id AND je.status = 'posted'
				AND je.period <= ?
			WHERE a.is_leaf = true
			GROUP BY a.id, a.code, a.name, a.type
			ORDER BY a.code
		`, period).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results, "period": period})
	})

	// GET /income-statement — 利润表
	rg.GET("/income-statement", func(c *gin.Context) {
		period := c.Query("period")
		if period == "" {
			period = time.Now().Format("2006-01")
		}
		type AccountSum struct {
			Type    string  `json:"type"`
			Amount  float64 `json:"amount"`
		}
		var results []AccountSum
		db.Raw(`
			SELECT a.type,
				CASE WHEN a.type = 'revenue' THEN COALESCE(SUM(jl.credit - jl.debit), 0)
				     ELSE COALESCE(SUM(jl.debit - jl.credit), 0) END as amount
			FROM erp_accounts a
			JOIN erp_journal_lines jl ON jl.account_id = a.id
			JOIN erp_journal_entries je ON je.id = jl.entry_id AND je.status = 'posted'
				AND je.period = ?
			WHERE a.type IN ('revenue', 'expense')
			GROUP BY a.type
		`, period).Scan(&results)

		var revenue, expense float64
		for _, r := range results {
			if r.Type == "revenue" {
				revenue = r.Amount
			} else {
				expense = r.Amount
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"period":  period,
			"revenue": revenue,
			"expense": expense,
			"profit":  revenue - expense,
		})
	})

	// GET /balance-sheet — 资产负债表
	rg.GET("/balance-sheet", func(c *gin.Context) {
		period := c.Query("period")
		if period == "" {
			period = time.Now().Format("2006-01")
		}
		type TypeSum struct {
			Type   string  `json:"type"`
			Amount float64 `json:"amount"`
		}
		var results []TypeSum
		db.Raw(`
			SELECT a.type,
				CASE WHEN a.type IN ('asset', 'expense') THEN COALESCE(SUM(jl.debit - jl.credit), 0)
				     ELSE COALESCE(SUM(jl.credit - jl.debit), 0) END as amount
			FROM erp_accounts a
			JOIN erp_journal_lines jl ON jl.account_id = a.id
			JOIN erp_journal_entries je ON je.id = jl.entry_id AND je.status = 'posted'
				AND je.period <= ?
			WHERE a.type IN ('asset', 'liability', 'equity')
			GROUP BY a.type
		`, period).Scan(&results)

		data := map[string]float64{}
		for _, r := range results {
			data[r.Type] = r.Amount
		}

		c.JSON(http.StatusOK, gin.H{
			"period":    period,
			"assets":    data["asset"],
			"liability": data["liability"],
			"equity":    data["equity"],
		})
	})

	// GET /ar-aging — 应收账龄
	rg.GET("/ar-aging", func(c *gin.Context) {
		var results []struct {
			CustomerID string  `json:"customer_id"`
			Current    float64 `json:"current"`
			Days30     float64 `json:"days_30"`
			Days60     float64 `json:"days_60"`
			Days90     float64 `json:"days_90"`
			Over90     float64 `json:"over_90"`
		}
		db.Raw(`
			SELECT customer_id,
				SUM(CASE WHEN due_date >= CURRENT_DATE THEN total - paid_amount ELSE 0 END) as current,
				SUM(CASE WHEN due_date < CURRENT_DATE AND due_date >= CURRENT_DATE - 30 THEN total - paid_amount ELSE 0 END) as days_30,
				SUM(CASE WHEN due_date < CURRENT_DATE - 30 AND due_date >= CURRENT_DATE - 60 THEN total - paid_amount ELSE 0 END) as days_60,
				SUM(CASE WHEN due_date < CURRENT_DATE - 60 AND due_date >= CURRENT_DATE - 90 THEN total - paid_amount ELSE 0 END) as days_90,
				SUM(CASE WHEN due_date < CURRENT_DATE - 90 THEN total - paid_amount ELSE 0 END) as over_90
			FROM erp_sales_invoices
			WHERE status NOT IN ('cancelled', 'paid')
			GROUP BY customer_id
		`).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results})
	})

	// GET /cost-analysis/:product_id — 产品成本分析
	rg.GET("/cost-analysis/:product_id", func(c *gin.Context) {
		productID := c.Param("product_id")
		// Get BOM items with costs
		var items []struct {
			MaterialID string  `json:"material_id"`
			Quantity   float64 `json:"quantity"`
			UnitCost   float64 `json:"unit_cost"`
			TotalCost  float64 `json:"total_cost"`
		}
		db.Raw(`
			SELECT bi.material_id, bi.quantity,
				COALESCE(i.unit_cost, 0) as unit_cost,
				bi.quantity * COALESCE(i.unit_cost, 0) as total_cost
			FROM plm_bom_items bi
			JOIN plm_boms b ON b.id = bi.bom_id AND b.status = 'released'
			LEFT JOIN (
				SELECT material_id, AVG(unit_cost) as unit_cost
				FROM erp_inventory WHERE quantity > 0
				GROUP BY material_id
			) i ON i.material_id = bi.material_id
			WHERE b.product_id = ?
		`, productID).Scan(&items)

		var totalCost float64
		for _, item := range items {
			totalCost += item.TotalCost
		}

		c.JSON(http.StatusOK, gin.H{
			"product_id": productID,
			"items":      items,
			"total_cost": totalCost,
		})
	})

	// ─── Quality ───

	// GET /oqc-dashboard — OQC 仪表盘
	rg.GET("/oqc-dashboard", func(c *gin.Context) {
		var passRate float64
		db.Raw(`
			SELECT COALESCE(
				SUM(CASE WHEN result = 'pass' THEN 1 ELSE 0 END)::float /
				NULLIF(COUNT(*), 0) * 100, 0
			) FROM erp_oqc_inspections
			WHERE inspected_at >= CURRENT_DATE - 30
		`).Scan(&passRate)

		var recentInspections []ErpOQCInspection
		db.Order("created_at DESC").Limit(10).Find(&recentInspections)

		var openNCRs int64
		db.Table("erp_ncr_reports").Where("status NOT IN ?", []string{"closed"}).Count(&openNCRs)

		c.JSON(http.StatusOK, gin.H{
			"pass_rate_30d":      fmt.Sprintf("%.1f", passRate),
			"recent_inspections": recentInspections,
			"open_ncrs":          openNCRs,
		})
	})

	// GET /quality-trend — 质量趋势
	rg.GET("/quality-trend", func(c *gin.Context) {
		var results []struct {
			Week     string  `json:"week"`
			PassRate float64 `json:"pass_rate"`
			Total    int     `json:"total"`
		}
		db.Raw(`
			SELECT TO_CHAR(DATE_TRUNC('week', inspected_at), 'YYYY-WW') as week,
				COALESCE(SUM(CASE WHEN result = 'pass' THEN 1 ELSE 0 END)::float / NULLIF(COUNT(*), 0) * 100, 0) as pass_rate,
				COUNT(*) as total
			FROM erp_oqc_inspections
			WHERE inspected_at >= CURRENT_DATE - 90
			GROUP BY DATE_TRUNC('week', inspected_at)
			ORDER BY week
		`).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results})
	})

	_ = strings.TrimSpace // suppress unused import
}
