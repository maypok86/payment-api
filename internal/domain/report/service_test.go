package report_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/maypok86/payment-api/internal/domain/report"
	"github.com/maypok86/payment-api/internal/pkg/logger"
	"github.com/stretchr/testify/require"
)

func mockService(t *testing.T) (*report.Service, *MockRepository, *MockCache) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := logger.New(os.Stdout, "debug")

	repository := NewMockRepository(mockCtrl)
	cache := NewMockCache(mockCtrl)
	service := report.NewService(repository, cache, l)

	return service, repository, cache
}

func newReportMap(t *testing.T, count int) map[int64]int64 {
	t.Helper()

	reportMap := make(map[int64]int64, count)

	for i := 0; i < count; i++ {
		reportMap[int64(i)] = int64(i)
	}

	return reportMap
}

func TestService_GetReportKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dto := report.GetMapDTO{
		Year:  2021,
		Month: 9,
	}
	fakeKey := fmt.Sprintf("%d-%d", dto.Year, dto.Month)
	fakeReportMap := newReportMap(t, 10)
	repositoryErr := errors.New("repository error")
	cacheErr := errors.New("cache error")

	type args struct {
		dto report.GetMapDTO
	}

	type mockBehavior func(repository *MockRepository, cache *MockCache)

	tests := []struct {
		name      string
		mock      mockBehavior
		args      args
		want      string
		wantedErr error
	}{
		{
			name: "success get report key",
			mock: func(repository *MockRepository, cache *MockCache) {
				cache.EXPECT().IsExist(fakeKey).Return(false)
				repository.EXPECT().GetReportMap(ctx, dto).Return(fakeReportMap, nil)
				cache.EXPECT().Set(fakeKey, gomock.Any()).Return(nil)
			},
			args: args{
				dto: dto,
			},
			want:      fakeKey,
			wantedErr: nil,
		},
		{
			name: "report is not available",
			mock: func(repository *MockRepository, cache *MockCache) {
			},
			args: args{
				dto: report.GetMapDTO{
					Year:  2025,
					Month: 9,
				},
			},
			want:      "",
			wantedErr: report.ErrIsNotAvailable,
		},
		{
			name: "key is exist in cache",
			mock: func(repository *MockRepository, cache *MockCache) {
				cache.EXPECT().IsExist(fakeKey).Return(true)
			},
			args: args{
				dto: dto,
			},
			want:      fakeKey,
			wantedErr: nil,
		},
		{
			name: "repository error",
			mock: func(repository *MockRepository, cache *MockCache) {
				cache.EXPECT().IsExist(fakeKey).Return(false)
				repository.EXPECT().GetReportMap(ctx, dto).Return(nil, repositoryErr)
			},
			args: args{
				dto: dto,
			},
			want:      "",
			wantedErr: repositoryErr,
		},
		{
			name: "empty report",
			mock: func(repository *MockRepository, cache *MockCache) {
				cache.EXPECT().IsExist(fakeKey).Return(false)
				repository.EXPECT().GetReportMap(ctx, dto).Return(nil, nil)
			},
			args: args{
				dto: dto,
			},
			want:      "",
			wantedErr: report.ErrNotFound,
		},
		{
			name: "cache set error",
			mock: func(repository *MockRepository, cache *MockCache) {
				cache.EXPECT().IsExist(fakeKey).Return(false)
				repository.EXPECT().GetReportMap(ctx, dto).Return(fakeReportMap, nil)
				cache.EXPECT().Set(fakeKey, gomock.Any()).Return(cacheErr)
			},
			args: args{
				dto: dto,
			},
			want:      "",
			wantedErr: cacheErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, repository, cache := mockService(t)

			tt.mock(repository, cache)

			got, err := service.GetReportKey(ctx, tt.args.dto)
			if err != nil {
				require.ErrorIs(t, err, tt.wantedErr)
			}
			require.True(t, reflect.DeepEqual(tt.want, got))
		})
	}
}

func TestService_GetReportContent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fakeKey := faker.Word()
	data := []byte(faker.Sentence())

	type args struct {
		key string
	}

	type mockBehavior func(cache *MockCache)

	tests := []struct {
		name    string
		mock    mockBehavior
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "success get report content",
			mock: func(cache *MockCache) {
				cache.EXPECT().Get(fakeKey).Return(data, nil)
			},
			args: args{
				key: fakeKey,
			},
			want:    data,
			wantErr: false,
		},
		{
			name: "cache error",
			mock: func(cache *MockCache) {
				cache.EXPECT().Get(fakeKey).Return(nil, errors.New("cache error"))
			},
			args: args{
				key: fakeKey,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, _, cache := mockService(t)

			tt.mock(cache)

			got, err := service.GetReportContent(ctx, tt.args.key)
			require.True(t, (err != nil) == tt.wantErr)
			require.True(t, reflect.DeepEqual(tt.want, got))
		})
	}
}
