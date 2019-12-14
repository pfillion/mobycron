package cron

import (
	context "context"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ContainerJob run a docker container on a schedule.
type ContainerJob struct {
	Schedule    string
	Action      string
	Timeout     string
	Command     string
	ServiceID   string
	ServiceName string
	TaskID      string
	TaskName    string
	Slot        int
	Created     int64
	Container   types.Container
	cron        *Cron
	cli         DockerClient
}

// Run a docker container and log the output.
func (j *ContainerJob) Run() {
	log := log.WithFields(log.Fields{
		"func":            "ContainerJob.Run",
		"schedule":        j.Schedule,
		"action":          j.Action,
		"timeout":         j.Timeout,
		"command":         j.Command,
		"container.ID":    j.Container.ID,
		"container.Names": strings.Join(j.Container.Names, ","),
	})

	j.cron.sync.Add(1)
	defer j.cron.sync.Done()
	defer j.cli.Close()
	var err error

	switch j.Action {
	case "start":
		err = j.start()
	case "restart":
		err = j.restart()
	case "stop":
		err = j.stop()
	case "exec":
		var out string
		if out, err = j.exec(); out != "" {
			log = log.WithField("output", out)
		}
	}

	if err != nil {
		log.WithError(err).Errorln("container job completed with error")
	} else {
		log.Infoln("container action completed successfully")
	}
}

func (j *ContainerJob) start() error {
	return j.cli.ContainerStart(context.Background(), j.Container.ID, types.ContainerStartOptions{})
}

func (j *ContainerJob) restart() error {
	return j.cli.ContainerRestart(context.Background(), j.Container.ID, j.getTimeout())
}

func (j *ContainerJob) stop() error {
	return j.cli.ContainerStop(context.Background(), j.Container.ID, j.getTimeout())
}

func (j *ContainerJob) exec() (string, error) {
	ctx := context.Background()
	cmd := strings.Fields(j.Command)

	// We need to inspect before we do the ContainerExecCreate, because
	// otherwise if we error out we will leak execIDs on the server (and
	// there's no easy way to clean those up). But also in order to make "not
	// exist" errors take precedence we do a dummy inspect first.
	if _, err := j.cli.ContainerInspect(ctx, j.Container.ID); err != nil {
		return "", err
	}

	createResp, err := j.cli.ContainerExecCreate(ctx, j.Container.ID, types.ExecConfig{AttachStdout: true, AttachStderr: true, Cmd: cmd})
	if err != nil {
		return "", err
	}

	if createResp.ID == "" {
		return "", errors.New("exec ID empty")
	}

	attachResp, err := j.cli.ContainerExecAttach(ctx, createResp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}

	defer attachResp.CloseWrite()
	defer attachResp.Close()

	var out strings.Builder
	if _, err = stdcopy.StdCopy(&out, &out, attachResp.Reader); err != nil {
		return "", err
	}

	inspectResp, err := j.cli.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return "", err
	}

	if inspectResp.ExitCode != 0 {
		return out.String(), errors.Errorf("exit status %d", inspectResp.ExitCode)
	}

	return out.String(), nil
}

func (j *ContainerJob) getTimeout() *time.Duration {
	var value = 10
	if j.Timeout != "" {
		value, _ = strconv.Atoi(j.Timeout)
	}
	timeout := time.Duration(value) * time.Second
	return &timeout
}
