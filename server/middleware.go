package server

import (
	"net/http"
)

type Handle func(http.ResponseWriter, *http.Request)

type Middleware func(Handle) Handle

func WrapParseMultipartForm(handle Handle) Handle {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(4096)
		if err != nil {
			http.Error(w, `{"error":"need multipart form"}`, http.StatusBadRequest)
			return
		}

		handle(w, req)
	}
}

// WrapMiddleware 将给定的中间件包裹在 handle 上。Slice 中最后一个中间件
// 最先处理 HTTP 请求
func WrapMiddleware(middleware []Middleware, handle Handle) Handle {
	for _, m := range middleware {
		handle = m(handle)
	}
	return handle
}
