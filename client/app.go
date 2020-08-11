package main

import (
	"fmt"
	"github.com/baagee/dmq/client/handler"
	"github.com/baagee/dmq/common"
	"github.com/justinas/alice"
	"log"
	"net/http"
)

type App struct {
	alice alice.Chain
}

// 初始化
func (app *App) Init() *App {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	var m Middleware
	app.alice = alice.New(m.LogHandler, m.RecoverHandler)
	return app
}

// 绑定路由
func (app *App) BindRouter() *App {
	http.Handle("/api/message/single", app.alice.ThenFunc(handler.SingleMessage))
	http.Handle("/api/message/batch", app.alice.ThenFunc(handler.BatchMessage))
	http.Handle("/api/message/status", app.alice.ThenFunc(handler.MessageStatus))
	http.Handle("/api/message/detail", app.alice.ThenFunc(handler.MessageDetail))
	return app
}

// 运行
func (app *App) Run(port uint) {
	log.Printf("http://127.0.0.1:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		common.ExitWithNotice(common.ThrowNotice(common.ErrorCodeDefault, err))
	}
}
