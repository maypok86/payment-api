package transaction_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	domain "github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/handler/http/v1/transaction"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"github.com/maypok86/payment-api/internal/pkg/logger"
	"github.com/maypok86/payment-api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func mockHandler(t *testing.T, w http.ResponseWriter) (*transaction.Handler, *MockService, *gin.Context) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	gin.SetMode(gin.TestMode)

	c, r := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}

	l := logger.New(os.Stdout, "debug")

	transactionService := NewMockService(mockCtrl)
	transactionHandler := transaction.NewHandler(transactionService, l)

	transactionHandler.InitAPI(r.Group("/"))

	return transactionHandler, transactionService, c
}

func newTransaction(t *testing.T, transactionID, senderID int64) domain.Transaction {
	t.Helper()

	return domain.Transaction{
		TransactionID: transactionID,
		Type:          domain.Transfer,
		SenderID:      senderID,
		ReceiverID:    2,
		Amount:        100,
		Description:   faker.Sentence(),
		CreatedAt:     time.Now(),
	}
}

func newTransactions(t *testing.T, senderID int64, count int) []domain.Transaction {
	t.Helper()

	var transactions []domain.Transaction

	for i := 0; i < count; i++ {
		transactions = append(transactions, newTransaction(t, int64(i+1), senderID))
	}

	return transactions
}

func newListResponse(
	t *testing.T,
	transactions []domain.Transaction,
	params pagination.Params,
	count int,
) transaction.ListResponse {
	t.Helper()

	fakeResponse := transaction.NewListResponse(transactions, params, count)

	// for fix time.Time in json
	var buffer bytes.Buffer
	require.NoError(t, json.NewEncoder(&buffer).Encode(fakeResponse))
	require.NoError(t, json.NewDecoder(&buffer).Decode(&fakeResponse))

	return fakeResponse
}

func TestHandler_GetTransactionBySenderID(t *testing.T) {
	ctx := context.Background()

	fakeSenderID := int64(1)
	fakeParam := fmt.Sprintf("%d", fakeSenderID)
	fakePaginationParams := pagination.Params{
		Limit:  1,
		Offset: 0,
	}
	fakeQueryParams := map[string]string{
		"limit":     fmt.Sprintf("%d", fakePaginationParams.Limit),
		"offset":    fmt.Sprintf("%d", fakePaginationParams.Offset),
		"sort":      "date",
		"direction": "asc",
	}
	fakeListParams, err := domain.NewListParams(
		fakeQueryParams["sort"],
		fakeQueryParams["direction"],
		fakePaginationParams,
	)
	require.NoError(t, err)
	fakeTransactions := newTransactions(t, fakeSenderID, 10)

	transactionServiceErr := errors.New("transaction service error")

	setupGin := func(c *gin.Context, param string, queryParams map[string]string) {
		c.Request.Method = http.MethodGet
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "sender_id", Value: param}}

		query := url.Values{}
		for k, v := range queryParams {
			query.Add(k, v)
		}
		c.Request.URL.RawQuery = query.Encode()
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		param       string
		queryParams map[string]string
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		response            transaction.ListResponse
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid sender_id param",
			mock: func(service *MockService) {
			},
			args: args{
				param: "invalid",
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Transactions not found. id is not valid",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "invalid pagination params",
			mock: func(service *MockService) {
			},
			args: args{
				param: fakeParam,
				queryParams: map[string]string{
					"limit": "invalid",
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Transactions not found. Pagination params is not valid",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "invalid sort param",
			mock: func(service *MockService) {
			},
			args: args{
				param: fakeParam,
				queryParams: map[string]string{
					"sort": "invalid",
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Transactions not found. Sort param is not valid",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "invalid direction param",
			mock: func(service *MockService) {
			},
			args: args{
				param: fakeParam,
				queryParams: map[string]string{
					"sort":      "date",
					"direction": "invalid",
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Transactions not found. Direction param is not valid",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "transaction service error",
			mock: func(service *MockService) {
				service.EXPECT().
					GetTransactionsBySenderID(ctx, fakeSenderID, fakeListParams).
					Return(nil, 0, transactionServiceErr)
			},
			args: args{
				param:       fakeParam,
				queryParams: fakeQueryParams,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Get transactions by sender id error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success get transactions by sender id",
			mock: func(service *MockService) {
				service.EXPECT().
					GetTransactionsBySenderID(ctx, fakeSenderID, fakeListParams).
					Return(fakeTransactions, len(fakeTransactions), nil)
			},
			args: args{
				param:       fakeParam,
				queryParams: fakeQueryParams,
			},
			response:   newListResponse(t, fakeTransactions, fakePaginationParams, len(fakeTransactions)),
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			transactionHandler, transactionService, c := mockHandler(t, w)

			setupGin(c, tt.args.param, tt.args.queryParams)
			tt.mock(transactionService)

			transactionHandler.GetTransactionsBySenderID(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			} else {
				var response transaction.ListResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				fmt.Println(response)
				fmt.Println(tt.response)
				require.True(t, reflect.DeepEqual(tt.response, response))
			}
		})
	}
}
