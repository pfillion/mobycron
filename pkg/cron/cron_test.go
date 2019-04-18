package cron

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	cron "gopkg.in/robfig/cron.v3"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func TestAddJob(t *testing.T) {
	type checkFunc func(*testing.T, *Cron, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockRunner, *Cron)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.NilError(t, err)
		}
	}

	hasOutput := func(want string) checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.Assert(t, is.Contains(out, want))
		}
	}

	tests := []struct {
		name   string
		job    *Job
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "valid job",
			job:  &Job{"3 * * * *", "/bin/bash", []string{"-c echo 1"}, nil},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * *", &Job{"3 * * * *", "/bin/bash", []string{"-c echo 1"}, c})
			},
			checks: check(
				hasOutput("add job to cron"),
				hasNilError(),
			),
		},
		{
			name: "job with empty schedule",
			job:  &Job{"", "/bin/bash", []string{"-c echo 1"}, nil},
			checks: check(
				hasError("schedule is required"),
			),
		},
		{
			name: "job with empty command",
			job:  &Job{"3 * * * *", "", []string{"-c echo 1"}, nil},
			checks: check(
				hasError("command is required"),
			),
		},
		{
			name: "job with empty args",
			job:  &Job{"3 * * * *", "/bin/bash", []string{""}, nil},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), &Job{"3 * * * *", "/bin/bash", []string{""}, c})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name: "job with nil args",
			job:  &Job{"3 * * * *", "/bin/bash", nil, nil},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * *", &Job{"3 * * * *", "/bin/bash", nil, c})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name: "nil jon",
			job:  nil,
			checks: check(
				hasError("job is required"),
			),
		},
		{
			name: "CronRunner.AddJob return error",
			job:  &Job{"3 * * * *", "/bin/bash", nil, nil},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(cron.EntryID(1), fmt.Errorf("a error"))
			},
			checks: check(
				hasError("a error"),
				hasError("failed to add job in cron"),
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
			r := NewMockRunner(ctrl)

			c := &Cron{r, nil, nil}
			if tt.mock != nil {
				tt.mock(r, c)
			}

			// Act
			err := c.AddJob(tt.job)

			// Assert
			for _, check := range tt.checks {
				check(t, c, out.String(), err)
			}
		})
	}
}

func TestAddJobs(t *testing.T) {
	type checkFunc func(*testing.T, *Cron, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockRunner, *Cron)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.NilError(t, err)
		}
	}

	tests := []struct {
		name   string
		jobs   []Job
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "zero jobs",
			jobs: []Job{},
			checks: check(
				hasNilError(),
			),
		},
		{
			name: "one job",
			jobs: []Job{{"3 * * * *", "echo", []string{"1"}, nil}},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * *", &Job{"3 * * * *", "echo", []string{"1"}, c})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name: "many jobs",
			jobs: []Job{
				{"1 * * * *", "echo1", []string{"1"}, nil},
				{"2 * * * *", "echo2", []string{"2"}, nil},
			},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("1 * * * *", &Job{"1 * * * *", "echo1", []string{"1"}, c})
				r.EXPECT().AddJob("2 * * * *", &Job{"2 * * * *", "echo2", []string{"2"}, c})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name: "AddJob return error",
			jobs: []Job{{"3 * * * *", "echo", []string{"1"}, nil}},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(cron.EntryID(1), fmt.Errorf("a error"))
			},
			checks: check(
				hasError("a error"),
			),
		},
		{
			name: "nil jobs",
			jobs: nil,
			checks: check(
				hasError("jobs is required"),
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
			r := NewMockRunner(ctrl)

			c := &Cron{r, nil, nil}
			if tt.mock != nil {
				tt.mock(r, c)
			}

			// Act
			err := c.AddJobs(tt.jobs)

			// Assert
			for _, check := range tt.checks {
				check(t, c, out.String(), err)
			}
		})
	}
}

func TestStart(t *testing.T) {
	type checkFunc func(*testing.T, string)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockRunner)

	hasOutput := func(want string) checkFunc {
		return func(t *testing.T, out string) {
			assert.Assert(t, is.Contains(out, want))
		}
	}

	tests := []struct {
		name   string
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "start",
			mock: func(r *MockRunner) {
				r.EXPECT().Start()
			},
			checks: check(
				hasOutput("start cron"),
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
			r := NewMockRunner(ctrl)
			if tt.mock != nil {
				tt.mock(r)
			}

			c := &Cron{r, nil, nil}

			// Act
			c.Start()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String())
			}
		})
	}
}

func TestStop(t *testing.T) {
	type checkFunc func(*testing.T, string)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockRunner, *MockJobSynchroniser)

	hasOutput := func(want string) checkFunc {
		return func(t *testing.T, out string) {
			assert.Assert(t, is.Contains(out, want))
		}
	}

	tests := []struct {
		name   string
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "stop",
			mock: func(r *MockRunner, s *MockJobSynchroniser) {
				r.EXPECT().Stop()
				s.EXPECT().Wait()
			},
			checks: check(
				hasOutput("stopping cron, wait for running jobs"),
				hasOutput("cron is stopped, all jobs are completed"),
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
			r := NewMockRunner(ctrl)
			s := NewMockJobSynchroniser(ctrl)
			if tt.mock != nil {
				tt.mock(r, s)
			}

			c := &Cron{r, s, nil}

			// Act
			c.Stop()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String())
			}
		})
	}
}

func TestNewCron(t *testing.T) {
	// Act
	c := NewCron()

	// Assert
	assert.Assert(t, c.runner != nil)
	assert.Assert(t, c.sync != nil)
	assert.Assert(t, c.fs != nil)
}

func TestLoadConfig(t *testing.T) {
	type checkFunc func(*testing.T, *Cron, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockRunner, *Cron)

	hasError := func(want string) checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.Assert(t, is.ErrorContains(err, want))
		}
	}

	hasNilError := func() checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.NilError(t, err)
		}
	}

	hasOutput := func(want string) checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.Assert(t, is.Contains(out, want))
		}
	}

	tests := []struct {
		name     string
		filename string
		config   string
		mock     mockFunc
		checks   []checkFunc
	}{
		{
			name:     "one job",
			filename: "/configs/config.json",
			config: `[
						{
							"schedule": "0/2 * * 12 *",
							"command": "echo",
							"args": [
								"boby"
							]
						}
					]`,
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("0/2 * * 12 *", &Job{"0/2 * * 12 *", "echo", []string{"boby"}, c})
			},
			checks: check(
				hasNilError(),
				hasOutput("load config file"),
			),
		},
		{
			name:     "many jobs",
			filename: "/configs/config.json",
			config: `[
						{
							"schedule": "0/2 * * 12 *",
							"command": "command1",
							"args": [
								"arg1"
							]
						},
						{
							"schedule": "5 5 * * *",
							"command": "command2",
							"args": [
								"arg2"
							]
						}
					]`,
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("0/2 * * 12 *", &Job{"0/2 * * 12 *", "command1", []string{"arg1"}, c})
				r.EXPECT().AddJob("5 5 * * *", &Job{"5 5 * * *", "command2", []string{"arg2"}, c})
			},
			checks: check(
				hasNilError(),
				hasOutput("load config file"),
			),
		},
		{
			name: "file not exist",
			checks: check(
				hasNilError(),
				hasOutput("no config was loaded, file not exist"),
			),
		},
		{
			name:     "json not valid",
			filename: "/configs/config.json",
			config: `[
						{
							error
						},
					]`,
			checks: check(
				hasError("failed to parse JSON data from config file"),
			),
		},
		{
			name:     "invalid job",
			filename: "/configs/config.json",
			config: `[
						{
							"schedule": "",
							"command": "echo",
							"args": [
								"boby"
							]
						}
					]`,
			checks: check(
				hasError("failed to add jobs fron config file"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			fs := afero.NewMemMapFs()
			if tt.filename != "" {
				afero.WriteFile(fs, tt.filename, []byte(tt.config), 0640)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := NewMockRunner(ctrl)

			c := &Cron{r, nil, fs}
			if tt.mock != nil {
				tt.mock(r, c)
			}

			// Act
			err := c.LoadConfig(tt.filename)

			// Assert
			for _, check := range tt.checks {
				check(t, c, out.String(), err)
			}
		})
	}
}
