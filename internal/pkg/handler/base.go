package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BaseHandler struct {
	logger *zap.Logger
}

func NewBaseHandler(logger *zap.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

func (bh *BaseHandler) ErrorResponse(c *gin.Context, status int, err error, message string) {
	bh.logger.Error(err.Error())

	c.AbortWithStatusJSON(status, errorResponse{
		Message: message,
	})
}

func (bh *BaseHandler) ParseIDFromPath(c *gin.Context, param string) (int64, error) {
	idParam := c.Param(param)
	if idParam == "" {
		return 0, errors.New("empty id param")
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, errors.New("invalid id param")
	}

	return id, nil
}
