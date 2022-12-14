// Code generated by MockGen. DO NOT EDIT.
// Source: handler.go

// Package account_test is a generated GoMock package.
package account_test

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	account "github.com/maypok86/payment-api/internal/domain/account"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// AddBalance mocks base method.
func (m *MockService) AddBalance(ctx context.Context, dto account.AddBalanceDTO) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddBalance", ctx, dto)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddBalance indicates an expected call of AddBalance.
func (mr *MockServiceMockRecorder) AddBalance(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddBalance", reflect.TypeOf((*MockService)(nil).AddBalance), ctx, dto)
}

// GetBalanceByID mocks base method.
func (m *MockService) GetBalanceByID(ctx context.Context, id int64) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalanceByID", ctx, id)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalanceByID indicates an expected call of GetBalanceByID.
func (mr *MockServiceMockRecorder) GetBalanceByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalanceByID", reflect.TypeOf((*MockService)(nil).GetBalanceByID), ctx, id)
}

// TransferBalance mocks base method.
func (m *MockService) TransferBalance(ctx context.Context, dto account.TransferBalanceDTO) (int64, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TransferBalance", ctx, dto)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// TransferBalance indicates an expected call of TransferBalance.
func (mr *MockServiceMockRecorder) TransferBalance(ctx, dto interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TransferBalance", reflect.TypeOf((*MockService)(nil).TransferBalance), ctx, dto)
}
