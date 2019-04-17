// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/spacekitty/ownCloud/dev/go/src/github.com/pfillion/mobycron/cmd/mobycron/main.go

// Package main is a generated GoMock package.
package main

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCron is a mock of Cron interface
type MockCron struct {
	ctrl     *gomock.Controller
	recorder *MockCronMockRecorder
}

// MockCronMockRecorder is the mock recorder for MockCron
type MockCronMockRecorder struct {
	mock *MockCron
}

// NewMockCron creates a new mock instance
func NewMockCron(ctrl *gomock.Controller) *MockCron {
	mock := &MockCron{ctrl: ctrl}
	mock.recorder = &MockCronMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCron) EXPECT() *MockCronMockRecorder {
	return m.recorder
}

// LoadConfig mocks base method
func (m *MockCron) LoadConfig(filename string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadConfig", filename)
	ret0, _ := ret[0].(error)
	return ret0
}

// LoadConfig indicates an expected call of LoadConfig
func (mr *MockCronMockRecorder) LoadConfig(filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadConfig", reflect.TypeOf((*MockCron)(nil).LoadConfig), filename)
}

// Start mocks base method
func (m *MockCron) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockCronMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockCron)(nil).Start))
}

// Stop mocks base method
func (m *MockCron) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockCronMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockCron)(nil).Stop))
}

// MockHandler is a mock of Handler interface
type MockHandler struct {
	ctrl     *gomock.Controller
	recorder *MockHandlerMockRecorder
}

// MockHandlerMockRecorder is the mock recorder for MockHandler
type MockHandlerMockRecorder struct {
	mock *MockHandler
}

// NewMockHandler creates a new mock instance
func NewMockHandler(ctrl *gomock.Controller) *MockHandler {
	mock := &MockHandler{ctrl: ctrl}
	mock.recorder = &MockHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHandler) EXPECT() *MockHandlerMockRecorder {
	return m.recorder
}

// Scan mocks base method
func (m *MockHandler) Scan() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Scan")
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan
func (mr *MockHandlerMockRecorder) Scan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*MockHandler)(nil).Scan))
}

// Listen mocks base method
func (m *MockHandler) Listen() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Listen")
	ret0, _ := ret[0].(error)
	return ret0
}

// Listen indicates an expected call of Listen
func (mr *MockHandlerMockRecorder) Listen() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Listen", reflect.TypeOf((*MockHandler)(nil).Listen))
}
