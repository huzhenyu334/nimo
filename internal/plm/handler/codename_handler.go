package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/repository"

	"github.com/gin-gonic/gin"
)

type CodenameHandler struct {
	repo *repository.CodenameRepository
}

func NewCodenameHandler(repo *repository.CodenameRepository) *CodenameHandler {
	return &CodenameHandler{repo: repo}
}

// List GET /codenames
func (h *CodenameHandler) List(c *gin.Context) {
	codenameType := c.Query("type")
	availableOnly := c.Query("available") == "true"

	codenames, err := h.repo.List(c.Request.Context(), codenameType, availableOnly)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, codenames)
}
