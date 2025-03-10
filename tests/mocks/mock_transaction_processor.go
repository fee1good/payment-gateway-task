// Code generated by MockGen. DO NOT EDIT.
// Source: internal/workers/transaction_processor.go
//
// Generated by this command:
//
//	mockgen -source=internal/workers/transaction_processor.go -destination=internal/tests/mocks/mock_transaction_processor.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	"context"
	"reflect"

	"payment-gateway/internal/models"

	"go.uber.org/mock/gomock"
)

// MockTransactionProcessor is a mock of TransactionProcessor interface.
type MockTransactionProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockTransactionProcessorMockRecorder
}

// MockTransactionProcessorMockRecorder is the mock recorder for MockTransactionProcessor.
type MockTransactionProcessorMockRecorder struct {
	mock *MockTransactionProcessor
}

// NewMockTransactionProcessor creates a new mock instance.
func NewMockTransactionProcessor(ctrl *gomock.Controller) *MockTransactionProcessor {
	mock := &MockTransactionProcessor{ctrl: ctrl}
	mock.recorder = &MockTransactionProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransactionProcessor) EXPECT() *MockTransactionProcessorMockRecorder {
	return m.recorder
}

// ProcessTransaction mocks base method.
func (m *MockTransactionProcessor) ProcessTransaction(tx models.Transaction) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ProcessTransaction", tx)
}

// ProcessTransaction indicates an expected call of ProcessTransaction.
func (mr *MockTransactionProcessorMockRecorder) ProcessTransaction(tx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessTransaction", reflect.TypeOf((*MockTransactionProcessor)(nil).ProcessTransaction), tx)
}

// Start mocks base method.
func (m *MockTransactionProcessor) Start(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start", ctx)
}

// Start indicates an expected call of Start.
func (mr *MockTransactionProcessorMockRecorder) Start(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockTransactionProcessor)(nil).Start), ctx)
}

// Stop mocks base method.
func (m *MockTransactionProcessor) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockTransactionProcessorMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockTransactionProcessor)(nil).Stop))
}
