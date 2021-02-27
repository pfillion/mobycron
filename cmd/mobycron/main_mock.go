// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/spacekitty/ownCloud/dev/go/src/github.com/pfillion/mobycron/cmd/mobycron/main.go

// Package main is a generated GoMock package.
package main

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCronner is a mock of Cronner interface.
type MockCronner struct {
	ctrl     *gomock.Controller
	recorder *MockCronnerMockRecorder
}

// MockCronnerMockRecorder is the mock recorder for MockCronner.
type MockCronnerMockRecorder struct {
	mock *MockCronner
}

// NewMockCronner creates a new mock instance.
func NewMockCronner(ctrl *gomock.Controller) *MockCronner {
	mock := &MockCronner{ctrl: ctrl}
	mock.recorder = &MockCronnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCronner) EXPECT() *MockCronnerMockRecorder {
	return m.recorder
}

// LoadConfig mocks base method.
func (m *MockCronner) LoadConfig(filename string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadConfig", filename)
	ret0, _ := ret[0].(error)
	return ret0
}

// LoadConfig indicates an expected call of LoadConfig.
func (mr *MockCronnerMockRecorder) LoadConfig(filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadConfig", reflect.TypeOf((*MockCronner)(nil).LoadConfig), filename)
}

// Start mocks base method.
func (m *MockCronner) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockCronnerMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockCronner)(nil).Start))
}

// Stop mocks base method.
func (m *MockCronner) Stop() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockCronnerMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockCronner)(nil).Stop))
}

// MockHandler is a mock of Handler interface.
type MockHandler struct {
	ctrl     *gomock.Controller
	recorder *MockHandlerMockRecorder
}

// MockHandlerMockRecorder is the mock recorder for MockHandler.
type MockHandlerMockRecorder struct {
	mock *MockHandler
}

// NewMockHandler creates a new mock instance.
func NewMockHandler(ctrl *gomock.Controller) *MockHandler {
	mock := &MockHandler{ctrl: ctrl}
	mock.recorder = &MockHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHandler) EXPECT() *MockHandlerMockRecorder {
	return m.recorder
}

// ListenContainer mocks base method.
func (m *MockHandler) ListenContainer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ListenContainer")
}

// ListenContainer indicates an expected call of ListenContainer.
func (mr *MockHandlerMockRecorder) ListenContainer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListenContainer", reflect.TypeOf((*MockHandler)(nil).ListenContainer))
}

// ListenService mocks base method.
func (m *MockHandler) ListenService() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ListenService")
}

// ListenService indicates an expected call of ListenService.
func (mr *MockHandlerMockRecorder) ListenService() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListenService", reflect.TypeOf((*MockHandler)(nil).ListenService))
}

// ScanContainer mocks base method.
func (m *MockHandler) ScanContainer() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScanContainer")
	ret0, _ := ret[0].(error)
	return ret0
}

// ScanContainer indicates an expected call of ScanContainer.
func (mr *MockHandlerMockRecorder) ScanContainer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScanContainer", reflect.TypeOf((*MockHandler)(nil).ScanContainer))
}

// ScanService mocks base method.
func (m *MockHandler) ScanService() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScanService")
	ret0, _ := ret[0].(error)
	return ret0
}

// ScanService indicates an expected call of ScanService.
func (mr *MockHandlerMockRecorder) ScanService() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScanService", reflect.TypeOf((*MockHandler)(nil).ScanService))
}
