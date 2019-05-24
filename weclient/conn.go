package weclient

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/liangdas/mqant/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"time"
	"wegate/common"
)

// prepareConn 准备连接（包括连接的认证等
func (m *WeClient) prepareConn(hostURL, password string) error {
	opts := m.mqttClient.GetDefaultOptions(hostURL)
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		log.Info("ConnectionLost: %s", err.Error())
	})
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		log.Info("OnConnectHandler")
	})
	opts.SetAutoReconnect(true)
	err := errors.WithStack(m.mqttClient.Connect(opts))
	if err != nil {
		return err
	}
	// 登陆
	pass := password + time.Now().Format(time.RFC822)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}
	resp, _ := m.mqttClient.Request("Login/HD_Login", []byte(`{"username":"wecraftManager","password":"`+string(hashedPass)+`"}`))
	if resp.Ret != common.RetCodeOK {
		errors.Errorf("登录失败: %s", resp.Msg)
	}
	return nil
}

// registerConn 注册连接
// 将本模块注册为wegate的plugin
func (m *WeClient) registerConn(loginEvent, modifyContact, newMessageEvent func(client MQTT.Client, msg MQTT.Message)) (token string, err error) {
	// 注册监听新消息方法
	m.mqttClient.On("loginEvent", loginEvent)
	m.mqttClient.On("modifyContact", modifyContact)
	m.mqttClient.On("newMessageEvent", newMessageEvent)
	resp, _ := m.mqttClient.Request("Wechat/HD_Wechat_RegisterMQTTPlugin", []byte(fmt.Sprintf(
		`{"name":"%s","description":"%s","loginListenerTopic":"%s","contactListenerTopic":"%s","msgListenerTopic":"%s","addPluginListenerTopic":"%s","removePluginListenerTopic":"%s"}`,
		"WeCraftManager",  // name
		"WeCraft插件管理模块",   // description
		"loginEvent",      // loginListenerTopic
		"modifyContact",   // contactListenerTopic
		"newMessageEvent", // msgListenerTopic
		"",                // addPluginListenerTopic
		"",                // removePluginListenerTopic
	)))
	if resp.Ret != common.RetCodeOK {
		err = errors.Errorf("注册plugin失败: %s", resp.Msg)
		return
	}
	token = resp.Msg
	return
	// output:
	// token: aaaa-bbbb-cccc
	// err: nil
}
