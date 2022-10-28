package transaction_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/golang/mock/gomock"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/pkg/logger"
	"github.com/maypok86/payment-api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func mockService(t *testing.T) (*transaction.Service, *MockRepository) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := logger.New(os.Stdout, "debug")

	repository := NewMockRepository(mockCtrl)
	service := transaction.NewService(repository, l)

	return service, repository
}

func newTransaction(t *testing.T, transactionID int64, senderID int64) transaction.Transaction {
	t.Helper()

	return transaction.Transaction{
		TransactionID: transactionID,
		Type:          transaction.Transfer,
		SenderID:      senderID,
		ReceiverID:    2,
		Amount:        100,
		Description:   faker.Sentence(),
		CreatedAt:     time.Now(),
	}
}

func newTransactionsList(t *testing.T, count int, senderID int64) []transaction.Transaction {
	t.Helper()

	var list []transaction.Transaction

	for i := 0; i < count; i++ {
		list = append(list, newTransaction(t, int64(i+1), senderID))
	}

	return list
}

func TestService_GetTransactionsBySenderID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fakeCount := 10
	fakeSenderID := int64(1)
	transactionList := newTransactionsList(t, fakeCount, fakeSenderID)
	listParams, err := transaction.NewListParams("", "", pagination.Params{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)

	type mockBehavior func(r *MockRepository)

	type want struct {
		transactions []transaction.Transaction
		count        int
	}

	tests := []struct {
		name    string
		mock    mockBehavior
		want    want
		wantErr bool
	}{
		{
			name: "success get transactions by account id",
			mock: func(repository *MockRepository) {
				repository.EXPECT().
					GetTransactionsByAccountID(ctx, fakeSenderID, listParams).
					Return(transactionList, fakeCount, nil)
			},
			want: want{
				transactions: transactionList,
				count:        fakeCount,
			},
			wantErr: false,
		},
		{
			name: "repository error",
			mock: func(repository *MockRepository) {
				repository.EXPECT().
					GetTransactionsByAccountID(ctx, fakeSenderID, listParams).
					Return(nil, 0, errors.New("repository error"))
			},
			want:    want{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, repository := mockService(t)

			tt.mock(repository)

			gotTransactionList, gotCount, err := service.GetTransactionsByAccountID(ctx, fakeSenderID, listParams)
			require.True(t, (err != nil) == tt.wantErr)
			require.True(t, reflect.DeepEqual(tt.want.transactions, gotTransactionList))
			require.Equal(t, tt.want.count, gotCount)
		})
	}
}
