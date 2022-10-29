package report_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	domain "github.com/maypok86/payment-api/internal/domain/report"
	"github.com/maypok86/payment-api/internal/handler/http/v1/report"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"github.com/maypok86/payment-api/internal/pkg/logger"
	"github.com/stretchr/testify/require"
)

func newFakeConfig() report.Config {
	return report.Config{
		ReportHost: "localhost",
		ReportPort: "8080",
	}
}

func mockHandler(t *testing.T, w http.ResponseWriter) (*report.Handler, *MockService, *gin.Context) {
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

	reportService := NewMockService(mockCtrl)
	reportHandler := report.NewHandler(newFakeConfig(), reportService, l)

	reportHandler.InitAPI(r.Group("/"))

	return reportHandler, reportService, c
}

func TestHandler_GetReportLink(t *testing.T) {
	ctx := context.Background()

	fakeRequest := report.GetReportLinkRequest{
		Month: 2,
		Year:  2022,
	}
	fakeKey := fmt.Sprintf("%d-%d", fakeRequest.Year, fakeRequest.Month)
	fakeCfg := newFakeConfig()
	fakeReportLink := fmt.Sprintf(
		"http://%s:%s/api/v1/report/download?key=%s",
		fakeCfg.ReportHost,
		fakeCfg.ReportPort,
		fakeKey,
	)
	reportServiceErr := errors.New("report service error")

	setupGin := func(c *gin.Context, content interface{}) {
		c.Request.Method = http.MethodPost
		c.Request.Header.Set("Content-Type", "application/json")

		data, err := json.Marshal(content)
		require.NoError(t, err)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		request report.GetReportLinkRequest
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		response            report.GetReportLinkResponse
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid request",
			mock: func(service *MockService) {
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Get report link error. Invalid request",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "not found report",
			mock: func(service *MockService) {
				service.EXPECT().GetReportKey(ctx, fakeRequest.ToDTO()).Return("", domain.ErrNotFound)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Get report link error. Report not found",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "report is not available yet",
			mock: func(service *MockService) {
				service.EXPECT().GetReportKey(ctx, fakeRequest.ToDTO()).Return("", domain.ErrIsNotAvailable)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Get report link error. Report is not available",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "report service error",
			mock: func(service *MockService) {
				service.EXPECT().GetReportKey(ctx, fakeRequest.ToDTO()).Return("", reportServiceErr)
			},
			args: args{
				request: fakeRequest,
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Get report link error. Internal server error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success get report link",
			mock: func(service *MockService) {
				service.EXPECT().GetReportKey(ctx, fakeRequest.ToDTO()).Return(fakeKey, nil)
			},
			args: args{
				request: fakeRequest,
			},
			response: report.GetReportLinkResponse{
				Link: fakeReportLink,
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			reportHandler, reportService, c := mockHandler(t, w)

			setupGin(c, tt.args.request)
			tt.mock(reportService)

			reportHandler.GetReportLink(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			} else {
				var response report.GetReportLinkResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.response, response))
			}
		})
	}
}

func TestHandler_DownloadReport(t *testing.T) {
	ctx := context.Background()

	fakeKey := "2022-2"
	reportServiceErr := errors.New("report service error")
	fakeReport := []byte("fake report")

	setupGin := func(c *gin.Context, queryParams map[string]string) {
		c.Request.Method = http.MethodGet
		c.Request.Header.Set("Content-Type", "application/json")

		query := url.Values{}
		for k, v := range queryParams {
			query.Add(k, v)
		}
		c.Request.URL.RawQuery = query.Encode()
	}

	type mockBehaviour func(service *MockService)

	type args struct {
		queryParams map[string]string
	}

	tests := []struct {
		name                string
		mock                mockBehaviour
		args                args
		response            []byte
		wantedErrorResponse *handler.ErrorResponse
		statusCode          int
	}{
		{
			name: "invalid request",
			mock: func(service *MockService) {
			},
			args: args{
				queryParams: map[string]string{},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Download report error. Invalid request",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "not found report",
			mock: func(service *MockService) {
				service.EXPECT().GetReportContent(ctx, fakeKey).Return(nil, domain.ErrNotFound)
			},
			args: args{
				queryParams: map[string]string{
					"key": fakeKey,
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Download report error. Report not found",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name: "report service error",
			mock: func(service *MockService) {
				service.EXPECT().GetReportContent(ctx, fakeKey).Return(nil, reportServiceErr)
			},
			args: args{
				queryParams: map[string]string{
					"key": fakeKey,
				},
			},
			wantedErrorResponse: &handler.ErrorResponse{
				Message: "Download report error. Internal server error",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success download report",
			mock: func(service *MockService) {
				service.EXPECT().GetReportContent(ctx, fakeKey).Return(fakeReport, nil)
			},
			args: args{
				queryParams: map[string]string{
					"key": fakeKey,
				},
			},
			response:   fakeReport,
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			reportHandler, reportService, c := mockHandler(t, w)

			setupGin(c, tt.args.queryParams)
			tt.mock(reportService)

			reportHandler.DownloadReport(c)

			require.Equal(t, tt.statusCode, w.Code)
			if tt.wantedErrorResponse != nil {
				var response handler.ErrorResponse
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
				require.True(t, reflect.DeepEqual(tt.wantedErrorResponse, &response))
			} else {
				got, err := io.ReadAll(w.Body)
				require.NoError(t, err)
				require.True(t, reflect.DeepEqual(tt.response, got))
			}
		})
	}
}
