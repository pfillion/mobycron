package main

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// Job todo
type Job struct {
	command string
	args    []string
	cron    *Cron
}

// Run todo
func (j *Job) Run() {
	log := log.WithFields(log.Fields{
		"func":    "Run",
		"command": j.command,
		"args":    j.args,
	})

	j.cron.sync.Add(1)

	// Expand env in all args
	args := make([]string, len(j.args))
	for i, arg := range j.args {
		args[i] = os.ExpandEnv(arg)
	}

	cmd := exec.Command(os.ExpandEnv(j.command), args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.WithField("output", string(out)).WithError(err).Errorln("job completed with error")
	} else {
		log.WithField("output", string(out)).Infoln("job completed sucessfully")
	}

	j.cron.sync.Done()
}
