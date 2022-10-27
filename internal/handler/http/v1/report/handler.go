package report

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/config"
	"github.com/maypok86/payment-api/internal/domain/report"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"go.uber.org/zap"
)

type Service interface {
	GetReportKey(ctx context.Context, dto report.GetMapDTO) (string, error)
	GetReportContent(ctx context.Context, filename string) ([]byte, error)
}

type Handler struct {
	*handler.BaseHandler
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		BaseHandler: handler.NewBaseHandler(logger),
		service:     service,
		logger:      logger,
	}
}

func (h *Handler) InitAPI(router *gin.RouterGroup) {
	reportGroup := router.Group("/report")
	{
		reportGroup.POST("/link", h.getReportLink)
		reportGroup.GET("/", h.downloadReport)
	}
}

type getReportLinkRequest struct {
	Month int64 `json:"month" binding:"required,min=1,max=12"`
	Year  int64 `json:"year"  binding:"required,min=2022"`
}

func (r getReportLinkRequest) toDTO() report.GetMapDTO {
	return report.GetMapDTO{
		Month: r.Month,
		Year:  r.Year,
	}
}

func (h *Handler) getReportLink(c *gin.Context) {
	var request getReportLinkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Get report link error. Invalid request")
		return
	}

	key, err := h.service.GetReportKey(c, request.toDTO())
	if err != nil {
		switch {
		case errors.Is(err, report.ErrNotFound):
			h.ErrorResponse(c, http.StatusNotFound, err, "Get report link error. Report not found")
			return
		case errors.Is(err, report.ErrIsNotAvailable):
			h.ErrorResponse(c, http.StatusNotFound, err, "Get report link error. Report is not available")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Get report link error. Internal server error")
		return
	}

	port := config.Get().HTTP.Port

	// I don't want to configure it
	link := fmt.Sprintf("http://localhost:%s/api/v1/report?key=%s", port, key)

	c.JSON(http.StatusOK, gin.H{
		"link": link,
	})
}

func (h *Handler) downloadReport(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		h.ErrorResponse(c, http.StatusBadRequest, nil, "Download report error. Invalid request")
		return
	}

	content, err := h.service.GetReportContent(c, key)
	if err != nil {
		if errors.Is(err, report.ErrNotFound) {
			h.ErrorResponse(c, http.StatusNotFound, err, "Download report error. Report not found")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Download report error. Internal server error")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s.csv", key))
	c.Data(http.StatusOK, "application/csv", content)
}
