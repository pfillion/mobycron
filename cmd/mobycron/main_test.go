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
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func TestInitApp(t *testing.T) {
	a := cli.NewApp()
	a.Before = initApp
	a.Action = func(ctx *cli.Context) error { return nil }
	os.Args = []string{"mobycron.test"}

	// Act
	err := a.Run(os.Args)

	// Assert
	assert.NilError(t, err)
	assert.Assert(t, app.cron != nil)
	assert.Assert(t, app.osChan != nil)
	// assert.Assert(t, app.handler != nil)
	assert.Assert(t, log.StandardLogger().Out == os.Stdout)
	assert.Assert(t, log.GetLevel() == log.InfoLevel)
	_, ok := log.StandardLogger().Formatter.(*log.JSONFormatter)
	assert.Assert(t, ok)
}

// func TestInitAppError(t *testing.T) {
// 	a := cli.NewApp()
// 	a.Before = initApp
// 	a.Action = func(ctx *cli.Context) error { return nil }
// 	os.Args = []string{"mobycron.test"}

// 	os.Setenv("DOCKER_HOST", "bad docker host")
// 	defer os.Unsetenv("DOCKER_HOST")

// 	// Act
// 	err := a.Run(os.Args)

// 	// Assert
// 	assert.ErrorContains(t, err, "unable to parse docker host")
// }

func TestMain(t *testing.T) {
	os.Args = []string{"mobycron"}
	var out = &bytes.Buffer{}
	exitCode := 0

	before = func(ctx *cli.Context) error {
		log.SetOutput(out)
		return nil
	}
	action = func(ctx *cli.Context) error {
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

	before = func(ctx *cli.Context) error {
		log.SetOutput(out)
		log.StandardLogger().ExitFunc = func(code int) {
			exitCode = code
		}
		return nil
	}
	action = func(ctx *cli.Context) error {
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

	type mockFunc func(*MockCron, *MockHandler)

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
		mock   mockFunc
		checks []checkFunc
	}{
		{
			name:   "run",
			osChan: make(chan os.Signal),
			sing:   syscall.SIGINT,
			mock: func(c *MockCron, h *MockHandler) {
				c.EXPECT().LoadConfig(gomock.Any()).Return(nil)
				// h.EXPECT().Scan()
				// h.EXPECT().Listen()
				c.EXPECT().Start()
				c.EXPECT().Stop()
			},
			checks: check(
				hasNilError(),
				hasOutput("cron is running and waiting signal for stop"),
			),
		},
		{
			name: "LoadConfig in error",
			mock: func(c *MockCron, h *MockHandler) {
				c.EXPECT().LoadConfig("/configs/config.json").Return(errors.New("config error"))
			},
			checks: check(
				hasError("config error"),
			),
		},
		// {
		// 	name: "Scan in error",
		// 	mock: func(c *MockCron, h *MockHandler) {
		// 		c.EXPECT().LoadConfig(gomock.Any())
		// 		h.EXPECT().Scan().Return(errors.New("scan error"))
		// 	},
		// 	checks: check(
		// 		hasError("scan error"),
		// 	),
		// },
		// {
		// 	name: "Listen in error",
		// 	mock: func(c *MockCron, h *MockHandler) {
		// 		c.EXPECT().LoadConfig(gomock.Any())
		// 		h.EXPECT().Scan()
		// 		h.EXPECT().Listen().Return(errors.New("listen error"))
		// 	},
		// 	checks: check(
		// 		hasError("listen error"),
		// 	),
		// },
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
			mc := NewMockCron(ctrl)
			mh := NewMockHandler(ctrl)
			if tt.mock != nil {
				tt.mock(mc, mh)
			}

			// Fake app
			a := cli.NewApp()
			a.Before = func(ctx *cli.Context) error {
				app = cronApp{
					cron:    mc,
					osChan:  tt.osChan,
					handler: mh,
				}

				return nil
			}
			a.Action = startApp
			os.Args = []string{"mobycron.test"}

			// Act
			err := a.Run(os.Args)

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), err)
			}
		})
	}
}
