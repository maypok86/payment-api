package pagination

const (
	DefaultLimit = 10
	MaxLimit     = 100

	DefaultOffset = 0
)

type Params struct {
	Limit  uint64
	Offset uint64
}

type ListRange struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

func NewListRange(params Params, count int) ListRange {
	return ListRange{
		Limit:  int(params.Limit),
		Offset: int(params.Offset),
		Count:  count,
	}
}
