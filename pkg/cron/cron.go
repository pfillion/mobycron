package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Cron keeps track of any number of jobs, invoking the associated Job as
// specified by the schedule. It may be started and stopped.
type Cron struct {
	runner     Runner
	sync       JobSynchroniser
	fs         afero.Fs
	jobEntries map[string]cronEntry
}

type cronEntry struct {
	id      cron.EntryID
	created int64
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
		make(map[string]cronEntry),
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
		"service.ID":      job.ServiceID,
		"service.Name":    job.ServiceName,
		"task.ID":         job.TaskID,
		"task.Name":       job.TaskName,
		"slot":            job.Slot,
		"created":         time.Unix(job.Created, 0).String(),
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

	// Check if replacement is needed
	key := fmt.Sprintf("%s.%d", job.ServiceID, job.Slot)
	if entry, ok := c.jobEntries[key]; ok {
		if entry.created > job.Created {
			log.Infoln("skip replacement, the container job is older")
			return nil
		}
		c.runner.Remove(entry.id)
		log.Infoln("replace container job in cron")
	} else {
		log.Infoln("add container job to cron")
	}

	job.cron = c
	id, err := c.runner.AddJob(job.Schedule, &job)
	if err != nil {
		return errors.Wrap(err, "failed to add container job in cron")
	}

	c.jobEntries[key] = cronEntry{id, job.Created}

	return nil
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
