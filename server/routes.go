package server

import (
	"log"
	"net/http"

	"html/template"

	"github.com/MSC-XDU/playground/dockerserver"
	"github.com/MSC-XDU/playground/lang/c"
	"github.com/MSC-XDU/playground/lang/golang"
	"github.com/MSC-XDU/playground/lang/python"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/MSC-XDU/playground/lang/python2"
)

var defaultMiddleware = []Middleware{
	WrapParseMultipartForm,
}

// Templates
var (
	editorTmpl = template.Must(template.New("editor.html").ParseFiles("assets/templates/editor.html"))
)

type editorData struct {
	Code       string
	ModeSelect bool
	Mode       string
}

func Start(client *client.Client) {
	r := mux.NewRouter()

	r.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte{})
	})

	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		editorTmpl.Execute(w, editorData{Code: "", Mode: "Go", ModeSelect: true})
	})

	r.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("assets/statics"))))

	// 代码运行路由配置
	factory := TempFileContainerFactory{
		Client: client,
		Getter: multipartFormGetter,
		Writer: dockerserver.Container.WriteResultStream,
	}
	run := r.PathPrefix("/run").Subrouter()
	run.HandleFunc("/go", WrapMiddleware(defaultMiddleware, factory.NewHandleFunc(golang.NewRunCreator)))
	run.HandleFunc("/py", WrapMiddleware(defaultMiddleware, factory.NewHandleFunc(python.NewRunCreator)))
	run.HandleFunc("/c", WrapMiddleware(defaultMiddleware, factory.NewHandleFunc(c.NewRunCreator)))
	run.HandleFunc("/py2", WrapMiddleware(defaultMiddleware, factory.NewHandleFunc(python2.NewRunCreator)))

	// 分享路由配置
	r.HandleFunc("/share/{id:[0-9a-zA-Z]+}", GetCodeHandle).Methods("GET")
	r.HandleFunc("/share", WrapMiddleware(defaultMiddleware, SaveCodeHandle)).Methods("POST")

	log.Println("启动 HTTP 服务")
	http.ListenAndServe(":80", r)
}
