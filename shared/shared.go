package shared

import (
	"bufio"
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
)

var InitImages []string

func init() {
	InitImages = make([]string, 0, 2)
	cpuTimeLimit, _ := units.ParseUlimit("cpu=5:5")
	ResourcesConfig.Ulimits = append(ResourcesConfig.Ulimits, cpuTimeLimit)
}

func AddImages(ref string) {
	InitImages = append(InitImages, ref)
}

var ResourcesConfig = container.Resources{
	Memory:    30 * 1024 * 1024,
	CPUPeriod: 1000000,
	CPUQuota:  200000,
}

type ContainerCreator struct {
	*client.Client
	Create func(context.Context, *os.File) (string, error)
}

//  PlayWithCode 将创建容器运行输入的代码字符串。返回值中的 Reader 用于获取容器的 Stderr 和 Stdout 输出。
//  返回的资源清理函数，需要由调用者在适当的时候调用，清理掉所有临时资源。
func PlayWithCode(ctx context.Context, source string, c ContainerCreator) (*bufio.Reader, func(), error) {
	tmp, err := ioutil.TempFile("/tmp", "playground_go_")
	if err != nil {
		return nil, nil, err
	}
	_, err = tmp.WriteString(source)
	if err != nil {
		return nil, nil, err
	}

	id, err := c.Create(ctx, tmp)
	if err != nil {
		return nil, nil, err
	}

	err = c.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return nil, nil, err
	}

	hijackResp, err := c.ContainerAttach(ctx, id, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return nil, nil, err
	}

	done := func() {
		hijackResp.Close()
		err := tmp.Close()
		if err != nil {
			log.Printf("临时文件 %s 关闭失败。错误信息: %s", tmp.Name(), err.Error())
		}
	}

	return hijackResp.Reader, done, nil
}
