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

func TestScan(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient)

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
				hasLogField("func", "Handler.Scan"),
				hasLogField("msg", "scan containers for cron schedule"),
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
			err := h.Scan()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}

func TestListen(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient, chan events.Message, chan error)
	type eventFunc func(chan events.Message, chan error)

	// hasError := func(want string) checkFunc {
	// 	return func(t *testing.T, out string, err error) {
	// 		assert.Assert(t, is.ErrorContains(err, want))
	// 	}
	// }

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
		events eventFunc
		checks []checkFunc
	}{
		{
			name: "container created",
			mock: func(sc *MockCronner, cli *MockDockerClient, eventChan chan events.Message, errChan chan error) {
				eventOpt := types.EventsOptions{Filters: filters.NewArgs()}
				eventOpt.Filters.Add("label", "mobycron.schedule")
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
				hasNilError(),
				hasLogField("level", "info"),
				hasLogField("func", "Handler.Listen"),
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
				hasNilError(),
				hasLogField("level", "info"),
				hasLogField("func", "Handler.Listen"),
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
			err := h.Listen()
			go tt.events(eventChan, errChan)
			time.Sleep(5 * time.Millisecond)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
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
				cli.EXPECT().Close()
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
				cli.EXPECT().Close()
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
				cli.EXPECT().Close()
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
				sc.EXPECT().AddContainerJob(ContainerJob{
					Schedule:    "2 * * * * *",
					Action:      "start",
					ServiceID:   "sid",
					ServiceName: "sname",
					TaskID:      "tid",
					TaskName:    "sname.1.tid",
					Slot:        1,
					Created:     123,
					Container:   containers[0],
					cli:         cli,
				})
				cli.EXPECT().Close()
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:    "invalid slot",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				containers := []types.Container{
					{
						Labels: map[string]string{
							"com.docker.swarm.task.name": "sname.invalidslot.tid",
						},
					},
				}
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(containers, nil)
				cli.EXPECT().Close()
			},
			checks: check(
				hasNilError(),
				hasLogField("level", "error"),
				hasLogField("msg", "failed to convert slot of label 'com.docker.swarm.task.name'"),
			),
		},
		{
			name:    "ContainerList in error",
			filters: filters.NewArgs(),
			mock: func(sc *MockCronner, cli *MockDockerClient, filters filters.Args) {
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(nil, errors.New("ContainerList in error"))
				cli.EXPECT().Close()
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
				cli.EXPECT().Close()
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
