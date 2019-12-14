package cron

import (
	context "context"
	"strconv"
	"strings"

	types "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

// Handler handle docker messages
type Handler struct {
	cron Cronner
	cli  DockerClient
}

// NewHandler returns a docker handler
func NewHandler(cron Cronner) (*Handler, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	cli.NegotiateAPIVersion(context.Background())

	return &Handler{cron, cli}, nil
}

// Scan current containers for cron schedule
func (h *Handler) Scan() error {
	log := log.WithFields(log.Fields{
		"func": "Handler.Scan",
	})
	log.Infoln("scan containers for cron schedule")

	defer h.cli.Close()

	f := filters.NewArgs()
	f.Add("label", "mobycron.schedule")
	containers, err := h.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		return err
	}

	for _, container := range containers {
		var slot int
		if tn, ok := container.Labels["com.docker.swarm.task.name"]; ok {
			if slot, err = strconv.Atoi(strings.Split(tn, ".")[1]); err != nil {
				log.WithError(err).Errorln("failed to convert slot of label 'com.docker.swarm.task.name'")
				continue
			}
		}
		j := ContainerJob{
			Schedule:    container.Labels["mobycron.schedule"],
			Action:      container.Labels["mobycron.action"],
			Timeout:     container.Labels["mobycron.timeout"],
			Command:     container.Labels["mobycron.command"],
			ServiceID:   container.Labels["com.docker.swarm.service.id"],
			ServiceName: container.Labels["com.docker.swarm.service.name"],
			TaskID:      container.Labels["com.docker.swarm.task.id"],
			TaskName:    container.Labels["com.docker.swarm.task.name"],
			Slot:        slot,
			Created:     container.Created,
			Container:   container,
			cli:         h.cli,
		}
		if err := h.cron.AddContainerJob(j); err != nil {
			log.WithError(err).Errorln("add container job to cron is in error")
		}
	}
	return nil
}

// Listen docker message for containers with cron schedule
func (h *Handler) Listen() error {
	// TODO: implement
	return nil

	// filterArgs := filters.NewArgs()
	// // Adds the cron job
	// filterArgs.Add("event", "start")
	// filterArgs.Add("event", "create")

	// filterArgs.Add("event", "stop")
	// filterArgs.Add("event", "die")

	// // removes from the cron queue
	// filterArgs.Add("event", "destroy")

	// eventOptions := types.EventsOptions{
	// 	Filters: filterArgs,
	// }
	// ctx, cancelFunc := context.WithCancel(context.Background())
	// eventChan, errChan := h.cli.Events(ctx, eventOptions)

	// listen := func() {
	// 	// loop:
	// 	for {
	// 		// eventStream, errChan := router.Listen(ctx)
	// 		for {
	// 			select {
	// 			case event := <-eventChan:
	// 				log.Infoln(event)
	// 				// handler.Handle(&event)
	// 			case err := <-errChan:
	// 				log.Errorln(err)
	// 				cancelFunc()
	// 				break
	// 				// continue loop
	// 				// TODO: replace loop by break ???
	// 			}
	// 		}
	// 	}
	// }
	// go listen()

	// return nil

	// Adding a cron.schedule label flags the container for deeper inspection
	// With this service
	// if _, ok := msg.Actor.Attributes["cron.schedule"]; ok {
	// 	if msg.Action == "start" || msg.Action == "create" {
	// 		logrus.Debugf("Processing %s event for container: %s", msg.Action, msg.ID)
	// 		dh.Crontab.AddJob(msg.ID, msg.Actor.Attributes, "docker")
	// 	}

	// 	if msg.Action == "stop" || msg.Action == "die" {
	// 		logrus.Debugf("Proccessing %s event for container: %s", msg.Action, msg.ID)
	// 		dh.Crontab.DeactivateJob(msg.ID, msg.Actor.Attributes)
	// 	}

	// 	if msg.Action == "destroy" {
	// 		logrus.Debugf("Processing destroy event for container: %s", msg.ID)
	// 		dh.Crontab.RemoveJob(msg.ID)
	// 	}
	// }
}
