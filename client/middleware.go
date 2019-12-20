package main

import (
	"log"
	"net/http"
	"time"
)

type Middleware struct {
}

// 记录请求时间log
func (m *Middleware) LogHandler(next http.Handler) http.Handler {
	fn := func(writer http.ResponseWriter, request *http.Request) {
		st := time.Now()
		next.ServeHTTP(writer, request)
		et := time.Now()
		//记录请求耗时 方法 url
		log.Printf("[%s] %q  cost time %v\n", request.Method, request.URL.String(), et.Sub(st))
	}
	return http.HandlerFunc(fn)
}

// panic 错误处理
func (m Middleware) RecoverHandler(next http.Handler) http.Handler {
	fn := func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("recover panic:%v\n", err)
				//返回500
				http.Error(writer, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(writer, request)
	}
	return http.HandlerFunc(fn)
}
