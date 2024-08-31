package cron

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/swarm"
	cron "github.com/robfig/cron/v3"
)

// Runner is an interface for testing robfig/cron
type Runner interface {
	AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
	Remove(id cron.EntryID)
	Start()
	Stop() context.Context
}

// JobSynchroniser is an interface for testing sync.WaitGroup
type JobSynchroniser interface {
	Add(delta int)
	Done()
	Wait()
}

// Cronner is an interface for adding job to cron
type Cronner interface {
	AddContainerJob(job ContainerJob) error
	AddServiceJob(job ServiceJob) error
	RemoveContainerJob(ID string)
	RemoveServiceJob(ID string)
}

// DockerClient is the client for docker
type DockerClient interface {
	Close() error
	ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error)
	ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (types.IDResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)
	ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerStart(ctx context.Context, container string, options container.StartOptions) error
	ContainerStop(ctx context.Context, container string, timeout container.StopOptions) error
	ContainerRestart(ctx context.Context, container string, options container.StopOptions) error
	Events(ctx context.Context, options events.ListOptions) (<-chan events.Message, <-chan error)
	ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
	ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (swarm.ServiceUpdateResponse, error)
}
