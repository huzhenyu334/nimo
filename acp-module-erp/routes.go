package erp

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
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

	// GET /stock-summary — 库存汇总（按物料/仓库），含 PLM 物料名 + 仓库名 join
	//
	// PRD v1 Sprint 1 升级：
	//   - 跨模块 LEFT JOIN plm_materials 出 material_code / material_name
	//   - JOIN erp_warehouses 出 warehouse_code / warehouse_name
	//   - 多维过滤：material_id / warehouse_id / status / abc_class / lot_number
	//   - 计算总价值 = SUM(quantity × unit_cost)
	//   - 兼容老调用方：query 参数全可选
	rg.GET("/stock-summary", func(c *gin.Context) {
		warehouseID := c.Query("warehouse_id")
		materialID := c.Query("material_id")
		status := c.Query("status")
		lotNumber := c.Query("lot_number")
		abcClass := c.Query("abc_class")

		query := db.Table("erp_inventory AS inv").
			Select(`
				inv.material_id,
				inv.warehouse_id,
				COALESCE(plm_materials.code, '') AS material_code,
				COALESCE(plm_materials.name, '') AS material_name,
				COALESCE(plm_materials.specs, '') AS material_spec,
				COALESCE(erp_warehouses.code, '') AS warehouse_code,
				COALESCE(erp_warehouses.name, '') AS warehouse_name,
				COALESCE(erp_material_inventory_attrs.abc_class, 'C') AS abc_class,
				COALESCE(erp_material_inventory_attrs.safety_stock, 0) AS safety_stock,
				COALESCE(erp_material_inventory_attrs.tracking_mode, 'lot') AS tracking_mode,
				COALESCE(erp_material_inventory_attrs.default_unit, 'pcs') AS unit,
				SUM(inv.quantity) AS total_qty,
				SUM(inv.reserved_qty) AS reserved_qty,
				SUM(inv.quantity - inv.reserved_qty) AS available_qty,
				MAX(inv.unit_cost) AS unit_cost,
				SUM(inv.quantity * inv.unit_cost) AS total_value,
				COUNT(*) AS lot_count,
				MIN(inv.received_at) AS oldest_received_at
			`).
			Joins("LEFT JOIN plm_materials ON plm_materials.id = inv.material_id").
			Joins("LEFT JOIN erp_warehouses ON erp_warehouses.id = inv.warehouse_id").
			Joins("LEFT JOIN erp_material_inventory_attrs ON erp_material_inventory_attrs.material_id = inv.material_id").
			Group("inv.material_id, inv.warehouse_id, plm_materials.code, plm_materials.name, plm_materials.specs, erp_warehouses.code, erp_warehouses.name, erp_material_inventory_attrs.abc_class, erp_material_inventory_attrs.safety_stock, erp_material_inventory_attrs.tracking_mode, erp_material_inventory_attrs.default_unit").
			Order("plm_materials.code ASC, erp_warehouses.code ASC")

		if warehouseID != "" {
			query = query.Where("inv.warehouse_id = ?", warehouseID)
		}
		if materialID != "" {
			query = query.Where("inv.material_id = ?", materialID)
		}
		if status != "" {
			query = query.Where("inv.status = ?", status)
		}
		if lotNumber != "" {
			query = query.Where("inv.lot_number = ?", lotNumber)
		}
		if abcClass != "" {
			query = query.Where("erp_material_inventory_attrs.abc_class = ?", abcClass)
		}

		var results []map[string]any
		query.Find(&results)

		// Top-level totals across the filtered result set
		var grandTotalQty, grandTotalValue float64
		var grandReserved float64
		for _, r := range results {
			if v, ok := r["total_qty"].(float64); ok {
				grandTotalQty += v
			}
			if v, ok := r["total_value"].(float64); ok {
				grandTotalValue += v
			}
			if v, ok := r["reserved_qty"].(float64); ok {
				grandReserved += v
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"items": results,
			"summary": gin.H{
				"row_count":     len(results),
				"total_qty":     grandTotalQty,
				"total_value":   grandTotalValue,
				"reserved_qty":  grandReserved,
				"available_qty": grandTotalQty - grandReserved,
			},
		})
	})

	// GET /stock-detail — 单物料 × 仓库的批次明细 (展开 stock-summary 的某一行)
	rg.GET("/stock-detail", func(c *gin.Context) {
		materialID := c.Query("material_id")
		warehouseID := c.Query("warehouse_id")
		if materialID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "material_id is required"})
			return
		}
		query := db.Table("erp_inventory AS inv").
			Select(`
				inv.*,
				COALESCE(plm_materials.code, '') AS material_code,
				COALESCE(plm_materials.name, '') AS material_name,
				COALESCE(erp_warehouses.code, '') AS warehouse_code,
				COALESCE(erp_warehouses.name, '') AS warehouse_name
			`).
			Joins("LEFT JOIN plm_materials ON plm_materials.id = inv.material_id").
			Joins("LEFT JOIN erp_warehouses ON erp_warehouses.id = inv.warehouse_id").
			Where("inv.material_id = ?", materialID).
			Order("inv.received_at ASC NULLS LAST")
		if warehouseID != "" {
			query = query.Where("inv.warehouse_id = ?", warehouseID)
		}
		var rows []map[string]any
		query.Find(&rows)
		c.JSON(http.StatusOK, gin.H{"items": rows})
	})

	// GET /stock-movements/:material_id — 物料出入库流水
	rg.GET("/stock-movements/:material_id", func(c *gin.Context) {
		materialID := c.Param("material_id")
		var txns []ErpInventoryTransaction
		db.Where("material_id = ?", materialID).Order("created_at DESC").Limit(100).Find(&txns)
		c.JSON(http.StatusOK, gin.H{"items": txns})
	})

	// GET /stock-movements — 全量流水（PRD v1 Sprint 1，含名称 join + 多维过滤）
	rg.GET("/stock-movements", func(c *gin.Context) {
		warehouseID := c.Query("warehouse_id")
		materialID := c.Query("material_id")
		txnType := c.Query("type")
		refType := c.Query("reference_type")
		from := c.Query("from")
		to := c.Query("to")
		limit := 200
		if l := c.Query("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}

		query := db.Table("erp_inventory_transactions AS t").
			Select(`
				t.*,
				COALESCE(plm_materials.code, '') AS material_code,
				COALESCE(plm_materials.name, '') AS material_name,
				COALESCE(from_wh.code, '') AS from_warehouse_code,
				COALESCE(from_wh.name, '') AS from_warehouse_name,
				COALESCE(to_wh.code, '') AS to_warehouse_code,
				COALESCE(to_wh.name, '') AS to_warehouse_name
			`).
			Joins("LEFT JOIN plm_materials ON plm_materials.id = t.material_id").
			Joins("LEFT JOIN erp_warehouses AS from_wh ON from_wh.id = t.from_warehouse_id").
			Joins("LEFT JOIN erp_warehouses AS to_wh ON to_wh.id = t.to_warehouse_id").
			Order("t.created_at DESC").
			Limit(limit)

		if materialID != "" {
			query = query.Where("t.material_id = ?", materialID)
		}
		if warehouseID != "" {
			query = query.Where("(t.from_warehouse_id = ? OR t.to_warehouse_id = ?)", warehouseID, warehouseID)
		}
		if txnType != "" {
			query = query.Where("t.type = ?", txnType)
		}
		if refType != "" {
			query = query.Where("t.reference_type = ?", refType)
		}
		if from != "" {
			query = query.Where("t.created_at >= ?", from)
		}
		if to != "" {
			query = query.Where("t.created_at <= ?", to)
		}

		var rows []map[string]any
		query.Find(&rows)
		c.JSON(http.StatusOK, gin.H{"items": rows, "count": len(rows)})
	})

	// GET /stock-value — 库存价值汇总（PRD v1 Sprint 2）
	//
	// 按仓库汇总，支持 as_of 时间点（v1 默认当前时刻）。返回每个仓库的库存价值
	// 和总价值，作为财务对账基础。
	rg.GET("/stock-value", func(c *gin.Context) {
		warehouseID := c.Query("warehouse_id")

		query := db.Table("erp_inventory AS inv").
			Select(`
				inv.warehouse_id,
				COALESCE(erp_warehouses.code, '') AS warehouse_code,
				COALESCE(erp_warehouses.name, '') AS warehouse_name,
				COALESCE(erp_warehouses.type, '') AS warehouse_type,
				COUNT(DISTINCT inv.material_id) AS material_count,
				SUM(inv.quantity) AS total_qty,
				SUM(inv.quantity * inv.unit_cost) AS total_value
			`).
			Joins("LEFT JOIN erp_warehouses ON erp_warehouses.id = inv.warehouse_id").
			Where("inv.status = ?", "available").
			Group("inv.warehouse_id, erp_warehouses.code, erp_warehouses.name, erp_warehouses.type").
			Order("erp_warehouses.code")

		if warehouseID != "" {
			query = query.Where("inv.warehouse_id = ?", warehouseID)
		}

		var rows []map[string]any
		query.Find(&rows)

		var grandTotal float64
		for _, r := range rows {
			if v, ok := r["total_value"].(float64); ok {
				grandTotal += v
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"items":          rows,
			"grand_total":    grandTotal,
			"as_of":          time.Now(),
			"warehouse_count": len(rows),
		})
	})

	// GET /stock-aging — 库龄分析
	// GET /stock-aging — 库龄报表（按 received_at 计算批次库龄）
	rg.GET("/stock-aging", func(c *gin.Context) {
		type agingRow struct {
			MaterialID    string  `json:"material_id"`
			MaterialName  string  `json:"material_name"`
			MaterialCode  string  `json:"material_code"`
			WarehouseID   string  `json:"warehouse_id"`
			WarehouseCode string  `json:"warehouse_code"`
			LotNumber     string  `json:"lot_number"`
			Quantity      float64 `json:"quantity"`
			UnitCost      float64 `json:"unit_cost"`
			Value         float64 `json:"value"`
			AgeDays       int     `json:"age_days"`
			AgeBucket     string  `json:"age_bucket"`
		}
		results := []agingRow{}
		db.Raw(`
			SELECT
				i.material_id,
				COALESCE(m.name, '') AS material_name,
				COALESCE(m.code, '') AS material_code,
				i.warehouse_id,
				COALESCE(w.code, '') AS warehouse_code,
				i.lot_number,
				i.quantity,
				i.unit_cost,
				i.quantity * i.unit_cost AS value,
				COALESCE(EXTRACT(DAY FROM NOW() - COALESCE(i.received_at, i.created_at))::int, 0) AS age_days,
				CASE
					WHEN COALESCE(EXTRACT(DAY FROM NOW() - COALESCE(i.received_at, i.created_at))::int, 0) <= 30 THEN '0-30'
					WHEN COALESCE(EXTRACT(DAY FROM NOW() - COALESCE(i.received_at, i.created_at))::int, 0) <= 90 THEN '31-90'
					WHEN COALESCE(EXTRACT(DAY FROM NOW() - COALESCE(i.received_at, i.created_at))::int, 0) <= 180 THEN '91-180'
					WHEN COALESCE(EXTRACT(DAY FROM NOW() - COALESCE(i.received_at, i.created_at))::int, 0) <= 365 THEN '181-365'
					ELSE '365+'
				END AS age_bucket
			FROM erp_inventory i
			LEFT JOIN plm_materials m ON m.id = i.material_id
			LEFT JOIN erp_warehouses w ON w.id = i.warehouse_id
			WHERE i.quantity > 0
			ORDER BY age_days DESC, value DESC
		`).Scan(&results)

		// Group stats
		buckets := map[string]struct {
			Qty   float64
			Value float64
			Count int
		}{}
		for _, r := range results {
			b := buckets[r.AgeBucket]
			b.Qty += r.Quantity
			b.Value += r.Value
			b.Count++
			buckets[r.AgeBucket] = b
		}
		c.JSON(http.StatusOK, gin.H{
			"items":   results,
			"buckets": buckets,
			"total":   len(results),
		})
	})

	// GET /stock-dead — 呆滞库存（X 天未出库）
	rg.GET("/stock-dead", func(c *gin.Context) {
		days := 90
		if d := c.Query("days"); d != "" {
			if parsed, err := strconv.Atoi(d); err == nil {
				days = parsed
			}
		}
		type deadRow struct {
			MaterialID   string  `json:"material_id"`
			MaterialName string  `json:"material_name"`
			MaterialCode string  `json:"material_code"`
			Quantity     float64 `json:"quantity"`
			Value        float64 `json:"value"`
			LastIssuedAt string  `json:"last_issued_at"`
			DaysSince    int     `json:"days_since"`
		}
		results := []deadRow{}
		db.Raw(`
			SELECT
				i.material_id,
				COALESCE(m.name, '') AS material_name,
				COALESCE(m.code, '') AS material_code,
				SUM(i.quantity) AS quantity,
				SUM(i.quantity * i.unit_cost) AS value,
				COALESCE(MAX(a.last_issued_at)::text, '') AS last_issued_at,
				COALESCE(EXTRACT(DAY FROM NOW() - MAX(a.last_issued_at))::int, 9999) AS days_since
			FROM erp_inventory i
			LEFT JOIN plm_materials m ON m.id = i.material_id
			LEFT JOIN erp_material_inventory_attrs a ON a.material_id = i.material_id
			WHERE i.quantity > 0
			GROUP BY i.material_id, m.name, m.code
			HAVING COALESCE(EXTRACT(DAY FROM NOW() - MAX(a.last_issued_at))::int, 9999) >= ?
			ORDER BY days_since DESC
		`, days).Scan(&results)

		var totalValue float64
		for _, r := range results {
			totalValue += r.Value
		}
		c.JSON(http.StatusOK, gin.H{
			"items":       results,
			"total":       len(results),
			"threshold":   days,
			"total_value": totalValue,
		})
	})

	// GET /stock-turnover — 库存周转率（按 30 天计算）
	rg.GET("/stock-turnover", func(c *gin.Context) {
		days := 30
		if d := c.Query("days"); d != "" {
			if parsed, err := strconv.Atoi(d); err == nil {
				days = parsed
			}
		}
		type turnoverRow struct {
			MaterialID   string  `json:"material_id"`
			MaterialName string  `json:"material_name"`
			MaterialCode string  `json:"material_code"`
			OnHand       float64 `json:"on_hand"`
			OnHandValue  float64 `json:"on_hand_value"`
			IssuedQty    float64 `json:"issued_qty"`
			IssuedValue  float64 `json:"issued_value"`
			TurnoverRate float64 `json:"turnover_rate"`
		}
		results := []turnoverRow{}
		sql := fmt.Sprintf(`
			SELECT
				i.material_id,
				COALESCE(m.name, '') AS material_name,
				COALESCE(m.code, '') AS material_code,
				SUM(i.quantity) AS on_hand,
				SUM(i.quantity * i.unit_cost) AS on_hand_value,
				COALESCE((
					SELECT SUM(t.quantity)
					FROM erp_inventory_transactions t
					WHERE t.material_id = i.material_id
					  AND t.type IN ('production_issue','sales_issue')
					  AND t.created_at >= NOW() - INTERVAL '%d days'
				), 0) AS issued_qty,
				COALESCE((
					SELECT SUM(t.quantity * t.unit_cost)
					FROM erp_inventory_transactions t
					WHERE t.material_id = i.material_id
					  AND t.type IN ('production_issue','sales_issue')
					  AND t.created_at >= NOW() - INTERVAL '%d days'
				), 0) AS issued_value,
				CASE
					WHEN SUM(i.quantity * i.unit_cost) > 0
					THEN COALESCE((
						SELECT SUM(t.quantity * t.unit_cost)
						FROM erp_inventory_transactions t
						WHERE t.material_id = i.material_id
						  AND t.type IN ('production_issue','sales_issue')
						  AND t.created_at >= NOW() - INTERVAL '%d days'
					), 0) / SUM(i.quantity * i.unit_cost)
					ELSE 0
				END AS turnover_rate
			FROM erp_inventory i
			LEFT JOIN plm_materials m ON m.id = i.material_id
			WHERE i.quantity > 0
			GROUP BY i.material_id, m.name, m.code
			ORDER BY turnover_rate DESC
		`, days, days, days)
		db.Raw(sql).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results, "days": days})
	})

	// GET /safety-stock-alerts — 安全库存预警
	rg.GET("/safety-stock-alerts", func(c *gin.Context) {
		type alertRow struct {
			MaterialID   string  `json:"material_id"`
			MaterialName string  `json:"material_name"`
			MaterialCode string  `json:"material_code"`
			OnHand       float64 `json:"on_hand"`
			Reserved     float64 `json:"reserved"`
			Available    float64 `json:"available"`
			SafetyStock  float64 `json:"safety_stock"`
			ReorderPoint float64 `json:"reorder_point"`
			Shortfall    float64 `json:"shortfall"`
			ABCClass     string  `json:"abc_class"`
			Severity     string  `json:"severity"`
		}
		results := []alertRow{}
		db.Raw(`
			SELECT
				i.material_id,
				COALESCE(m.name, '') AS material_name,
				COALESCE(m.code, '') AS material_code,
				SUM(i.quantity) AS on_hand,
				SUM(i.reserved_qty) AS reserved,
				SUM(i.quantity - i.reserved_qty) AS available,
				COALESCE(MAX(a.safety_stock), 0) AS safety_stock,
				COALESCE(MAX(a.reorder_point), 0) AS reorder_point,
				GREATEST(COALESCE(MAX(a.safety_stock), 0) - SUM(i.quantity - i.reserved_qty), 0) AS shortfall,
				COALESCE(MAX(a.abc_class), 'C') AS abc_class,
				CASE
					WHEN SUM(i.quantity - i.reserved_qty) <= 0 THEN 'stockout'
					WHEN SUM(i.quantity - i.reserved_qty) < COALESCE(MAX(a.safety_stock), 0) THEN 'below_safety'
					WHEN SUM(i.quantity - i.reserved_qty) < COALESCE(MAX(a.reorder_point), 0) THEN 'below_reorder'
					ELSE 'ok'
				END AS severity
			FROM erp_inventory i
			LEFT JOIN plm_materials m ON m.id = i.material_id
			LEFT JOIN erp_material_inventory_attrs a ON a.material_id = i.material_id
			WHERE i.status = 'available'
			GROUP BY i.material_id, m.name, m.code
			HAVING SUM(i.quantity - i.reserved_qty) < COALESCE(MAX(a.reorder_point), 0)
			   OR SUM(i.quantity - i.reserved_qty) <= 0
			ORDER BY severity, shortfall DESC
		`).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results, "total": len(results)})
	})

	// GET /serial-trace/:serial — 序列号追溯
	// GET /count-detail/:count_id — 盘点单明细 + 统计
	rg.GET("/count-detail/:count_id", func(c *gin.Context) {
		countID := c.Param("count_id")
		var count ErpInventoryCount
		if err := db.First(&count, "id = ?", countID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "count not found"})
			return
		}
		var lines []ErpInventoryCountLine
		db.Where("count_id = ?", countID).Order("material_id, lot_number").Find(&lines)

		// Enrich with material name + warehouse name
		type enrichedLine struct {
			ErpInventoryCountLine
			MaterialName  string `json:"material_name"`
			MaterialCode  string `json:"material_code"`
			WarehouseName string `json:"warehouse_name"`
			WarehouseCode string `json:"warehouse_code"`
		}
		enriched := make([]enrichedLine, 0, len(lines))
		for _, l := range lines {
			el := enrichedLine{ErpInventoryCountLine: l}
			var mat struct {
				Name string
				Code string
			}
			db.Table("plm_materials").Select("name, code").Where("id = ?", l.MaterialID).Scan(&mat)
			el.MaterialName = mat.Name
			el.MaterialCode = mat.Code
			var wh struct {
				Name string
				Code string
			}
			db.Table("erp_warehouses").Select("name, code").Where("id = ?", l.WarehouseID).Scan(&wh)
			el.WarehouseName = wh.Name
			el.WarehouseCode = wh.Code
			enriched = append(enriched, el)
		}

		// Stats
		var countedLines, variantLines int
		var totalVariance float64
		for _, l := range lines {
			if l.CountedQty >= 0 {
				countedLines++
			}
			if l.Variance != 0 {
				variantLines++
				totalVariance += l.VarianceAmt
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"count": count,
			"lines": enriched,
			"stats": gin.H{
				"total_lines":    len(lines),
				"counted_lines":  countedLines,
				"variance_lines": variantLines,
				"total_variance": totalVariance,
			},
		})
	})

	// GET /counts — 盘点单列表
	rg.GET("/counts", func(c *gin.Context) {
		var counts []ErpInventoryCount
		q := db.Order("created_at DESC")
		if status := c.Query("status"); status != "" {
			q = q.Where("status = ?", status)
		}
		if t := c.Query("type"); t != "" {
			q = q.Where("type = ?", t)
		}
		if wh := c.Query("warehouse_id"); wh != "" {
			q = q.Where("warehouse_id = ?", wh)
		}
		q.Limit(200).Find(&counts)
		c.JSON(http.StatusOK, gin.H{"counts": counts, "total": len(counts)})
	})

	// GET /pda/scan/:barcode — PDA 扫码查询（统一入口）
	// 输入可以是序列号、批次号或物料编码，自动识别
	rg.GET("/pda/scan/:barcode", func(c *gin.Context) {
		barcode := c.Param("barcode")
		result := gin.H{"barcode": barcode}

		// 1) Try as serial number
		var sn ErpSerialNumber
		if err := db.First(&sn, "serial_number = ?", barcode).Error; err == nil {
			var mat struct {
				Name string
				Code string
			}
			db.Table("plm_materials").Select("name, code").Where("id = ?", sn.MaterialID).Scan(&mat)
			var wh ErpWarehouse
			db.First(&wh, "id = ?", sn.WarehouseID)
			result["type"] = "serial"
			result["serial"] = sn
			result["material_name"] = mat.Name
			result["material_code"] = mat.Code
			result["warehouse_code"] = wh.Code
			c.JSON(http.StatusOK, result)
			return
		}

		// 2) Try as lot number
		var lotRows []ErpInventory
		if err := db.Where("lot_number = ?", barcode).Find(&lotRows).Error; err == nil && len(lotRows) > 0 {
			result["type"] = "lot"
			result["lot_number"] = barcode
			result["inventory_rows"] = lotRows
			result["row_count"] = len(lotRows)
			c.JSON(http.StatusOK, result)
			return
		}

		// 3) Try as material code (from plm_materials)
		var matID string
		db.Table("plm_materials").Select("id").Where("code = ?", barcode).Scan(&matID)
		if matID != "" {
			var invs []ErpInventory
			db.Where("material_id = ? AND quantity > 0", matID).Find(&invs)
			var totalQty float64
			for _, inv := range invs {
				totalQty += inv.Quantity
			}
			result["type"] = "material"
			result["material_id"] = matID
			result["material_code"] = barcode
			result["total_qty"] = totalQty
			result["location_count"] = len(invs)
			c.JSON(http.StatusOK, result)
			return
		}

		// 4) Not found
		c.JSON(http.StatusNotFound, gin.H{"error": "barcode not found", "barcode": barcode})
	})

	// POST /pda/quick-issue — PDA 快速出库（扫码后直接扣减）
	rg.POST("/pda/quick-issue", func(c *gin.Context) {
		var req struct {
			SerialNumber string  `json:"serial_number"`
			LotNumber    string  `json:"lot_number"`
			Quantity     float64 `json:"quantity"`
			Reason       string  `json:"reason"`
			OperatorID   string  `json:"operator_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid json"})
			return
		}
		if req.SerialNumber != "" {
			// Locate inventory via serial number → material+warehouse
			var sn ErpSerialNumber
			if err := db.First(&sn, "serial_number = ?", req.SerialNumber).Error; err != nil {
				c.JSON(404, gin.H{"error": "serial not found"})
				return
			}
			// Update serial to scrapped/issued
			db.Model(&sn).Update("status", "shipped")
			c.JSON(200, gin.H{"ok": true, "serial": req.SerialNumber, "status": "shipped"})
			return
		}
		c.JSON(400, gin.H{"error": "serial_number or lot_number required"})
	})

	// GET /wo-kitting/:wo_id — 工单齐套检查
	rg.GET("/wo-kitting/:wo_id", func(c *gin.Context) {
		woID := c.Param("wo_id")
		var wo ErpWorkOrder
		if err := db.First(&wo, "id = ?", woID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "work order not found"})
			return
		}

		// Get BOM items
		type bomItem struct {
			MaterialID   string
			MaterialName string
			MaterialCode string
			RequiredQty  float64
			OnHand       float64
			Reserved     float64
			Available    float64
			Shortage     float64
			KittingOK    bool
		}
		items := []bomItem{}
		if wo.BomID != "" {
			type bomRow struct {
				MaterialID string
				Quantity   float64
				Name       string
				Code       string
			}
			var bomRows []bomRow
			db.Raw(`
				SELECT bi.material_id, bi.quantity, m.name, m.code
				FROM plm_bom_items bi
				LEFT JOIN plm_materials m ON m.id = bi.material_id
				WHERE bi.bom_id = ?
			`, wo.BomID).Scan(&bomRows)

			for _, row := range bomRows {
				required := row.Quantity * wo.PlannedQty
				var onHand, reserved float64
				db.Raw(`
					SELECT COALESCE(SUM(quantity), 0) AS total_qty,
					       COALESCE(SUM(reserved_qty), 0) AS reserved_qty
					FROM erp_inventory
					WHERE material_id = ? AND status = 'available'
				`, row.MaterialID).Row().Scan(&onHand, &reserved)
				available := onHand - reserved
				shortage := required - available
				if shortage < 0 {
					shortage = 0
				}
				items = append(items, bomItem{
					MaterialID:   row.MaterialID,
					MaterialName: row.Name,
					MaterialCode: row.Code,
					RequiredQty:  required,
					OnHand:       onHand,
					Reserved:     reserved,
					Available:    available,
					Shortage:     shortage,
					KittingOK:    shortage == 0,
				})
			}
		}

		// Overall kitting status
		allOK := true
		shortCount := 0
		for _, it := range items {
			if !it.KittingOK {
				allOK = false
				shortCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"work_order":  wo,
			"items":       items,
			"kitting_ok":  allOK,
			"total_items": len(items),
			"short_items": shortCount,
		})
	})

	// GET /reservations — 库存预留列表
	rg.GET("/reservations", func(c *gin.Context) {
		var reservations []ErpInventoryReservation
		q := db.Order("created_at DESC")
		if status := c.Query("status"); status != "" {
			q = q.Where("status = ?", status)
		} else {
			q = q.Where("status = ?", "active")
		}
		if mid := c.Query("material_id"); mid != "" {
			q = q.Where("material_id = ?", mid)
		}
		if src := c.Query("source_id"); src != "" {
			q = q.Where("source_id = ?", src)
		}
		q.Limit(200).Find(&reservations)
		c.JSON(http.StatusOK, gin.H{"reservations": reservations, "total": len(reservations)})
	})

	// GET /serial-trace/:serial — 序列号完整追溯链
	// 返回: serial + material + work_order + shipment + return + timeline
	rg.GET("/serial-trace/:serial", func(c *gin.Context) {
		serial := c.Param("serial")
		var sn ErpSerialNumber
		if err := db.First(&sn, "serial_number = ?", serial).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "serial number not found"})
			return
		}

		// Material
		var material struct {
			Name string `json:"name"`
			Code string `json:"code"`
			Unit string `json:"unit"`
		}
		if sn.MaterialID != "" {
			db.Table("plm_materials").Select("name, code, unit").Where("id = ?", sn.MaterialID).Scan(&material)
		}

		// Warehouse
		var warehouse ErpWarehouse
		if sn.WarehouseID != "" {
			db.First(&warehouse, "id = ?", sn.WarehouseID)
		}

		// Work order
		var workOrder ErpWorkOrder
		if sn.WorkOrderID != "" {
			db.First(&workOrder, "id = ?", sn.WorkOrderID)
		}

		// Shipment
		var shipment ErpShipment
		if sn.ShipmentID != "" {
			db.First(&shipment, "id = ?", sn.ShipmentID)
		}

		// Customer
		var customer ErpCustomer
		if sn.CustomerID != "" {
			db.First(&customer, "id = ?", sn.CustomerID)
		}

		// Return
		var ret ErpReturn
		if sn.ReturnID != "" {
			db.First(&ret, "id = ?", sn.ReturnID)
		}

		// Related transactions (same lot, same material)
		var txns []ErpInventoryTransaction
		if sn.LotNumber != "" {
			db.Where("lot_number = ? AND material_id = ?", sn.LotNumber, sn.MaterialID).
				Order("created_at").Find(&txns)
		}

		// Build timeline
		type timelineItem struct {
			Time  string `json:"time"`
			Event string `json:"event"`
			Ref   string `json:"ref"`
			Label string `json:"label"`
		}
		timeline := []timelineItem{}
		if sn.ManufacturedAt != nil {
			timeline = append(timeline, timelineItem{
				Time: sn.ManufacturedAt.Format("2006-01-02 15:04"),
				Event: "生产",
				Ref: workOrder.Code,
				Label: fmt.Sprintf("工单 %s 入库", workOrder.Code),
			})
		}
		if sn.SoldAt != nil {
			timeline = append(timeline, timelineItem{
				Time: sn.SoldAt.Format("2006-01-02 15:04"),
				Event: "销售",
				Ref: shipment.Code,
				Label: fmt.Sprintf("发货 %s 客户 %s", shipment.Code, customer.Name),
			})
		}
		if ret.ID != "" {
			timeline = append(timeline, timelineItem{
				Time: ret.CreatedAt.Format("2006-01-02 15:04"),
				Event: "退货",
				Ref: ret.Code,
				Label: fmt.Sprintf("退货单 %s (%s)", ret.Code, ret.Reason),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"serial":       sn,
			"material":     material,
			"warehouse":    warehouse,
			"work_order":   workOrder,
			"shipment":     shipment,
			"customer":     customer,
			"return":       ret,
			"transactions": txns,
			"timeline":     timeline,
		})
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

	// ─── Dashboard / Analytics ───

	// GET /activity-log — 最近操作日志
	rg.GET("/activity-log", func(c *gin.Context) {
		limit := 10
		if l := c.Query("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
				limit = n
			}
		}

		type Activity struct {
			Time    time.Time `json:"time"`
			Type    string    `json:"type"`
			Summary string    `json:"summary"`
			RefID   string    `json:"ref_id"`
		}

		var activities []Activity

		// Recent inventory transactions
		var txns []ErpInventoryTransaction
		db.Order("created_at DESC").Limit(limit).Find(&txns)
		for _, t := range txns {
			activities = append(activities, Activity{
				Time: t.CreatedAt, Type: "inventory",
				Summary: fmt.Sprintf("%s %s 数量%v", t.Type, t.MaterialID, t.Quantity),
				RefID:   t.ID,
			})
		}

		// Recent orders
		var orders []ErpSalesOrder
		db.Order("created_at DESC").Limit(5).Find(&orders)
		for _, o := range orders {
			activities = append(activities, Activity{
				Time: o.CreatedAt, Type: "sales_order",
				Summary: fmt.Sprintf("销售订单 %s %s", o.Code, o.Status),
				RefID:   o.ID,
			})
		}

		// Sort by time desc
		sort.Slice(activities, func(i, j int) bool {
			return activities[i].Time.After(activities[j].Time)
		})
		if len(activities) > limit {
			activities = activities[:limit]
		}

		c.JSON(http.StatusOK, gin.H{"items": activities})
	})

	// GET /sales-trend — 月度销售趋势
	rg.GET("/sales-trend", func(c *gin.Context) {
		months := 6
		if m := c.Query("months"); m != "" {
			if n, err := strconv.Atoi(m); err == nil && n > 0 {
				months = n
			}
		}
		var results []struct {
			Month string  `json:"month"`
			Total float64 `json:"total"`
			Count int     `json:"count"`
		}
		db.Raw(`
			SELECT TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') as month,
				COALESCE(SUM(total), 0) as total,
				COUNT(*) as count
			FROM erp_sales_orders
			WHERE created_at >= DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '1 month' * ?
				AND status NOT IN ('cancelled', 'draft')
			GROUP BY DATE_TRUNC('month', created_at)
			ORDER BY month
		`, months).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results})
	})

	// POST /mrp-run — 执行MRP计算
	rg.POST("/mrp-run", func(c *gin.Context) {
		var req map[string]any
		c.ShouldBindJSON(&req)
		if req == nil {
			req = map[string]any{}
		}
		if _, ok := req["demand_source"]; !ok {
			req["demand_source"] = "so" // default
		}

		handler, ok := commandHandlers["run_mrp"]
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "run_mrp command not found"})
			return
		}
		result, err := handler(db, nil, "", "", req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	// GET /pipeline-summary — 销售管道汇总（按状态聚合）
	rg.GET("/pipeline-summary", func(c *gin.Context) {
		var results []struct {
			Status string  `json:"status"`
			Count  int64   `json:"count"`
			Total  float64 `json:"total"`
		}
		db.Table("erp_sales_orders").
			Select("status, COUNT(*) as count, COALESCE(SUM(total), 0) as total").
			Where("status NOT IN ?", []string{"cancelled"}).
			Group("status").Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results})
	})

	// GET /stock-alerts — 库存预警（低于安全库存）
	rg.GET("/stock-alerts", func(c *gin.Context) {
		var results []struct {
			MaterialID  string  `json:"material_id"`
			Available   float64 `json:"available"`
			SafetyStock float64 `json:"safety_stock"`
			Shortfall   float64 `json:"shortfall"`
		}
		// Safety stock = 100 (hardcoded threshold since we don't have safety_stock field)
		db.Raw(`
			SELECT material_id,
				COALESCE(SUM(quantity - reserved_qty), 0) as available,
				100 as safety_stock,
				GREATEST(100 - COALESCE(SUM(quantity - reserved_qty), 0), 0) as shortfall
			FROM erp_inventory
			GROUP BY material_id
			HAVING COALESCE(SUM(quantity - reserved_qty), 0) < 100
		`).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"items": results})
	})

	// ─── Cash Flow ───

	// GET /cash-flow — 现金流量表
	rg.GET("/cash-flow", func(c *gin.Context) {
		period := c.Query("period")
		if period == "" {
			period = time.Now().Format("2006-01")
		}

		// Operating: sales receipts - purchase payments
		var salesReceipts float64
		db.Table("erp_receipts").
			Where("status = 'confirmed'").
			Where("TO_CHAR(received_date, 'YYYY-MM') = ?", period).
			Select("COALESCE(SUM(amount), 0)").Row().Scan(&salesReceipts)

		// For AP payments, estimate from journal entries
		var apPayments float64
		db.Raw(`
			SELECT COALESCE(SUM(jl.debit), 0) FROM erp_journal_lines jl
			JOIN erp_journal_entries je ON je.id = jl.entry_id
			JOIN erp_accounts a ON a.id = jl.account_id
			WHERE je.status = 'posted' AND je.period = ? AND a.code = '2202'
		`, period).Scan(&apPayments)

		operatingNet := salesReceipts - apPayments

		// Cash balance from bank account (1002)
		var cashBalance float64
		db.Raw(`
			SELECT COALESCE(SUM(jl.debit - jl.credit), 0) FROM erp_journal_lines jl
			JOIN erp_journal_entries je ON je.id = jl.entry_id
			JOIN erp_accounts a ON a.id = jl.account_id
			WHERE je.status = 'posted' AND je.period <= ? AND a.code = '1002'
		`, period).Scan(&cashBalance)

		c.JSON(http.StatusOK, gin.H{
			"period":         period,
			"operating_in":   salesReceipts,
			"operating_out":  apPayments,
			"operating_net":  operatingNet,
			"investing_net":  0, // placeholder
			"financing_net":  0, // placeholder
			"net_change":     operatingNet,
			"ending_balance": cashBalance,
		})
	})

	// ─── Audit Log ───

	// GET /audit-log/:entity/:id — 实体变更记录
	rg.GET("/audit-log/:entity/:id", func(c *gin.Context) {
		entityType := c.Param("entity")
		entityID := c.Param("id")
		var logs []ErpAuditLog
		db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
			Order("created_at DESC").Limit(50).Find(&logs)
		c.JSON(http.StatusOK, gin.H{"items": logs})
	})

	_ = strings.TrimSpace // suppress unused import
}
