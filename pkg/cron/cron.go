package cron

import (
	"encoding/json"
	"os"
	"os/signal"
	"sync"

	"github.com/pkg/errors"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Runner is an interface for testing robfig/cron
type Runner interface {
	AddJob(spec string, cmd cron.Job) error
	Start()
	Stop()
}

// JobSynchroniser is an interface for testing sync.WaitGroup
type JobSynchroniser interface {
	Add(delta int)
	Done()
	Wait()
}

// Cron keeps track of any number of entries, invoking the associated Job as
// specified by the schedule. It may be started and stopped.
type Cron struct {
	runner Runner
	sync   JobSynchroniser
	fs     afero.Fs
}

// Entry consists of a schedule and the command to execute on that schedule.
type Entry struct {
	Schedule string   `json:"schedule"`
	Command  string   `json:"command"`
	Args     []string `json:"args"`
}

// NewCron return a new Cron job runner.
func NewCron(fs afero.Fs) *Cron {
	return &Cron{cron.New(), &sync.WaitGroup{}, fs}
}

// AddJob adds a Entry to the Cron to be run on the given schedule.
func (c *Cron) AddJob(entry *Entry) error {
	if entry == nil {
		return errors.New("entry is required")
	}

	log.WithFields(log.Fields{
		"func":     "AddJob",
		"schedule": entry.Schedule,
		"command":  entry.Command,
		"args":     entry.Args,
	}).Infoln("add job to cron")

	if entry.Schedule == "" {
		return errors.New("schedule is required")
	}

	if entry.Command == "" {
		return errors.New("command is required")
	}

	if err := c.runner.AddJob(entry.Schedule, &Job{entry.Command, entry.Args, c}); err != nil {
		return errors.Wrap(err, "failed to add job in cron")
	}

	return nil
}

// AddJobs adds entries to the Cron.
func (c *Cron) AddJobs(entries []Entry) error {
	if entries == nil {
		return errors.New("entries is required")
	}
	for _, entry := range entries {
		if err := c.AddJob(&entry); err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig read Entry from file in JSON format and add them to Cron.
func (c *Cron) LoadConfig(filename string) error {
	log.WithFields(log.Fields{
		"func":     "LoadConfig",
		"filename": filename,
	}).Infoln("load config file")

	config, err := afero.ReadFile(c.fs, filename)
	if err != nil {
		return errors.Wrap(err, "failed to read config file")
	}

	e := []Entry{}
	if err = json.Unmarshal([]byte(config), &e); err != nil {
		return errors.Wrap(err, "failed to parse JSON data from config file")
	}

	if err = c.AddJobs(e); err != nil {
		return errors.Wrap(err, "failed to add jobs entries fron config file")
	}
	return nil
}

// Run the Cron scheduler and wait for specific os.Signal to stop.
func (c *Cron) Run(cs chan os.Signal, sigs ...os.Signal) error {
	if cs == nil {
		return errors.New("channel is required")
	}

	c.Start()

	log.WithFields(log.Fields{
		"func":   "Run",
		"signal": sigs,
	}).Infoln("cron is running and waiting signal for stop")

	signal.Notify(cs, sigs...)
	<-cs

	c.Stop()

	return nil
}

// Start the Cron scheduler.
func (c *Cron) Start() {
	log.WithFields(log.Fields{"func": "Start"}).Infoln("start cron")
	c.runner.Start()
}

// Stop the Cron scheduler.
func (c *Cron) Stop() {
	log := log.WithFields(log.Fields{"func": "Stop"})

	log.Infoln("stopping cron, wait for running jobs")
	c.runner.Stop()

	c.sync.Wait()
	log.Infoln("cron is stopped, all jobs are completed")
}
