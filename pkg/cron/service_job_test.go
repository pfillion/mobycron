package cron

import (
	"bytes"
	context "context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestServiceJobRun(t *testing.T) {
	type checkFunc func(*testing.T, string)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockJobSynchroniser, *MockDockerClient)

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, out string) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name           string
		schedule       string
		action         string
		timeout        string
		command        string
		serviceID      string
		serviceName    string
		serviceVersion swarm.Version
		service        swarm.Service
		mock           mockFunc
		checks         []checkFunc
	}{
		{
			name:           "service update",
			schedule:       "1 * * * 5",
			action:         "update",
			timeout:        "30",
			serviceID:      "ID1",
			serviceName:    "s1",
			serviceVersion: swarm.Version{Index: 1},
			service:        swarm.Service{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ServiceUpdate(context.Background(), "ID1", swarm.Version{Index: 1}, swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ForceUpdate: 1}}, types.ServiceUpdateOptions{})
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasLogField("level", "info"),
				hasLogField("func", "ServiceJob.Run"),
				hasLogField("schedule", "1 * * * 5"),
				hasLogField("service.ID", "ID1"),
				hasLogField("service.Name", "s1"),
				hasLogField("msg", "service action completed successfully"),
			),
		},
		{
			name:    "service update error",
			action:  "update",
			service: swarm.Service{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				s.EXPECT().Add(1)
				cli.EXPECT().ServiceUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(swarm.ServiceUpdateResponse{}, errors.New("update error"))
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasLogField("level", "error"),
				hasLogField("msg", "service job completed with error"),
				hasLogField("error", "update error"),
			),
		},
		{
			name:    "service update warnings",
			action:  "update",
			service: swarm.Service{},
			mock: func(s *MockJobSynchroniser, cli *MockDockerClient) {
				r := swarm.ServiceUpdateResponse{
					Warnings: []string{
						"w1",
						"w2",
					},
				}

				s.EXPECT().Add(1)
				cli.EXPECT().ServiceUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(r, nil)
				cli.EXPECT().Close()
				s.EXPECT().Done()
			},
			checks: check(
				hasLogField("level", "warning"),
				hasLogField("msg", "w1"),
				hasLogField("msg", "w2"),
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

			c := &Cron{nil, s, nil, nil, nil}
			j := &ServiceJob{
				Schedule:       tt.schedule,
				Action:         tt.action,
				ServiceID:      tt.serviceID,
				ServiceName:    tt.serviceName,
				ServiceVersion: tt.serviceVersion,
				Service:        tt.service,
				cron:           c,
				cli:            cli,
			}

			// Act
			j.Run()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String())
			}
		})
	}
}
