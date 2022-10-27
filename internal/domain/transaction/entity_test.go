package transaction_test

import (
	"database/sql/driver"
	"reflect"
	"testing"

	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/pkg/pagination"
	"github.com/maypok86/payment-api/internal/pkg/sort"
	"github.com/stretchr/testify/require"
)

func TestType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		transactionType transaction.Type
		want            string
	}{
		{
			name:            "check enrollment",
			transactionType: transaction.Enrollment,
			want:            "enrollment",
		},
		{
			name:            "check transfer",
			transactionType: transaction.Transfer,
			want:            "transfer",
		},
		{
			name:            "check reservation",
			transactionType: transaction.Reservation,
			want:            "reservation",
		},
		{
			name:            "check cancel_reservation",
			transactionType: transaction.CancelReservation,
			want:            "cancel_reservation",
		},
		{
			name:            "check empty",
			transactionType: transaction.Type{},
			want:            "",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tt.transactionType.String())
		})
	}
}

func TestType_Scan(t *testing.T) {
	t.Parallel()

	type args struct {
		value interface{}
	}

	tests := []struct {
		name    string
		args    args
		want    transaction.Type
		wantErr bool
	}{
		{
			name: "success scan",
			args: args{
				value: "enrollment",
			},
			want:    transaction.Enrollment,
			wantErr: false,
		},
		{
			name: "wrong value for Type",
			args: args{
				value: "wrong",
			},
			want:    transaction.Type{},
			wantErr: true,
		},
		{
			name: "scan source is not string",
			args: args{
				value: 1,
			},
			want:    transaction.Type{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var transactionType transaction.Type

			err := transactionType.Scan(tt.args.value)
			require.True(t, (err != nil) == tt.wantErr)
			require.True(t, reflect.DeepEqual(transactionType, tt.want))
		})
	}
}

func TestType_Value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		transactionType transaction.Type
		want            driver.Value
		wantErr         bool
	}{
		{
			name:            "check enrollment",
			transactionType: transaction.Enrollment,
			want:            "enrollment",
			wantErr:         false,
		},
		{
			name:            "check transfer",
			transactionType: transaction.Transfer,
			want:            "transfer",
			wantErr:         false,
		},
		{
			name:            "check reservation",
			transactionType: transaction.Reservation,
			want:            "reservation",
			wantErr:         false,
		},
		{
			name:            "check cancel_reservation",
			transactionType: transaction.CancelReservation,
			want:            "cancel_reservation",
			wantErr:         false,
		},
		{
			name:            "check empty",
			transactionType: transaction.Type{},
			want:            nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.transactionType.Value()
			require.True(t, (err != nil) == tt.wantErr)
			require.True(t, reflect.DeepEqual(got, tt.want))
		})
	}
}

func TestListParams_NewListParams(t *testing.T) {
	t.Parallel()

	type args struct {
		sortParam      string
		directionParam string
		params         pagination.Params
	}

	tests := []struct {
		name      string
		args      args
		want      transaction.ListParams
		wantedErr error
	}{
		{
			name: "success new list params",
			args: args{
				sortParam: "",
				params: pagination.Params{
					Limit:  0,
					Offset: 0,
				},
			},
			want: transaction.ListParams{
				Sort: nil,
				Pagination: pagination.Params{
					Limit:  pagination.DefaultLimit,
					Offset: 0,
				},
			},
			wantedErr: nil,
		},
		{
			name: "invalid sort param",
			args: args{
				sortParam: "invalid",
			},
			want:      transaction.ListParams{},
			wantedErr: transaction.ErrInvalidSortParam,
		},
		{
			name: "invalid direction param",
			args: args{
				sortParam:      "date",
				directionParam: "invalid",
			},
			want:      transaction.ListParams{},
			wantedErr: transaction.ErrInvalidDirectionParam,
		},
		{
			name: "success new list params with sort",
			args: args{
				sortParam:      "date",
				directionParam: "",
			},
			want: transaction.ListParams{
				Sort: sort.New("created_at", "asc"),
				Pagination: pagination.Params{
					Limit: pagination.DefaultLimit,
				},
			},
			wantedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := transaction.NewListParams(tt.args.sortParam, tt.args.directionParam, tt.args.params)
			if err != nil {
				require.ErrorIs(t, err, tt.wantedErr)
			}
			require.True(t, reflect.DeepEqual(got, tt.want))
		})
	}
}
