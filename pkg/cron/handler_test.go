package cron

import (
	"bytes"
	context "context"
	"encoding/json"
	"os"
	"testing"

	types "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	gomock "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
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
	type checkFunc func(*testing.T, log.Fields, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockDockerClient)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, fields log.Fields, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, fields log.Fields, err error) {
			assert.NilError(t, err)
			// assert.Assert(t, is.Equal(fields["level"], "info"), "log fields: %v", fields)
			// assert.Assert(t, is.Nil(fields["error"]), "log fields: %v", fields)
		}
	}

	hasLogError := func(want string) checkFunc {
		return func(t *testing.T, fields log.Fields, err error) {
			assert.Assert(t, is.Equal(fields["level"], "error"), "log fields: %v", fields)
			assert.Assert(t, is.Equal(fields["error"], want), "log fields: %v", fields)
		}
	}

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, fields log.Fields, err error) {
			assert.Assert(t, is.Equal(fields[field], want), "log fields: %v", fields)
		}
	}

	tests := []struct {
		name   string
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "zero container",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return([]types.Container{}, nil)
				cli.EXPECT().Close()
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name: "one container",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				args := filters.NewArgs()
				args.Add("label", "mobycron.schedule")

				ctx := context.Background()
				opt := types.ContainerListOptions{All: true, Filters: args}
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
				// TODO: cleanup
				// hasLogField("func", "Handler.Scan"),
				// hasLogField("msg", "scan containers for cron schedule"),
			),
		},
		{
			name: "many containers",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				args := filters.NewArgs()
				args.Add("label", "mobycron.schedule")

				ctx := context.Background()
				opt := types.ContainerListOptions{All: true, Filters: args}
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
			name: "ContainerList in error",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(nil, errors.New("ContainerList in error"))
				cli.EXPECT().Close()
			},
			checks: check(
				hasError("ContainerList in error"),
			),
		},
		{
			name: "AddContainerJob in error",
			mock: func(sc *MockCronner, cli *MockDockerClient) {
				containers := []types.Container{{ID: "1"}}
				cli.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(containers, nil)
				sc.EXPECT().AddContainerJob(gomock.Any()).Return(errors.New("AddContainerJob in error"))
				cli.EXPECT().Close()
			},
			checks: check(
				hasLogError("AddContainerJob in error"),
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
				tt.mock(cron, cli)
			}

			// Act
			err := h.Scan()

			// Assert
			var fields log.Fields
			if out.Len() > 0 {
				errLog := json.Unmarshal(out.Bytes(), &fields)
				assert.NilError(t, errLog)
			}

			for _, check := range tt.checks {
				check(t, fields, err)
			}
		})
	}
}

func TestListen(t *testing.T) {
	// Arrange
	out := &bytes.Buffer{}
	log.SetOutput(out)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cron := NewMockCronner(ctrl)
	cli := NewMockDockerClient(ctrl)

	h := &Handler{cron, cli}
	// if tt.mock != nil {
	// 	tt.mock(cron, cli)
	// }

	// Act
	err := h.Listen()

	// Assert
	assert.NilError(t, err)
	// var fields log.Fields
	// if out.Len() > 0 {
	// 	errLog := json.Unmarshal(out.Bytes(), &fields)
	// 	assert.NilError(t, errLog)
	// }

	// for _, check := range tt.checks {
	// 	check(t, fields, err)
	// }
}
