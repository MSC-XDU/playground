package dockerserver

import (
	"context"
	"io"

	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/versions"
)

// ContainerCreator 创建 container
type ContainerCreator interface {
	// 创建 container 并返回 ID
	Create(context.Context, *client.Client) (string, error)
}

type Container struct {
	*client.Client
	Id string
}

func CreateContainer(ctx context.Context, client *client.Client, creator ContainerCreator) (Container, error) {
	id, err := creator.Create(ctx, client)
	if err != nil {
		return Container{}, err
	}
	return Container{
		Client: client,
		Id:     id,
	}, nil
}

// 启动并 attach 到 creator 创建的 container
func (c Container) Attach(ctx context.Context) (types.HijackedResponse, error) {
	err := c.ContainerStart(ctx, c.Id, types.ContainerStartOptions{})
	if err != nil {
		return types.HijackedResponse{}, err
	}

	return c.ContainerAttach(ctx, c.Id, types.ContainerAttachOptions{
		Logs:   true,
		Stream: true,
		Stderr: true,
		Stdout: true,
	})
}

// 启动并 等待 creator 创建的 container 退出
func (c Container) Wait(ctx context.Context) (io.ReadCloser, error) {
	err := c.ContainerStart(ctx, c.Id, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	_, err = c.ContainerWait(ctx, c.Id)
	if err != nil {
		return nil, err
	}

	return c.ContainerLogs(ctx, c.Id, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
	})
}

// ResultWriter 将 container 的运行结果写入 writer
type ResultWriter func(Container, context.Context, io.Writer) error

// WriteResultStream 将结果即时写入 writer
func (c Container) WriteResultStream(ctx context.Context, w io.Writer) error {
	hijacked, err := c.Attach(ctx)
	if err != nil {
		return err
	}

	hijacked.Reader.WriteTo(w)
	hijacked.Close()
	go c.remove()
	return nil
}

// WriteResultBlock 将等待容器退出，一次将所有结果写入 writer
func (c Container) WriteResultBlock(ctx context.Context, w io.Writer) error {
	result, err := c.Wait(ctx)
	if err != nil {
		return err
	}

	io.Copy(w, result)
	result.Close()
	go c.remove()
	return nil
}

func (c Container) remove() {
	if versions.LessThan(c.ClientVersion(), "1.25") {
		err := c.ContainerRemove(context.Background(), c.Id, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("contaier %s 移除\n", c.Id)
		}
	}
}
