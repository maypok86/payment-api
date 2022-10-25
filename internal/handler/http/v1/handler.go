package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/domain"
	"github.com/maypok86/payment-api/internal/handler/http/v1/account"
	"github.com/maypok86/payment-api/internal/handler/http/v1/transaction"
	"go.uber.org/zap"
)

type Handler struct {
	services *domain.Services
	logger   *zap.Logger
}

func NewHandler(services *domain.Services, logger *zap.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) InitAPI(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		account.NewHandler(h.services.Account, h.logger).InitAPI(v1)
		transaction.NewHandler(h.services.Transaction, h.logger).InitAPI(v1)
	}
}
