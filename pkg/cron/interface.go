package cron

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
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
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
	ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStart(ctx context.Context, container string, options types.ContainerStartOptions) error
	ContainerStop(ctx context.Context, container string, timeout *time.Duration) error
	ContainerRestart(ctx context.Context, container string, timeout *time.Duration) error
	Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error)
	ServiceInspectWithRaw(ctx context.Context, serviceID string, options types.ServiceInspectOptions) (swarm.Service, []byte, error)
	ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
	ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error)
	TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error)
}
