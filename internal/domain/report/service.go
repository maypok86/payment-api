package report

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

//go:generate mockgen -source=service.go -destination=mock_test.go -package=report_test

type Repository interface {
	GetReportMap(ctx context.Context, dto GetMapDTO) (map[int64]int64, error)
}

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	IsExist(key string) bool
}

type Service struct {
	repository Repository
	cache      Cache
	logger     *zap.Logger
}

func NewService(repository Repository, cache Cache, logger *zap.Logger) *Service {
	return &Service{
		repository: repository,
		cache:      cache,
		logger:     logger,
	}
}

func reportMapToCSV(reportMap map[int64]int64) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	ids := make([]int64, 0, len(reportMap))
	for id := range reportMap {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	if err := writer.Write([]string{"service_id", "amount"}); err != nil {
		return nil, fmt.Errorf("report map to csv: %w", err)
	}
	for _, serviceID := range ids {
		if err := writer.Write([]string{fmt.Sprintf("%d", serviceID), fmt.Sprintf("%d", reportMap[serviceID])}); err != nil {
			return nil, fmt.Errorf("report map to csv: %w", err)
		}
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("report map to csv: %w", err)
	}

	return buffer.Bytes(), nil
}

func (s *Service) GetReportKey(ctx context.Context, dto GetMapDTO) (string, error) {
	year, month, _ := time.Now().Date()
	if dto.Year > int64(year) || (dto.Year == int64(year) && dto.Month > int64(month)) {
		return "", fmt.Errorf("get report key: %w", ErrIsNotAvailable)
	}

	key := fmt.Sprintf("%d-%d", dto.Year, dto.Month)
	if s.cache.IsExist(key) {
		return key, nil
	}

	reportMap, err := s.repository.GetReportMap(ctx, dto)
	if err != nil {
		return "", fmt.Errorf("get report key: %w", err)
	}
	if len(reportMap) == 0 {
		return "", ErrNotFound
	}

	reportContent, err := reportMapToCSV(reportMap)
	if err != nil {
		return "", fmt.Errorf("get report key: %w", err)
	}

	if err := s.cache.Set(key, reportContent); err != nil {
		return "", fmt.Errorf("set report data: %w", err)
	}

	return key, nil
}

func (s *Service) GetReportContent(ctx context.Context, key string) ([]byte, error) {
	reportContent, err := s.cache.Get(key)
	if err != nil {
		return nil, fmt.Errorf("get report content: %w", err)
	}

	return reportContent, nil
}
