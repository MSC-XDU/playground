package python

import "github.com/MSC-XDU/playground/dockerserver"

const imageRef = "python:alpine"

func init() {
	dockerserver.AddImage(imageRef)
}
