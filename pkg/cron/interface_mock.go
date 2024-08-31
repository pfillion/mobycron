// Code generated by MockGen. DO NOT EDIT.
// Source: /home/spacekitty/ownCloud/dev/go/src/github.com/pfillion/mobycron/pkg/cron/interface.go

// Package cron is a generated GoMock package.
package cron

import (
	context "context"
	reflect "reflect"

	types "github.com/docker/docker/api/types"
	container "github.com/docker/docker/api/types/container"
	events "github.com/docker/docker/api/types/events"
	swarm "github.com/docker/docker/api/types/swarm"
	gomock "github.com/golang/mock/gomock"
	v3 "github.com/robfig/cron/v3"
)

// MockRunner is a mock of Runner interface.
type MockRunner struct {
	ctrl     *gomock.Controller
	recorder *MockRunnerMockRecorder
}

// MockRunnerMockRecorder is the mock recorder for MockRunner.
type MockRunnerMockRecorder struct {
	mock *MockRunner
}

// NewMockRunner creates a new mock instance.
func NewMockRunner(ctrl *gomock.Controller) *MockRunner {
	mock := &MockRunner{ctrl: ctrl}
	mock.recorder = &MockRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRunner) EXPECT() *MockRunnerMockRecorder {
	return m.recorder
}

// AddJob mocks base method.
func (m *MockRunner) AddJob(spec string, cmd v3.Job) (v3.EntryID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddJob", spec, cmd)
	ret0, _ := ret[0].(v3.EntryID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddJob indicates an expected call of AddJob.
func (mr *MockRunnerMockRecorder) AddJob(spec, cmd interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddJob", reflect.TypeOf((*MockRunner)(nil).AddJob), spec, cmd)
}

// Remove mocks base method.
func (m *MockRunner) Remove(id v3.EntryID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Remove", id)
}

// Remove indicates an expected call of Remove.
func (mr *MockRunnerMockRecorder) Remove(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockRunner)(nil).Remove), id)
}

// Start mocks base method.
func (m *MockRunner) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockRunnerMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockRunner)(nil).Start))
}

// Stop mocks base method.
func (m *MockRunner) Stop() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockRunnerMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockRunner)(nil).Stop))
}

// MockJobSynchroniser is a mock of JobSynchroniser interface.
type MockJobSynchroniser struct {
	ctrl     *gomock.Controller
	recorder *MockJobSynchroniserMockRecorder
}

// MockJobSynchroniserMockRecorder is the mock recorder for MockJobSynchroniser.
type MockJobSynchroniserMockRecorder struct {
	mock *MockJobSynchroniser
}

// NewMockJobSynchroniser creates a new mock instance.
func NewMockJobSynchroniser(ctrl *gomock.Controller) *MockJobSynchroniser {
	mock := &MockJobSynchroniser{ctrl: ctrl}
	mock.recorder = &MockJobSynchroniserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJobSynchroniser) EXPECT() *MockJobSynchroniserMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockJobSynchroniser) Add(delta int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Add", delta)
}

// Add indicates an expected call of Add.
func (mr *MockJobSynchroniserMockRecorder) Add(delta interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockJobSynchroniser)(nil).Add), delta)
}

// Done mocks base method.
func (m *MockJobSynchroniser) Done() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Done")
}

// Done indicates an expected call of Done.
func (mr *MockJobSynchroniserMockRecorder) Done() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Done", reflect.TypeOf((*MockJobSynchroniser)(nil).Done))
}

// Wait mocks base method.
func (m *MockJobSynchroniser) Wait() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Wait")
}

// Wait indicates an expected call of Wait.
func (mr *MockJobSynchroniserMockRecorder) Wait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockJobSynchroniser)(nil).Wait))
}

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

// AddContainerJob mocks base method.
func (m *MockCronner) AddContainerJob(job ContainerJob) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddContainerJob", job)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddContainerJob indicates an expected call of AddContainerJob.
func (mr *MockCronnerMockRecorder) AddContainerJob(job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddContainerJob", reflect.TypeOf((*MockCronner)(nil).AddContainerJob), job)
}

// AddServiceJob mocks base method.
func (m *MockCronner) AddServiceJob(job ServiceJob) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddServiceJob", job)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddServiceJob indicates an expected call of AddServiceJob.
func (mr *MockCronnerMockRecorder) AddServiceJob(job interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddServiceJob", reflect.TypeOf((*MockCronner)(nil).AddServiceJob), job)
}

// RemoveContainerJob mocks base method.
func (m *MockCronner) RemoveContainerJob(ID string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveContainerJob", ID)
}

// RemoveContainerJob indicates an expected call of RemoveContainerJob.
func (mr *MockCronnerMockRecorder) RemoveContainerJob(ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveContainerJob", reflect.TypeOf((*MockCronner)(nil).RemoveContainerJob), ID)
}

// RemoveServiceJob mocks base method.
func (m *MockCronner) RemoveServiceJob(ID string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveServiceJob", ID)
}

// RemoveServiceJob indicates an expected call of RemoveServiceJob.
func (mr *MockCronnerMockRecorder) RemoveServiceJob(ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveServiceJob", reflect.TypeOf((*MockCronner)(nil).RemoveServiceJob), ID)
}

// MockDockerClient is a mock of DockerClient interface.
type MockDockerClient struct {
	ctrl     *gomock.Controller
	recorder *MockDockerClientMockRecorder
}

// MockDockerClientMockRecorder is the mock recorder for MockDockerClient.
type MockDockerClientMockRecorder struct {
	mock *MockDockerClient
}

// NewMockDockerClient creates a new mock instance.
func NewMockDockerClient(ctrl *gomock.Controller) *MockDockerClient {
	mock := &MockDockerClient{ctrl: ctrl}
	mock.recorder = &MockDockerClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDockerClient) EXPECT() *MockDockerClientMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockDockerClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockDockerClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDockerClient)(nil).Close))
}

// ContainerExecAttach mocks base method.
func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerExecAttach", ctx, execID, config)
	ret0, _ := ret[0].(types.HijackedResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ContainerExecAttach indicates an expected call of ContainerExecAttach.
func (mr *MockDockerClientMockRecorder) ContainerExecAttach(ctx, execID, config interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerExecAttach", reflect.TypeOf((*MockDockerClient)(nil).ContainerExecAttach), ctx, execID, config)
}

// ContainerExecCreate mocks base method.
func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (types.IDResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerExecCreate", ctx, container, config)
	ret0, _ := ret[0].(types.IDResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ContainerExecCreate indicates an expected call of ContainerExecCreate.
func (mr *MockDockerClientMockRecorder) ContainerExecCreate(ctx, container, config interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerExecCreate", reflect.TypeOf((*MockDockerClient)(nil).ContainerExecCreate), ctx, container, config)
}

// ContainerExecInspect mocks base method.
func (m *MockDockerClient) ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerExecInspect", ctx, execID)
	ret0, _ := ret[0].(container.ExecInspect)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ContainerExecInspect indicates an expected call of ContainerExecInspect.
func (mr *MockDockerClientMockRecorder) ContainerExecInspect(ctx, execID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerExecInspect", reflect.TypeOf((*MockDockerClient)(nil).ContainerExecInspect), ctx, execID)
}

// ContainerInspect mocks base method.
func (m *MockDockerClient) ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerInspect", ctx, container)
	ret0, _ := ret[0].(types.ContainerJSON)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ContainerInspect indicates an expected call of ContainerInspect.
func (mr *MockDockerClientMockRecorder) ContainerInspect(ctx, container interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerInspect", reflect.TypeOf((*MockDockerClient)(nil).ContainerInspect), ctx, container)
}

// ContainerList mocks base method.
func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerList", ctx, options)
	ret0, _ := ret[0].([]types.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ContainerList indicates an expected call of ContainerList.
func (mr *MockDockerClientMockRecorder) ContainerList(ctx, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerList", reflect.TypeOf((*MockDockerClient)(nil).ContainerList), ctx, options)
}

// ContainerRestart mocks base method.
func (m *MockDockerClient) ContainerRestart(ctx context.Context, container string, options container.StopOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerRestart", ctx, container, options)
	ret0, _ := ret[0].(error)
	return ret0
}

// ContainerRestart indicates an expected call of ContainerRestart.
func (mr *MockDockerClientMockRecorder) ContainerRestart(ctx, container, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerRestart", reflect.TypeOf((*MockDockerClient)(nil).ContainerRestart), ctx, container, options)
}

// ContainerStart mocks base method.
func (m *MockDockerClient) ContainerStart(ctx context.Context, container string, options container.StartOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerStart", ctx, container, options)
	ret0, _ := ret[0].(error)
	return ret0
}

// ContainerStart indicates an expected call of ContainerStart.
func (mr *MockDockerClientMockRecorder) ContainerStart(ctx, container, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerStart", reflect.TypeOf((*MockDockerClient)(nil).ContainerStart), ctx, container, options)
}

// ContainerStop mocks base method.
func (m *MockDockerClient) ContainerStop(ctx context.Context, container string, timeout container.StopOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerStop", ctx, container, timeout)
	ret0, _ := ret[0].(error)
	return ret0
}

// ContainerStop indicates an expected call of ContainerStop.
func (mr *MockDockerClientMockRecorder) ContainerStop(ctx, container, timeout interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerStop", reflect.TypeOf((*MockDockerClient)(nil).ContainerStop), ctx, container, timeout)
}

// Events mocks base method.
func (m *MockDockerClient) Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Events", ctx, options)
	ret0, _ := ret[0].(<-chan events.Message)
	ret1, _ := ret[1].(<-chan error)
	return ret0, ret1
}

// Events indicates an expected call of Events.
func (mr *MockDockerClientMockRecorder) Events(ctx, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Events", reflect.TypeOf((*MockDockerClient)(nil).Events), ctx, options)
}

// ServiceList mocks base method.
func (m *MockDockerClient) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServiceList", ctx, options)
	ret0, _ := ret[0].([]swarm.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServiceList indicates an expected call of ServiceList.
func (mr *MockDockerClientMockRecorder) ServiceList(ctx, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServiceList", reflect.TypeOf((*MockDockerClient)(nil).ServiceList), ctx, options)
}

// ServiceUpdate mocks base method.
func (m *MockDockerClient) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (swarm.ServiceUpdateResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServiceUpdate", ctx, serviceID, version, service, options)
	ret0, _ := ret[0].(swarm.ServiceUpdateResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServiceUpdate indicates an expected call of ServiceUpdate.
func (mr *MockDockerClientMockRecorder) ServiceUpdate(ctx, serviceID, version, service, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServiceUpdate", reflect.TypeOf((*MockDockerClient)(nil).ServiceUpdate), ctx, serviceID, version, service, options)
}
