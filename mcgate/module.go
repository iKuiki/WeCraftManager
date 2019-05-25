// Package mcgate mc网关模块，负责处理WeCraft插件的连接
package mcgate

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	basegate "github.com/liangdas/mqant/gate/base"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
)

// Module 模块实例化
func Module() module.Module {
	mgt := new(MCGate)
	mgt.sessionMap = make(map[string]gate.Session)
	return mgt
}

// MCGate 网关
type MCGate struct {
	basegate.Gate
	sessionMap map[string]gate.Session // 已经连接的session map, sessionID => session
}

// GetType 返回Type
func (mgt *MCGate) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "MCGate"
}

// Version 返回Version
func (mgt *MCGate) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

// OnInit 模块初始化
func (mgt *MCGate) OnInit(app module.App, settings *conf.ModuleSettings) {
	//注意这里一定要用 gate.Gate 而不是 module.BaseModule
	mgt.Gate.OnInit(mgt, app, settings)
	mgt.Gate.SetSessionLearner(mgt)
	mgt.GetServer().RegisterGO("HD_Say", mgt.hdSay)
	mgt.GetServer().RegisterGO("HD_PlayerJoin", mgt.hdPlayerJoin)
	mgt.GetServer().RegisterGO("HD_PlayerLeave", mgt.hdPlayerLeave)
	mgt.GetServer().RegisterGO("HD_PlayerDeath", mgt.hdPlayerDeath)
	mgt.GetServer().RegisterGO("HD_PlayerChat", mgt.hdPlayerChat)
	mgt.GetServer().RegisterGO("BroadcastToMC", mgt.broadcastToMC)
}

// Connect 当连接建立  并且MQTT协议握手成功
func (mgt *MCGate) Connect(session gate.Session) {
	log.Info("客户端建立了链接")
	mgt.sessionMap[session.GetSessionId()] = session
}

// DisConnect 当连接关闭	或者客户端主动发送MQTT DisConnect命令 ,这个函数中Session无法再继续后续的设置操作，只能读取部分配置内容了
func (mgt *MCGate) DisConnect(session gate.Session) {
	log.Info("客户端断开了链接")
	delete(mgt.sessionMap, session.GetSessionId())
}
