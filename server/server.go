package server

import (
	"net/http"

	"log"

	"github.com/MSC-XDU/playground/c"
	"github.com/MSC-XDU/playground/golang"
	"github.com/MSC-XDU/playground/python"
	"github.com/gorilla/mux"
)

func Start() {
	router := mux.NewRouter()
	router.HandleFunc("/go", wrapMiddleware(golang.ServerHandle)).Methods(http.MethodPost)
	router.HandleFunc("/python", wrapMiddleware(python.ServerHandle)).Methods(http.MethodPost)
	router.HandleFunc("/c", wrapMiddleware(c.ServerHandle)).Methods(http.MethodPost)
	//router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	//	http.ServeFile(w, req, "test.html")
	//}).Methods(http.MethodGet)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":9090", router)
}

type Handle func(http.ResponseWriter, *http.Request)
type Middleware func(Handle) Handle

func ParseMultipartForm(handle Handle) Handle {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(4096)
		if err != nil {
			w.WriteHeader(400)
			log.Println(err.Error())
			w.Write([]byte("异常请求"))
			return
		}
		handle(w, req)
	}
}

var middleware = []Middleware{
	ParseMultipartForm,
}

func wrapMiddleware(handle Handle) Handle {
	for _, m := range middleware {
		handle = m(handle)
	}
	return handle
}
