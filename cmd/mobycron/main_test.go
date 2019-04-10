package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"syscall"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func TestMain(t *testing.T) {
	type checkFunc func(*testing.T, string, int)
	check := func(fns ...checkFunc) []checkFunc { return fns }

	hasOutput := func(want string) checkFunc {
		return func(t *testing.T, out string, exitCode int) {
			assert.Assert(t, is.Contains(out, want))
		}
	}

	hasExitCode := func(want int) checkFunc {
		return func(t *testing.T, out string, exitCode int) {
			assert.Assert(t, is.Equal(exitCode, want))
		}
	}

	hasLogLevel := func(want log.Level) checkFunc {
		return func(t *testing.T, out string, exitCode int) {
			assert.Assert(t, is.Equal(log.GetLevel(), want))
		}
	}

	hasJSONOutput := func() checkFunc {
		return func(t *testing.T, out string, exitCode int) {
			for i, line := range strings.Split(out, "\n") {
				if line != "" {
					var js interface{}
					assert.NilError(t, json.Unmarshal([]byte(line), &js), "line %d: %s", i, line)
				}
			}
		}
	}

	tests := []struct {
		name   string
		osChan chan os.Signal
		checks []checkFunc
	}{
		{
			name:   "run",
			osChan: make(chan os.Signal),
			checks: check(
				hasExitCode(0),
				hasLogLevel(log.InfoLevel),
				hasJSONOutput(),
				hasOutput("cron is stopped, all jobs are completed"),
			),
		},
		{
			name:   "error on run",
			osChan: nil,
			checks: check(
				hasExitCode(1),
				hasJSONOutput(),
				hasOutput("channel is required"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// Replace log output
			var out = &bytes.Buffer{}
			output = out

			// Replace exiter func in real code by a fake one
			var exitCode int
			oldExiter := exiter
			exiter = func(code int) {
				exitCode = code
			}
			defer func() {
				exiter = oldExiter
			}()

			// Fake config file
			fs = afero.NewMemMapFs()

			// Send terminating signal
			osChan = tt.osChan
			go func() {
				tt.osChan <- syscall.SIGTERM
			}()

			// Act
			main()

			// Assert
			for _, check := range tt.checks {
				check(t, out.String(), exitCode)
			}
		})
	}
}
