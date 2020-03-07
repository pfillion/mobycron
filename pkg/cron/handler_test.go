package cron

import (
	"bytes"
	context "context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestNewHandler(t *testing.T) {
	// Arrange
	c := &Cron{}

	// Act
	h, err := NewHandler(c)

	// Assert
	assert.Assert(t, h.cron == c)
	assert.Assert(t, h.cli != nil)
	assert.NilError(t, err)
}

func TestNewHandlerError(t *testing.T) {
	// Arrange
	c := &Cron{}
	os.Setenv("DOCKER_HOST", "bad docker host")
	defer os.Unsetenv("DOCKER_HOST")

	// Act
	h, err := NewHandler(c)

	// Assert
	assert.Assert(t, is.Nil(h))
	assert.ErrorContains(t, err, "unable to parse docker host")
}

func TestScanContainer(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.NilError(t, err)
		}
	}

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name   string
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "scan filtered by 'mobycron.schedule' label",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				args := filters.NewArgs()
				args.Add("label", "mobycron.schedule")

				opt := types.ContainerListOptions{All: true, Filters: args}
				cli.EXPECT().ContainerList(gomock.Any(), opt).Return(nil, nil)
				cli.EXPECT().Close()
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "info"),
				hasLogField("func", "Handler.ScanContainer"),
				hasLogField("msg", "scan containers for cron schedule"),
			),
		},
		{
			name: "addContainers in error",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(nil, errors.New("addContainers in error"))
				cli.EXPECT().Close()
			},
			checks: check(
				hasError("addContainers in error"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cron := NewMockCronner(ctrl)
			cli := NewMockDockerClient(ctrl)

			h := &Handler{cron, cli}
			if tt.mock != nil {
				tt.mock(cron, cli)
			}

			// Act
			err := h.ScanContainer()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}

func TestScanService(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.NilError(t, err)
		}
	}

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name   string
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "scan filtered by 'mobycron.schedule' label",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				args := filters.NewArgs()
				args.Add("label", "mobycron.schedule")

				opt := types.ServiceListOptions{Filters: args}
				cli.EXPECT().ServiceList(gomock.Any(), opt).Return(nil, nil)
				cli.EXPECT().Close()
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "info"),
				hasLogField("func", "Handler.ScanService"),
				hasLogField("msg", "scan services for cron schedule"),
			),
		},
		{
			name: "addServices in error",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				cli.EXPECT().ServiceList(gomock.Any(), gomock.Any()).Return(nil, errors.New("addServices in error"))
				cli.EXPECT().Close()
			},
			checks: check(
				hasError("addServices in error"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cron := NewMockCronner(ctrl)
			cli := NewMockDockerClient(ctrl)

			h := &Handler{cron, cli}
			if tt.mock != nil {
				tt.mock(cron, cli)
			}

			// Act
			err := h.ScanService()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}

func TestListenContainer(t *testing.T) {
	type checkFunc func(*testing.T, string)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient, chan events.Message, chan error)
	type eventFunc func(chan events.Message, chan error)

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, out string) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name   string
		mock   mockFunc
		events eventFunc
		checks []checkFunc
	}{
		{
			name: "container created",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				eventOpt := types.EventsOptions{Filters: filters.NewArgs()}
				eventOpt.Filters.Add("label", "mobycron.schedule")
				eventOpt.Filters.Add("type", "container")
				eventOpt.Filters.Add("event", "create")
				eventOpt.Filters.Add("event", "destroy")

				listOpt := types.ContainerListOptions{All: true, Filters: filters.NewArgs()}
				listOpt.Filters.Add("id", "1")

				cli.EXPECT().Events(gomock.Any(), eventOpt).Return(eventChan, errChan)
				cli.EXPECT().ContainerList(gomock.Any(), listOpt)
				cli.EXPECT().Close()
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "create", Actor: events.Actor{ID: "1"}}
			},
			checks: check(
				hasLogField("level", "info"),
				hasLogField("func", "Handler.ListenContainer"),
				hasLogField("msg", "event message from server"),
				hasLogField("msg", "add containers from filters"),
			),
		},
		{
			name: "addContainers in error when container created",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan)
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(nil, errors.New("addContainers in error"))
				cli.EXPECT().Close()
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "create", Actor: events.Actor{ID: "1"}}
			},
			checks: check(
				hasLogField("level", "error"),
				hasLogField("msg", "addContainers in error"),
			),
		},
		{
			name: "container destroyed",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan)
				sc.EXPECT().RemoveContainerJob("1")
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "destroy", Actor: events.Actor{ID: "1"}}
			},
			checks: check(
				hasLogField("level", "info"),
				hasLogField("func", "Handler.ListenContainer"),
				hasLogField("msg", "event message from server"),
			),
		},
		{
			name: "error on channel",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan).Times(2)
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				errChan <- errors.New("error on channel")
			},
			checks: check(
				hasLogField("level", "error"),
				hasLogField("msg", "error from server"),
				hasLogField("error", "error on channel"),
			),
		},
		{
			name: "mix of message and error on channels",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan).Times(2)
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Times(2)
				cli.EXPECT().Close().Times(2)
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "create", Actor: events.Actor{ID: "1"}}
				errChan <- errors.New("error on channel")
				eventChan <- events.Message{Action: "create", Actor: events.Actor{ID: "2"}}
			},
			checks: check(
				hasLogField("actor.ID", "1"),
				hasLogField("actor.ID", "2"),
				hasLogField("error", "error on channel"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			eventChan := make(chan events.Message)
			errChan := make(chan error)

			// Mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cron := NewMockCronner(ctrl)
			cli := NewMockDockerClient(ctrl)

			h := &Handler{cron, cli}
			if tt.mock != nil {
				tt.mock(cron, cli, eventChan, errChan)
			}

			// Act
			h.ListenContainer()
			go tt.events(eventChan, errChan)
			time.Sleep(5 * time.Millisecond)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String())
			}
		})
	}
}

func TestListenService(t *testing.T) {
	type checkFunc func(*testing.T, string)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient, chan events.Message, chan error)
	type eventFunc func(chan events.Message, chan error)

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, out string) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name   string
		mock   mockFunc
		events eventFunc
		checks []checkFunc
	}{
		{
			name: "service create - add service",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				eventOpt := types.EventsOptions{Filters: filters.NewArgs()}
				eventOpt.Filters.Add("type", "service")
				eventOpt.Filters.Add("event", "create")
				eventOpt.Filters.Add("event", "remove")
				eventOpt.Filters.Add("event", "update")

				listOpt := types.ServiceListOptions{Filters: filters.NewArgs()}
				listOpt.Filters.Add("id", "S1")

				cli.EXPECT().Events(gomock.Any(), eventOpt).Return(eventChan, errChan)
				cli.EXPECT().ServiceList(gomock.Any(), listOpt).Return(nil, nil)
				cli.EXPECT().Close()
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "create", Actor: events.Actor{ID: "S1"}}
			},
		},
		{
			name: "service create - add services in error",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan)
				cli.EXPECT().ServiceList(gomock.Any(), gomock.Any()).Return(nil, errors.New("add in error"))
				cli.EXPECT().Close()
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "create", Actor: events.Actor{ID: "S1"}}
			},
			checks: check(
				hasLogField("level", "error"),
				hasLogField("msg", "add in error"),
			),
		},
		{
			name: "service remove",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan)
				sc.EXPECT().RemoveServiceJob("1")
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "remove", Actor: events.Actor{ID: "1"}}
			},
			checks: check(
				hasLogField("level", "info"),
				hasLogField("func", "Handler.ListenService"),
				hasLogField("msg", "event message from server"),
			),
		},
		{
			name: "service update",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				listOpt := types.ServiceListOptions{Filters: filters.NewArgs()}
				listOpt.Filters.Add("id", "S1")

				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan)
				sc.EXPECT().RemoveServiceJob("S1")
				cli.EXPECT().ServiceList(gomock.Any(), listOpt).Return(nil, nil)
				cli.EXPECT().Close()
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "update", Actor: events.Actor{ID: "S1"}}
			},
			checks: check(
				hasLogField("level", "info"),
				hasLogField("func", "Handler.ListenService"),
				hasLogField("msg", "event message from server"),
			),
		},
		{
			name: "service update - add service in error",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan)
				sc.EXPECT().RemoveServiceJob(gomock.Any())
				cli.EXPECT().ServiceList(gomock.Any(), gomock.Any()).Return(nil, errors.New("add in error"))
				cli.EXPECT().Close()
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "update", Actor: events.Actor{ID: "S1"}}
			},
			checks: check(
				hasLogField("level", "error"),
				hasLogField("msg", "add in error"),
			),
		},
		{
			name: "error on channel",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan).Times(2)
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				errChan <- errors.New("error on channel")
			},
			checks: check(
				hasLogField("level", "error"),
				hasLogField("msg", "error from server"),
				hasLogField("error", "error on channel"),
			),
		},
		{
			name: "mix of message and error on channels",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				cli.EXPECT().Events(gomock.Any(), gomock.Any()).Return(eventChan, errChan).Times(2)
				cli.EXPECT().ServiceList(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
				cli.EXPECT().Close().Times(2)
			},
			events: func(eventChan chan events.Message, errChan chan error) {
				eventChan <- events.Message{Action: "create", Actor: events.Actor{ID: "1"}}
				errChan <- errors.New("error on channel")
				eventChan <- events.Message{Action: "create", Actor: events.Actor{ID: "2"}}
			},
			checks: check(
				hasLogField("actor.ID", "1"),
				hasLogField("actor.ID", "2"),
				hasLogField("error", "error on channel"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			eventChan := make(chan events.Message)
			errChan := make(chan error)

			// Mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cron := NewMockCronner(ctrl)
			cli := NewMockDockerClient(ctrl)

			h := &Handler{cron, cli}
			if tt.mock != nil {
				tt.mock(cron, cli, eventChan, errChan)
			}

			// Act
			h.ListenService()
			go tt.events(eventChan, errChan)
			time.Sleep(5 * time.Millisecond)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String())
			}
		})
	}
}

func TestAddContainers(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient, filters.Args)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.NilError(t, err)
		}
	}

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name    string
		filters filters.Args
		mock    mockFunc
		checks  []checkFunc
	}{
		{
			name:    "zero container",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return([]types.Container{}, nil)
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:    "one container",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				ctx := context.Background()
				opt := types.ContainerListOptions{All: true, Filters: filters}
				containers := []types.Container{
					{
						ID: "12345",
						Labels: map[string]string{
							"mobycron.schedule": "3 * * * * *",
							"mobycron.action":   "exec",
							"mobycron.timeout":  "30",
							"mobycron.command":  "echo 'do job'",
						},
					},
				}
				cli.EXPECT().ContainerList(ctx, opt).Return(containers, nil)
				sc.EXPECT().AddContainerJob(ContainerJob{
					Schedule:  "3 * * * * *",
					Action:    "exec",
					Timeout:   "30",
					Command:   "echo 'do job'",
					Container: containers[0],
					cron:      nil,
					cli:       cli,
				})
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "info"),
				hasLogField("func", "Handler.addContainers"),
				hasLogField("msg", "add containers from filters"),
			),
		},
		{
			name:    "many containers",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				ctx := context.Background()
				opt := types.ContainerListOptions{All: true, Filters: filters}
				containers := []types.Container{
					{
						ID: "1",
						Labels: map[string]string{
							"mobycron.schedule": "1 * * * * *",
							"mobycron.action":   "exec1",
							"mobycron.timeout":  "30",
							"mobycron.command":  "echo 'do job 1'",
						},
					},
					{
						ID: "2",
						Labels: map[string]string{
							"mobycron.schedule": "2 * * * * *",
							"mobycron.action":   "exec2",
							"mobycron.timeout":  "5",
							"mobycron.command":  "echo 'do job 2'",
						},
					},
				}
				cli.EXPECT().ContainerList(ctx, opt).Return(containers, nil)
				sc.EXPECT().AddContainerJob(ContainerJob{
					Schedule:  "1 * * * * *",
					Action:    "exec1",
					Timeout:   "30",
					Command:   "echo 'do job 1'",
					Container: containers[0],
					cron:      nil,
					cli:       cli,
				})
				sc.EXPECT().AddContainerJob(ContainerJob{
					Schedule:  "2 * * * * *",
					Action:    "exec2",
					Timeout:   "5",
					Command:   "echo 'do job 2'",
					Container: containers[1],
					cron:      nil,
					cli:       cli,
				})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:    "container from service/task",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				containers := []types.Container{
					{
						Created: 123,
						Labels: map[string]string{
							"com.docker.swarm.service.id":   "sid",
							"com.docker.swarm.service.name": "sname",
							"com.docker.swarm.task.id":      "tid",
							"com.docker.swarm.task.name":    "sname.1.tid",
							"mobycron.schedule":             "2 * * * * *",
							"mobycron.action":               "start",
						},
					},
				}
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(containers, nil)
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "error"),
				hasLogField("msg", "mobycron label must be set on service, not directly on the container"),
			),
		},
		{
			name:    "ContainerList in error",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(nil, errors.New("ContainerList in error"))
			},
			checks: check(
				hasError("ContainerList in error"),
			),
		},
		{
			name:    "AddContainerJob in error",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				containers := []types.Container{{ID: "1"}}
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(containers, nil)
				sc.EXPECT().AddContainerJob(gomock.Any()).Return(errors.New("AddContainerJob in error"))
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "error"),
				hasLogField("error", "AddContainerJob in error"),
				hasLogField("msg", "add container job to cron is in error"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cron := NewMockCronner(ctrl)
			cli := NewMockDockerClient(ctrl)

			h := &Handler{cron, cli}
			if tt.mock != nil {
				tt.mock(cron, cli, tt.filters)
			}

			// Act
			err := h.addContainers(tt.filters)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}

func TestAddServices(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient, filters.Args)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.NilError(t, err)
		}
	}

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name    string
		filters filters.Args
		mock    mockFunc
		checks  []checkFunc
	}{
		{
			name:    "zero service",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				cli.EXPECT().ServiceList(gomock.Any(), gomock.Any()).Return([]swarm.Service{}, nil)
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:    "one service",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				ctx := context.Background()
				opt := types.ServiceListOptions{Filters: filters}
				services := []swarm.Service{
					{
						ID: "12345",
						Spec: swarm.ServiceSpec{
							Annotations: swarm.Annotations{
								Name: "name1",
								Labels: map[string]string{
									"mobycron.schedule": "3 * * * * *",
									"mobycron.action":   "exec",
									"mobycron.timeout":  "30",
									"mobycron.command":  "echo 'do job'",
								},
							},
						},
						Meta: swarm.Meta{
							Version:   swarm.Version{Index: 111},
							CreatedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
				}
				cli.EXPECT().ServiceList(ctx, opt).Return(services, nil)
				sc.EXPECT().AddServiceJob(ServiceJob{
					Schedule:         "3 * * * * *",
					Action:           "exec",
					Timeout:          "30",
					Command:          "echo 'do job'",
					ServiceID:        "12345",
					ServiceName:      "name1",
					ServiceVersion:   swarm.Version{Index: 111},
					ServiceCreatedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
					Service:          services[0],
					cron:             nil,
					cli:              cli,
				})
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "info"),
				hasLogField("func", "Handler.addServices"),
				hasLogField("msg", "add services from filters"),
			),
		},
		{
			name:    "many containers",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				ctx := context.Background()
				opt := types.ServiceListOptions{Filters: filters}
				services := []swarm.Service{
					{
						ID: "12345",
						Spec: swarm.ServiceSpec{
							Annotations: swarm.Annotations{
								Name: "name1",
								Labels: map[string]string{
									"mobycron.schedule": "3 * * * * *",
									"mobycron.action":   "exec",
									"mobycron.timeout":  "30",
									"mobycron.command":  "echo 'do job'",
								},
							},
						},
						Meta: swarm.Meta{
							Version:   swarm.Version{Index: 111},
							CreatedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
					{
						ID: "2222",
						Spec: swarm.ServiceSpec{
							Annotations: swarm.Annotations{
								Name: "name2",
								Labels: map[string]string{
									"mobycron.schedule": "2 * * * * *",
									"mobycron.action":   "exec",
									"mobycron.timeout":  "2",
									"mobycron.command":  "echo 'do job2'",
								},
							},
						},
						Meta: swarm.Meta{
							Version:   swarm.Version{Index: 222},
							CreatedAt: time.Date(2002, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
				}
				cli.EXPECT().ServiceList(ctx, opt).Return(services, nil)
				sc.EXPECT().AddServiceJob(ServiceJob{
					Schedule:         "3 * * * * *",
					Action:           "exec",
					Timeout:          "30",
					Command:          "echo 'do job'",
					ServiceID:        "12345",
					ServiceName:      "name1",
					ServiceVersion:   swarm.Version{Index: 111},
					ServiceCreatedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
					Service:          services[0],
					cron:             nil,
					cli:              cli,
				})
				sc.EXPECT().AddServiceJob(ServiceJob{
					Schedule:         "2 * * * * *",
					Action:           "exec",
					Timeout:          "2",
					Command:          "echo 'do job2'",
					ServiceID:        "2222",
					ServiceName:      "name2",
					ServiceVersion:   swarm.Version{Index: 222},
					ServiceCreatedAt: time.Date(2002, 1, 1, 0, 0, 0, 0, time.UTC),
					Service:          services[1],
					cron:             nil,
					cli:              cli,
				})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:    "ServiceList in error",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				cli.EXPECT().ServiceList(gomock.Any(), gomock.Any()).Return(nil, errors.New("ServiceList in error"))
			},
			checks: check(
				hasError("ServiceList in error"),
			),
		},
		{
			name:    "AddServiceJob in error",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				services := []swarm.Service{
					{
						Spec: swarm.ServiceSpec{
							Annotations: swarm.Annotations{
								Labels: map[string]string{
									"mobycron.schedule": "3 * * * * *",
								},
							},
						},
					},
				}

				cli.EXPECT().ServiceList(gomock.Any(), gomock.Any()).Return(services, nil)
				sc.EXPECT().AddServiceJob(gomock.Any()).Return(errors.New("AddServiceJob in error"))
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "error"),
				hasLogField("error", "AddServiceJob in error"),
				hasLogField("msg", "add service job to cron is in error"),
			),
		},
		{
			name:    "skipped - no label",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				services := []swarm.Service{{ID: "111"}}

				cli.EXPECT().ServiceList(gomock.Any(), gomock.Any()).Return(services, nil)
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "info"),
				hasLogField("msg", "skipped, mobycron label not found"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cron := NewMockCronner(ctrl)
			cli := NewMockDockerClient(ctrl)

			h := &Handler{cron, cli}
			if tt.mock != nil {
				tt.mock(cron, cli, tt.filters)
			}

			// Act
			err := h.addServices(tt.filters)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}

func TestPruneContainersFromService(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.NilError(t, err)
		}
	}

	tests := []struct {
		name      string
		serviceID string
		mock      mockFunc
		checks    []checkFunc
	}{
		{
			name:      "one container to prune",
			serviceID: "s1",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				serviceOpt := types.ServiceInspectOptions{}
				service := swarm.Service{Spec: swarm.ServiceSpec{Annotations: swarm.Annotations{Name: "s1"}}}

				ctx := context.Background()
				filters := filters.NewArgs()
				filters.Add("service", "s1")
				taskOpt := types.TaskListOptions{Filters: filters}

				tasks := []swarm.Task{
					{
						Meta:      swarm.Meta{CreatedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)},
						ServiceID: "s1",
						Slot:      1,
						Status:    swarm.TaskStatus{ContainerStatus: &swarm.ContainerStatus{ContainerID: "c1"}},
					},
					{
						Meta:      swarm.Meta{CreatedAt: time.Date(2001, 1, 2, 0, 0, 0, 0, time.UTC)},
						ServiceID: "s1",
						Slot:      1,
						Status:    swarm.TaskStatus{ContainerStatus: &swarm.ContainerStatus{ContainerID: "c2"}},
					},
				}

				cli.EXPECT().ServiceInspectWithRaw(ctx, "s1", serviceOpt).Return(service, nil, nil)
				cli.EXPECT().TaskList(ctx, taskOpt).Return(tasks, nil)
				cron.EXPECT().RemoveContainerJob("c1")
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "many containers to prune",
			serviceID: "s1",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				tasks := []swarm.Task{
					{
						Meta:      swarm.Meta{CreatedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)},
						ServiceID: "s1",
						Slot:      1,
						Status:    swarm.TaskStatus{ContainerStatus: &swarm.ContainerStatus{ContainerID: "c1"}},
					},
					{
						Meta:      swarm.Meta{CreatedAt: time.Date(2001, 1, 3, 0, 0, 0, 0, time.UTC)},
						ServiceID: "s1",
						Slot:      1,
						Status:    swarm.TaskStatus{ContainerStatus: &swarm.ContainerStatus{ContainerID: "c2"}},
					},
					{
						Meta:      swarm.Meta{CreatedAt: time.Date(2001, 1, 2, 0, 0, 0, 0, time.UTC)},
						ServiceID: "s1",
						Slot:      1,
						Status:    swarm.TaskStatus{ContainerStatus: &swarm.ContainerStatus{ContainerID: "c3"}},
					},
				}

				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).Return(swarm.Service{}, nil, nil)
				cli.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(tasks, nil)
				cron.EXPECT().RemoveContainerJob("c1")
				cron.EXPECT().RemoveContainerJob("c3")
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "no container to prune - not same slot",
			serviceID: "s1",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				tasks := []swarm.Task{
					{
						Meta:      swarm.Meta{CreatedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)},
						ServiceID: "s1",
						Slot:      1,
						Status:    swarm.TaskStatus{ContainerStatus: &swarm.ContainerStatus{ContainerID: "c1"}},
					},
					{
						Meta:      swarm.Meta{CreatedAt: time.Date(2001, 1, 2, 0, 0, 0, 0, time.UTC)},
						ServiceID: "s1",
						Slot:      2,
						Status:    swarm.TaskStatus{ContainerStatus: &swarm.ContainerStatus{ContainerID: "c2"}},
					},
				}

				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).Return(swarm.Service{}, nil, nil)
				cli.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(tasks, nil)
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "no container to prune - only one task",
			serviceID: "s1",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				tasks := []swarm.Task{
					{
						Meta:      swarm.Meta{CreatedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)},
						ServiceID: "s1",
						Slot:      1,
						Status:    swarm.TaskStatus{ContainerStatus: &swarm.ContainerStatus{ContainerID: "c1"}},
					},
				}

				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).Return(swarm.Service{}, nil, nil)
				cli.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(tasks, nil)
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "no task found",
			serviceID: "s1",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).Return(swarm.Service{}, nil, nil)
				cli.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return([]swarm.Task{}, nil)
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "TaskList error",
			serviceID: "s1",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).Return(swarm.Service{}, nil, nil)
				cli.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return([]swarm.Task{}, errors.New("TaskList error"))
			},
			checks: check(
				hasError("TaskList error"),
			),
		},
		{
			name:      "ServiceInspectWithRaw error",
			serviceID: "s1",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).Return(swarm.Service{}, nil, errors.New("ServiceInspectWithRaw error"))
			},
			checks: check(
				hasError("ServiceInspectWithRaw error"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cron := NewMockCronner(ctrl)
			cli := NewMockDockerClient(ctrl)

			h := &Handler{cron, cli}
			if tt.mock != nil {
				tt.mock(cron, cli)
			}

			// Act
			err := h.pruneContainersFromService(tt.serviceID)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}

func TestPruneContainersFromAllServices(t *testing.T) {
	type checkFunc func(*testing.T, string)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient)

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, out string) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name   string
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "one service to prune",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				ctx := context.Background()
				services := []swarm.Service{
					{
						ID: "s1",
					},
				}

				cli.EXPECT().ServiceList(ctx, types.ServiceListOptions{}).Return(services, nil)
				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), "s1", gomock.Any()).Return(swarm.Service{}, nil, nil)
				cli.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return([]swarm.Task{}, nil)
			},
		},
		{
			name: "many services to prune",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				ctx := context.Background()
				services := []swarm.Service{
					{
						ID: "s1",
					},
					{
						ID: "s2",
					},
				}

				cli.EXPECT().ServiceList(ctx, types.ServiceListOptions{}).Return(services, nil)
				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), "s1", gomock.Any()).Return(swarm.Service{}, nil, nil)
				cli.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return([]swarm.Task{}, nil)
				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), "s2", gomock.Any()).Return(swarm.Service{}, nil, nil)
				cli.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return([]swarm.Task{}, nil)
			},
		},
		{
			name: "no services to prune",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				ctx := context.Background()
				services := []swarm.Service{}
				cli.EXPECT().ServiceList(ctx, types.ServiceListOptions{}).Return(services, nil)
			},
		},
		{
			name: "ServiceList in error",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				ctx := context.Background()
				services := []swarm.Service{
					{
						ID: "s1",
					},
				}

				cli.EXPECT().ServiceList(ctx, types.ServiceListOptions{}).Return(services, errors.New("ServiceList in error"))
			},
			checks: check(
				hasLogField("level", "error"),
				hasLogField("msg", "ServiceList in error"),
			),
		},
		{
			name: "pruneContainersFromService in error",
			mock: func(cron *MockCronner, cli *MockDockerClient) {
				ctx := context.Background()
				services := []swarm.Service{
					{
						ID: "s1",
					},
				}

				cli.EXPECT().ServiceList(ctx, types.ServiceListOptions{}).Return(services, nil)
				cli.EXPECT().ServiceInspectWithRaw(gomock.Any(), gomock.Any(), gomock.Any()).Return(swarm.Service{}, nil, errors.New("pruneContainersFromService in error"))
			},
			checks: check(
				hasLogField("level", "error"),
				hasLogField("msg", "pruneContainersFromService in error"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cron := NewMockCronner(ctrl)
			cli := NewMockDockerClient(ctrl)

			h := &Handler{cron, cli}
			if tt.mock != nil {
				tt.mock(cron, cli)
			}

			// Act
			h.pruneContainersFromAllServices()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String())
			}
		})
	}
}
