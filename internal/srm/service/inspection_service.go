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
	activityLogRepo *repository.ActivityLogRepository
	feishuClient    *feishu.FeishuClient
}

func NewInspectionService(repo *repository.InspectionRepository, prRepo *repository.PRRepository) *InspectionService {
	return &InspectionService{
		repo:   repo,
		prRepo: prRepo,
	}
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
	Result          string           `json:"result" binding:"required"` // passed/failed/conditional
	InspectionItems *json.RawMessage `json:"inspection_items"`
	Notes           string           `json:"notes"`
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
	inspection.InspectorID = &userID
	inspection.InspectedAt = &now
	if req.InspectionItems != nil {
		inspection.InspectionItems = *req.InspectionItems
	}
	if req.Notes != "" {
		inspection.Notes = req.Notes
	}

	if err := s.repo.Update(ctx, inspection); err != nil {
		return nil, err
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
