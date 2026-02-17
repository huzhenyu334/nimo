package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/bitfantasy/nimo/internal/shared/feishu"
	"github.com/bitfantasy/nimo/internal/srm/entity"
	"github.com/bitfantasy/nimo/internal/srm/repository"
	"github.com/google/uuid"
)

// InspectionService 检验服务
type InspectionService struct {
	repo            *repository.InspectionRepository
	prRepo          *repository.PRRepository
	poRepo          *repository.PORepository
	inventorySvc    *InventoryService
	activityLogRepo *repository.ActivityLogRepository
	feishuClient    *feishu.FeishuClient
}

func NewInspectionService(repo *repository.InspectionRepository, prRepo *repository.PRRepository) *InspectionService {
	return &InspectionService{
		repo:   repo,
		prRepo: prRepo,
	}
}

// SetPORepo 注入PO仓库
func (s *InspectionService) SetPORepo(repo *repository.PORepository) {
	s.poRepo = repo
}

// SetInventoryService 注入库存服务
func (s *InspectionService) SetInventoryService(svc *InventoryService) {
	s.inventorySvc = svc
}

// SetActivityLogRepo 注入操作日志仓库
func (s *InspectionService) SetActivityLogRepo(repo *repository.ActivityLogRepository) {
	s.activityLogRepo = repo
}

// SetFeishuClient 注入飞书客户端
func (s *InspectionService) SetFeishuClient(fc *feishu.FeishuClient) {
	s.feishuClient = fc
}

// ListInspections 获取检验列表
func (s *InspectionService) ListInspections(ctx context.Context, page, pageSize int, filters map[string]string) ([]entity.Inspection, int64, error) {
	return s.repo.FindAll(ctx, page, pageSize, filters)
}

// GetInspection 获取检验详情
func (s *InspectionService) GetInspection(ctx context.Context, id string) (*entity.Inspection, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateInspectionRequest 更新检验请求
type UpdateInspectionRequest struct {
	InspectorID     *string          `json:"inspector_id"`
	SampleQty       *int             `json:"sample_qty"`
	InspectionItems *json.RawMessage `json:"inspection_items"`
	ReportURL       *string          `json:"report_url"`
	Notes           *string          `json:"notes"`
}

// UpdateInspection 更新检验
func (s *InspectionService) UpdateInspection(ctx context.Context, id string, req *UpdateInspectionRequest) (*entity.Inspection, error) {
	inspection, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.InspectorID != nil {
		inspection.InspectorID = req.InspectorID
	}
	if req.SampleQty != nil {
		inspection.SampleQty = req.SampleQty
	}
	if req.InspectionItems != nil {
		inspection.InspectionItems = *req.InspectionItems
	}
	if req.ReportURL != nil {
		inspection.ReportURL = *req.ReportURL
	}
	if req.Notes != nil {
		inspection.Notes = *req.Notes
	}

	// 如果有检验员分配，状态改为进行中
	if inspection.InspectorID != nil && inspection.Status == entity.InspectionStatusPending {
		inspection.Status = entity.InspectionStatusInProgress
	}

	if err := s.repo.Update(ctx, inspection); err != nil {
		return nil, err
	}
	return inspection, nil
}

// CompleteInspectionRequest 完成检验请求
type CompleteInspectionRequest struct {
	Result          string                      `json:"result" binding:"required"` // passed/failed/conditional
	InspectionItems *json.RawMessage            `json:"inspection_items"`
	Items           []CompleteInspectionItemReq `json:"items"`
	Notes           string                      `json:"notes"`
}

// CompleteInspectionItemReq 行项结果
type CompleteInspectionItemReq struct {
	ID           string  `json:"id"`
	InspectedQty float64 `json:"inspected_quantity"`
	QualifiedQty float64 `json:"qualified_quantity"`
	DefectQty    float64 `json:"defect_quantity"`
	DefectDesc   string  `json:"defect_description"`
	Result       string  `json:"result"`
}

// CompleteInspection 完成检验
func (s *InspectionService) CompleteInspection(ctx context.Context, id, userID string, req *CompleteInspectionRequest) (*entity.Inspection, error) {
	inspection, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	inspection.Status = entity.InspectionStatusCompleted
	inspection.Result = req.Result
	inspection.OverallResult = req.Result
	inspection.InspectorID = &userID
	inspection.InspectedAt = &now
	inspection.InspectionDate = &now
	if req.InspectionItems != nil {
		inspection.InspectionItems = *req.InspectionItems
	}
	if req.Notes != "" {
		inspection.Notes = req.Notes
	}

	// Update inspection item results if provided
	if len(req.Items) > 0 {
		for _, itemReq := range req.Items {
			for i := range inspection.Items {
				if inspection.Items[i].ID == itemReq.ID {
					inspection.Items[i].QualifiedQty = itemReq.QualifiedQty
					inspection.Items[i].DefectQty = itemReq.DefectQty
					inspection.Items[i].DefectDesc = itemReq.DefectDesc
					inspection.Items[i].InspectedQty = itemReq.InspectedQty
					inspection.Items[i].Result = itemReq.Result
				}
			}
		}
	}

	if err := s.repo.Update(ctx, inspection); err != nil {
		return nil, err
	}

	// Save updated items
	for _, item := range inspection.Items {
		s.repo.UpdateItem(ctx, &item)
	}

	// Update PO received quantities for passed/conditional items
	if (req.Result == "passed" || req.Result == "conditional") && s.poRepo != nil {
		for _, item := range inspection.Items {
			if item.Result == "failed" {
				continue
			}
			if item.POItemID != nil && item.QualifiedQty > 0 {
				s.poRepo.ReceiveItem(ctx, *item.POItemID, item.QualifiedQty)
			}

			// Auto stock-in for qualified quantity
			if s.inventorySvc != nil && item.QualifiedQty > 0 {
				s.inventorySvc.StockInFromInspection(ctx, inspection.ID, item.MaterialName, item.MaterialCode, inspection.SupplierID, item.QualifiedQty, "pcs")
			}
		}
	}

	// 记录操作日志
	if s.activityLogRepo != nil {
		action := "inspect_pass"
		content := fmt.Sprintf("检验通过: %s", inspection.MaterialName)
		if req.Result == "failed" {
			action = "inspect_fail"
			content = fmt.Sprintf("检验不通过: %s", inspection.MaterialName)
		} else if req.Result == "conditional" {
			action = "inspect_conditional"
			content = fmt.Sprintf("让步接收: %s", inspection.MaterialName)
		}
		s.activityLogRepo.LogActivity(ctx, "inspection", inspection.ID, inspection.InspectionCode,
			action, "in_progress", "completed", content, userID, "")
	}

	// 检验不合格时发送飞书通知
	if req.Result == "failed" {
		go s.sendInspectionFailedNotification(context.Background(), inspection)
	}

	return inspection, nil
}

// sendInspectionFailedNotification 发送检验不合格飞书通知
func (s *InspectionService) sendInspectionFailedNotification(ctx context.Context, inspection *entity.Inspection) {
	if s.feishuClient == nil {
		return
	}

	// 硬编码管理员用户ID（采购+品质负责人）
	adminUserID := "ou_5b159fc157d4042f1e8088b1ffebb2da"

	// SRM检验页面链接
	rawURL := "http://43.134.86.237:8080/srm/inspections"
	detailURL := fmt.Sprintf("https://applink.feishu.cn/client/web_url/open?url=%s&mode=window", url.QueryEscape(rawURL))

	notes := inspection.Notes
	if notes == "" {
		notes = "无"
	}

	card := feishu.InteractiveCard{
		Config: &feishu.CardConfig{WideScreenMode: true},
		Header: &feishu.CardHeader{
			Title:    feishu.CardText{Tag: "plain_text", Content: "⚠️ 来料检验不合格"},
			Template: "red",
		},
		Elements: []feishu.CardElement{
			{
				Tag: "div",
				Fields: []feishu.CardField{
					{IsShort: true, Text: feishu.CardText{Tag: "lark_md", Content: fmt.Sprintf("**检验编码**\n%s", inspection.InspectionCode)}},
					{IsShort: true, Text: feishu.CardText{Tag: "lark_md", Content: fmt.Sprintf("**物料名称**\n%s", inspection.MaterialName)}},
					{IsShort: true, Text: feishu.CardText{Tag: "lark_md", Content: fmt.Sprintf("**物料编码**\n%s", inspection.MaterialCode)}},
					{IsShort: true, Text: feishu.CardText{Tag: "lark_md", Content: fmt.Sprintf("**检验结果**\n❌ 不合格")}},
				},
			},
			{
				Tag:  "div",
				Text: &feishu.CardText{Tag: "lark_md", Content: fmt.Sprintf("**备注**\n%s", notes)},
			},
			{Tag: "hr"},
			{
				Tag: "action",
				Actions: []feishu.CardAction{
					{
						Tag:  "button",
						Text: feishu.CardText{Tag: "plain_text", Content: "查看检验详情"},
						Type: "primary",
						URL:  detailURL,
					},
				},
			},
		},
	}

	if err := s.feishuClient.SendUserCard(ctx, adminUserID, card); err != nil {
		log.Printf("[SRM] 发送飞书检验不合格通知失败: %v", err)
	} else {
		log.Printf("[SRM] 飞书检验不合格通知已发送: %s", inspection.InspectionCode)
	}
}

// CreateInspectionFromPO 从采购订单创建质检单（包含所有PO行项）
func (s *InspectionService) CreateInspectionFromPO(ctx context.Context, poID string) (*entity.Inspection, error) {
	if s.poRepo == nil {
		return nil, fmt.Errorf("PO仓库未注入")
	}
	po, err := s.poRepo.FindByID(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("采购订单不存在")
	}

	code, err := s.repo.GenerateCode(ctx)
	if err != nil {
		return nil, err
	}

	var totalQty float64
	var items []entity.InspectionItem
	for i, poItem := range po.Items {
		totalQty += poItem.Quantity
		itemID := poItem.ID
		items = append(items, entity.InspectionItem{
			ID:           uuid.New().String()[:32],
			POItemID:     &itemID,
			MaterialName: poItem.MaterialName,
			MaterialCode: poItem.MaterialCode,
			InspectedQty: poItem.Quantity,
			SortOrder:    i + 1,
		})
	}

	inspection := &entity.Inspection{
		ID:             uuid.New().String()[:32],
		InspectionCode: code,
		POID:           &poID,
		SupplierID:     &po.SupplierID,
		MaterialName:   fmt.Sprintf("%s 等%d项", po.POCode, len(po.Items)),
		Quantity:       &totalQty,
		Status:         entity.InspectionStatusPending,
		Items:          items,
	}

	if err := s.repo.Create(ctx, inspection); err != nil {
		return nil, err
	}
	// Re-read with preloads
	created, _ := s.repo.FindByID(ctx, inspection.ID)
	if created != nil {
		return created, nil
	}
	return inspection, nil
}

// CreateInspectionFromPOItem 从PO行项创建检验任务
func (s *InspectionService) CreateInspectionFromPOItem(ctx context.Context, poID, poItemID, supplierID, materialID, materialCode, materialName string, quantity float64) (*entity.Inspection, error) {
	code, err := s.repo.GenerateCode(ctx)
	if err != nil {
		return nil, err
	}

	inspection := &entity.Inspection{
		ID:             uuid.New().String()[:32],
		InspectionCode: code,
		POID:           strPtr(poID),
		POItemID:       strPtr(poItemID),
		SupplierID:     strPtr(supplierID),
		MaterialID:     strPtr(materialID),
		MaterialCode:   materialCode,
		MaterialName:   materialName,
		Quantity:       &quantity,
		Status:         entity.InspectionStatusPending,
	}

	if err := s.repo.Create(ctx, inspection); err != nil {
		return nil, err
	}
	return inspection, nil
}
