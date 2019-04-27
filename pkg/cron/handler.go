package cron

import (
	context "context"

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
	// TODO: Refactor log desing and unittesting
	// log := log.WithFields(log.Fields{
	// 	"func": "Handler.Scan",
	// })
	// log.Infoln("scan containers for cron schedule")

	f := filters.NewArgs()
	f.Add("label", "mobycron.schedule")
	containers, err := h.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		return err
	}

	for _, container := range containers {
		j := ContainerJob{
			Schedule:  container.Labels["mobycron.schedule"],
			Action:    container.Labels["mobycron.action"],
			Timeout:   container.Labels["mobycron.timeout"],
			Command:   container.Labels["mobycron.command"],
			Container: container,
			cli:       h.cli,
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
}
