package cron

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
	"gotest.tools/env"
)

func TestJobRun(t *testing.T) {
	type checkFunc func(*testing.T, string)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockJobSynchroniser)

	hasOutput := func(want string) checkFunc {
		return func(t *testing.T, out string) {
			assert.Assert(t, is.Contains(out, want))
		}
	}

	tests := []struct {
		name           string
		command        string
		args           []string
		envs           map[string]string
		secret         string
		secretFilename string
		mock           mockFunc
		checks         []checkFunc
	}{
		{
			name:    "run job",
			command: "echo",
			args:    []string{"1"},
			mock: func(s *MockJobSynchroniser) {
				s.EXPECT().Add(1)
				s.EXPECT().Done()
			},
			checks: check(
				hasOutput("job completed successfully"),
			),
		},
		{
			name:    "command with env variable",
			command: "$CMD",
			args:    []string{"hello bob"},
			envs: map[string]string{
				"CMD": "echo",
			},
			mock: func(s *MockJobSynchroniser) {
				s.EXPECT().Add(1)
				s.EXPECT().Done()
			},
			checks: check(
				hasOutput("job completed successfully"),
				hasOutput("hello bob"),
			),
		},
		{
			name:    "command with secret env variable",
			command: "$CMD__FILE",
			args:    []string{"hello bob"},
			envs: map[string]string{
				"CMD__FILE": "/run/secret/secretFile",
			},
			secret:         "echo",
			secretFilename: "/run/secret/secretFile",
			mock: func(s *MockJobSynchroniser) {
				s.EXPECT().Add(1)
				s.EXPECT().Done()
			},
			checks: check(
				hasOutput("job completed successfully"),
				hasOutput("hello bob"),
			),
		},
		{
			name:    "command with invalid secret env variable",
			command: "$CMD__FILE",
			args:    []string{"hello bob"},
			envs: map[string]string{
				"CMD__FILE": "/path/not/exists",
			},
			mock: func(s *MockJobSynchroniser) {
				s.EXPECT().Add(1)
				s.EXPECT().Done()
			},
			checks: check(
				hasOutput("invalid secret environment variable"),
			),
		},
		{
			name:    "args with env variable",
			command: "echo",
			args:    []string{"hello $NAME"},
			envs: map[string]string{
				"NAME": "bob",
			},
			mock: func(s *MockJobSynchroniser) {
				s.EXPECT().Add(1)
				s.EXPECT().Done()
			},
			checks: check(
				hasOutput("job completed successfully"),
				hasOutput("hello bob"),
			),
		},
		{
			name:    "args with secret env variable",
			command: "echo",
			args:    []string{"hello $NAME__FILE"},
			envs: map[string]string{
				"NAME__FILE": "/run/secret/secretFile",
			},
			secret:         "bob",
			secretFilename: "/run/secret/secretFile",
			mock: func(s *MockJobSynchroniser) {
				s.EXPECT().Add(1)
				s.EXPECT().Done()
			},
			checks: check(
				hasOutput("job completed successfully"),
				hasOutput("hello bob"),
			),
		},
		{
			name:    "args with invalid secret env variable",
			command: "echo",
			args:    []string{"hello $NAME__FILE"},
			envs: map[string]string{
				"NAME__FILE": "/path/not/exists",
			},
			mock: func(s *MockJobSynchroniser) {
				s.EXPECT().Add(1)
				s.EXPECT().Done()
			},
			checks: check(
				hasOutput("invalid secret environment variable"),
			),
		},
		{
			name:    "invalid command",
			command: "invalid command",
			args:    nil,
			mock: func(s *MockJobSynchroniser) {
				s.EXPECT().Add(1)
				s.EXPECT().Done()
			},
			checks: check(
				hasOutput("job completed with error"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// Env variables
			for k, v := range tt.envs {
				f := env.Patch(t, k, v)
				defer f()
			}

			// File system
			fs := afero.NewMemMapFs()
			if tt.secret != "" {
				afero.WriteFile(fs, tt.secretFilename, []byte(tt.secret), 0640)
			}

			// Log
			out := &bytes.Buffer{}
			log.SetOutput(out)

			// Mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := NewMockJobSynchroniser(ctrl)
			if tt.mock != nil {
				tt.mock(s)
			}

			c := &Cron{nil, s, fs}
			j := &Job{"3 * * * * *", tt.command, tt.args, c}

			// Act
			j.Run()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String())
			}
		})
	}
}
