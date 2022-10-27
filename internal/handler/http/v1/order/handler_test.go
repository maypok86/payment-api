package order_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/maypok86/payment-api/internal/domain/account"
	domain "github.com/maypok86/payment-api/internal/domain/order"
	"github.com/maypok86/payment-api/internal/handler/http/v1/order"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"github.com/maypok86/payment-api/internal/pkg/logger"
	"github.com/stretchr/testify/require"
)

func mockHandler(t *testing.T, w http.ResponseWriter) (*order.Handler, *MockService, *gin.Context) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gin.SetMode(gin.TestMode)

	c, r := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: make(http.Header),
	}

	l := logger.New(os.Stdout, "debug")

	orderService := NewMockService(mockCtrl)
	orderHandler := order.NewHandler(orderService, l)

	orderHandler.InitAPI(r.Group("/"))

	return orderHandler, orderService, c
}

func newCreateOrderResponse(t *testing.T, fakeOrder domain.Order, fakeBalance int64) order.CreateOrderResponse {
	t.Helper()

	fakeResponse := order.CreateOrderResponse{
		Order: order.Response{
			OrderID:     fakeOrder.OrderID,
			AccountID:   fakeOrder.AccountID,
			ServiceID:   fakeOrder.ServiceID,
			Amount:      fakeOrder.Amount,
			IsPaid:      fakeOrder.IsPaid,
			IsCancelled: fakeOrder.IsCancelled,
			CreatedAt:   fakeOrder.CreatedAt,
			UpdatedAt:   fakeOrder.UpdatedAt,
		},
		Balance: fakeBalance,
	}

	// for fix time.Time in json
	var buffer bytes.Buffer
	require.NoError(t, json.NewEncoder(&buffer).Encode(fakeResponse))
	require.NoError(t, json.NewDecoder(&buffer).Decode(&fakeResponse))

	return fakeResponse
}

func TestHandler_CreateOrder(t *testing.T) {
	ctx := context.Background()

	fakeRequest := order.CreateOrderRequest{
		OrderID:   1,
		AccountID: 1,
		ServiceID: 1,
		Amount:    100,
	}
	fakeBalance := int64(100)
	fakeOrder := domain.Order{
		OrderID:     fakeRequest.OrderID,
		AccountID:   fakeRequest.AccountID,
		ServiceID:   fakeRequest.ServiceID,
		Amount:      fakeRequest.Amount,
		IsPaid:      false,
		IsCancelled: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	fakeResponse := newCreateOrderResponse(t, fakeOrder, fakeBalance)
	orderServiceErr := errors.New("order service error")

	setupGin := func(c *gin.Context, content interface{}) {
		c.Request.Method = http.MethodPost
		c.Request.Header.Set("Content-Type", "application/json")

		data, err := json.Marshal(content)
		require.NoError(t, err)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		request order.CreateOrderRequest
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		response            order.CreateOrderResponse
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid request",
			mock: func(service *MockService) {
			},
			args: args{
				request: order.CreateOrderRequest{
					OrderID:   1,
					AccountID: 1,
					ServiceID: 1,
					Amount:    -1,
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Create order error. Invalid request",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "account not found",
			mock: func(service *MockService) {
				service.EXPECT().
					CreateOrder(ctx, fakeRequest.ToDTO()).
					Return(domain.Order{}, int64(0), account.ErrNotFound)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Create order error. Account not found",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "order already exists",
			mock: func(service *MockService) {
				service.EXPECT().
					CreateOrder(ctx, fakeRequest.ToDTO()).
					Return(domain.Order{}, int64(0), domain.ErrAlreadyExist)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Create order error. Order already exist",
			},
			statusCode: http.StatusConflict,
		},
		{
			name: "order service error",
			mock: func(service *MockService) {
				service.EXPECT().CreateOrder(ctx, fakeRequest.ToDTO()).Return(domain.Order{}, int64(0), orderServiceErr)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Create order error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success create order",
			mock: func(service *MockService) {
				service.EXPECT().
					CreateOrder(ctx, fakeRequest.ToDTO()).
					Return(fakeOrder, fakeBalance, nil)
			},
			args: args{
				request: fakeRequest,
			},
			response:   fakeResponse,
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			orderHandler, orderService, c := mockHandler(t, w)

			setupGin(c, tt.args.request)
			tt.mock(orderService)

			orderHandler.CreateOrder(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			} else {
				var response order.CreateOrderResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.response, response))
			}
		})
	}
}

func TestHandler_PayForOrder(t *testing.T) {
	ctx := context.Background()

	fakeRequest := order.PayForOrderRequest{
		OrderID:   1,
		AccountID: 1,
		ServiceID: 1,
		Amount:    100,
	}
	orderServiceErr := errors.New("order service error")

	setupGin := func(c *gin.Context, content interface{}) {
		c.Request.Method = http.MethodPost
		c.Request.Header.Set("Content-Type", "application/json")

		data, err := json.Marshal(content)
		require.NoError(t, err)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		request order.PayForOrderRequest
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid request",
			mock: func(service *MockService) {
			},
			args: args{
				request: order.PayForOrderRequest{
					OrderID:   1,
					AccountID: 1,
					ServiceID: 1,
					Amount:    -1,
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Pay for order error. Invalid request",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "order not found",
			mock: func(service *MockService) {
				service.EXPECT().
					PayForOrder(ctx, fakeRequest.ToDTO()).
					Return(domain.ErrNotFound)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Pay for order error. Order not found",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "account service error",
			mock: func(service *MockService) {
				service.EXPECT().PayForOrder(ctx, fakeRequest.ToDTO()).Return(orderServiceErr)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Pay for order error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success pay for order",
			mock: func(service *MockService) {
				service.EXPECT().
					PayForOrder(ctx, fakeRequest.ToDTO()).
					Return(nil)
			},
			args: args{
				request: fakeRequest,
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			orderHandler, orderService, c := mockHandler(t, w)

			setupGin(c, tt.args.request)
			tt.mock(orderService)

			orderHandler.PayForOrder(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			}
		})
	}
}

func TestHandler_CancelOrder(t *testing.T) {
	ctx := context.Background()

	fakeRequest := order.CancelOrderRequest{
		OrderID:   1,
		AccountID: 1,
		ServiceID: 1,
		Amount:    100,
	}
	fakeBalance := int64(100)
	fakeResponse := order.CancelOrderResponse{
		Balance: fakeBalance,
	}
	orderServiceErr := errors.New("order service error")

	setupGin := func(c *gin.Context, content interface{}) {
		c.Request.Method = http.MethodPost
		c.Request.Header.Set("Content-Type", "application/json")

		data, err := json.Marshal(content)
		require.NoError(t, err)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		request order.CancelOrderRequest
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		response            order.CancelOrderResponse
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid request",
			mock: func(service *MockService) {
			},
			args: args{
				request: order.CancelOrderRequest{
					OrderID:   1,
					AccountID: 1,
					ServiceID: 1,
					Amount:    -1,
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Cancel order error. Invalid request",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "account not found",
			mock: func(service *MockService) {
				service.EXPECT().
					CancelOrder(ctx, fakeRequest.ToDTO()).
					Return(int64(0), account.ErrNotFound)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Cancel order error. Account not found",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "order not found",
			mock: func(service *MockService) {
				service.EXPECT().
					CancelOrder(ctx, fakeRequest.ToDTO()).
					Return(int64(0), domain.ErrNotFound)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Cancel order error. Order not found",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "order service error",
			mock: func(service *MockService) {
				service.EXPECT().CancelOrder(ctx, fakeRequest.ToDTO()).Return(int64(0), orderServiceErr)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Cancel order error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success cancel order",
			mock: func(service *MockService) {
				service.EXPECT().
					CancelOrder(ctx, fakeRequest.ToDTO()).
					Return(fakeBalance, nil)
			},
			args: args{
				request: fakeRequest,
			},
			response:   fakeResponse,
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			orderHandler, orderService, c := mockHandler(t, w)

			setupGin(c, tt.args.request)
			tt.mock(orderService)

			orderHandler.CancelOrder(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			} else {
				var response order.CancelOrderResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.response, response))
			}
		})
	}
}
