package python2

import "github.com/MSC-XDU/playground/dockerserver"

const imageRef = "python:2.7-alpine"

func init() {
	dockerserver.AddImage(imageRef)
}
