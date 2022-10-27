package account_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/maypok86/payment-api/internal/domain/account"
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

func mockService(t *testing.T, txErr error) (*account.Service, *MockRepository, *MockTransactionRepository) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	l := logger.New(os.Stdout, "debug")

	repository := NewMockRepository(mockCtrl)

	transactor := newFakeTransactor(txErr)
	transactionRepository := NewMockTransactionRepository(mockCtrl)
	service := account.NewService(transactor, repository, transactionRepository, l)

	return service, repository, transactionRepository
}

func TestService_GetBalanceByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fakeAccount := account.Account{
		AccountID: 1,
		Balance:   100,
	}

	type mockBehavior func(r *MockRepository)

	type args struct {
		accountID int64
	}

	tests := []struct {
		name    string
		mock    mockBehavior
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "success get balance by id",
			mock: func(repository *MockRepository) {
				repository.EXPECT().GetAccountByID(ctx, fakeAccount.AccountID).Return(fakeAccount, nil)
			},
			args: args{
				accountID: fakeAccount.AccountID,
			},
			want: fakeAccount.Balance,
		},
		{
			name: "repository error",
			mock: func(repository *MockRepository) {
				repository.EXPECT().
					GetAccountByID(ctx, fakeAccount.AccountID).
					Return(account.Account{}, errors.New("get account by id repository error"))
			},
			args: args{
				accountID: fakeAccount.AccountID,
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, repository, _ := mockService(t, nil)

			tt.mock(repository)

			got, err := service.GetBalanceByID(ctx, tt.args.accountID)
			require.True(t, (err != nil) == tt.wantErr)
			require.True(t, reflect.DeepEqual(tt.want, got))
		})
	}
}

func TestService_AddBalance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	fakeAccount := account.Account{
		AccountID: 1,
		Balance:   1,
	}
	dto := account.AddBalanceDTO{
		AccountID: 1,
		Amount:    100,
	}
	transactionDTO := transaction.CreateDTO{
		Type:        transaction.Enrollment,
		SenderID:    dto.AccountID,
		ReceiverID:  dto.AccountID,
		Amount:      dto.Amount,
		Description: fmt.Sprintf("Add %d kopecks to account with id = %d", dto.Amount, dto.AccountID),
	}
	txErr := errors.New("transaction error")
	repositoryErr := errors.New("repository error")
	transactionRepositoryErr := errors.New("transaction repository error")

	type args struct {
		dto            account.AddBalanceDTO
		transactionDTO transaction.CreateDTO
	}

	type mockBehavior func(r *MockRepository, tr *MockTransactionRepository)

	tests := []struct {
		name      string
		mock      mockBehavior
		args      args
		want      int64
		wantedErr error
		txErr     error
	}{
		{
			name: "success add balance",
			mock: func(repository *MockRepository, transactionRepository *MockTransactionRepository) {
				repository.EXPECT().AddBalance(ctx, dto).Return(fakeAccount.Balance+dto.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(nil)
			},
			args: args{
				dto:            dto,
				transactionDTO: transactionDTO,
			},
			want:      fakeAccount.Balance + dto.Amount,
			wantedErr: nil,
			txErr:     nil,
		},
		{
			name: "repository error",
			mock: func(repository *MockRepository, transactionRepository *MockTransactionRepository) {
				repository.EXPECT().AddBalance(ctx, dto).Return(int64(0), repositoryErr)
			},
			args: args{
				dto:            dto,
				transactionDTO: transactionDTO,
			},
			want:      0,
			wantedErr: repositoryErr,
			txErr:     nil,
		},
		{
			name: "transaction repository error",
			mock: func(repository *MockRepository, transactionRepository *MockTransactionRepository) {
				repository.EXPECT().AddBalance(ctx, dto).Return(fakeAccount.Balance+dto.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(transactionRepositoryErr)
			},
			args: args{
				dto:            dto,
				transactionDTO: transactionDTO,
			},
			want:      0,
			wantedErr: transactionRepositoryErr,
			txErr:     nil,
		},
		{
			name: "transaction error",
			mock: func(repository *MockRepository, transactionRepository *MockTransactionRepository) {
				repository.EXPECT().AddBalance(ctx, dto).Return(fakeAccount.Balance+dto.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(nil)
			},
			args: args{
				dto:            dto,
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

			service, repository, transactionRepository := mockService(t, tt.txErr)

			tt.mock(repository, transactionRepository)
			got, err := service.AddBalance(ctx, tt.args.dto)
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

func TestService_TransferBalance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	senderAccount := account.Account{
		AccountID: 1,
		Balance:   100,
	}
	receiverAccount := account.Account{
		AccountID: 2,
		Balance:   10,
	}
	dto := account.TransferBalanceDTO{
		SenderID:   senderAccount.AccountID,
		ReceiverID: receiverAccount.AccountID,
		Amount:     50,
	}
	transactionDTO := transaction.CreateDTO{
		Type:       transaction.Transfer,
		SenderID:   dto.SenderID,
		ReceiverID: dto.ReceiverID,
		Amount:     dto.Amount,
		Description: fmt.Sprintf(
			"Transfer %d kopecks from account with id = %d to account with id = %d",
			dto.Amount,
			dto.SenderID,
			dto.ReceiverID,
		),
	}
	txErr := errors.New("transaction error")
	repositoryErr := errors.New("repository error")
	transactionRepositoryErr := errors.New("transaction repository error")

	type args struct {
		dto            account.TransferBalanceDTO
		transactionDTO transaction.CreateDTO
	}

	type mockBehavior func(r *MockRepository, tr *MockTransactionRepository)

	type balancies struct {
		senderBalance   int64
		receiverBalance int64
	}

	tests := []struct {
		name      string
		mock      mockBehavior
		args      args
		want      balancies
		wantedErr error
		txErr     error
	}{
		{
			name: "success transfer balance",
			mock: func(repository *MockRepository, transactionRepository *MockTransactionRepository) {
				repository.EXPECT().
					TransferBalance(ctx, dto).
					Return(senderAccount.Balance-dto.Amount, receiverAccount.Balance+dto.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(nil)
			},
			args: args{
				dto:            dto,
				transactionDTO: transactionDTO,
			},
			want: balancies{
				senderBalance:   senderAccount.Balance - dto.Amount,
				receiverBalance: receiverAccount.Balance + dto.Amount,
			},
			wantedErr: nil,
			txErr:     nil,
		},
		{
			name: "repository error",
			mock: func(repository *MockRepository, transactionRepository *MockTransactionRepository) {
				repository.EXPECT().TransferBalance(ctx, dto).Return(int64(0), int64(0), repositoryErr)
			},
			args: args{
				dto:            dto,
				transactionDTO: transactionDTO,
			},
			want:      balancies{},
			wantedErr: repositoryErr,
			txErr:     nil,
		},
		{
			name: "transaction repository error",
			mock: func(repository *MockRepository, transactionRepository *MockTransactionRepository) {
				repository.EXPECT().
					TransferBalance(ctx, dto).
					Return(senderAccount.Balance-dto.Amount, receiverAccount.Balance+dto.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(transactionRepositoryErr)
			},
			args: args{
				dto:            dto,
				transactionDTO: transactionDTO,
			},
			want:      balancies{},
			wantedErr: transactionRepositoryErr,
			txErr:     nil,
		},
		{
			name: "transaction error",
			mock: func(repository *MockRepository, transactionRepository *MockTransactionRepository) {
				repository.EXPECT().
					TransferBalance(ctx, dto).
					Return(senderAccount.Balance-dto.Amount, receiverAccount.Balance+dto.Amount, nil)
				transactionRepository.EXPECT().CreateTransaction(ctx, transactionDTO).Return(nil)
			},
			args: args{
				dto:            dto,
				transactionDTO: transactionDTO,
			},
			want:      balancies{},
			wantedErr: nil,
			txErr:     txErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, repository, transactionRepository := mockService(t, tt.txErr)

			tt.mock(repository, transactionRepository)
			gotSenderBalance, gotReceiverBalance, err := service.TransferBalance(ctx, tt.args.dto)
			if tt.wantedErr != nil {
				require.ErrorIs(t, err, tt.wantedErr)
			}
			if tt.txErr != nil {
				require.ErrorIs(t, err, tt.txErr)
			}
			require.True(t, reflect.DeepEqual(tt.want.senderBalance, gotSenderBalance))
			require.True(t, reflect.DeepEqual(tt.want.receiverBalance, gotReceiverBalance))
		})
	}
}
