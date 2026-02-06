package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// FeishuIntegrationService 飞书集成服务
type FeishuIntegrationService struct {
	appID       string
	appSecret   string
	tokenCache  string
	tokenExpire time.Time
	mu          sync.RWMutex
}

// NewFeishuIntegrationService 创建飞书集成服务
func NewFeishuIntegrationService(appID, appSecret string) *FeishuIntegrationService {
	return &FeishuIntegrationService{appID: appID, appSecret: appSecret}
}

// GetAppAccessToken 获取应用访问令牌
func (s *FeishuIntegrationService) GetAppAccessToken(ctx context.Context) (string, error) {
	// 先尝试从缓存获取
	s.mu.RLock()
	if s.tokenCache != "" && time.Now().Before(s.tokenExpire) {
		token := s.tokenCache
		s.mu.RUnlock()
		return token, nil
	}
	s.mu.RUnlock()

	// 请求新token
	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查
	if s.tokenCache != "" && time.Now().Before(s.tokenExpire) {
		return s.tokenCache, nil
	}

	reqBody := map[string]string{
		"app_id":     s.appID,
		"app_secret": s.appSecret,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		"https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("request feishu: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code           int    `json:"code"`
		Msg            string `json:"msg"`
		AppAccessToken string `json:"app_access_token"`
		Expire         int    `json:"expire"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("feishu error: %s", result.Msg)
	}

	// 缓存token
	s.tokenCache = result.AppAccessToken
	s.tokenExpire = time.Now().Add(time.Duration(result.Expire-60) * time.Second)

	return result.AppAccessToken, nil
}

// SendMessage 发送消息给用户
func (s *FeishuIntegrationService) SendMessage(ctx context.Context, userID string, content string) error {
	token, err := s.GetAppAccessToken(ctx)
	if err != nil {
		return err
	}

	reqBody := map[string]interface{}{
		"receive_id_type": "user_id",
		"receive_id":      userID,
		"msg_type":        "text",
		"content":         fmt.Sprintf(`{"text":"%s"}`, content),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://open.feishu.cn/open-apis/im/v1/messages",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("feishu error: %s", result.Msg)
	}

	return nil
}

// CreateTask 创建飞书任务
func (s *FeishuIntegrationService) CreateTask(ctx context.Context, summary, description string, assigneeID string, dueTime *time.Time) (string, error) {
	token, err := s.GetAppAccessToken(ctx)
	if err != nil {
		return "", err
	}

	reqBody := map[string]interface{}{
		"summary":     summary,
		"description": description,
	}

	if assigneeID != "" {
		reqBody["members"] = []map[string]interface{}{
			{
				"id":   assigneeID,
				"role": "assignee",
			},
		}
	}

	if dueTime != nil {
		reqBody["due"] = map[string]interface{}{
			"time":       dueTime.Unix() * 1000,
			"is_all_day": false,
		}
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://open.feishu.cn/open-apis/task/v2/tasks",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create task: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Task struct {
				Guid string `json:"guid"`
			} `json:"task"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("feishu error: %s", result.Msg)
	}

	return result.Data.Task.Guid, nil
}

// CreateCalendarEvent 创建日历事件
func (s *FeishuIntegrationService) CreateCalendarEvent(ctx context.Context, summary, description string, startTime, endTime time.Time, attendeeIDs []string) (string, error) {
	token, err := s.GetAppAccessToken(ctx)
	if err != nil {
		return "", err
	}

	attendees := make([]map[string]interface{}, len(attendeeIDs))
	for i, id := range attendeeIDs {
		attendees[i] = map[string]interface{}{
			"type":    "user",
			"user_id": id,
		}
	}

	reqBody := map[string]interface{}{
		"summary":     summary,
		"description": description,
		"start_time": map[string]interface{}{
			"timestamp": fmt.Sprintf("%d", startTime.Unix()),
		},
		"end_time": map[string]interface{}{
			"timestamp": fmt.Sprintf("%d", endTime.Unix()),
		},
		"attendees":         attendees,
		"need_notification": true,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://open.feishu.cn/open-apis/calendar/v4/calendars/primary/events",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create event: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Event struct {
				EventID string `json:"event_id"`
			} `json:"event"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("feishu error: %s", result.Msg)
	}

	return result.Data.Event.EventID, nil
}
