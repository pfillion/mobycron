package main

import (
	"bytes"
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestInitAppModeSwarm(t *testing.T) {
	cmdRoot.Before = initApp
	cmdRoot.Action = func(ctx *cli.Context) error { return nil }
	args := []string{"mobycron", "--docker-mode=swarm"}

	// Act
	err := cmdRoot.Run(args)

	// Assert
	assert.NilError(t, err)
	assert.Assert(t, cronner != nil)
	assert.Assert(t, osChan != nil)
	assert.Assert(t, handler != nil)
	assert.Assert(t, log.StandardLogger().Out == os.Stdout)
	assert.Assert(t, log.GetLevel() == log.InfoLevel)
	_, ok := log.StandardLogger().Formatter.(*log.JSONFormatter)
	assert.Assert(t, ok)
}

func TestInitAppModeContainer(t *testing.T) {
	cmdRoot.Before = initApp
	cmdRoot.Action = func(ctx *cli.Context) error { return nil }
	args := []string{"mobycron", "--docker-mode=container"}

	// Act
	err := cmdRoot.Run(args)

	// Assert
	assert.NilError(t, err)
	assert.Assert(t, cronner != nil)
	assert.Assert(t, osChan != nil)
	assert.Assert(t, handler != nil)
	assert.Assert(t, log.StandardLogger().Out == os.Stdout)
	assert.Assert(t, log.GetLevel() == log.InfoLevel)
	_, ok := log.StandardLogger().Formatter.(*log.JSONFormatter)
	assert.Assert(t, ok)
}

func TestInitAppHandlerError(t *testing.T) {
	cmdRoot.Before = initApp
	cmdRoot.Action = func(ctx *cli.Context) error { return nil }
	args := []string{"mobycron", "--docker-mode=swarm"}

	os.Setenv("DOCKER_HOST", "bad docker host")
	defer os.Unsetenv("DOCKER_HOST")

	// Act
	err := cmdRoot.Run(args)

	// Assert
	assert.ErrorContains(t, err, "unable to parse docker host")
}

func TestInitAppModeNone(t *testing.T) {
	cmdRoot.Before = initApp
	cmdRoot.Action = func(ctx *cli.Context) error { return nil }
	args := []string{"mobycron", "--docker-mode=none"}

	// Act
	err := cmdRoot.Run(args)

	// Assert
	assert.NilError(t, err)
	assert.Assert(t, is.Nil(handler))
}

func TestInitAppModeInvalid(t *testing.T) {
	cmdRoot.Before = initApp
	cmdRoot.Action = func(ctx *cli.Context) error { return nil }
	args := []string{"mobycron", "--docker-mode=sdfsdf"}

	// Act
	err := cmdRoot.Run(args)

	// Assert
	assert.Error(t, err, "docker-mode flag is invalid")
}

func TestMain(t *testing.T) {
	os.Args = []string{"mobycron"}
	var out = &bytes.Buffer{}
	exitCode := 0

	cmdRoot.Before = func(ctx *cli.Context) error {
		log.SetOutput(out)
		return nil
	}
	cmdRoot.Action = func(ctx *cli.Context) error {
		log.Infoln("completed")
		return nil
	}

	// Act
	main()

	// Assert
	assert.Assert(t, is.Equal(0, exitCode))
	assert.Assert(t, is.Contains(out.String(), "completed"))
}

func TestMainError(t *testing.T) {
	os.Args = []string{"mobycron"}
	var out = &bytes.Buffer{}
	exitCode := 0

	cmdRoot.Before = func(ctx *cli.Context) error {
		log.SetOutput(out)
		log.StandardLogger().ExitFunc = func(code int) {
			exitCode = code
		}
		return nil
	}
	cmdRoot.Action = func(ctx *cli.Context) error {
		return errors.New("error in main app")
	}

	// Act
	main()

	// Assert
	assert.Assert(t, is.Equal(1, exitCode))
	assert.Assert(t, is.Contains(out.String(), "fatal"))
	assert.Assert(t, is.Contains(out.String(), "error in main app"))
}

func TestStartApp(t *testing.T) {
	type checkFunc func(*testing.T, string, error)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	type mockFunc func(*MockCronner, *MockHandler)

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

	tests := []struct {
		name   string
		osChan chan os.Signal
		sing   os.Signal
		args   []string
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name:   "run docker mode - swarm",
			osChan: make(chan os.Signal),
			sing:   syscall.SIGINT,
			args:   []string{"mobycron", "--docker-mode=swarm"},
			mock: func(c *MockCronner, h *MockHandler) {
				h.EXPECT().ScanService()
				h.EXPECT().ListenService()
				c.EXPECT().Start()
				c.EXPECT().Stop()
			},
			checks: check(
				hasNilError(),
				hasOutput("cron is running and waiting signal for stop"),
			),
		},
		{
			name:   "run docker mode - container",
			osChan: make(chan os.Signal),
			sing:   syscall.SIGINT,
			args:   []string{"mobycron", "--docker-mode=container"},
			mock: func(c *MockCronner, h *MockHandler) {
				h.EXPECT().ScanContainer()
				h.EXPECT().ListenContainer()
				c.EXPECT().Start()
				c.EXPECT().Stop()
			},
			checks: check(
				hasNilError(),
				hasOutput("cron is running and waiting signal for stop"),
			),
		},
		{
			name:   "run docker mode - none",
			osChan: make(chan os.Signal),
			sing:   syscall.SIGINT,
			args:   []string{"mobycron", "--docker-mode=none"},
			mock: func(c *MockCronner, h *MockHandler) {
				c.EXPECT().Start()
				c.EXPECT().Stop()
			},
			checks: check(
				hasNilError(),
				hasOutput("cron is running and waiting signal for stop"),
			),
		},
		{
			name:   "run config file",
			osChan: make(chan os.Signal),
			sing:   syscall.SIGINT,
			args:   []string{"mobycron", "--docker-mode=none", "--config-file=/etc/mobycron/config.json"},
			mock: func(c *MockCronner, h *MockHandler) {
				c.EXPECT().LoadConfig("/etc/mobycron/config.json").Return(nil)
				c.EXPECT().Start()
				c.EXPECT().Stop()
			},
			checks: check(
				hasNilError(),
				hasOutput("cron is running and waiting signal for stop"),
			),
		},
		{
			name: "run config file in error",
			args: []string{"mobycron", "--docker-mode=none", "--config-file=/etc/mobycron/config.json"},
			mock: func(c *MockCronner, h *MockHandler) {
				c.EXPECT().LoadConfig(gomock.Any()).Return(errors.New("config error"))
			},
			checks: check(
				hasError("config error"),
			),
		},
		{
			name: "ScanContainer in error",
			args: []string{"mobycron", "--docker-mode=container"},
			mock: func(c *MockCronner, h *MockHandler) {
				h.EXPECT().ScanContainer().Return(errors.New("scan error"))
			},
			checks: check(
				hasError("scan error"),
			),
		},
		{
			name: "ScanService in error",
			args: []string{"mobycron", "--docker-mode=swarm"},
			mock: func(c *MockCronner, h *MockHandler) {
				h.EXPECT().ScanService().Return(errors.New("scan error"))
			},
			checks: check(
				hasError("scan error"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// Replace log output
			var out = &bytes.Buffer{}
			log.SetOutput(out)

			// Send terminating signal
			go func() {
				tt.osChan <- tt.sing
			}()

			// Mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mc := NewMockCronner(ctrl)
			mh := NewMockHandler(ctrl)
			if tt.mock != nil {
				tt.mock(mc, mh)
			}

			// Inject mocks
			cmdRoot.Before = func(ctx *cli.Context) error {
				cronner = mc
				osChan = tt.osChan
				handler = mh
				return nil
			}
			cmdRoot.Action = startApp

			// Act
			err := cmdRoot.Run(tt.args)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}
