package handler

import (
	"github.com/bitfantasy/nimo/internal/srm/service"
	"github.com/gin-gonic/gin"
)

// InventoryHandler 库存处理器
type InventoryHandler struct {
	svc *service.InventoryService
}

func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

// ListInventory 库存列表
// GET /api/v1/srm/inventory
func (h *InventoryHandler) ListInventory(c *gin.Context) {
	page, pageSize := GetPagination(c)
	filters := map[string]string{
		"search":      c.Query("search"),
		"warehouse":   c.Query("warehouse"),
		"supplier_id": c.Query("supplier_id"),
		"low_stock":   c.Query("low_stock"),
	}

	items, total, err := h.svc.ListInventory(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		InternalError(c, "获取库存列表失败: "+err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	Success(c, ListResponse{
		Items: items,
		Pagination: &Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      int(total),
			TotalPages: totalPages,
		},
	})
}

// GetTransactions 库存流水
// GET /api/v1/srm/inventory/:id/transactions
func (h *InventoryHandler) GetTransactions(c *gin.Context) {
	id := c.Param("id")
	page, pageSize := GetPagination(c)

	items, total, err := h.svc.GetTransactions(c.Request.Context(), id, page, pageSize)
	if err != nil {
		InternalError(c, "获取库存流水失败: "+err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	Success(c, ListResponse{
		Items: items,
		Pagination: &Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      int(total),
			TotalPages: totalPages,
		},
	})
}

// StockIn 入库
// POST /api/v1/srm/inventory/in
func (h *InventoryHandler) StockIn(c *gin.Context) {
	var req service.InRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Operator = GetUserID(c)

	record, err := h.svc.StockIn(c.Request.Context(), &req)
	if err != nil {
		InternalError(c, "入库失败: "+err.Error())
		return
	}
	Success(c, record)
}

// StockOut 出库
// POST /api/v1/srm/inventory/out
func (h *InventoryHandler) StockOut(c *gin.Context) {
	var req service.OutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Operator = GetUserID(c)

	record, err := h.svc.StockOut(c.Request.Context(), &req)
	if err != nil {
		BadRequest(c, "出库失败: "+err.Error())
		return
	}
	Success(c, record)
}

// StockAdjust 库存调整
// POST /api/v1/srm/inventory/adjust
func (h *InventoryHandler) StockAdjust(c *gin.Context) {
	var req service.AdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Operator = GetUserID(c)

	record, err := h.svc.StockAdjust(c.Request.Context(), &req)
	if err != nil {
		BadRequest(c, "库存调整失败: "+err.Error())
		return
	}
	Success(c, record)
}
