package report

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/domain/report"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"go.uber.org/zap"
)

//go:generate mockgen -source=handler.go -destination=mock_test.go -package=report_test

type Service interface {
	GetReportKey(ctx context.Context, dto report.GetMapDTO) (string, error)
	GetReportContent(ctx context.Context, filename string) ([]byte, error)
}

type Config struct {
	ReportHost string
	ReportPort string
}

type Handler struct {
	*handler.BaseHandler
	cfg     Config
	service Service
	logger  *zap.Logger
}

func NewHandler(cfg Config, service Service, logger *zap.Logger) *Handler {
	return &Handler{
		BaseHandler: handler.NewBaseHandler(logger),
		cfg:         cfg,
		service:     service,
		logger:      logger,
	}
}

func (h *Handler) InitAPI(router *gin.RouterGroup) {
	reportGroup := router.Group("/report")
	{
		reportGroup.POST("/link", h.GetReportLink)
		reportGroup.GET("/download", h.DownloadReport)
	}
}

func (h *Handler) GetReportLink(c *gin.Context) {
	var request GetReportLinkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Get report link error. Invalid request")
		return
	}

	key, err := h.service.GetReportKey(c.Request.Context(), request.ToDTO())
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

	link := fmt.Sprintf("http://%s:%s/api/v1/report/download?key=%s", h.cfg.ReportHost, h.cfg.ReportPort, key)

	c.JSON(http.StatusOK, GetReportLinkResponse{
		Link: link,
	})
}

func (h *Handler) DownloadReport(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		h.ErrorResponse(
			c,
			http.StatusBadRequest,
			errors.New("query key is empty"),
			"Download report error. Invalid request",
		)
		return
	}

	content, err := h.service.GetReportContent(c.Request.Context(), key)
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
