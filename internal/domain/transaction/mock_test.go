// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package transaction_test is a generated GoMock package.
package transaction_test

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	transaction "github.com/maypok86/payment-api/internal/domain/transaction"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// GetTransactionsBySenderID mocks base method.
func (m *MockRepository) GetTransactionsBySenderID(ctx context.Context, senderID int64, listParams transaction.ListParams) ([]transaction.Transaction, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTransactionsBySenderID", ctx, senderID, listParams)
	ret0, _ := ret[0].([]transaction.Transaction)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetTransactionsBySenderID indicates an expected call of GetTransactionsBySenderID.
func (mr *MockRepositoryMockRecorder) GetTransactionsBySenderID(ctx, senderID, listParams interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTransactionsBySenderID", reflect.TypeOf((*MockRepository)(nil).GetTransactionsBySenderID), ctx, senderID, listParams)
}
