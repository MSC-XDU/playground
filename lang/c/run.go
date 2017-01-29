package c

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/MSC-XDU/playground/dockerserver"
	"github.com/MSC-XDU/playground/lang/internal"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	sourcePath = "/source/source.c"
)

// C 代码运行容器 Creator
type RunCreator struct {
	tempfile *os.File
}

// 创建 C 代码运行容器
func (config RunCreator) Create(ctx context.Context, client *client.Client) (string, error) {
	limits := lang.DefaultResourceLimits()
	hostConfig := container.HostConfig{
		AutoRemove: true,
		Binds:      []string{fmt.Sprintf("%s:%s:rw", config.tempfile.Name(), sourcePath)},
		Resources:  limits,
	}
	containerConfig := container.Config{
		Image:           imageRef,
		AttachStderr:    true,
		AttachStdout:    true,
		NetworkDisabled: true,
		Tty:             true,
	}

	result, err := client.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, "")
	if err != nil {
		return "", err
	}
	log.Printf("创建 C 容器, ID: %s, 警告信息: %v", result.ID, result.Warnings)
	return result.ID, nil
}

func NewRunCreator(f *os.File) dockerserver.ContainerCreator {
	return RunCreator{
		tempfile: f,
	}
}
