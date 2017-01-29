package main

import (
	"github.com/MSC-XDU/playground/dockerserver"
	"github.com/MSC-XDU/playground/server"
	"github.com/docker/docker/client"
)

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	dockerserver.Pull(cli)
	server.Start(cli)
}
