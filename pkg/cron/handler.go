package cron

import (
	context "context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
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

	defer h.cli.Close()

	err := h.addContainers(f)
	if err != nil {
		return err
	}
	h.pruneContainersFromAllServices()
	return nil
}

// ListenContainer listen docker message for containers with cron schedule
func (h *Handler) ListenContainer() {
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "mobycron.schedule")
	filterArgs.Add("type", "container")
	filterArgs.Add("event", "create")
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

					if event.Action == "create" {
						f := filters.NewArgs()
						f.Add("id", event.Actor.ID)

						if err := h.addContainers(f); err != nil {
							log.Errorln(err)
						}
						h.cli.Close()
					}
					if event.Action == "destroy" {
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
}

// ListenService listen docker message for services with cron schedule
func (h *Handler) ListenService() {
	filterArgs := filters.NewArgs()
	filterArgs.Add("type", "service")
	filterArgs.Add("event", "update")

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
						"func":     "Handler.ListenService",
						"Status":   event.Status,
						"ID":       event.ID,
						"From":     event.From,
						"Type":     event.Type,
						"action":   event.Action,
						"actor.ID": event.Actor.ID,
						"Scope":    event.Scope,
					})
					log.Infoln("event message from server")

					if err := h.pruneContainersFromService(event.Actor.ID); err != nil {
						log.Errorln(err)
					}
					h.cli.Close()

				case err := <-errChan:
					log.WithFields(log.Fields{
						"func": "Handler.ListenService",
					}).WithError(err).Errorln("error from server")
					cancelFunc()
					continue listenLoop
				}
			}
		}
	}
	go listen()
}

func (h *Handler) addContainers(filters filters.Args) error {
	log := log.WithFields(log.Fields{
		"func": "Handler.addContainers",
	})
	log.Infoln("add containers from filters")

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

func (h *Handler) pruneContainersFromService(serviceID string) error {
	service, _, err := h.cli.ServiceInspectWithRaw(context.Background(), serviceID, types.ServiceInspectOptions{})
	if err != nil {
		return err
	}

	f := filters.NewArgs()
	f.Add("service", service.Spec.Name)
	opt := types.TaskListOptions{Filters: f}

	tasks, err := h.cli.TaskList(context.Background(), opt)
	if err != nil {
		return err
	}

	tasksBySlot := make(map[string]swarm.Task)

	for _, task := range tasks {
		key := fmt.Sprintf("%s.%d", task.ServiceID, task.Slot)
		latestTask, ok := tasksBySlot[key]

		if !ok {
			tasksBySlot[key] = task
		} else if latestTask.CreatedAt.Before(task.CreatedAt) {
			tasksBySlot[key] = task
			h.cron.RemoveContainerJob(latestTask.Status.ContainerStatus.ContainerID)
		} else {
			h.cron.RemoveContainerJob(task.Status.ContainerStatus.ContainerID)
		}
	}

	return nil
}

func (h *Handler) pruneContainersFromAllServices() {
	log := log.WithFields(log.Fields{
		"func": "Handler.pruneContainersFromAllServices",
	})

	services, err := h.cli.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		log.Errorln(err)
		return
	}

	for _, service := range services {
		err = h.pruneContainersFromService(service.ID)
		if err != nil {
			log.WithField("serviceID", service.ID).Errorln(err)
		}
	}
}
