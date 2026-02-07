package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	contactSyncSvc *service.ContactSyncService
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(contactSyncSvc *service.ContactSyncService) *AdminHandler {
	return &AdminHandler{contactSyncSvc: contactSyncSvc}
}

// SyncContacts 同步飞书通讯录
// POST /api/v1/admin/sync-contacts
func (h *AdminHandler) SyncContacts(c *gin.Context) {
	result, err := h.contactSyncSvc.SyncContacts(c.Request.Context())
	if err != nil {
		InternalError(c, "通讯录同步失败: "+err.Error())
		return
	}
	Success(c, result)
}
