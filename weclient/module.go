// Package weclient 连接wegate的客户端模块，处理wegate事宜
// 包括认证微信联系人中哪个是mc群，谁是管理员
// 以及处理管理员发来的命令，并与MCGate模块沟通
package weclient

import (
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/ikuiki/go-component/language"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"time"
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
	// 初始化与wegate的连接

	opts := m.mqttClient.GetDefaultOptions(common.ForceString(settings.Settings["HostURL"]))
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		log.Info("ConnectionLost: %s", err.Error())
	})
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		log.Info("OnConnectHandler")
	})
	opts.SetAutoReconnect(true)
	err := m.mqttClient.Connect(opts)
	if err != nil {
		panic(err)
	}
	// 登陆
	pass := common.ForceString(settings.Settings["Password"]) + time.Now().Format(time.RFC822)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	resp, _ := m.mqttClient.Request("Login/HD_Login", []byte(`{"username":"wecraftManager","password":"`+string(hashedPass)+`"}`))
	if resp.Ret != common.RetCodeOK {
		panic(fmt.Sprintf("登录失败: %s", resp.Msg))
	}
	// 注册监听新消息方法
	m.mqttClient.On("loginEvent", m.loginEvent)
	m.mqttClient.On("modifyContact", m.modifyContact)
	m.mqttClient.On("newMessageEvent", m.newMessageEvent)
	resp, _ = m.mqttClient.Request("Wechat/HD_Wechat_RegisterMQTTPlugin", []byte(fmt.Sprintf(
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
		panic(fmt.Sprintf("注册plugin失败: %s", resp.Msg))
	}
	m.wegateToken = resp.Msg
	log.Info("获取到token：%s\n", m.wegateToken)
	m.GetServer().RegisterGO("McSay", m.mcSay)
}

func (m *WeClient) loginEvent(client MQTT.Client, msg MQTT.Message) {
	var loginStatus wwdk.LoginChannelItem
	e := json.Unmarshal(msg.Payload(), &loginStatus)
	if e != nil {
		log.Error("loginEvent: json.Unmarshal(msg.Payload(),&loginStatus) error: %v", e)
		return
	}
	if loginStatus.Code == wwdk.LoginStatusGotBatchContact {
		log.Info("检测到登陆成功开始获取星标联系人")
		m.starContacts = []string{}                         // 清空旧联系人
		m.chatroomMap = make(map[string]datastruct.Contact) // 重新整理联系人列表
		if contacts, err := m.getContacts(); err != nil {
			log.Error("m.getContacts error: %+v", err)
		} else {
			for _, contact := range contacts {
				if contact.IsStar() {
					m.starContacts = append(m.starContacts, contact.UserName)
				}
				if contact.IsChatroom() {
					m.chatroomMap[contact.UserName] = contact
				}
			}
		}
		m.starContacts = language.ArrayUnique(m.starContacts).([]string)
		log.Info("共找到%d位星标联系人与%d个聊天室", len(m.starContacts), len(m.chatroomMap))
		// 清理旧的mc聊天室
		m.mcChatrooms = []string{}
		// 清理完成后通知星标联系人
		m.broadcastStaredContact("Minecraft聊天室列表已初始化")
	}
}

func (m *WeClient) modifyContact(client MQTT.Client, msg MQTT.Message) {
	var contact datastruct.Contact
	e := json.Unmarshal(msg.Payload(), &contact)
	if e != nil {
		log.Error("modifyContact: json.Unmarshal(msg.Payload(),&contact) error: %v", e)
		return
	}
	log.Info("modify contact: %s", contact.NickName)
	if contact.IsStar() {
		if language.ArrayIn(m.starContacts, contact.UserName) == -1 {
			// 找到新的星标联系人
			log.Info("发现新的星标联系人：%s", contact.NickName)
			m.starContacts = append(m.starContacts, contact.UserName)
		}
	} else {
		if language.ArrayIn(m.starContacts, contact.UserName) != -1 {
			// 发现已经移除的星标联系人
			log.Info("发现已经移除的星标联系人：%s", contact.NickName)
			olds := m.starContacts
			m.starContacts = []string{}
			for _, old := range olds {
				if old != contact.UserName {
					m.starContacts = append(m.starContacts, old)
				}
			}
		}
	}
	// 如果是群联系人，则记录
	if contact.IsChatroom() {
		m.chatroomMap[contact.UserName] = contact
	}
}

func (m *WeClient) newMessageEvent(client MQTT.Client, msg MQTT.Message) {
	var message datastruct.Message
	e := json.Unmarshal(msg.Payload(), &message)
	if e != nil {
		log.Error("newMessageEvent: json.Unmarshal(msg.Payload(),&message) error: %v", e)
		return
	}
	switch message.MsgType {
	case datastruct.TextMsg:
		// 目前只处理文字消息
		if message.IsChatroom() {
			content, _ := message.GetMemberMsgContent()
			memberUserName, _ := message.GetMemberUserName()
			if language.ArrayIn(m.starContacts, memberUserName) != -1 {
				log.Info("new chatroom starContact message: %s", content)
				switch content {
				case "set mc chatroom":
					mcChatrooms := language.ArrayUnique(append(m.mcChatrooms, message.FromUserName)).([]string)
					if len(mcChatrooms) > len(m.mcChatrooms) {
						m.mcChatrooms = mcChatrooms
						m.sayToContact(message.FromUserName, "已设置为mc聊天室")
						chatroom := m.chatroomMap[message.FromUserName]
						m.broadcastStaredContact(fmt.Sprintf("已添加新的mc聊天室[%s]，当前mc聊天室数量为%d", chatroom.NickName, len(mcChatrooms)))
					}
				case "unset mc chatroom":
					mcChatrooms := language.ArrayDiff(m.mcChatrooms, []string{message.FromUserName}).([]string)
					if len(mcChatrooms) < len(m.mcChatrooms) {
						m.mcChatrooms = mcChatrooms
						m.sayToContact(message.FromUserName, "已从mc聊天室中移除")
						chatroom := m.chatroomMap[message.FromUserName]
						m.broadcastStaredContact(fmt.Sprintf("已移除mc聊天室[%s]，当前mc聊天室数量为%d", chatroom.NickName, len(mcChatrooms)))
					}
				default:
					m.processMessage(message)
				}
			} else {
				log.Info("new chatroom message: %s", content)
				m.processMessage(message)
			}
		} else {
			if language.ArrayIn(m.starContacts, message.FromUserName) != -1 {
				log.Info("new starContact message: %s", message.Content)
			} else {
				log.Info("new message: %s", message.Content)
			}
		}
	}
}

// 处理消息，检查是否为mc聊天室的消息，如果是则发送给mc
func (m *WeClient) processMessage(message datastruct.Message) {
	if language.ArrayIn(m.mcChatrooms, message.FromUserName) != -1 {
		// 是mc聊天室
		if chatroom, ok := m.chatroomMap[message.FromUserName]; ok {
			// 找到这个聊天室联系人对象
			memberUserName, _ := message.GetMemberUserName()
			contact, _ := chatroom.GetMember(memberUserName)
			content, _ := message.GetMemberMsgContent()
			// 是mc聊天室，则将消息发送到群内
			text := contact.NickName + ": " + content
			log.Info("转发mc聊天室信息到Minefraft服务器: %s", text)
			m.sayToMC(text)
		}
	}
}

func (m *WeClient) getContacts() (contacts []datastruct.Contact, err error) {
	resp, _ := m.mqttClient.Request("Wechat/HD_Wechat_CallWechat", []byte(`{"fnName":"GetContactList","token":"`+m.wegateToken+`"}`))
	if resp.Ret != common.RetCodeOK {
		err = errors.Errorf("GetContactList失败: %s", resp.Msg)
		return
	}
	err = json.Unmarshal([]byte(resp.Msg), &contacts)
	return
}

// Run 运行主函数
func (m *WeClient) Run(closeSig chan bool) {
	log.Debug("weclient模块开始运行，")
	// 关闭信号
	<-closeSig
	m.broadcastStaredContact("wecraftmanager正在停止")
}
