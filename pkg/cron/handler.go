package cron

import (
	context "context"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
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

	f := filters.NewArgs()
	f.Add("label", "mobycron.schedule")
	return h.addContainers(f)
}

// Listen docker message for containers with cron schedule
func (h *Handler) Listen() error {
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "mobycron.schedule")

	// // Adds the cron job
	// filterArgs.Add("event", "start")
	filterArgs.Add("event", "create")

	// filterArgs.Add("event", "stop")
	// filterArgs.Add("event", "die")

	// // removes from the cron queue
	filterArgs.Add("event", "destroy")

	eventOptions := types.EventsOptions{Filters: filterArgs}

	listen := func() {
	listenLoop:
		for {
			ctx, cancelFunc := context.WithCancel(context.Background())
			eventChan, errChan := h.cli.Events(ctx, eventOptions)
			for {
				select {
				case event := <-eventChan:
					log := log.WithFields(log.Fields{
						"func":     "Handler.Listen",
						"Status":   event.Status,
						"ID":       event.ID,
						"From":     event.From,
						"Type":     event.Type,
						"action":   event.Action,
						"actor.ID": event.Actor.ID,
						"Scope":    event.Scope,
					})
					log.Infoln("event message from server")

					f := filters.NewArgs()
					f.Add("id", event.Actor.ID)

					switch event.Action {
					case "create":
						if err := h.addContainers(f); err != nil {
							log.Errorln(err)
						}
					case "destroy":
						h.cron.RemoveContainerJob(event.Actor.ID)
					}

				case err := <-errChan:
					log.WithFields(log.Fields{
						"func": "Handler.Listen",
					}).WithError(err).Errorln("error from server")
					cancelFunc()
					continue listenLoop
				}
			}
		}
	}
	go listen()

	return nil
}

func (h *Handler) addContainers(filters filters.Args) error {
	log := log.WithFields(log.Fields{
		"func": "Handler.addContainers",
	})
	log.Infoln("add containers from filters")
	// TODO: this log is useless

	defer h.cli.Close()

	containers, err := h.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: filters})
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
