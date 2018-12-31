package main

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func TestAddJob(t *testing.T) {
	type checkFunc func(*testing.T, *Cron, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronRunner, *Cron)

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
		entry  *Entry
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name:  "valid cron parameters",
			entry: &Entry{"3 * * * * *", "/bin/bash", []string{"-c echo 1"}},
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * * *", &Job{"/bin/bash", []string{"-c echo 1"}, c})
			},
			checks: check(
				hasOutput("add job to cron"),
				hasNilError(),
			),
		},
		{
			name:  "empty schedule",
			entry: &Entry{"", "/bin/bash", []string{"-c echo 1"}},
			checks: check(
				hasError("schedule is required"),
			),
		},
		{
			name:  "empty command",
			entry: &Entry{"3 * * * * *", "", []string{"-c echo 1"}},
			checks: check(
				hasError("command is required"),
			),
		},
		{
			name:  "empty args",
			entry: &Entry{"3 * * * * *", "/bin/bash", []string{""}},
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), &Job{"/bin/bash", []string{""}, c})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:  "nil args",
			entry: &Entry{"3 * * * * *", "/bin/bash", nil},
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * * *", &Job{"/bin/bash", nil, c})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:  "nil entry",
			entry: nil,
			checks: check(
				hasError("entry is required"),
			),
		},
		{
			name:  "CronRunner.AddJob return error",
			entry: &Entry{"3 * * * * *", "/bin/bash", nil},
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(fmt.Errorf("a error"))
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
			r := NewMockCronRunner(ctrl)

			c := &Cron{r, nil, nil}
			if tt.mock != nil {
				tt.mock(r, c)
			}

			// Act
			err := c.AddJob(tt.entry)

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

	type mockFunc func(*MockCronRunner, *Cron)

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
		name    string
		entries []Entry
		mock    mockFunc
		checks  []checkFunc
	}{
		{
			name:    "zero entries",
			entries: []Entry{},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:    "one entries",
			entries: []Entry{{"3 * * * * *", "echo", []string{"1"}}},
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * * *", &Job{"echo", []string{"1"}, c})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name: "many entries",
			entries: []Entry{
				{"1 * * * * *", "echo1", []string{"1"}},
				{"2 * * * * *", "echo2", []string{"2"}},
			},
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob("1 * * * * *", &Job{"echo1", []string{"1"}, c})
				r.EXPECT().AddJob("2 * * * * *", &Job{"echo2", []string{"2"}, c})
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name:    "AddJob return error",
			entries: []Entry{{"3 * * * * *", "echo", []string{"1"}}},
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(fmt.Errorf("a error"))
			},
			checks: check(
				hasError("a error"),
			),
		},
		{
			name:    "nil entries",
			entries: nil,
			checks: check(
				hasError("entries is required"),
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
			r := NewMockCronRunner(ctrl)

			c := &Cron{r, nil, nil}
			if tt.mock != nil {
				tt.mock(r, c)
			}

			// Act
			err := c.AddJobs(tt.entries)

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

	type mockFunc func(*MockCronRunner)

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
			mock: func(r *MockCronRunner) {
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
			r := NewMockCronRunner(ctrl)
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

	type mockFunc func(*MockCronRunner, *MockJobSynchroniser)

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
			mock: func(r *MockCronRunner, s *MockJobSynchroniser) {
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
			r := NewMockCronRunner(ctrl)
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
	c := NewCron(afero.NewOsFs())

	// Assert
	assert.Assert(t, c.runner != nil)
	assert.Assert(t, c.sync != nil)
	assert.Assert(t, c.fs != nil)
}

func TestRun(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	hasOutput := func(want string) checkFunc {
		return func(t *testing.T, out string, err error) {
			assert.Assert(t, is.Contains(out, want))
		}
	}

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

	type args struct {
		c   chan os.Signal
		sig []os.Signal
	}

	tests := []struct {
		name   string
		args   *args
		sing   os.Signal
		checks []checkFunc
	}{
		{
			name: "stop when signal match",
			args: &args{make(chan os.Signal), []os.Signal{syscall.SIGINT}},
			sing: syscall.SIGINT,
			checks: check(
				hasNilError(),
				hasOutput("start cron"),
				hasOutput("cron is running and waiting signal for stop"),
				hasOutput("cron is stopped, all jobs are completed"),
			),
		},
		{
			name: "stop when any signal",
			args: &args{make(chan os.Signal), []os.Signal{}},
			sing: syscall.SIGINT,
			checks: check(
				hasNilError(),
				hasOutput("start cron"),
				hasOutput("cron is running and waiting signal for stop"),
				hasOutput("cron is stopped, all jobs are completed"),
			),
		},
		{
			name: "nil sig",
			args: &args{make(chan os.Signal), nil},
			sing: syscall.SIGINT,
			checks: check(
				hasNilError(),
				hasOutput("start cron"),
				hasOutput("cron is running and waiting signal for stop"),
				hasOutput("cron is stopped, all jobs are completed"),
			),
		},
		{
			name: "nil ch",
			args: &args{nil, []os.Signal{syscall.SIGINT}},
			checks: check(
				hasError("channel is required"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			if tt.args.c != nil {
				go func() {
					tt.args.c <- tt.sing
				}()
			}

			c := NewCron(afero.NewOsFs())

			// Act
			err := c.Run(tt.args.c, tt.args.sig...)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	type checkFunc func(*testing.T, *Cron, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronRunner, *Cron)

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
			name:     "one entry",
			filename: "/config/config.json",
			config: `[
						{
							"schedule": "0/2 * * 12 * *",
							"command": "echo",
							"args": [
								"boby"
							]
						}
					]`,
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob("0/2 * * 12 * *", &Job{"echo", []string{"boby"}, c})
			},
			checks: check(
				hasNilError(),
				hasOutput("load config file"),
			),
		},
		{
			name:     "many entries",
			filename: "/config/config.json",
			config: `[
						{
							"schedule": "0/2 * * 12 * *",
							"command": "command1",
							"args": [
								"arg1"
							]
						},
						{
							"schedule": "5 5 * * * *",
							"command": "command2",
							"args": [
								"arg2"
							]
						}
					]`,
			mock: func(r *MockCronRunner, c *Cron) {
				r.EXPECT().AddJob("0/2 * * 12 * *", &Job{"command1", []string{"arg1"}, c})
				r.EXPECT().AddJob("5 5 * * * *", &Job{"command2", []string{"arg2"}, c})
			},
			checks: check(
				hasNilError(),
				hasOutput("load config file"),
			),
		},
		{
			name:     "file not exist",
			filename: "/config/config.json",
			config:   "",
			checks: check(
				hasError("failed to read config file"),
			),
		},
		{
			name:     "json not valid",
			filename: "/config/config.json",
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
			name:     "invalid entry",
			filename: "/config/config.json",
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
				hasError("failed to add jobs entries fron config file"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			out := &bytes.Buffer{}
			log.SetOutput(out)

			fs := afero.NewMemMapFs()
			if tt.config != "" {
				afero.WriteFile(fs, tt.filename, []byte(tt.config), 0640)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := NewMockCronRunner(ctrl)

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
