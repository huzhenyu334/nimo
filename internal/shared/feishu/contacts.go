package feishu

import (
	"context"
	"fmt"
	"log"
)

// FeishuDepartment 飞书部门
type FeishuDepartment struct {
	DepartmentID     string `json:"department_id"`
	OpenDepartmentID string `json:"open_department_id"`
	Name             string `json:"name"`
	ParentID         string `json:"parent_department_id"`
	LeaderUserID     string `json:"leader_user_id"`
	MemberCount      int    `json:"member_count"`
}

// FeishuUser 飞书用户
type FeishuUser struct {
	UserID        string   `json:"user_id"`
	OpenID        string   `json:"open_id"`
	UnionID       string   `json:"union_id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Mobile        string   `json:"mobile"`
	Avatar        struct {
		URL string `json:"avatar_origin"`
	} `json:"avatar"`
	DepartmentIDs []string `json:"department_ids"`
	EmployeeNo    string   `json:"employee_no"`
}

// ListDepartments 获取所有部门列表（递归获取所有子部门）
// API: GET /open-apis/contact/v3/departments?parent_department_id=0&fetch_child=true
func (c *FeishuClient) ListDepartments(ctx context.Context) ([]FeishuDepartment, error) {
	var allDepts []FeishuDepartment
	pageToken := ""

	for {
		path := "/open-apis/contact/v3/departments?parent_department_id=0&fetch_child=true&page_size=50"
		if pageToken != "" {
			path += "&page_token=" + pageToken
		}

		var resp struct {
			BaseResponse
			Data struct {
				HasMore   bool               `json:"has_more"`
				PageToken string             `json:"page_token"`
				Items     []FeishuDepartment `json:"items"`
			} `json:"data"`
		}

		if err := c.doRequest(ctx, "GET", path, nil, &resp); err != nil {
			return nil, fmt.Errorf("获取部门列表失败: %w", err)
		}

		allDepts = append(allDepts, resp.Data.Items...)

		if !resp.Data.HasMore || resp.Data.PageToken == "" {
			break
		}
		pageToken = resp.Data.PageToken
	}

	log.Printf("[FeishuContacts] 获取到 %d 个部门", len(allDepts))
	return allDepts, nil
}

// ListDepartmentUsers 获取指定部门的成员列表
// API: GET /open-apis/contact/v3/users?department_id=xxx&page_size=50
func (c *FeishuClient) ListDepartmentUsers(ctx context.Context, deptID string) ([]FeishuUser, error) {
	var allUsers []FeishuUser
	pageToken := ""

	for {
		path := fmt.Sprintf("/open-apis/contact/v3/users?department_id=%s&page_size=50", deptID)
		if pageToken != "" {
			path += "&page_token=" + pageToken
		}

		var resp struct {
			BaseResponse
			Data struct {
				HasMore   bool         `json:"has_more"`
				PageToken string       `json:"page_token"`
				Items     []FeishuUser `json:"items"`
			} `json:"data"`
		}

		if err := c.doRequest(ctx, "GET", path, nil, &resp); err != nil {
			return nil, fmt.Errorf("获取部门[%s]成员失败: %w", deptID, err)
		}

		allUsers = append(allUsers, resp.Data.Items...)

		if !resp.Data.HasMore || resp.Data.PageToken == "" {
			break
		}
		pageToken = resp.Data.PageToken
	}

	return allUsers, nil
}

// ListAllUsers 遍历所有部门获取全部用户（自动去重）
func (c *FeishuClient) ListAllUsers(ctx context.Context) ([]FeishuUser, error) {
	depts, err := c.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var allUsers []FeishuUser

	for _, dept := range depts {
		deptID := dept.OpenDepartmentID
		if deptID == "" {
			deptID = dept.DepartmentID
		}
		if deptID == "" {
			continue
		}

		users, err := c.ListDepartmentUsers(ctx, deptID)
		if err != nil {
			log.Printf("[FeishuContacts] 获取部门[%s]成员失败，跳过: %v", dept.Name, err)
			continue
		}

		for _, u := range users {
			if u.OpenID != "" && !seen[u.OpenID] {
				seen[u.OpenID] = true
				allUsers = append(allUsers, u)
			}
		}
	}

	log.Printf("[FeishuContacts] 去重后共获取到 %d 个用户", len(allUsers))
	return allUsers, nil
}
