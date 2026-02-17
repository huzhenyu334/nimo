package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/plm/repository"
	"github.com/google/uuid"
)

type BOMECNService struct {
	bomRepo      *repository.ProjectBOMRepository
	draftRepo    *repository.BOMDraftRepository
	ecnRepo      *repository.BOMECNRepository
	bomItemRepo  *repository.ProjectBOMRepository // 用于访问BOM items
}

func NewBOMECNService(
	bomRepo *repository.ProjectBOMRepository,
	draftRepo *repository.BOMDraftRepository,
	ecnRepo *repository.BOMECNRepository,
) *BOMECNService {
	return &BOMECNService{
		bomRepo:     bomRepo,
		draftRepo:   draftRepo,
		ecnRepo:     ecnRepo,
		bomItemRepo: bomRepo,
	}
}

// DraftData BOM草稿数据结构
type DraftData struct {
	Items       []entity.ProjectBOMItem `json:"items"`
	Name        string                  `json:"name,omitempty"`
	Description string                  `json:"description,omitempty"`
}

// SaveDraft 保存或更新草稿
func (s *BOMECNService) SaveDraft(ctx context.Context, bomID string, draftData *DraftData, userID string) (*entity.BOMDraft, error) {
	// 验证BOM存在且状态允许编辑
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return nil, fmt.Errorf("BOM not found: %w", err)
	}

	if bom.Status != "released" && bom.Status != "editing" && bom.Status != "frozen" {
		return nil, fmt.Errorf("只有已发布或冻结的BOM才能编辑")
	}

	// 序列化草稿数据为JSONB
	var draftDataMap entity.JSONB
	dataBytes, err := json.Marshal(draftData)
	if err != nil {
		return nil, fmt.Errorf("marshal draft data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &draftDataMap); err != nil {
		return nil, fmt.Errorf("convert to JSONB: %w", err)
	}

	draft := &entity.BOMDraft{
		ID:        uuid.New().String()[:32],
		BOMID:     bomID,
		DraftData: draftDataMap,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.draftRepo.Upsert(ctx, draft); err != nil {
		return nil, fmt.Errorf("save draft: %w", err)
	}

	// 如果BOM不是editing状态，更新为editing
	if bom.Status != "editing" {
		bom.Status = "editing"
		bom.UpdatedAt = time.Now()
		if err := s.bomRepo.Update(ctx, bom); err != nil {
			return nil, fmt.Errorf("update bom status: %w", err)
		}
	}

	return draft, nil
}

// GetDraft 获取草稿
func (s *BOMECNService) GetDraft(ctx context.Context, bomID string) (*entity.BOMDraft, error) {
	draft, err := s.draftRepo.FindByBOMID(ctx, bomID)
	if err != nil {
		return nil, err
	}
	return draft, nil
}

// DiscardDraft 撤销编辑，删除草稿，BOM状态回到released
func (s *BOMECNService) DiscardDraft(ctx context.Context, bomID string) error {
	// 删除草稿
	if err := s.draftRepo.Delete(ctx, bomID); err != nil {
		return fmt.Errorf("delete draft: %w", err)
	}

	// BOM状态回到released
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return fmt.Errorf("find bom: %w", err)
	}

	if bom.FrozenAt != nil {
		bom.Status = "frozen"
	} else {
		bom.Status = "released"
	}
	bom.UpdatedAt = time.Now()

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return fmt.Errorf("update bom status: %w", err)
	}

	return nil
}

// StartEditing BOM状态改为editing
func (s *BOMECNService) StartEditing(ctx context.Context, bomID string) (*entity.ProjectBOM, error) {
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return nil, fmt.Errorf("BOM not found: %w", err)
	}

	if bom.Status != "released" && bom.Status != "frozen" {
		return nil, fmt.Errorf("只有已发布或冻结的BOM才能开始编辑")
	}

	bom.Status = "editing"
	bom.UpdatedAt = time.Now()

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("update bom: %w", err)
	}

	return bom, nil
}

// SubmitECN 提交ECN，计算diff，创建ECN记录，BOM状态改为ecn_pending
func (s *BOMECNService) SubmitECN(ctx context.Context, bomID string, title string, userID string) (*entity.BOMECN, error) {
	// 验证BOM处于editing状态
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return nil, fmt.Errorf("BOM not found: %w", err)
	}

	if bom.Status != "editing" {
		return nil, fmt.Errorf("只有编辑中的BOM才能提交ECN")
	}

	// 获取草稿数据
	draft, err := s.draftRepo.FindByBOMID(ctx, bomID)
	if err != nil {
		return nil, fmt.Errorf("未找到草稿数据")
	}

	// 计算变更diff
	var draftData DraftData
	draftBytes, err := json.Marshal(draft.DraftData)
	if err != nil {
		return nil, fmt.Errorf("marshal draft data: %w", err)
	}
	if err := json.Unmarshal(draftBytes, &draftData); err != nil {
		return nil, fmt.Errorf("parse draft data: %w", err)
	}

	changeSummary := s.calculateDiff(bom.Items, draftData.Items)
	var changeSummaryMap entity.JSONB
	summaryBytes, err := json.Marshal(changeSummary)
	if err != nil {
		return nil, fmt.Errorf("marshal change summary: %w", err)
	}
	if err := json.Unmarshal(summaryBytes, &changeSummaryMap); err != nil {
		return nil, fmt.Errorf("convert summary to JSONB: %w", err)
	}

	// 生成ECN编号
	ecnNumber, err := s.ecnRepo.GenerateECNNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate ecn number: %w", err)
	}

	// 创建ECN记录
	ecn := &entity.BOMECN{
		ID:            uuid.New().String()[:32],
		ECNNumber:     ecnNumber,
		BOMID:         bomID,
		Title:         title,
		Description:   fmt.Sprintf("BOM变更申请: %s", bom.Name),
		ChangeSummary: changeSummaryMap,
		Status:        entity.BOMECNStatusPending,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.ecnRepo.Create(ctx, ecn); err != nil {
		return nil, fmt.Errorf("create ecn: %w", err)
	}

	// BOM状态改为ecn_pending
	bom.Status = "ecn_pending"
	bom.UpdatedAt = time.Now()
	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("update bom status: %w", err)
	}

	return ecn, nil
}

// ApproveECN 审批通过，应用变更到BOM，状态回released
func (s *BOMECNService) ApproveECN(ctx context.Context, ecnID string, approverID string) (*entity.BOMECN, error) {
	ecn, err := s.ecnRepo.FindByID(ctx, ecnID)
	if err != nil {
		return nil, fmt.Errorf("ECN not found: %w", err)
	}

	if ecn.Status != entity.BOMECNStatusPending {
		return nil, fmt.Errorf("只有待审批的ECN才能审批")
	}

	// 获取草稿数据并应用到BOM
	draft, err := s.draftRepo.FindByBOMID(ctx, ecn.BOMID)
	if err == nil {
		var draftData DraftData
		draftBytes, err := json.Marshal(draft.DraftData)
		if err == nil {
			if err := json.Unmarshal(draftBytes, &draftData); err == nil {
				// 应用变更
				if err := s.applyDraftToBOM(ctx, ecn.BOMID, &draftData); err != nil {
					return nil, fmt.Errorf("apply changes: %w", err)
				}
			}
		}
	}

	// 更新ECN状态
	now := time.Now()
	ecn.Status = entity.BOMECNStatusApproved
	ecn.ApprovedBy = &approverID
	ecn.ApprovedAt = &now
	ecn.UpdatedAt = now

	if err := s.ecnRepo.Update(ctx, ecn); err != nil {
		return nil, fmt.Errorf("update ecn: %w", err)
	}

	// BOM状态回released，版本号+1
	bom, err := s.bomRepo.FindByID(ctx, ecn.BOMID)
	if err != nil {
		return nil, fmt.Errorf("find bom: %w", err)
	}

	bom.VersionMinor++
	bom.Version = fmt.Sprintf("v%d.%d", bom.VersionMajor, bom.VersionMinor)
	if bom.FrozenAt != nil {
		bom.Status = "frozen"
	} else {
		bom.Status = "released"
	}
	bom.UpdatedAt = time.Now()

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("update bom: %w", err)
	}

	// 删除草稿
	s.draftRepo.Delete(ctx, ecn.BOMID)

	return ecn, nil
}

// RejectECN 审批拒绝，BOM状态回released
func (s *BOMECNService) RejectECN(ctx context.Context, ecnID string, rejecterID string, note string) (*entity.BOMECN, error) {
	ecn, err := s.ecnRepo.FindByID(ctx, ecnID)
	if err != nil {
		return nil, fmt.Errorf("ECN not found: %w", err)
	}

	if ecn.Status != entity.BOMECNStatusPending {
		return nil, fmt.Errorf("只有待审批的ECN才能拒绝")
	}

	// 更新ECN状态
	now := time.Now()
	ecn.Status = entity.BOMECNStatusRejected
	ecn.RejectedBy = &rejecterID
	ecn.RejectedAt = &now
	ecn.RejectionNote = note
	ecn.UpdatedAt = now

	if err := s.ecnRepo.Update(ctx, ecn); err != nil {
		return nil, fmt.Errorf("update ecn: %w", err)
	}

	// BOM状态回released，保留草稿以便重新编辑
	bom, err := s.bomRepo.FindByID(ctx, ecn.BOMID)
	if err != nil {
		return nil, fmt.Errorf("find bom: %w", err)
	}

	if bom.FrozenAt != nil {
		bom.Status = "frozen"
	} else {
		bom.Status = "released"
	}
	bom.UpdatedAt = time.Now()

	if err := s.bomRepo.Update(ctx, bom); err != nil {
		return nil, fmt.Errorf("update bom: %w", err)
	}

	return ecn, nil
}

// ListECNs 获取ECN列表
func (s *BOMECNService) ListECNs(ctx context.Context, bomID string, status string) ([]entity.BOMECN, error) {
	return s.ecnRepo.List(ctx, bomID, status)
}

// GetECN 获取ECN详情
func (s *BOMECNService) GetECN(ctx context.Context, ecnID string) (*entity.BOMECN, error) {
	return s.ecnRepo.FindByID(ctx, ecnID)
}

// calculateDiff 计算BOM变更diff
func (s *BOMECNService) calculateDiff(originalItems []entity.ProjectBOMItem, draftItems []entity.ProjectBOMItem) map[string]interface{} {
	diff := map[string]interface{}{
		"added":    []entity.ProjectBOMItem{},
		"removed":  []entity.ProjectBOMItem{},
		"modified": []map[string]interface{}{},
	}

	originalMap := make(map[string]entity.ProjectBOMItem)
	for _, item := range originalItems {
		originalMap[item.ID] = item
	}

	draftMap := make(map[string]entity.ProjectBOMItem)
	for _, item := range draftItems {
		draftMap[item.ID] = item
	}

	// 找出新增和修改的项
	for id, draftItem := range draftMap {
		if _, exists := originalMap[id]; !exists {
			diff["added"] = append(diff["added"].([]entity.ProjectBOMItem), draftItem)
		} else {
			// 简化的对比逻辑，实际应该深度对比字段
			originalItem := originalMap[id]
			if !s.itemsEqual(originalItem, draftItem) {
				diff["modified"] = append(diff["modified"].([]map[string]interface{}), map[string]interface{}{
					"id":       id,
					"before":   originalItem,
					"after":    draftItem,
					"changes": s.getItemChanges(originalItem, draftItem),
				})
			}
		}
	}

	// 找出删除的项
	for id, originalItem := range originalMap {
		if _, exists := draftMap[id]; !exists {
			diff["removed"] = append(diff["removed"].([]entity.ProjectBOMItem), originalItem)
		}
	}

	return diff
}

// itemsEqual 简化的item对比
func (s *BOMECNService) itemsEqual(a, b entity.ProjectBOMItem) bool {
	return a.Name == b.Name &&
		a.Quantity == b.Quantity &&
		a.Unit == b.Unit &&
		a.Supplier == b.Supplier
}

// getItemChanges 获取item变更字段
func (s *BOMECNService) getItemChanges(before, after entity.ProjectBOMItem) []string {
	changes := []string{}
	if before.Name != after.Name {
		changes = append(changes, "name")
	}
	if before.Quantity != after.Quantity {
		changes = append(changes, "quantity")
	}
	if before.Unit != after.Unit {
		changes = append(changes, "unit")
	}
	if before.Supplier != after.Supplier {
		changes = append(changes, "supplier")
	}
	return changes
}

// applyDraftToBOM 应用草稿到BOM
func (s *BOMECNService) applyDraftToBOM(ctx context.Context, bomID string, draftData *DraftData) error {
	bom, err := s.bomRepo.FindByID(ctx, bomID)
	if err != nil {
		return err
	}

	// 更新BOM基本信息
	if draftData.Name != "" {
		bom.Name = draftData.Name
	}
	if draftData.Description != "" {
		bom.Description = draftData.Description
	}

	// 删除旧的items
	if err := s.bomItemRepo.DB().WithContext(ctx).Where("bom_id = ?", bomID).Delete(&entity.ProjectBOMItem{}).Error; err != nil {
		return err
	}

	// 创建新的items
	for _, item := range draftData.Items {
		item.BOMID = bomID
		item.UpdatedAt = time.Now()
		if err := s.bomItemRepo.DB().WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}

	bom.TotalItems = len(draftData.Items)
	bom.UpdatedAt = time.Now()

	return s.bomRepo.Update(ctx, bom)
}
