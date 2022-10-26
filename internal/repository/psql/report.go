package psql

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
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
		tableName: "reports",
		db:        db,
		logger:    logger,
	}
}

func (rr *ReportRepository) CreateReport(ctx context.Context, dto report.CreateDTO) (report.Report, error) {
	sql, args, err := rr.db.Builder.Insert(rr.tableName).
		Columns("service_id", "amount").
		Values(dto.ServiceID, dto.Amount).
		Suffix("RETURNING report_id, created_at").
		ToSql()
	if err != nil {
		return report.Report{}, fmt.Errorf("build create report query: %w", err)
	}

	rr.logger.Debug("create report query", zap.String("sql", sql), zap.Any("args", args))

	entity := report.Report{
		ServiceID: dto.ServiceID,
		Amount:    dto.Amount,
	}
	if err := rr.db.QueryRow(ctx, sql, args...).Scan(&entity.ReportID, &entity.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return report.Report{}, fmt.Errorf("insert report: %w", report.ErrAlreadyExist)
			}
		}

		return report.Report{}, fmt.Errorf("insert report: %w", err)
	}

	return entity, nil
}

func (rr *ReportRepository) addAmount(ctx context.Context, dto report.AddAmountDTO) (int64, error) {
	sql, args, err := rr.db.Builder.Update(rr.tableName).
		Set("amount", sq.Expr("amount + ?", dto.Amount)).
		Where(sq.And{
			sq.Eq{"service_id": dto.ServiceID},
			sq.Eq{"DATE_PART('year', created_at)": dto.Year},
			sq.Eq{"DATE_PART('month', created_at)": dto.Month},
		}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build add amount query: %w", err)
	}

	rr.logger.Debug("add amount query", zap.String("sql", sql), zap.Any("args", args))

	var amount int64
	if err := rr.db.QueryRow(ctx, sql, args...).Scan(&amount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, report.ErrNotFound
		}

		return 0, err
	}

	return amount, nil
}

func (rr *ReportRepository) AddAmount(ctx context.Context, dto report.AddAmountDTO) (int64, error) {
	amount, err := rr.addAmount(ctx, dto)
	if err == nil {
		return amount, nil
	}

	if errors.Is(err, report.ErrNotFound) {
		entity, err := rr.CreateReport(ctx, report.CreateDTO(dto))
		if err != nil {
			return 0, err
		}
		return entity.Amount, nil
	}
	return 0, err
}

/*
func (rr *ReportRepository) GetReportMap(ctx context.Context, dto report.GetDTO) (map[int64]int64, error) {
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

*/
