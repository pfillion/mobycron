package cron

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Cron keeps track of any number of jobs, invoking the associated Job as
// specified by the schedule. It may be started and stopped.
type Cron struct {
	runner   Runner
	sync     JobSynchroniser
	fs       afero.Fs
	cEntries map[string]cron.EntryID
	sEntries map[string]cron.EntryID
}

// NewCron return a new Cron job runner.
func NewCron(parseSecond bool) *Cron {
	option := cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor
	if parseSecond {
		option = option | cron.Second
	}

	return &Cron{
		cron.New(cron.WithParser(cron.NewParser(option))),
		&sync.WaitGroup{},
		afero.NewOsFs(),
		make(map[string]cron.EntryID),
		make(map[string]cron.EntryID),
	}
}

// AddJob adds a Job to the Cron to be run on the given schedule.
func (c *Cron) AddJob(job Job) error {
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

	if _, err := c.runner.AddJob(job.Schedule, &job); err != nil {
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
		if err := c.AddJob(job); err != nil {
			return err
		}
	}
	return nil
}

// AddContainerJob add container job to the Cron to be run on the given schedule.
func (c *Cron) AddContainerJob(job ContainerJob) error {
	log := log.WithFields(log.Fields{
		"func":            "Cron.AddContainerJob",
		"schedule":        job.Schedule,
		"action":          job.Action,
		"timeout":         job.Timeout,
		"command":         job.Command,
		"container.ID":    job.Container.ID,
		"container.Names": job.Container.Names,
	})

	if job.Schedule == "" {
		return errors.New("schedule is required")
	}

	if job.Timeout != "" {
		if _, err := strconv.ParseInt(job.Timeout, 10, 0); err != nil {
			return errors.New("invalid container timeout, only integer are permitted")
		}
	}

	switch job.Action {
	case "start", "restart", "stop":
		if job.Command != "" {
			return errors.New("a command can be specified only with 'exec' action")
		}
	case "exec":
		if job.Command == "" {
			return errors.New("command is required")
		}
	default:
		return errors.New("invalid container action, only 'start', 'restart', 'stop' and 'exec' are permitted")
	}

	log.Infoln("add container job to cron")

	job.cron = c
	ID, err := c.runner.AddJob(job.Schedule, &job)
	if err != nil {
		return errors.Wrap(err, "failed to add container job in cron")
	}

	c.cEntries[job.Container.ID] = ID

	return nil
}

// AddServiceJob add service job to the Cron to be run on the given schedule.
func (c *Cron) AddServiceJob(job ServiceJob) error {
	log := log.WithFields(log.Fields{
		"func":              "Cron.AddServiceJob",
		"schedule":          job.Schedule,
		"action":            job.Action,
		"timeout":           job.Timeout,
		"command":           job.Command,
		"service.ID":        job.ServiceID,
		"service.Name":      job.ServiceName,
		"service.Version":   job.ServiceVersion,
		"service.CreatedAt": job.ServiceCreatedAt,
	})

	if job.Schedule == "" {
		return errors.New("schedule is required")
	}

	if job.Timeout != "" {
		if _, err := strconv.ParseInt(job.Timeout, 10, 0); err != nil {
			return errors.New("invalid container timeout, only integer are permitted")
		}
	}

	switch job.Action {
	case "update":
		if job.Command != "" {
			return errors.New("a command can be specified only with 'exec' action")
		}
	case "exec":
		if job.Command == "" {
			return errors.New("command is required")
		}
	default:
		return errors.New("invalid service action, only 'update' and 'exec' are permitted")
	}

	log.Infoln("add service job to cron")

	job.cron = c
	ID, err := c.runner.AddJob(job.Schedule, &job)
	if err != nil {
		return errors.Wrap(err, "failed to add service job in cron")
	}

	c.sEntries[job.ServiceID] = ID

	return nil
}

// RemoveContainerJob remove container job from Cron.
func (c *Cron) RemoveContainerJob(ID string) {
	if entry, ok := c.cEntries[ID]; ok {
		delete(c.cEntries, ID)
		c.runner.Remove(entry)

		log := log.WithFields(log.Fields{
			"func":         "Cron.RemoveContainerJob",
			"container.ID": ID,
		})
		log.Infoln("remove container job from cron")
	}
}

// RemoveServiceJob remove service job from Cron.
func (c *Cron) RemoveServiceJob(ID string) {
	if entry, ok := c.sEntries[ID]; ok {
		delete(c.sEntries, ID)
		c.runner.Remove(entry)

		log := log.WithFields(log.Fields{
			"func":       "Cron.RemoveServiceJob",
			"service.ID": ID,
		})
		log.Infoln("remove service job from cron")
	}
}

// LoadConfig read Job from file in JSON format and add them to Cron.
func (c *Cron) LoadConfig(filename string) error {
	log := log.WithFields(log.Fields{
		"func":     "Cron.LoadConfig",
		"filename": filename,
	})
	log.Infoln("load config file")

	config, err := afero.ReadFile(c.fs, filename)
	if err != nil {
		return errors.Wrap(err, "failed to read config file")
	}

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
func (c *Cron) Stop() context.Context {
	log := log.WithFields(log.Fields{"func": "Cron.Stop"})

	log.Infoln("stopping cron, wait for running jobs")
	ctx := c.runner.Stop()

	c.sync.Wait()
	log.Infoln("cron is stopped, all jobs are completed")

	// TODO: See if the new stop context can be handy
	return ctx
}
