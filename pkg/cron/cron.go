package cron

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	cron "gopkg.in/robfig/cron.v3"
)

// Runner is an interface for testing robfig/cron
type Runner interface {
	AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
	Start()
	Stop()
}

// JobSynchroniser is an interface for testing sync.WaitGroup
type JobSynchroniser interface {
	Add(delta int)
	Done()
	Wait()
}

// Cron keeps track of any number of jobs, invoking the associated Job as
// specified by the schedule. It may be started and stopped.
type Cron struct {
	runner Runner
	sync   JobSynchroniser
	fs     afero.Fs
}

// NewCron return a new Cron job runner.
func NewCron() *Cron {
	return &Cron{cron.New(), &sync.WaitGroup{}, afero.NewOsFs()}
}

// AddJob adds a Job to the Cron to be run on the given schedule.
func (c *Cron) AddJob(job *Job) error {
	if job == nil {
		return errors.New("job is required")
	}

	log.WithFields(log.Fields{
		"func":     "Cron.AddJob",
		"schedule": job.Schedule,
		"command":  job.Command,
		"args":     strings.Join(job.Args, " "),
	}).Infoln("add job to cron")

	if job.Schedule == "" {
		return errors.New("schedule is required")
	}

	if job.Command == "" {
		return errors.New("command is required")
	}

	job.cron = c

	if _, err := c.runner.AddJob(job.Schedule, job); err != nil {
		return errors.Wrap(err, "failed to add job in cron")
	}

	return nil
}

// AddJobs adds jobs to the Cron.
func (c *Cron) AddJobs(jobs []Job) error {
	if jobs == nil {
		return errors.New("jobs is required")
	}
	for _, job := range jobs {
		if err := c.AddJob(&job); err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig read Job from file in JSON format and add them to Cron.
func (c *Cron) LoadConfig(filename string) error {
	log := log.WithFields(log.Fields{
		"func":     "Cron.LoadConfig",
		"filename": filename,
	})
	log.Infoln("load config file")

	if ok, _ := afero.Exists(c.fs, "/configs/config.json"); !ok {
		log.Warningln("no config was loaded, file not exist")
		return nil
	}

	config, _ := afero.ReadFile(c.fs, filename)
	j := []Job{}
	if err := json.Unmarshal([]byte(config), &j); err != nil {
		return errors.Wrap(err, "failed to parse JSON data from config file")
	}

	if err := c.AddJobs(j); err != nil {
		return errors.Wrap(err, "failed to add jobs fron config file")
	}
	return nil
}

// Start the Cron scheduler.
func (c *Cron) Start() {
	log.WithFields(log.Fields{"func": "Cron.Start"}).Infoln("start cron")
	c.runner.Start()
}

// Stop the Cron scheduler.
func (c *Cron) Stop() {
	log := log.WithFields(log.Fields{"func": "Cron.Stop"})

	log.Infoln("stopping cron, wait for running jobs")
	c.runner.Stop()

	c.sync.Wait()
	log.Infoln("cron is stopped, all jobs are completed")
}
