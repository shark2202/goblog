package main

import (
	"github.com/astaxie/beego"
	_ "github.com/zd04/goblog/routers"
)

func main() {

	//beego.AddAPPStartHook(startWsServer)

	beego.Run()
}
