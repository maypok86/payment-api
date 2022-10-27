package account_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	account_domain "github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/handler/http/v1/account"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"github.com/maypok86/payment-api/internal/pkg/logger"
	"github.com/stretchr/testify/require"
)

func mockHandler(t *testing.T, w http.ResponseWriter) (*account.Handler, *MockService, *gin.Context) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gin.SetMode(gin.TestMode)

	c, r := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: make(http.Header),
	}

	l := logger.New(os.Stdout, "debug")

	accountService := NewMockService(mockCtrl)
	accountHandler := account.NewHandler(accountService, l)

	accountHandler.InitAPI(r.Group("/"))

	return accountHandler, accountService, c
}

func TestHandler_GetBalance(t *testing.T) {
	ctx := context.Background()

	fakeAccountID := int64(1)
	fakeParam := fmt.Sprintf("%d", fakeAccountID)
	fakeBalance := int64(100)
	accountServiceErr := errors.New("account service error")

	setupGin := func(c *gin.Context, param string) {
		c.Request.Method = http.MethodGet
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "account_id", Value: param}}
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		param string
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		response            account.GetBalanceResponse
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid account_id param",
			mock: func(service *MockService) {
			},
			args: args{
				param: "invalid",
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Balance not found. id is not valid",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "account not found",
			mock: func(service *MockService) {
				service.EXPECT().GetBalanceByID(ctx, fakeAccountID).Return(int64(0), account_domain.ErrNotFound)
			},
			args: args{
				param: fakeParam,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Get balance by error. Account not found",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "account service error",
			mock: func(service *MockService) {
				service.EXPECT().GetBalanceByID(ctx, fakeAccountID).Return(int64(0), accountServiceErr)
			},
			args: args{
				param: fakeParam,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Get balance by id error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success get balance",
			mock: func(service *MockService) {
				service.EXPECT().GetBalanceByID(ctx, fakeAccountID).Return(fakeBalance, nil)
			},
			args: args{
				param: fakeParam,
			},
			response: account.GetBalanceResponse{
				Balance: fakeBalance,
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			accountHandler, accountService, c := mockHandler(t, w)

			setupGin(c, tt.args.param)
			tt.mock(accountService)

			accountHandler.GetBalance(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			} else {
				var response account.GetBalanceResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.response, response))
			}
		})
	}
}

func TestHandler_AddBalance(t *testing.T) {
	ctx := context.Background()

	fakeRequest := account.AddBalanceRequest{
		AccountID: 1,
		Amount:    100,
	}
	fakeBalance := int64(100)
	accountServiceErr := errors.New("account service error")

	setupGin := func(c *gin.Context, content interface{}) {
		c.Request.Method = http.MethodPost
		c.Request.Header.Set("Content-Type", "application/json")

		data, err := json.Marshal(content)
		require.NoError(t, err)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		request account.AddBalanceRequest
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		response            account.AddBalanceResponse
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid request",
			mock: func(service *MockService) {
			},
			args: args{
				request: account.AddBalanceRequest{
					AccountID: 1,
					Amount:    -1,
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Amount not added. request is not valid",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "account service error",
			mock: func(service *MockService) {
				service.EXPECT().AddBalance(ctx, fakeRequest.ToDTO()).Return(int64(0), accountServiceErr)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Add balance error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success add balance",
			mock: func(service *MockService) {
				service.EXPECT().AddBalance(ctx, fakeRequest.ToDTO()).Return(fakeBalance, nil)
			},
			args: args{
				request: fakeRequest,
			},
			response: account.AddBalanceResponse{
				Balance: fakeBalance,
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			accountHandler, accountService, c := mockHandler(t, w)

			setupGin(c, tt.args.request)
			tt.mock(accountService)

			accountHandler.AddBalance(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			} else {
				var response account.AddBalanceResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.response, response))
			}
		})
	}
}

func TestHandler_TransferBalance(t *testing.T) {
	ctx := context.Background()

	fakeRequest := account.TransferBalanceRequest{
		SenderID:   1,
		ReceiverID: 2,
		Amount:     100,
	}
	fakeResponse := account.TransferBalanceResponse{
		SenderBalance:   100,
		ReceiverBalance: 200,
	}
	accountServiceErr := errors.New("account service error")

	setupGin := func(c *gin.Context, content interface{}) {
		c.Request.Method = http.MethodPost
		c.Request.Header.Set("Content-Type", "application/json")

		data, err := json.Marshal(content)
		require.NoError(t, err)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		request account.TransferBalanceRequest
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		response            account.TransferBalanceResponse
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid request",
			mock: func(service *MockService) {
			},
			args: args{
				request: account.TransferBalanceRequest{
					SenderID:   1,
					ReceiverID: 2,
					Amount:     -1,
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Amount not transferred. request is not valid",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "account not found",
			mock: func(service *MockService) {
				service.EXPECT().
					TransferBalance(ctx, fakeRequest.ToDTO()).
					Return(int64(0), int64(0), account_domain.ErrNotFound)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Sender or receiver not found",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "account service error",
			mock: func(service *MockService) {
				service.EXPECT().TransferBalance(ctx, fakeRequest.ToDTO()).Return(int64(0), int64(0), accountServiceErr)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Transfer balance error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success get balance",
			mock: func(service *MockService) {
				service.EXPECT().
					TransferBalance(ctx, fakeRequest.ToDTO()).
					Return(fakeResponse.SenderBalance, fakeResponse.ReceiverBalance, nil)
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
			accountHandler, accountService, c := mockHandler(t, w)

			setupGin(c, tt.args.request)
			tt.mock(accountService)

			accountHandler.TransferBalance(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			} else {
				var response account.TransferBalanceResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.response, response))
			}
		})
	}
}
