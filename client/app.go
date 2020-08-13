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
	//提交单个消息
	http.Handle("/api/message/single", app.alice.ThenFunc(handler.SingleMessage))
	//批量提交消息
	http.Handle("/api/message/batch", app.alice.ThenFunc(handler.BatchMessage))
	//查看消息消费状态
	http.Handle("/api/message/status", app.alice.ThenFunc(handler.MessageStatus))
	//查看消息详情
	http.Handle("/api/message/detail", app.alice.ThenFunc(handler.MessageDetail))
	//查看消息失败待处理ID列表
	http.Handle("/api/message/pending", app.alice.ThenFunc(handler.PendingMessageIdList))
	//设置消息消费成功已处理
	http.Handle("/api/message/solved", app.alice.ThenFunc(handler.MessageSolved))
	return app
}

// 运行
func (app *App) Run(port uint) {
	log.Printf("http://127.0.0.1:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		common.ExitWithNotice(common.ThrowNotice(common.ErrorCodeDefault, err))
	}
}
