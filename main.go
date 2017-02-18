package main

import (
	"log"

	"github.com/MSC-XDU/playground/dockerserver"
	"github.com/MSC-XDU/playground/server"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/client"
)

func main() {
	cli, err := client.NewEnvClient()
	if versions.GreaterThanOrEqualTo(cli.ClientVersion(), "1.25") {
		log.Println("容器移除将由 docker 自动完成")
	}

	if err != nil {
		panic(err)
	}
	dockerserver.Pull(cli)
	server.Start(cli)
}
