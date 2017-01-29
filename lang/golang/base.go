package golang

import "github.com/MSC-XDU/playground/dockerserver"

const imageRef = "golang:alpine"

func init() {
	dockerserver.AddImage(imageRef)
}
