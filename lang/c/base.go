package c

import "github.com/MSC-XDU/playground/dockerserver"

const imageRef = "tuzili/gcc-run"

func init() {
	dockerserver.AddImage(imageRef)
}
