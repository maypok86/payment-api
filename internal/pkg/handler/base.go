package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/pkg/pagination"
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

var (
	ErrEmptyIDParam       = errors.New("empty id param")
	ErrInvalidID          = errors.New("invalid id param")
	ErrInvalidLimitParam  = errors.New("invalid limit param")
	ErrInvalidOffsetParam = errors.New("invalid offset param")
)

func (bh *BaseHandler) ParseIDFromPath(c *gin.Context, param string) (int64, error) {
	idParam := c.Param(param)
	if idParam == "" {
		return 0, ErrEmptyIDParam
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, ErrInvalidID
	}

	return id, nil
}

func (bh *BaseHandler) ParsePaginationParams(c *gin.Context) (pagination.Params, error) {
	var (
		params pagination.Params
		err    error
	)

	limit := c.Query("limit")
	params.Limit, err = strconv.ParseUint(limit, 10, 64)
	if err != nil {
		if limit != "" {
			return pagination.Params{}, ErrInvalidLimitParam
		}
		params.Limit = pagination.DefaultLimit
	}

	offset := c.Query("offset")
	params.Offset, err = strconv.ParseUint(offset, 10, 64)
	if err != nil {
		if offset != "" {
			return pagination.Params{}, ErrInvalidOffsetParam
		}
		params.Offset = pagination.DefaultOffset
	}

	return params, nil
}
