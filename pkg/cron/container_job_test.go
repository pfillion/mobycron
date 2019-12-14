package cron

import (
	"bufio"
	"bytes"
	context "context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestContainerJobRun(t *testing.T) {
	type checkFunc func(*testing.T, log.Fields)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockJobSynchroniser, *MockDockerClient)

	hasNilError := func() checkFunc {
		return func(t *testing.T, fields log.Fields) {
			assert.Assert(t, is.Equal(fields["level"], "info"), "log fields: %v", fields)
			assert.Assert(t, is.Nil(fields["error"]), "log fields: %v", fields)
		}
	}

	hasError := func(want string) checkFunc {
		return func(t *testing.T, fields log.Fields) {
			assert.Assert(t, is.Equal(fields["level"], "error"), "log fields: %v", fields)
			assert.Assert(t, is.Equal(fields["error"], want), "log fields: %v", fields)
		}
	}

	hasUnknowError := func() checkFunc {
		return func(t *testing.T, fields log.Fields) {
			assert.Assert(t, is.Equal(fields["level"], "error"), "log fields: %v", fields)
			assert.Assert(t, fields["error"] != "", "log fields: %v", fields)
		}
	}

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, fields log.Fields) {
			assert.Assert(t, is.Equal(fields[field], want), "log fields: %v", fields)
		}
	}

	tests := []struct {
		name      string
		schedule  string
		action    string
		timeout   string
		command   string
		container types.Container
		mock      mockFunc
		checks    []checkFunc
	}{
		{
			name:      "ContainerStart",
			schedule:  "1 * * * 5",
			action:    "start",
			timeout:   "30",
			container: types.Container{ID: "id1", Names: []string{"name1", "name2"}},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerStart(context.Background(), "id1", types.ContainerStartOptions{})
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasNilError(),
				hasLogField("func", "ContainerJob.Run"),
				hasLogField("schedule", "1 * * * 5"),
				hasLogField("action", "start"),
				hasLogField("timeout", "30"),
				hasLogField("container.ID", "id1"),
				hasLogField("container.Names", "name1,name2"),
				hasLogField("msg", "container action completed successfully"),
			),
		},
		{
			name:      "ContainerStart error",
			action:    "start",
			container: types.Container{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerStart(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("container error"))
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("container error"),
				hasLogField("msg", "container job completed with error"),
			),
		},
		{
			name:      "ContainerRestart default timeout",
			action:    "restart",
			container: types.Container{ID: "id1"},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				timeout := 10 * time.Second
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerRestart(context.Background(), "id1", &timeout)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "ContainerRestart specific timeout",
			action:    "restart",
			timeout:   "30",
			container: types.Container{ID: "id1"},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				timeout := 30 * time.Second
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerRestart(context.Background(), "id1", &timeout)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "ContainerRestart error",
			action:    "restart",
			container: types.Container{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerRestart(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("container error"))
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("container error"),
			),
		},
		{
			name:      "ContainerStop default timeout",
			action:    "stop",
			container: types.Container{ID: "id1"},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				timeout := 10 * time.Second
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerStop(context.Background(), "id1", &timeout)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "ContainerStop specific timeout",
			action:    "stop",
			timeout:   "30",
			container: types.Container{ID: "id1"},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				timeout := 30 * time.Second
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerStop(context.Background(), "id1", &timeout)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:      "ContainerStop error",
			action:    "stop",
			container: types.Container{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerStop(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("container error"))
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("container error"),
			),
		},
		{
			name:      "ContainerExec",
			action:    "exec",
			command:   "echo 'hello bob'",
			container: types.Container{ID: "id1"},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {

				server, client := net.Pipe()
				buf := bufio.NewReader(client)

				go func() {
					server.Write([]byte{1, 0, 0, 0, 0, 0, 0, 12})
					server.Write([]byte("exec stdout "))
					server.Write([]byte{2, 0, 0, 0, 0, 0, 0, 11})
					server.Write([]byte("exec stderr"))
					server.Close()
				}()

				s.EXPECT().Add(1)
				cli.EXPECT().ContainerInspect(context.Background(), "id1").Return(types.ContainerJSON{}, nil)
				cli.EXPECT().ContainerExecCreate(context.Background(), "id1", types.ExecConfig{AttachStdout: true, AttachStderr: true, Cmd: []string{"echo", "'hello", "bob'"}}).Return(types.IDResponse{ID: "execid1"}, nil)
				cli.EXPECT().ContainerExecAttach(context.Background(), "execid1", types.ExecStartCheck{}).Return(types.HijackedResponse{Conn: client, Reader: buf}, nil)
				cli.EXPECT().ContainerExecInspect(context.Background(), "execid1").Return(types.ContainerExecInspect{ExitCode: 0}, nil)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasNilError(),
				hasLogField("output", "exec stdout exec stderr"),
			),
		},
		{
			name:      "ContainerInspect error",
			action:    "exec",
			command:   "echo 'hello bob'",
			container: types.Container{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerInspect(gomock.Any(), gomock.Any()).Return(types.ContainerJSON{}, errors.New("error inspect"))
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("error inspect"),
			),
		},
		{
			name:      "ContainerExecCreate error",
			action:    "exec",
			command:   "echo 'hello bob'",
			container: types.Container{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerInspect(gomock.Any(), gomock.Any()).Return(types.ContainerJSON{}, nil)
				cli.EXPECT().ContainerExecCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.IDResponse{}, errors.New("error create"))
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("error create"),
			),
		},
		{
			name:      "ContainerExecCreate empty ID",
			action:    "exec",
			command:   "echo 'hello bob'",
			container: types.Container{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerInspect(gomock.Any(), gomock.Any()).Return(types.ContainerJSON{}, nil)
				cli.EXPECT().ContainerExecCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.IDResponse{ID: ""}, nil)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("exec ID empty"),
			),
		},
		{
			name:      "ContainerExecAttach error",
			action:    "exec",
			command:   "echo 'hello bob'",
			container: types.Container{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ContainerInspect(gomock.Any(), gomock.Any()).Return(types.ContainerJSON{}, nil)
				cli.EXPECT().ContainerExecCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.IDResponse{ID: "1"}, nil)
				cli.EXPECT().ContainerExecAttach(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.HijackedResponse{}, errors.New("error attach"))
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("error attach"),
			),
		},
		{
			name:      "StdCopy in error",
			action:    "exec",
			command:   "echo 'hello bob'",
			container: types.Container{ID: "id1"},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {

				server, client := net.Pipe()
				buf := bufio.NewReader(client)

				go func() {
					server.Write([]byte("StdCopy error header "))
					server.Close()
				}()

				s.EXPECT().Add(1)
				cli.EXPECT().ContainerInspect(gomock.Any(), gomock.Any()).Return(types.ContainerJSON{}, nil)
				cli.EXPECT().ContainerExecCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.IDResponse{ID: "1"}, nil)
				cli.EXPECT().ContainerExecAttach(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.HijackedResponse{Conn: client, Reader: buf}, nil)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasUnknowError(),
			),
		},
		{
			name:      "ContainerExecInspect error",
			action:    "exec",
			command:   "echo 'hello bob'",
			container: types.Container{ID: "id1"},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {

				server, client := net.Pipe()
				buf := bufio.NewReader(client)

				go func() {
					server.Write([]byte{1, 0, 0, 0, 0, 0, 0, 12})
					server.Write([]byte("exec stdout "))
					server.Write([]byte{2, 0, 0, 0, 0, 0, 0, 11})
					server.Write([]byte("exec stderr"))
					server.Close()
				}()

				s.EXPECT().Add(1)
				cli.EXPECT().ContainerInspect(gomock.Any(), gomock.Any()).Return(types.ContainerJSON{}, nil)
				cli.EXPECT().ContainerExecCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.IDResponse{ID: "1"}, nil)
				cli.EXPECT().ContainerExecAttach(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.HijackedResponse{Conn: client, Reader: buf}, nil)
				cli.EXPECT().ContainerExecInspect(gomock.Any(), gomock.Any()).Return(types.ContainerExecInspect{}, errors.New("error inspect"))
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("error inspect"),
			),
		},
		{
			name:      "ExitCode != 0",
			action:    "exec",
			command:   "echo 'hello bob'",
			container: types.Container{ID: "id1"},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {

				server, client := net.Pipe()
				buf := bufio.NewReader(client)

				go func() {
					server.Write([]byte{1, 0, 0, 0, 0, 0, 0, 12})
					server.Write([]byte("exec stdout "))
					server.Write([]byte{2, 0, 0, 0, 0, 0, 0, 11})
					server.Write([]byte("exec stderr"))
					server.Close()
				}()

				s.EXPECT().Add(1)
				cli.EXPECT().ContainerInspect(gomock.Any(), gomock.Any()).Return(types.ContainerJSON{}, nil)
				cli.EXPECT().ContainerExecCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.IDResponse{ID: "1"}, nil)
				cli.EXPECT().ContainerExecAttach(gomock.Any(), gomock.Any(), gomock.Any()).Return(types.HijackedResponse{Conn: client, Reader: buf}, nil)
				cli.EXPECT().ContainerExecInspect(gomock.Any(), gomock.Any()).Return(types.ContainerExecInspect{ExitCode: 1}, nil)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasError("exit status 1"),
				hasLogField("output", "exec stdout exec stderr"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// Log
			out := &bytes.Buffer{}
			log.SetOutput(out)
			log.SetFormatter(&log.JSONFormatter{})

			// Mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := NewMockJobSynchroniser(ctrl)
			cli := NewMockDockerClient(ctrl)
			if tt.mock != nil {
				tt.mock(s, cli)
			}

			c := &Cron{nil, s, nil, nil}
			j := &ContainerJob{
				Schedule:  tt.schedule,
				Action:    tt.action,
				Timeout:   tt.timeout,
				Command:   tt.command,
				Container: tt.container,
				cron:      c,
				cli:       cli,
			}

			// Act
			j.Run()

			var fields log.Fields
			err := json.Unmarshal(out.Bytes(), &fields)
			assert.NilError(t, err)

			// Assert
			for _, check := range tt.checks {
				check(t, fields)
			}
		})
	}
}
