package report

import "github.com/maypok86/payment-api/internal/domain/report"

type GetReportLinkRequest struct {
	Month int64 `json:"month" binding:"required,min=1,max=12"`
	Year  int64 `json:"year"  binding:"required,min=2022"`
}

func (r GetReportLinkRequest) ToDTO() report.GetMapDTO {
	return report.GetMapDTO{
		Month: r.Month,
		Year:  r.Year,
	}
}
