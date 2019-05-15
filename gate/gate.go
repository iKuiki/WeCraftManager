package gate

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/gate/base"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
)

// Module 模块实例化
func Module() module.Module {
	gate := new(Gate)
	return gate
}

// Gate 网关
type Gate struct {
	basegate.Gate
}

// GetType 返回Type
func (gt *Gate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "Gate"
}

// Version 返回Version
func (gt *Gate) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

// OnInit 模块初始化
func (gt *Gate) OnInit(app module.App, settings *conf.ModuleSettings) {
	//注意这里一定要用 gate.Gate 而不是 module.BaseModule
	gt.Gate.OnInit(gt, app, settings)

	gt.Gate.SetSessionLearner(gt)
}

// Connect 当连接建立  并且MQTT协议握手成功
func (gt *Gate) Connect(session gate.Session) {
	log.Info("客户端建立了链接")
}

// DisConnect 当连接关闭	或者客户端主动发送MQTT DisConnect命令 ,这个函数中Session无法再继续后续的设置操作，只能读取部分配置内容了
func (gt *Gate) DisConnect(session gate.Session) {
	log.Info("客户端断开了链接")
}
