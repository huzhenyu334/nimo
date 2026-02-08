package service

import (
	"context"
	"fmt"
	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/plm/repository"
	"time"

	"github.com/google/uuid"
)

type ProjectBOMService struct {
	bomRepo         *repository.ProjectBOMRepository
	projectRepo     *repository.ProjectRepository
	deliverableRepo *repository.DeliverableRepository
}

func NewProjectBOMService(bomRepo *repository.ProjectBOMRepository, projectRepo *repository.ProjectRepository, deliverableRepo *repository.DeliverableRepository) *ProjectBOMService {
	return &ProjectBOMService{
		bomRepo:         bomRepo,
		projectRepo:     projectRepo,
		deliverableRepo: deliverableRepo,
	}
}

// CreateBOM 创建BOM（草稿状态）
func (s *ProjectBOMService) CreateBOM(ctx context.Context, projectID string, input *CreateBOMInput, createdBy string) (*entity.ProjectBOM, error) {
	bom := &entity.ProjectBOM{
		ID:          uuid.New().String()[:32],
		ProjectID:   projectID,
		PhaseID:     input.PhaseID,
		BOMType:     input.BOMType,
		Version:     input.Version,
		Name:        input.Name,
		Status:      "draft",
		Description: input.Description,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if bom.Version == "" {
		bom.Version = "v1.0"
	}

	if err := s.bomRepo.Create(ctx, bom); err != nil {
		return nil, fmt.Errorf("create bom: %w", err)
	}

	return bom, nil
}

// GetBOM 获取BOM详情（含行项）
func (s *ProjectBOMService) GetBOM(ctx context.Context, id string) (*entity.ProjectBOM, error) {
	return s.bomRepo.FindByID(ctx, id)
}

// ListBOMs 获取项目BOM列表
func (s *ProjectBOMService) ListBOMs(ctx context.Context, projectID, bomType, status string) ([]entity.ProjectBOM, error) {
	return s.bomRepo.ListByProject(ctx, projectID, bomType, status)
}

// UpdateBOM 更新BOM基本信息（仅草稿状态可改）
func (s *ProjectBOMService) UpdateBOM(ctx context.Context, id string, input *UpdateBOMInput) (*entity.ProjectBOM, error) {
	bom, err := s.bomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bom not found: %w", err)
	}

	if bom.Status != "draft" && bom.Status != "rejected" {
		return nil, fmt.Errorf("只有草稿或被驳回的BOM才能编辑")
	}

	if input.Name != "" {
		bom.Name = input.Name
	}
	if input.Description != "" {
		bom.Description = input.Description
	}
	if input.Version != "" {
		bom.Version = input.Version
	}

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("update bom: %w", err)
	}
	return bom, nil
}

// SubmitBOM 提交BOM审批
func (s *ProjectBOMService) SubmitBOM(ctx context.Context, id, submitterID string) (*entity.ProjectBOM, error) {
	bom, err := s.bomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bom not found: %w", err)
	}

	if bom.Status != "draft" && bom.Status != "rejected" {
		return nil, fmt.Errorf("只有草稿或被驳回的BOM才能提交审批")
	}

	// 检查是否有行项
	count, _ := s.bomRepo.CountItems(ctx, id)
	if count == 0 {
		return nil, fmt.Errorf("BOM没有物料行项，无法提交")
	}

	now := time.Now()
	bom.Status = "pending_review"
	bom.SubmittedBy = &submitterID
	bom.SubmittedAt = &now

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("submit bom: %w", err)
	}
	return bom, nil
}

// ApproveBOM 审批通过BOM
func (s *ProjectBOMService) ApproveBOM(ctx context.Context, id, reviewerID, comment string) (*entity.ProjectBOM, error) {
	bom, err := s.bomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bom not found: %w", err)
	}

	if bom.Status != "pending_review" {
		return nil, fmt.Errorf("只有待审批的BOM才能审批")
	}

	now := time.Now()
	bom.Status = "published"
	bom.ReviewedBy = &reviewerID
	bom.ReviewedAt = &now
	bom.ReviewComment = comment
	bom.ApprovedBy = &reviewerID
	bom.ApprovedAt = &now

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("approve bom: %w", err)
	}
	return bom, nil
}

// RejectBOM 驳回BOM
func (s *ProjectBOMService) RejectBOM(ctx context.Context, id, reviewerID, comment string) (*entity.ProjectBOM, error) {
	bom, err := s.bomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bom not found: %w", err)
	}

	if bom.Status != "pending_review" {
		return nil, fmt.Errorf("只有待审批的BOM才能驳回")
	}

	now := time.Now()
	bom.Status = "rejected"
	bom.ReviewedBy = &reviewerID
	bom.ReviewedAt = &now
	bom.ReviewComment = comment

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("reject bom: %w", err)
	}
	return bom, nil
}

// FreezeBOM 冻结BOM（阶段门评审通过后）
func (s *ProjectBOMService) FreezeBOM(ctx context.Context, id, frozenByID string) (*entity.ProjectBOM, error) {
	bom, err := s.bomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bom not found: %w", err)
	}

	if bom.Status != "published" {
		return nil, fmt.Errorf("只有已发布的BOM才能冻结")
	}

	now := time.Now()
	bom.Status = "frozen"
	bom.FrozenAt = &now
	bom.FrozenBy = &frozenByID

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("freeze bom: %w", err)
	}
	return bom, nil
}

// AddItem 添加BOM行项
func (s *ProjectBOMService) AddItem(ctx context.Context, bomID string, input *BOMItemInput) (*entity.ProjectBOMItem, error) {
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return nil, fmt.Errorf("bom not found: %w", err)
	}
	if bom.Status != "draft" && bom.Status != "rejected" {
		return nil, fmt.Errorf("只有草稿状态的BOM才能添加物料")
	}

	item := &entity.ProjectBOMItem{
		ID:              uuid.New().String()[:32],
		BOMID:           bomID,
		ItemNumber:      input.ItemNumber,
		ParentItemID:    input.ParentItemID,
		Level:           input.Level,
		MaterialID:      input.MaterialID,
		Category:        input.Category,
		Name:            input.Name,
		Specification:   input.Specification,
		Quantity:        input.Quantity,
		Unit:            input.Unit,
		Reference:       input.Reference,
		Manufacturer:    input.Manufacturer,
		ManufacturerPN:  input.ManufacturerPN,
		Supplier:        input.Supplier,
		SupplierPN:      input.SupplierPN,
		UnitPrice:       input.UnitPrice,
		LeadTimeDays:    input.LeadTimeDays,
		ProcurementType: input.ProcurementType,
		MOQ:             input.MOQ,
		LifecycleStatus: input.LifecycleStatus,
		IsCritical:      input.IsCritical,
		IsAlternative:   input.IsAlternative,
		Notes:           input.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if item.Unit == "" {
		item.Unit = "pcs"
	}
	if item.ProcurementType == "" {
		item.ProcurementType = "buy"
	}
	if item.LifecycleStatus == "" {
		item.LifecycleStatus = "active"
	}

	// 计算小计
	if input.UnitPrice != nil {
		extCost := input.Quantity * *input.UnitPrice
		item.ExtendedCost = &extCost
	}

	if err := s.bomRepo.CreateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("create bom item: %w", err)
	}

	// 更新BOM统计
	s.updateBOMCost(ctx, bomID)

	return item, nil
}

// BatchAddItems 批量添加BOM行项
func (s *ProjectBOMService) BatchAddItems(ctx context.Context, bomID string, items []BOMItemInput) (int, error) {
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return 0, fmt.Errorf("bom not found: %w", err)
	}
	if bom.Status != "draft" && bom.Status != "rejected" {
		return 0, fmt.Errorf("只有草稿状态的BOM才能添加物料")
	}

	var entities []entity.ProjectBOMItem
	for i, input := range items {
		item := entity.ProjectBOMItem{
			ID:             uuid.New().String()[:32],
			BOMID:          bomID,
			ItemNumber:     i + 1,
			Category:       input.Category,
			Name:           input.Name,
			Specification:  input.Specification,
			Quantity:        input.Quantity,
			Unit:           input.Unit,
			Reference:      input.Reference,
			Manufacturer:   input.Manufacturer,
			ManufacturerPN: input.ManufacturerPN,
			Supplier:       input.Supplier,
			UnitPrice:      input.UnitPrice,
			LeadTimeDays:   input.LeadTimeDays,
			IsCritical:     input.IsCritical,
			IsAlternative:  input.IsAlternative,
			Notes:          input.Notes,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if item.Unit == "" {
			item.Unit = "pcs"
		}
		if input.MaterialID != nil {
			item.MaterialID = input.MaterialID
		}
		entities = append(entities, item)
	}

	if err := s.bomRepo.BatchCreateItems(ctx, entities); err != nil {
		return 0, fmt.Errorf("batch create items: %w", err)
	}

	count, _ := s.bomRepo.CountItems(ctx, bomID)
	s.bomRepo.DB().Model(&entity.ProjectBOM{}).Where("id = ?", bomID).Update("total_items", count)

	return len(entities), nil
}

// DeleteItem 删除BOM行项
func (s *ProjectBOMService) DeleteItem(ctx context.Context, bomID, itemID string) error {
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return fmt.Errorf("bom not found: %w", err)
	}
	if bom.Status != "draft" && bom.Status != "rejected" {
		return fmt.Errorf("只有草稿状态的BOM才能删除物料")
	}

	if err := s.bomRepo.DeleteItem(ctx, itemID); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	s.updateBOMCost(ctx, bomID)

	return nil
}

// UpdateItem 更新单个BOM行项
func (s *ProjectBOMService) UpdateItem(ctx context.Context, bomID, itemID string, input *BOMItemInput) (*entity.ProjectBOMItem, error) {
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return nil, fmt.Errorf("bom not found: %w", err)
	}
	if bom.Status != "draft" && bom.Status != "rejected" {
		return nil, fmt.Errorf("只有草稿状态的BOM才能编辑物料")
	}

	item, err := s.bomRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}
	if item.BOMID != bomID {
		return nil, fmt.Errorf("item does not belong to this BOM")
	}

	if input.Name != "" {
		item.Name = input.Name
	}
	if input.Category != "" {
		item.Category = input.Category
	}
	if input.Specification != "" {
		item.Specification = input.Specification
	}
	if input.MaterialID != nil {
		item.MaterialID = input.MaterialID
	}
	item.Quantity = input.Quantity
	if input.Unit != "" {
		item.Unit = input.Unit
	}
	item.Reference = input.Reference
	item.Manufacturer = input.Manufacturer
	item.ManufacturerPN = input.ManufacturerPN
	item.Supplier = input.Supplier
	item.SupplierPN = input.SupplierPN
	item.UnitPrice = input.UnitPrice
	item.LeadTimeDays = input.LeadTimeDays
	item.IsCritical = input.IsCritical
	item.IsAlternative = input.IsAlternative
	item.Notes = input.Notes
	item.ProcurementType = input.ProcurementType
	item.MOQ = input.MOQ
	item.LifecycleStatus = input.LifecycleStatus
	item.ParentItemID = input.ParentItemID
	item.Level = input.Level

	// 计算小计
	if input.UnitPrice != nil {
		extCost := input.Quantity * *input.UnitPrice
		item.ExtendedCost = &extCost
	}

	item.UpdatedAt = time.Now()

	if err := s.bomRepo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("update item: %w", err)
	}

	// 更新BOM总成本
	s.updateBOMCost(ctx, bomID)

	return item, nil
}

// ReorderItems 拖拽排序
func (s *ProjectBOMService) ReorderItems(ctx context.Context, bomID string, itemIDs []string) error {
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return fmt.Errorf("bom not found: %w", err)
	}
	if bom.Status != "draft" && bom.Status != "rejected" {
		return fmt.Errorf("只有草稿状态的BOM才能排序")
	}

	for i, id := range itemIDs {
		s.bomRepo.DB().Model(&entity.ProjectBOMItem{}).Where("id = ? AND bom_id = ?", id, bomID).Update("item_number", i+1)
	}
	return nil
}

// updateBOMCost 更新BOM总成本统计
func (s *ProjectBOMService) updateBOMCost(ctx context.Context, bomID string) {
	var totalCost float64
	s.bomRepo.DB().Model(&entity.ProjectBOMItem{}).
		Where("bom_id = ?", bomID).
		Select("COALESCE(SUM(extended_cost), 0)").
		Scan(&totalCost)
	count, _ := s.bomRepo.CountItems(ctx, bomID)
	s.bomRepo.DB().Model(&entity.ProjectBOM{}).Where("id = ?", bomID).
		Updates(map[string]interface{}{"estimated_cost": totalCost, "total_items": count})
}

// ---- Input DTOs ----

type CreateBOMInput struct {
	PhaseID     *string `json:"phase_id"`
	BOMType     string  `json:"bom_type" binding:"required"`
	Version     string  `json:"version"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
}

type UpdateBOMInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type BOMItemInput struct {
	MaterialID      *string  `json:"material_id"`
	ParentItemID    *string  `json:"parent_item_id"`
	Level           int      `json:"level"`
	Category        string   `json:"category"`
	Name            string   `json:"name" binding:"required"`
	Specification   string   `json:"specification"`
	Quantity        float64  `json:"quantity"`
	Unit            string   `json:"unit"`
	Reference       string   `json:"reference"`
	Manufacturer    string   `json:"manufacturer"`
	ManufacturerPN  string   `json:"manufacturer_pn"`
	Supplier        string   `json:"supplier"`
	SupplierPN      string   `json:"supplier_pn"`
	UnitPrice       *float64 `json:"unit_price"`
	LeadTimeDays    *int     `json:"lead_time_days"`
	ProcurementType string   `json:"procurement_type"`
	MOQ             *int     `json:"moq"`
	LifecycleStatus string   `json:"lifecycle_status"`
	IsCritical      bool     `json:"is_critical"`
	IsAlternative   bool     `json:"is_alternative"`
	Notes           string   `json:"notes"`
	ItemNumber      int      `json:"item_number"`
}

type ReorderItemsInput struct {
	ItemIDs []string `json:"item_ids" binding:"required"`
}
