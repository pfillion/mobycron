package cron

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Job run a command with specified args.
type Job struct {
	command string
	args    []string
	cron    *Cron
}

// Run a Job and log the output.
func (j *Job) Run() {
	log := log.WithFields(log.Fields{
		"func":    "Run",
		"command": j.command,
		"args":    strings.Join(j.args, " "),
	})

	j.cron.sync.Add(1)

	// Secret mapping
	secretMapper := func(key string) string {
		env := os.Getenv(key)
		if strings.HasSuffix(key, "__FILE") {
			data, err := afero.ReadFile(j.cron.fs, env)
			if err != nil {
				log.WithField("env", key).WithError(err).Errorln("invalid secret environment variable")
				env = ""
			} else {
				env = string(data)
			}
		}
		return env
	}

	// Expand env in all args
	args := make([]string, len(j.args))
	for i, arg := range j.args {
		args[i] = os.Expand(arg, secretMapper)
	}

	cmd := exec.Command(os.Expand(j.command, secretMapper), args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.WithField("output", string(out)).WithError(err).Errorln("job completed with error")
	} else {
		log.WithField("output", string(out)).Infoln("job completed successfully")
	}

	j.cron.sync.Done()
}
