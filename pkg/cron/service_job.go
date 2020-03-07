package cron

import (
	context "context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	log "github.com/sirupsen/logrus"
)

// ServiceJob run a docker service task on a schedule.
type ServiceJob struct {
	Schedule         string
	Action           string
	Timeout          string
	Command          string
	ServiceID        string
	ServiceName      string
	ServiceVersion   swarm.Version
	ServiceCreatedAt time.Time
	Service          swarm.Service
	cron             *Cron
	cli              DockerClient
}

// Run a docker container and log the output.
func (j *ServiceJob) Run() {
	log := log.WithFields(log.Fields{
		"func":         "ServiceJob.Run",
		"schedule":     j.Schedule,
		"action":       j.Action,
		"timeout":      j.Timeout,
		"command":      j.Command,
		"service.ID":   j.ServiceID,
		"service.Name": j.ServiceName,
	})
	// TODO: add all property of Service in log Fields

	j.cron.sync.Add(1)
	defer j.cron.sync.Done()
	defer j.cli.Close()
	var err error

	switch j.Action {
	case "update":
		var r types.ServiceUpdateResponse
		j.Service.Spec.TaskTemplate.ForceUpdate = j.ServiceVersion.Index
		r, err = j.cli.ServiceUpdate(context.Background(), j.ServiceID, j.ServiceVersion, j.Service.Spec, types.ServiceUpdateOptions{})
		for _, w := range r.Warnings {
			log.Warning(w)
		}
	}

	if err != nil {
		log.WithError(err).Errorln("service job completed with error")
	} else {
		log.Infoln("service action completed successfully")
	}
}
