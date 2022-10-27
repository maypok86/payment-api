package psql

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/maypok86/payment-api/internal/domain/report"
	"github.com/maypok86/payment-api/internal/pkg/postgres"
	"go.uber.org/zap"
)

type ReportRepository struct {
	tableName string
	db        *postgres.Client
	logger    *zap.Logger
}

func NewReportRepository(db *postgres.Client, logger *zap.Logger) *ReportRepository {
	return &ReportRepository{
		tableName: "orders",
		db:        db,
		logger:    logger,
	}
}

func (rr *ReportRepository) GetReportMap(ctx context.Context, dto report.GetMapDTO) (map[int64]int64, error) {
	sql, args, err := rr.db.Builder.Select("service_id", "SUM(amount)").From(rr.tableName).Where(sq.And{
		sq.Eq{"is_paid": true},
		sq.Eq{"is_cancelled": false},
		sq.Eq{"DATE_PART('year', created_at)": dto.Year},
		sq.Eq{"DATE_PART('month', created_at)": dto.Month},
	}).GroupBy("service_id").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get report query: %w", err)
	}

	rr.logger.Debug("get report query", zap.String("sql", sql), zap.Any("args", args))

	rows, err := rr.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("run get report query: %w", err)
	}

	reportMap := make(map[int64]int64)
	for rows.Next() {
		var serviceID int64
		var amount int64
		if err := rows.Scan(&serviceID, &amount); err != nil {
			return nil, fmt.Errorf("scan report row: %w", err)
		}

		reportMap[serviceID] = amount
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read report: %w", err)
	}

	return reportMap, nil
}
