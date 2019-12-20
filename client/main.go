package main

import "github.com/baagee/dmq/common"

func main() {
	//// 设置log 记录时间 文件和行号
	//log.SetFlags(log.LstdFlags | log.Llongfile)
	//// 设置中间件
	//m := Middleware{}
	//mm := alice.New(m.LogHandler, m.RecoverHandler, m.CheckParamsHandler, m.CheckProductHandler)
	//// 设置路由
	//http.Handle("/api/message/single", mm.ThenFunc(handler.Single))
	//http.Handle("/api/message/batch", mm.ThenFunc(handler.Batch))
	//
	//log.Printf("http://127.0.0.1:%d\n", common.Config.HttpPort)
	//http.ListenAndServe(fmt.Sprintf(":%d", common.Config.HttpPort), nil)
	var app App
	app.Init().BindRouter().Run(common.Config.HttpPort)
}
