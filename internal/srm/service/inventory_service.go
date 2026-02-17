package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bitfantasy/nimo/internal/srm/entity"
	"github.com/bitfantasy/nimo/internal/srm/repository"
	"github.com/google/uuid"
)

// InventoryService 库存服务
type InventoryService struct {
	repo *repository.InventoryRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

// ListInventory 库存列表
func (s *InventoryService) ListInventory(ctx context.Context, page, pageSize int, filters map[string]string) ([]entity.InventoryRecord, int64, error) {
	return s.repo.FindAll(ctx, page, pageSize, filters)
}

// GetInventory 库存详情
func (s *InventoryService) GetInventory(ctx context.Context, id string) (*entity.InventoryRecord, error) {
	return s.repo.FindByID(ctx, id)
}

// GetTransactions 库存流水
func (s *InventoryService) GetTransactions(ctx context.Context, inventoryID string, page, pageSize int) ([]entity.InventoryTransaction, int64, error) {
	return s.repo.FindTransactions(ctx, inventoryID, page, pageSize)
}

// InRequest 入库请求
type InRequest struct {
	MaterialName string  `json:"material_name" binding:"required"`
	MaterialCode string  `json:"material_code"`
	MPN          string  `json:"mpn"`
	SupplierID   *string `json:"supplier_id"`
	Quantity     float64 `json:"quantity" binding:"required"`
	Unit         string  `json:"unit"`
	Warehouse    string  `json:"warehouse"`
	Operator     string  `json:"operator"`
	Notes        string  `json:"notes"`
}

// StockIn 入库
func (s *InventoryService) StockIn(ctx context.Context, req *InRequest) (*entity.InventoryRecord, error) {
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("入库数量必须大于0")
	}

	record, err := s.findOrCreateRecord(ctx, req.MaterialName, req.MaterialCode, req.MPN, req.SupplierID, req.Unit, req.Warehouse)
	if err != nil {
		return nil, err
	}

	record.Quantity += req.Quantity
	now := time.Now()
	record.LastInDate = &now
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}

	tx := &entity.InventoryTransaction{
		ID:            uuid.New().String()[:32],
		InventoryID:   record.ID,
		Type:          entity.InventoryTxTypeIn,
		Quantity:      req.Quantity,
		ReferenceType: entity.InventoryRefManual,
		Operator:      req.Operator,
		Notes:         req.Notes,
	}
	s.repo.CreateTransaction(ctx, tx)

	return record, nil
}

// OutRequest 出库请求
type OutRequest struct {
	InventoryID string  `json:"inventory_id" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required"`
	Operator    string  `json:"operator"`
	Notes       string  `json:"notes"`
}

// StockOut 出库
func (s *InventoryService) StockOut(ctx context.Context, req *OutRequest) (*entity.InventoryRecord, error) {
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("出库数量必须大于0")
	}

	record, err := s.repo.FindByID(ctx, req.InventoryID)
	if err != nil {
		return nil, err
	}

	if record.Quantity < req.Quantity {
		return nil, fmt.Errorf("库存不足，当前库存 %.2f", record.Quantity)
	}

	record.Quantity -= req.Quantity
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}

	tx := &entity.InventoryTransaction{
		ID:            uuid.New().String()[:32],
		InventoryID:   record.ID,
		Type:          entity.InventoryTxTypeOut,
		Quantity:      -req.Quantity,
		ReferenceType: entity.InventoryRefManual,
		Operator:      req.Operator,
		Notes:         req.Notes,
	}
	s.repo.CreateTransaction(ctx, tx)

	return record, nil
}

// AdjustRequest 调整请求
type AdjustRequest struct {
	InventoryID string  `json:"inventory_id" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required"` // 调整后的数量
	Operator    string  `json:"operator"`
	Notes       string  `json:"notes"`
}

// StockAdjust 库存调整
func (s *InventoryService) StockAdjust(ctx context.Context, req *AdjustRequest) (*entity.InventoryRecord, error) {
	record, err := s.repo.FindByID(ctx, req.InventoryID)
	if err != nil {
		return nil, err
	}

	diff := req.Quantity - record.Quantity
	record.Quantity = req.Quantity
	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}

	tx := &entity.InventoryTransaction{
		ID:            uuid.New().String()[:32],
		InventoryID:   record.ID,
		Type:          entity.InventoryTxTypeAdjust,
		Quantity:      diff,
		ReferenceType: entity.InventoryRefAdjust,
		Operator:      req.Operator,
		Notes:         fmt.Sprintf("调整: %.2f → %.2f (%s)", record.Quantity-diff, req.Quantity, req.Notes),
	}
	s.repo.CreateTransaction(ctx, tx)

	return record, nil
}

// StockInFromInspection 质检通过后自动入库
func (s *InventoryService) StockInFromInspection(ctx context.Context, inspectionID string, materialName, materialCode string, supplierID *string, qty float64, unit string) error {
	record, err := s.findOrCreateRecord(ctx, materialName, materialCode, "", supplierID, unit, "")
	if err != nil {
		return err
	}

	record.Quantity += qty
	now := time.Now()
	record.LastInDate = &now
	if err := s.repo.Update(ctx, record); err != nil {
		return err
	}

	tx := &entity.InventoryTransaction{
		ID:            uuid.New().String()[:32],
		InventoryID:   record.ID,
		Type:          entity.InventoryTxTypeIn,
		Quantity:      qty,
		ReferenceType: entity.InventoryRefInspection,
		ReferenceID:   inspectionID,
		Operator:      "system",
		Notes:         "质检通过自动入库",
	}
	return s.repo.CreateTransaction(ctx, tx)
}

func (s *InventoryService) findOrCreateRecord(ctx context.Context, name, code, mpn string, supplierID *string, unit, warehouse string) (*entity.InventoryRecord, error) {
	if code != "" {
		existing, err := s.repo.FindByMaterialAndSupplier(ctx, code, supplierID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	record := &entity.InventoryRecord{
		ID:           uuid.New().String()[:32],
		MaterialName: name,
		MaterialCode: code,
		MPN:          mpn,
		SupplierID:   supplierID,
		Unit:         unit,
		Warehouse:    warehouse,
	}
	if record.Unit == "" {
		record.Unit = "pcs"
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}
