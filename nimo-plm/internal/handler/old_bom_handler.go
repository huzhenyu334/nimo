package handler

import "github.com/gin-gonic/gin"

// OldBOMHandler 旧BOM处理器（产品关联BOM，保留兼容）
type OldBOMHandler struct{}

func (h *OldBOMHandler) Get(c *gin.Context)          { Success(c, gin.H{"items": []interface{}{}, "total": 0}) }
func (h *OldBOMHandler) ListVersions(c *gin.Context)  { Success(c, gin.H{"versions": []interface{}{}}) }
func (h *OldBOMHandler) AddItem(c *gin.Context)       { BadRequest(c, "请使用新的项目BOM接口 /projects/:id/boms") }
func (h *OldBOMHandler) UpdateItem(c *gin.Context)    { BadRequest(c, "请使用新的项目BOM接口") }
func (h *OldBOMHandler) DeleteItem(c *gin.Context)    { BadRequest(c, "请使用新的项目BOM接口") }
func (h *OldBOMHandler) Release(c *gin.Context)       { BadRequest(c, "请使用新的项目BOM接口") }
func (h *OldBOMHandler) Compare(c *gin.Context)       { Success(c, gin.H{"changes": []interface{}{}}) }
