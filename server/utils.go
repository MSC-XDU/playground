package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/MSC-XDU/playground/dockerserver"
	"github.com/docker/docker/client"
)

var tempDir string

func init() {
	tempDir = os.Getenv("PLAYGROUND_TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp"
	}
}

type TempFileContainerFactory struct {
	Client *client.Client
	Getter CodeGetter
	Writer dockerserver.ResultWriter
}

func (factory TempFileContainerFactory) NewHandleFunc(creator func(*os.File) dockerserver.ContainerCreator) Handle {
	return func(w http.ResponseWriter, req *http.Request) {
		code, err := factory.Getter(req)
		if err != nil {
			log.Println(err)
			http.Error(w, `{"error": "can't find code"}`, http.StatusBadRequest)
		}

		tmp, err := createTempFile(code)
		if err != nil {
			log.Println(err.Error())
			internalError(w)
			return
		}
		defer tmp.Close()

		c, err := dockerserver.CreateContainer(req.Context(), factory.Client, creator(tmp))
		if err != nil {
			log.Println(err.Error())
			internalError(w)
			return
		}

		err = factory.Writer(c, req.Context(), w)
		if err != nil {
			log.Println(err.Error())
			internalError(w)
			return
		}
	}
}

func createTempFile(content string) (*os.File, error) {
	f, err := ioutil.TempFile(tempDir, "playground_")
	if err != nil {
		return nil, err
	}
	_, err = f.WriteString(content)
	return f, err
}

func internalError(w http.ResponseWriter) {
	http.Error(w, `{"error", "server error"}`, http.StatusInternalServerError)
}

// CodeGetter 从 request 中获取用户上传的代码
type CodeGetter func(*http.Request) (string, error)

func multipartFormGetter(req *http.Request) (string, error) {
	return req.MultipartForm.Value["code"][0], nil
}
