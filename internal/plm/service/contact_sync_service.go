package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/shared/feishu"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactSyncService 通讯录同步服务
type ContactSyncService struct {
	db           *gorm.DB
	feishuClient *feishu.FeishuClient
}

// NewContactSyncService 创建通讯录同步服务
func NewContactSyncService(db *gorm.DB, fc *feishu.FeishuClient) *ContactSyncService {
	return &ContactSyncService{db: db, feishuClient: fc}
}

// SyncResult 同步结果
type SyncResult struct {
	DepartmentsCreated int      `json:"departments_created"`
	DepartmentsUpdated int      `json:"departments_updated"`
	UsersCreated       int      `json:"users_created"`
	UsersUpdated       int      `json:"users_updated"`
	Errors             []string `json:"errors,omitempty"`
}

// SyncContacts 执行通讯录同步（部门+用户）
func (s *ContactSyncService) SyncContacts(ctx context.Context) (*SyncResult, error) {
	if s.feishuClient == nil {
		return nil, fmt.Errorf("飞书客户端未初始化")
	}

	result := &SyncResult{}

	// 1. 同步部门
	depts, err := s.feishuClient.ListDepartments(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取飞书部门列表失败: %w", err)
	}

	for _, dept := range depts {
		created, err := s.syncDepartment(ctx, dept)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("同步部门[%s]失败: %v", dept.Name, err))
			continue
		}
		if created {
			result.DepartmentsCreated++
		} else {
			result.DepartmentsUpdated++
		}
	}

	// 2. 同步用户
	users, err := s.feishuClient.ListAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取飞书用户列表失败: %w", err)
	}

	for _, user := range users {
		created, err := s.syncUser(ctx, user)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("同步用户[%s]失败: %v", user.Name, err))
			continue
		}
		if created {
			result.UsersCreated++
		} else {
			result.UsersUpdated++
		}
	}

	log.Printf("[ContactSync] 同步完成: 部门新建=%d 更新=%d, 用户新建=%d 更新=%d, 错误=%d",
		result.DepartmentsCreated, result.DepartmentsUpdated,
		result.UsersCreated, result.UsersUpdated, len(result.Errors))

	return result, nil
}

// syncDepartment 同步单个部门 (upsert by feishu_dept_id)
func (s *ContactSyncService) syncDepartment(ctx context.Context, dept feishu.FeishuDepartment) (created bool, err error) {
	deptID := dept.OpenDepartmentID
	if deptID == "" {
		deptID = dept.DepartmentID
	}
	if deptID == "" {
		return false, fmt.Errorf("部门ID为空")
	}

	var existing entity.Department
	err = s.db.WithContext(ctx).Where("feishu_dept_id = ?", deptID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新部门（不设 parent_id，避免外键约束）
		newDept := entity.Department{
			ID:           uuid.New().String()[:32],
			FeishuDeptID: deptID,
			Name:         dept.Name,
			ParentID:     "", // 不设外键，避免父部门尚未创建
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := s.db.WithContext(ctx).Create(&newDept).Error; err != nil {
			return false, fmt.Errorf("创建部门失败: %w", err)
		}
		return true, nil
	} else if err != nil {
		return false, err
	}

	// 更新现有部门
	existing.Name = dept.Name
	existing.UpdatedAt = time.Now()
	if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return false, fmt.Errorf("更新部门失败: %w", err)
	}
	return false, nil
}

// syncUser 同步单个用户 (upsert by feishu_open_id)
func (s *ContactSyncService) syncUser(ctx context.Context, user feishu.FeishuUser) (created bool, err error) {
	if user.OpenID == "" {
		return false, fmt.Errorf("用户 OpenID 为空")
	}

	// 查找对应的部门ID
	var departmentID string
	if len(user.DepartmentIDs) > 0 {
		var dept entity.Department
		for _, deptFeishuID := range user.DepartmentIDs {
			if err := s.db.WithContext(ctx).Where("feishu_dept_id = ?", deptFeishuID).First(&dept).Error; err == nil {
				departmentID = dept.ID
				break
			}
		}
	}

	var existing entity.User
	err = s.db.WithContext(ctx).Where("feishu_open_id = ?", user.OpenID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新用户
		username := user.Email
		if username == "" {
			username = "feishu_" + user.OpenID[:16]
		}

		// 检查 username 是否已存在，避免唯一约束冲突
		var count int64
		s.db.WithContext(ctx).Model(&entity.User{}).Where("username = ?", username).Count(&count)
		if count > 0 {
			username = username + "_" + uuid.New().String()[:4]
		}

		// 处理 email：空 email 设为唯一占位符，避免唯一约束冲突
		email := user.Email
		if email == "" {
			email = "feishu_" + user.OpenID[:16] + "@placeholder.local"
		}

		// 检查 email 唯一性
		s.db.WithContext(ctx).Model(&entity.User{}).Where("email = ? AND feishu_open_id != ?", email, user.OpenID).Count(&count)
		if count > 0 {
			// email 冲突，尝试更新已有用户的 feishu_open_id
			var conflictUser entity.User
			if err := s.db.WithContext(ctx).Where("email = ?", email).First(&conflictUser).Error; err == nil {
				conflictUser.FeishuOpenID = user.OpenID
				conflictUser.FeishuUserID = user.UserID
				conflictUser.FeishuUnionID = user.UnionID
				conflictUser.Name = user.Name
				conflictUser.Mobile = user.Mobile
				conflictUser.AvatarURL = user.Avatar.URL
				if departmentID != "" {
					conflictUser.DepartmentID = departmentID
				}
				conflictUser.UpdatedAt = time.Now()
				if err := s.db.WithContext(ctx).Save(&conflictUser).Error; err != nil {
					return false, fmt.Errorf("更新已有用户飞书信息失败: %w", err)
				}
				return false, nil
			}
		}

		// 检查 employee_no 唯一性
		employeeNo := user.EmployeeNo
		if employeeNo == "" {
			employeeNo = "FS-" + user.OpenID[:8]
		} else {
			s.db.WithContext(ctx).Model(&entity.User{}).Where("employee_no = ?", employeeNo).Count(&count)
			if count > 0 {
				employeeNo = employeeNo + "-" + uuid.New().String()[:4]
			}
		}

		newUser := entity.User{
			ID:            uuid.New().String()[:32],
			FeishuOpenID:  user.OpenID,
			FeishuUnionID: user.UnionID,
			FeishuUserID:  user.UserID,
			EmployeeNo:    employeeNo,
			Username:      username,
			Name:          user.Name,
			Email:         email,
			Mobile:        user.Mobile,
			AvatarURL:     user.Avatar.URL,
			DepartmentID:  departmentID,
			Status:        "active",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := s.db.WithContext(ctx).Create(&newUser).Error; err != nil {
			return false, fmt.Errorf("创建用户失败: %w", err)
		}
		return true, nil
	} else if err != nil {
		return false, err
	}

	// 更新现有用户
	existing.Name = user.Name
	if user.Email != "" {
		existing.Email = user.Email
	}
	existing.Mobile = user.Mobile
	existing.AvatarURL = user.Avatar.URL
	existing.FeishuUserID = user.UserID
	existing.FeishuUnionID = user.UnionID
	if user.EmployeeNo != "" {
		existing.EmployeeNo = user.EmployeeNo
	}
	if departmentID != "" {
		existing.DepartmentID = departmentID
	}
	existing.Status = "active"
	existing.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return false, fmt.Errorf("更新用户失败: %w", err)
	}
	return false, nil
}
