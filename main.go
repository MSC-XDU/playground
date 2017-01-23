package main

import (
	"context"
	"log"
	"sync"

	"bufio"
	"encoding/json"

	"github.com/MSC-XDU/playground/server"
	"github.com/MSC-XDU/playground/shared"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var cli *client.Client

func pullImage(ctx context.Context, ref string, wg *sync.WaitGroup) {
	var err error
	out, err := cli.ImagePull(ctx, ref, types.ImagePullOptions{})
	if err != nil {
		log.Panic(err)
	}
	defer out.Close()
	sc := bufio.NewScanner(out)
	for sc.Scan() {
		var prog interface{}
		json.Unmarshal(sc.Bytes(), &prog)
		if prog.(map[string]interface{})["status"].(string) != "Downloading" {
			log.Printf("%v\n", prog)
		}
	}
	wg.Done()
}

func initImage() {
	var err error
	cli, err = client.NewEnvClient()
	if err != nil {
		log.Fatal("docker client 初始化失败。错误信息：", err)
	}

	var wg sync.WaitGroup
	ctx := context.Background()

	total := len(shared.InitImages)
	wg.Add(total)
	for i, ref := range shared.InitImages {
		log.Printf("pulling image %d of %d %s", i+1, total, ref)
		go pullImage(ctx, ref, &wg)
	}

	wg.Wait()
	log.Println("pull 完成")
	log.Println("准备完成，开始启动服务")
}

func main() {
	initImage()
	log.Println("启动 HTTP 服务")
	server.Start()
}
