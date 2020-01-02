package main

import "github.com/baagee/dmq/common"

func main() {
	var app App
	app.Init().BindRouter().Run(common.Config.HttpPort)
}
