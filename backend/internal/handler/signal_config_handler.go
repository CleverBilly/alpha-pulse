package handler

import (
	"net/http"

	signalpkg "alpha-pulse/backend/internal/signal"
	"alpha-pulse/backend/models"
	"alpha-pulse/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// SignalConfigRepo 定义 SignalConfigHandler 所需的持久化操作接口，
// 使测试可通过 stub 注入而无需真实数据库。
type SignalConfigRepo interface {
	Upsert(cfg models.SignalConfig) error
	GetAll() ([]models.SignalConfig, error)
}

// SignalConfigHandler 处理信号配置热更新请求。
type SignalConfigHandler struct {
	provider *signalpkg.DBConfigProvider
	repo     SignalConfigRepo
}

// NewSignalConfigHandler 创建 SignalConfigHandler。
func NewSignalConfigHandler(provider *signalpkg.DBConfigProvider, repo SignalConfigRepo) *SignalConfigHandler {
	return &SignalConfigHandler{provider: provider, repo: repo}
}

// patchSignalConfigRequest 定义 PATCH 请求体结构。
type patchSignalConfigRequest struct {
	Symbol   string `json:"symbol"   binding:"required"`
	Interval string `json:"interval" binding:"required"`
	Key      string `json:"key"      binding:"required"`
	Value    string `json:"value"    binding:"required"`
}

// Patch 更新单条信号配置：先写 DB，再热更新内存 ConfigProvider。
// PATCH /api/signal-configs
func (h *SignalConfigHandler) Patch(c *gin.Context) {
	var req patchSignalConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Error(http.StatusBadRequest, err.Error()))
		return
	}

	cfg := models.SignalConfig{
		Symbol:   req.Symbol,
		Interval: req.Interval,
		Key:      req.Key,
		Value:    req.Value,
	}
	if err := h.repo.Upsert(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(http.StatusInternalServerError, "db upsert failed"))
		return
	}

	h.provider.Update(req.Symbol, req.Interval, req.Key, req.Value)

	c.JSON(http.StatusOK, utils.Success(cfg))
}

// List 返回所有信号配置。
// GET /api/signal-configs
func (h *SignalConfigHandler) List(c *gin.Context) {
	configs, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Success(configs))
}
