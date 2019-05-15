package main

import (
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/conf"
	"wecraftmanager/gate"
)

func main() {
	conf.LoadConfig("server.json")
	app := mqant.CreateApp(true, conf.Conf) // 只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
	app.Run(gate.Module())
}
