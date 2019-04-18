package cron

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Job run a command with specified args on a schedule.
type Job struct {
	Schedule string   `json:"schedule"`
	Command  string   `json:"command"`
	Args     []string `json:"args"`
	cron     *Cron
}

// Run a Job and log the output.
func (j *Job) Run() {
	log := log.WithFields(log.Fields{
		"func":     "Job.Run",
		"schedule": j.Schedule,
		"command":  j.Command,
		"args":     strings.Join(j.Args, " "),
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
	args := make([]string, len(j.Args))
	for i, arg := range j.Args {
		args[i] = os.Expand(arg, secretMapper)
	}

	cmd := exec.Command(os.Expand(j.Command, secretMapper), args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.WithField("output", string(out)).WithError(err).Errorln("job completed with error")
	} else {
		log.WithField("output", string(out)).Infoln("job completed successfully")
	}

	j.cron.sync.Done()
}
