package dockerserver

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
)

var initImages []string

// AddImage 添加一个 image 到初始化拉取列表
func AddImage(ref string) {
	initImages = append(initImages, ref)
}

func pull(ctx context.Context, client *client.Client, ref string) {
	prog, err := client.ImagePull(ctx, ref, types.ImagePullOptions{})
	if err != nil {
		log.Fatalf("拉取 %s 失败，错误信息: %s\n", ref, err.Error())
	}

	dec := json.NewDecoder(prog)
	for {
		var message jsonmessage.JSONMessage
		if err := dec.Decode(&message); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		if message.Error != nil {
			log.Fatal(message.Error)
		}

	}

	prog.Close()
}

// Pull 拉取所有的初始化 image
func Pull(client *client.Client) {
	total := len(initImages)
	var wg sync.WaitGroup
	wg.Add(total)
	for i, ref := range initImages {
		log.Printf("开始拉取 %s，%d of %d", ref, i+1, total)
		go func(ref string) {
			pull(context.Background(), client, ref)
			log.Printf("%s 拉取完成\n", ref)
			wg.Done()
		}(ref)
	}
	wg.Wait()
}
