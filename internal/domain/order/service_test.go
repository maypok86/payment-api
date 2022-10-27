package order_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/domain/order"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/pkg/logger"
	"github.com/stretchr/testify/require"
)

type fakeTransactor struct {
	txErr error
}

func newFakeTransactor(txErr error) fakeTransactor {
	return fakeTransactor{txErr: txErr}
}

func (ft fakeTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	err := fn(ctx)

	if ft.txErr != nil {
		return ft.txErr
	}

	return err
}

func mockService(
	t *testing.T,
	txErr error,
) (*order.Service, *MockRepository, *MockTransactionRepository, *MockAccountRepository) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := logger.New(os.Stdout, "debug")

	transactor := newFakeTransactor(txErr)
	repository := NewMockRepository(mockCtrl)
	accountRepository := NewMockAccountRepository(mockCtrl)
	transactionRepository := NewMockTransactionRepository(mockCtrl)
	service := order.NewService(transactor, repository, transactionRepository, accountRepository, l)

	return service, repository, transactionRepository, accountRepository
}

func TestService_CreateOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fakeOrder := order.Order{
		OrderID:     1,
		AccountID:   1,
		ServiceID:   1,
		Amount:      100,
		IsPaid:      false,
		IsCancelled: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	fakeBalance := int64(1000)
	dto := order.CreateDTO{
		OrderID:   fakeOrder.OrderID,
		AccountID: fakeOrder.AccountID,
		ServiceID: fakeOrder.ServiceID,
		Amount:    fakeOrder.Amount,
	}
	accountDTO := account.ReserveBalanceDTO{
		AccountID: fakeOrder.AccountID,
		Amount:    fakeOrder.Amount,
	}
	transactionDTO := transaction.CreateDTO{
		Type:        transaction.Reservation,
		SenderID:    dto.AccountID,
		ReceiverID:  dto.AccountID,
		Amount:      dto.Amount,
		Description: fmt.Sprintf("Reserve %d kopecks for order with id = %d", dto.Amount, dto.OrderID),
	}
	txErr := errors.New("transaction error")
	repositoryErr := errors.New("repository error")
	accountRepositoryErr := errors.New("account repository error")
	transactionRepositoryErr := errors.New("transaction repository error")

	type args struct {
		dto            order.CreateDTO
		accountDTO     account.ReserveBalanceDTO
		transactionDTO transaction.CreateDTO
	}

	type mockBehavior func(r *MockRepository, tr *MockTransactionRepository, ar *MockAccountRepository)

	type want struct {
		order   order.Order
		balance int64
	}

	tests := []struct {
		name      string
		mock      mockBehavior
		args      args
		want      want
		wantedErr error
		txErr     error
	}{
		{
			name: "success create order",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				accountRepository.EXPECT().ReserveBalance(ctx, accountDTO).Return(fakeBalance-accountDTO.Amount, nil)
				repository.EXPECT().CreateOrder(ctx, dto).Return(fakeOrder, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(nil)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want: want{
				order:   fakeOrder,
				balance: fakeBalance - accountDTO.Amount,
			},
			wantedErr: nil,
			txErr:     nil,
		},
		{
			name: "account repository error",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				accountRepository.EXPECT().ReserveBalance(ctx, accountDTO).Return(int64(0), accountRepositoryErr)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      want{},
			wantedErr: accountRepositoryErr,
			txErr:     nil,
		},
		{
			name: "repository error",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				accountRepository.EXPECT().ReserveBalance(ctx, accountDTO).Return(fakeBalance-accountDTO.Amount, nil)
				repository.EXPECT().CreateOrder(ctx, dto).Return(order.Order{}, repositoryErr)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      want{},
			wantedErr: repositoryErr,
			txErr:     nil,
		},
		{
			name: "transaction repository error",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				accountRepository.EXPECT().ReserveBalance(ctx, accountDTO).Return(fakeBalance-accountDTO.Amount, nil)
				repository.EXPECT().CreateOrder(ctx, dto).Return(fakeOrder, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(transactionRepositoryErr)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      want{},
			wantedErr: transactionRepositoryErr,
			txErr:     nil,
		},
		{
			name: "transaction error",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				accountRepository.EXPECT().ReserveBalance(ctx, accountDTO).Return(fakeBalance-accountDTO.Amount, nil)
				repository.EXPECT().CreateOrder(ctx, dto).Return(fakeOrder, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(nil)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      want{},
			wantedErr: nil,
			txErr:     txErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, repository, transactionRepository, accountRepository := mockService(t, tt.txErr)

			tt.mock(repository, transactionRepository, accountRepository)
			gotOrder, gotBalance, err := service.CreateOrder(ctx, tt.args.dto)
			if tt.wantedErr != nil {
				require.ErrorIs(t, err, tt.wantedErr)
			}
			if tt.txErr != nil {
				require.ErrorIs(t, err, tt.txErr)
			}
			require.True(t, reflect.DeepEqual(tt.want.order, gotOrder))
			require.True(t, reflect.DeepEqual(tt.want.balance, gotBalance))
		})
	}
}

func TestService_PayForOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dto := order.PayForDTO{
		OrderID:   1,
		AccountID: 1,
		ServiceID: 1,
		Amount:    100,
	}

	type mockBehavior func(r *MockRepository)

	type args struct {
		dto order.PayForDTO
	}

	tests := []struct {
		name    string
		mock    mockBehavior
		args    args
		wantErr bool
	}{
		{
			name: "success pay for order",
			mock: func(repository *MockRepository) {
				repository.EXPECT().PayForOrder(ctx, dto).Return(nil)
			},
			args: args{
				dto: dto,
			},
			wantErr: false,
		},
		{
			name: "repository error",
			mock: func(repository *MockRepository) {
				repository.EXPECT().PayForOrder(ctx, dto).Return(errors.New("pay for order repository error"))
			},
			args: args{
				dto: dto,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, repository, _, _ := mockService(t, nil)

			tt.mock(repository)

			err := service.PayForOrder(ctx, tt.args.dto)
			require.True(t, (err != nil) == tt.wantErr)
		})
	}
}

func TestService_CancelOrder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fakeBalance := int64(1000)
	dto := order.CancelDTO{
		OrderID:   1,
		AccountID: 1,
		ServiceID: 1,
		Amount:    100,
	}
	accountDTO := account.ReturnBalanceDTO{
		AccountID: dto.AccountID,
		Amount:    dto.Amount,
	}
	transactionDTO := transaction.CreateDTO{
		Type:        transaction.CancelReservation,
		SenderID:    dto.AccountID,
		ReceiverID:  dto.AccountID,
		Amount:      dto.Amount,
		Description: fmt.Sprintf("Cancel reservation %d kopecks for order with id = %d", dto.Amount, dto.OrderID),
	}
	txErr := errors.New("transaction error")
	repositoryErr := errors.New("repository error")
	accountRepositoryErr := errors.New("account repository error")
	transactionRepositoryErr := errors.New("transaction repository error")

	type args struct {
		dto            order.CancelDTO
		accountDTO     account.ReturnBalanceDTO
		transactionDTO transaction.CreateDTO
	}

	type mockBehavior func(r *MockRepository, tr *MockTransactionRepository, ar *MockAccountRepository)

	tests := []struct {
		name      string
		mock      mockBehavior
		args      args
		want      int64
		wantedErr error
		txErr     error
	}{
		{
			name: "success cancel order",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				repository.EXPECT().CancelOrder(ctx, dto).Return(nil)
				accountRepository.EXPECT().ReturnBalance(ctx, accountDTO).Return(fakeBalance+accountDTO.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(nil)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      fakeBalance + accountDTO.Amount,
			wantedErr: nil,
			txErr:     nil,
		},
		{
			name: "repository error",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				repository.EXPECT().CancelOrder(ctx, dto).Return(repositoryErr)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      0,
			wantedErr: repositoryErr,
			txErr:     nil,
		},
		{
			name: "account repository error",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				repository.EXPECT().CancelOrder(ctx, dto).Return(nil)
				accountRepository.EXPECT().ReturnBalance(ctx, accountDTO).Return(int64(0), accountRepositoryErr)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      0,
			wantedErr: accountRepositoryErr,
			txErr:     nil,
		},
		{
			name: "transaction repository error",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				repository.EXPECT().CancelOrder(ctx, dto).Return(nil)
				accountRepository.EXPECT().ReturnBalance(ctx, accountDTO).Return(fakeBalance+accountDTO.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(transactionRepositoryErr)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      0,
			wantedErr: transactionRepositoryErr,
			txErr:     nil,
		},
		{
			name: "transaction error",
			mock: func(
				repository *MockRepository,
				transactionRepository *MockTransactionRepository,
				accountRepository *MockAccountRepository,
			) {
				repository.EXPECT().CancelOrder(ctx, dto).Return(nil)
				accountRepository.EXPECT().ReturnBalance(ctx, accountDTO).Return(fakeBalance+accountDTO.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(nil)
			},
			args: args{
				dto:            dto,
				accountDTO:     accountDTO,
				transactionDTO: transactionDTO,
			},
			want:      0,
			wantedErr: nil,
			txErr:     txErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, repository, transactionRepository, accountRepository := mockService(t, tt.txErr)

			tt.mock(repository, transactionRepository, accountRepository)
			got, err := service.CancelOrder(ctx, tt.args.dto)
			if tt.wantedErr != nil {
				require.ErrorIs(t, err, tt.wantedErr)
			}
			if tt.txErr != nil {
				require.ErrorIs(t, err, tt.txErr)
			}
			require.True(t, reflect.DeepEqual(tt.want, got))
		})
	}
}
