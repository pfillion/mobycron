package cron

import (
	context "context"
	"time"

	"github.com/docker/docker/api/types"
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

// Cronner is an interface for adding job to cron
type Cronner interface {
	AddContainerJob(job ContainerJob) error
}

// DockerClient is the client for docker
type DockerClient interface {
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
	ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStart(ctx context.Context, container string, options types.ContainerStartOptions) error
	ContainerStop(ctx context.Context, container string, timeout *time.Duration) error
	ContainerRestart(ctx context.Context, container string, timeout *time.Duration) error
}