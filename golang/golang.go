package golang

import (
	"context"
	"fmt"
	"log"
	"os"

	"net/http"

	"github.com/MSC-XDU/playground/shared"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	sourceFileName = "/source.go"
	imageName      = "golang:alpine"
)

var (
	cli                    *client.Client
	defaultContainerConfig = container.Config{
		Image:           imageName,
		Cmd:             []string{"go", "run", sourceFileName},
		NetworkDisabled: true,
	}
)

func init() {
	var err error
	cli, err = client.NewEnvClient()
	if err != nil {
		log.Panic(err)
	}
	shared.AddImages(imageName)
}

func createContainer(ctx context.Context, tmp *os.File) (string, error) {
	hostConfig := &container.HostConfig{
		Binds:     []string{fmt.Sprintf("%s:%s", tmp.Name(), sourceFileName)},
		Resources: shared.ResourcesConfig,
		//AutoRemove: true,
	}
	resp, err := cli.ContainerCreate(ctx, &defaultContainerConfig, hostConfig, nil, "")
	if err != nil {
		return "", err
	}
	log.Printf("创建 container for go, 警告信息：%v", resp.Warnings)
	return resp.ID, nil
}

//  Golang 的 playground HTTP HandleFunc
func ServerHandle(w http.ResponseWriter, req *http.Request) {
	code := req.MultipartForm.Value["code"][0]
	if code == "" {
		w.Write([]byte{})
		return
	}

	result, done, err := shared.PlayWithCode(req.Context(), code, shared.ContainerCreator{
		Client: cli,
		Create: createContainer,
	})
	if err != nil {
		w.WriteHeader(500)
		log.Println(err.Error())
		w.Write([]byte("服务器异常"))
	}

	result.WriteTo(w)
	// 新建 Goroutine 进行资源清理
	go done()
}
