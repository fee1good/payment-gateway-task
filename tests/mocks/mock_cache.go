// Code generated by MockGen. DO NOT EDIT.
// Source: internal/cache/redis.go
//
// Generated by this command:
//
//	mockgen -source=internal/cache/redis.go -destination=internal/tests/mocks/mock_cache.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	"context"
	"reflect"
	"time"

	"payment-gateway/db"

	"go.uber.org/mock/gomock"
)

// MockCache is a mock of Cache interface.
type MockCache struct {
	ctrl     *gomock.Controller
	recorder *MockCacheMockRecorder
}

// MockCacheMockRecorder is the mock recorder for MockCache.
type MockCacheMockRecorder struct {
	mock *MockCache
}

// NewMockCache creates a new mock instance.
func NewMockCache(ctrl *gomock.Controller) *MockCache {
	mock := &MockCache{ctrl: ctrl}
	mock.recorder = &MockCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCache) EXPECT() *MockCacheMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, key)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockCacheMockRecorder) Get(ctx, key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCache)(nil).Get), ctx, key)
}

// GetGatewaysByCountry mocks base method.
func (m *MockCache) GetGatewaysByCountry(ctx context.Context, dbHandler db.Storage, countryID int) ([]db.Gateway, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGatewaysByCountry", ctx, dbHandler, countryID)
	ret0, _ := ret[0].([]db.Gateway)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGatewaysByCountry indicates an expected call of GetGatewaysByCountry.
func (mr *MockCacheMockRecorder) GetGatewaysByCountry(ctx, dbHandler, countryID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGatewaysByCountry", reflect.TypeOf((*MockCache)(nil).GetGatewaysByCountry), ctx, dbHandler, countryID)
}

// Set mocks base method.
func (m *MockCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, key, value, expiration)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockCacheMockRecorder) Set(ctx, key, value, expiration any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCache)(nil).Set), ctx, key, value, expiration)
}
