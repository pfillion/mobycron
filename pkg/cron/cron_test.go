package cron

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name   string
		job    Job
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "valid job",
			job:  Job{"3 * * * *", "/bin/bash", []string{"-c echo 1"}, nil},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * *", &Job{"3 * * * *", "/bin/bash", []string{"-c echo 1"}, c})
			},
			checks: check(
				hasNilError(),
				hasLogField("msg", "add job to cron"),
			),
		},
		{
			name: "job with empty schedule",
			job:  Job{"", "/bin/bash", []string{"-c echo 1"}, nil},
			checks: check(
				hasError("schedule is required"),
			),
		},
		{
			name: "job with empty command",
			job:  Job{"3 * * * *", "", []string{"-c echo 1"}, nil},
			checks: check(
				hasError("command is required"),
			),
		},
		{
			name: "job with empty args",
			job:  Job{"3 * * * *", "/bin/bash", []string{""}, nil},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), &Job{"3 * * * *", "/bin/bash", []string{""}, c})
			},
			checks: check(
				hasNilError(),
				hasLogField("msg", "add job to cron"),
			),
		},
		{
			name: "job with nil args",
			job:  Job{"3 * * * *", "/bin/bash", nil, nil},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * *", &Job{"3 * * * *", "/bin/bash", nil, c})
			},
			checks: check(
				hasNilError(),
				hasLogField("msg", "add job to cron"),
			),
		},
		{
			name: "CronRunner.AddJob return error",
			job:  Job{"3 * * * *", "/bin/bash", nil, nil},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(cron.EntryID(0), fmt.Errorf("a error"))
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

			c := &Cron{r, nil, nil, nil}
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

			c := &Cron{r, nil, nil, nil}
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

func TestAddContainerJob(t *testing.T) {
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

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
		}
	}

	tests := []struct {
		name   string
		job1   ContainerJob
		job2   *ContainerJob
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name: "simple container job",
			job1: ContainerJob{Schedule: "3 * * * *", Action: "start"},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * *", &ContainerJob{Schedule: "3 * * * *", Action: "start", cron: c})
			},
			checks: check(
				hasNilError(),
				hasLogField("msg", "add container job to cron"),
			),
		},
		{
			name: "job with empty schedule",
			job1: ContainerJob{Schedule: ""},
			checks: check(
				hasError("schedule is required"),
			),
		},
		{
			name: "job with empty timeout",
			job1: ContainerJob{Schedule: "3 * * * *", Action: "restart", Timeout: ""},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("3 * * * *", &ContainerJob{Schedule: "3 * * * *", Action: "restart", cron: c})
			},
			checks: check(
				hasNilError(),
				hasLogField("msg", "add container job to cron"),
			),
		},
		{
			name: "invalid timeout",
			job1: ContainerJob{Schedule: "3 * * * *", Action: "restart", Timeout: "invalid"},
			checks: check(
				hasError("invalid container timeout, only integer are permitted"),
			),
		},
		{
			name: "invalid action",
			job1: ContainerJob{Schedule: "3 * * * *", Action: "invalid"},
			checks: check(
				hasError("invalid container action, only 'start', 'restart', 'stop' and 'exec' are permitted"),
			),
		},
		{
			name: "invalid command when action is start",
			job1: ContainerJob{Schedule: "* * * * *", Action: "start", Command: "ls"},
			checks: check(
				hasError("a command can be specified only with 'exec' action"),
			),
		},
		{
			name: "invalid command when action is restart",
			job1: ContainerJob{Schedule: "* * * * *", Action: "restart", Command: "ls"},
			checks: check(
				hasError("a command can be specified only with 'exec' action"),
			),
		},
		{
			name: "invalid command when action is stop",
			job1: ContainerJob{Schedule: "* * * * *", Action: "stop", Command: "ls"},
			checks: check(
				hasError("a command can be specified only with 'exec' action"),
			),
		},
		{
			name: "valid command when action is exec",
			job1: ContainerJob{Schedule: "* * * * *", Action: "exec", Command: "ls"},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), gomock.Any())
			},
			checks: check(
				hasNilError(),
			),
		},
		{
			name: "command required when action is exec",
			job1: ContainerJob{Schedule: "* * * * *", Action: "exec", Command: ""},
			checks: check(
				hasError("command is required"),
			),
		},
		{
			name: "CronRunner.AddJob return error",
			job1: ContainerJob{Schedule: "3 * * * *", Action: "start"},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob(gomock.Any(), gomock.Any()).Return(cron.EntryID(0), errors.New("a error"))
			},
			checks: check(
				hasError("a error"),
				hasError("failed to add container job in cron"),
			),
		},
		{
			name: "replace job from service by one in same slot",
			job1: ContainerJob{Schedule: "1 * * * *", Action: "start", ServiceID: "s1", Slot: 1, Created: 1},
			job2: &ContainerJob{Schedule: "2 * * * *", Action: "start", ServiceID: "s1", Slot: 1, Created: 2},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("1 * * * *", &ContainerJob{Schedule: "1 * * * *", Action: "start", ServiceID: "s1", Slot: 1, Created: 1, cron: c}).Return(cron.EntryID(1), nil)
				r.EXPECT().Remove(cron.EntryID(1))
				r.EXPECT().AddJob("2 * * * *", &ContainerJob{Schedule: "2 * * * *", Action: "start", ServiceID: "s1", Slot: 1, Created: 2, cron: c})
			},
			checks: check(
				hasNilError(),
				hasLogField("msg", "replace container job in cron"),
			),
		},
		{
			name: "don't replace job from service when is older",
			job1: ContainerJob{Schedule: "1 * * * *", Action: "start", ServiceID: "s1", Slot: 1, Created: 2},
			job2: &ContainerJob{Schedule: "2 * * * *", Action: "start", ServiceID: "s1", Slot: 1, Created: 1},
			mock: func(r *MockRunner, c *Cron) {
				r.EXPECT().AddJob("1 * * * *", &ContainerJob{Schedule: "1 * * * *", Action: "start", ServiceID: "s1", Slot: 1, Created: 2, cron: c}).Return(cron.EntryID(1), nil)
			},
			checks: check(
				hasNilError(),
				hasLogField("msg", "skip replacement, the container job is older"),
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

			c := &Cron{r, nil, nil, make(map[string]cronEntry)}
			if tt.mock != nil {
				tt.mock(r, c)
			}

			// Act
			err := c.AddContainerJob(tt.job1)
			if tt.job2 != nil {
				err = c.AddContainerJob(*tt.job2)
			}

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
			name: "start",
			mock: func(r *MockRunner) {
				r.EXPECT().Start()
			},
			checks: check(
				hasLogField("msg", "start cron"),
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

			c := &Cron{r, nil, nil, nil}

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
			name: "stop",
			mock: func(r *MockRunner, s *MockJobSynchroniser) {
				r.EXPECT().Stop()
				s.EXPECT().Wait()
			},
			checks: check(
				hasLogField("msg", "stopping cron, wait for running jobs"),
				hasLogField("msg", "cron is stopped, all jobs are completed"),
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

			c := &Cron{r, s, nil, nil}

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
	c := NewCron(false)

	// Assert
	assert.Assert(t, c.runner != nil)
	assert.Assert(t, c.sync != nil)
	assert.Assert(t, c.fs != nil)
	assert.Assert(t, len(c.jobEntries) == 0)

	err := c.AddJob(Job{
		Schedule: "* * * * * *",
		Command:  "sh",
	})
	assert.Assert(t, is.ErrorContains(err, "expected exactly 5 fields, found 6"))
}

func TestNewCronParseSecond(t *testing.T) {
	// Act
	c := NewCron(true)

	// Assert
	assert.Assert(t, c.runner != nil)
	assert.Assert(t, c.sync != nil)
	assert.Assert(t, c.fs != nil)
	assert.Assert(t, len(c.jobEntries) == 0)

	err := c.AddJob(Job{
		Schedule: "* * * * * *",
		Command:  "sh",
	})
	assert.NilError(t, err)
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

	hasLogField := func(field string, want string) checkFunc {
		return func(t *testing.T, c *Cron, out string, err error) {
			assert.Assert(t, is.Contains(out, fmt.Sprintf("\"%s\":\"%s\"", field, want)))
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
				hasLogField("msg", "load config file"),
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
				hasLogField("msg", "load config file"),
			),
		},
		{
			name:     "error read config file",
			filename: "/configs/config.json",
			config:   "",
			checks: check(
				hasError("failed to read config file"),
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
			if tt.config != "" {
				afero.WriteFile(fs, tt.filename, []byte(tt.config), 0640)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := NewMockRunner(ctrl)

			c := &Cron{r, nil, fs, nil}
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
