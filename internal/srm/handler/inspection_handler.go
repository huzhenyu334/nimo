package handler

import (
	"github.com/bitfantasy/nimo/internal/srm/service"
	"github.com/gin-gonic/gin"
)

// InspectionHandler 检验处理器
type InspectionHandler struct {
	svc *service.InspectionService
}

func NewInspectionHandler(svc *service.InspectionService) *InspectionHandler {
	return &InspectionHandler{svc: svc}
}

// CreateInspection 创建检验单
// POST /api/v1/srm/inspections
func (h *InspectionHandler) CreateInspection(c *gin.Context) {
	var req struct {
		POID         string  `json:"po_id"`
		POItemID     string  `json:"po_item_id"`
		SupplierID   string  `json:"supplier_id"`
		MaterialID   string  `json:"material_id"`
		MaterialCode string  `json:"material_code" binding:"required"`
		MaterialName string  `json:"material_name" binding:"required"`
		Quantity     float64 `json:"quantity"`
		SampleQty    int     `json:"sample_qty"`
		Notes        string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	inspection, err := h.svc.CreateInspectionFromPOItem(
		c.Request.Context(),
		req.POID, req.POItemID, req.SupplierID,
		req.MaterialID, req.MaterialCode, req.MaterialName,
		req.Quantity,
	)
	if err != nil {
		InternalError(c, "创建检验失败: "+err.Error())
		return
	}

	// 补充sample_qty和notes
	if req.SampleQty > 0 || req.Notes != "" {
		updateReq := service.UpdateInspectionRequest{
			SampleQty: &req.SampleQty,
			Notes:     &req.Notes,
		}
		if updated, err := h.svc.UpdateInspection(c.Request.Context(), inspection.ID, &updateReq); err == nil {
			inspection = updated
		}
	}

	Success(c, inspection)
}

// ListInspections 检验列表
// GET /api/v1/srm/inspections?supplier_id=xxx&status=xxx&result=xxx&po_id=xxx
func (h *InspectionHandler) ListInspections(c *gin.Context) {
	page, pageSize := GetPagination(c)
	filters := map[string]string{
		"supplier_id": c.Query("supplier_id"),
		"status":      c.Query("status"),
		"result":      c.Query("result"),
		"po_id":       c.Query("po_id"),
	}

	items, total, err := h.svc.ListInspections(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		InternalError(c, "获取检验列表失败: "+err.Error())
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

// GetInspection 检验详情
// GET /api/v1/srm/inspections/:id
func (h *InspectionHandler) GetInspection(c *gin.Context) {
	id := c.Param("id")
	inspection, err := h.svc.GetInspection(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "检验记录不存在")
		return
	}
	Success(c, inspection)
}

// UpdateInspection 更新检验
// PUT /api/v1/srm/inspections/:id
func (h *InspectionHandler) UpdateInspection(c *gin.Context) {
	id := c.Param("id")
	var req service.UpdateInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	inspection, err := h.svc.UpdateInspection(c.Request.Context(), id, &req)
	if err != nil {
		InternalError(c, "更新检验失败: "+err.Error())
		return
	}

	Success(c, inspection)
}

// CompleteInspection 完成检验
// POST /api/v1/srm/inspections/:id/complete
func (h *InspectionHandler) CompleteInspection(c *gin.Context) {
	id := c.Param("id")
	userID := GetUserID(c)

	var req service.CompleteInspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	inspection, err := h.svc.CompleteInspection(c.Request.Context(), id, userID, &req)
	if err != nil {
		InternalError(c, "完成检验失败: "+err.Error())
		return
	}

	Success(c, inspection)
}
