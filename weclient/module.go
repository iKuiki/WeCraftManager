// Package weclient 连接wegate的客户端模块，处理wegate事宜
// 包括认证微信联系人中哪个是mc群，谁是管理员
// 以及处理管理员发来的命令，并与MCGate模块沟通
package weclient

import (
	"github.com/ikuiki/storer"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"wegate/common"
	"wegate/common/test"
)

// Module 模块实例化
func Module() module.Module {
	m := new(WeClient)
	m.chatroomMap = make(map[string]datastruct.Contact)
	return m
}

// WeClient weclient模块
type WeClient struct {
	basemodule.BaseModule
	wegateToken  string          // 连接wegate注册后的token
	mqttClient   commontest.Work // 连接wegate的mqtt客户端
	conf         config
	starContacts []string
	chatroomMap  map[string]datastruct.Contact // 群列表（需要维护
	mcChatrooms  []string                      // 属于mc的群
}

// GetType 获取模块类型
func (m *WeClient) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "WeClient"
}

// Version 获取模块Version
func (m *WeClient) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}

// OnInit 模块初始化
func (m *WeClient) OnInit(app module.App, settings *conf.ModuleSettings) {
	m.BaseModule.OnInit(m, app, settings)
	filepath := common.ForceString(settings.Settings["SavePath"])
	if filepath != "" {
		m.conf.storer = storer.MustNewFileStorer(filepath)
	}
	// 初始化与wegate的连接
	err := m.prepareConn(
		common.ForceString(settings.Settings["HostURL"]),
		common.ForceString(settings.Settings["Password"]),
	)
	if err != nil {
		panic(err)
	}
	token, err := m.registerConn(m.loginEvent, m.modifyContact, m.newMessageEvent)
	if err != nil {
		panic(err)
	}
	m.wegateToken = token
	log.Info("获取到token：%s\n", m.wegateToken)
	m.GetServer().RegisterGO("McSay", m.mcSay)
}

// Run 运行主函数
func (m *WeClient) Run(closeSig chan bool) {
	log.Debug("weclient模块开始运行，")
	// 关闭信号
	<-closeSig
	m.broadcastStaredContact("wecraftmanager正在停止")
	// 关闭conf的storer
	m.conf.Close()
}
