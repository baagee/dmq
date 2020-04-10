package main

import "github.com/baagee/dmq/common"

func main() {
	var app App
	err := common.LoadLuaScript()
	if err != nil {
		common.ExitWithNotice(common.ThrowNotice(common.ErrorCodeRedisLoadLuaFailed, err))
	}
	app.Init().BindRouter().Run(common.Config.HttpPort)
}
