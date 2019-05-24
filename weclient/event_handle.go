package weclient

import (
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/ikuiki/go-component/language"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/log"
	"strings"
)

func (m *WeClient) loginEvent(client MQTT.Client, msg MQTT.Message) {
	var loginStatus wwdk.LoginChannelItem
	e := json.Unmarshal(msg.Payload(), &loginStatus)
	if e != nil {
		log.Error("loginEvent: json.Unmarshal(msg.Payload(),&loginStatus) error: %v", e)
		return
	}
	if loginStatus.Code == wwdk.LoginStatusGotBatchContact {
		log.Info("检测到登陆成功")
		// 载入配置，检查配置是否适用（依据登陆用户是否为当前获取到的登陆用户的userName
		m.conf.Load()
		// 检查是否有读取到配置
		user, err := m.getUser()
		if err != nil {
			log.Error("loginEvent loadConf | getUser error: %+v", err)
			m.conf.Reset()
		} else if user.UserName != m.conf.UserName {
			// user可用，则检查当前是否为当前用户
			log.Info("loginEvent loadConf | conf is timeout, reset")
			m.conf.Reset()
			m.conf.UserName = user.UserName
		} else {
			// 读取到的配置可用，则不重置了
			log.Info("read reliable conf")
		}
		log.Info("检测到登陆成功开始获取星标联系人")
		m.starContacts = []string{}                         // 清空旧联系人
		m.chatroomMap = make(map[string]datastruct.Contact) // 重新整理联系人列表
		m.mcChatrooms = []string{}                          // 清理旧的mc聊天室
		if contacts, err := m.getContacts(); err != nil {
			log.Error("m.getContacts error: %+v", err)
		} else {
			for _, contact := range contacts {
				// 检查是否为星标联系人
				if contact.IsStar() {
					m.starContacts = append(m.starContacts, contact.UserName)
				}
				// 检查是否为群聊，如果为群聊则需要缓存了
				if contact.IsChatroom() {
					m.chatroomMap[contact.UserName] = contact
					// 检查是否在读取的mc聊天室中
					if language.ArrayIn(m.conf.McChatrooms, contact.UserName) != -1 {
						m.mcChatrooms = append(m.mcChatrooms, contact.UserName)
					}
				}
			}
		}
		m.starContacts = language.ArrayUnique(m.starContacts).([]string)
		log.Info("共找到%d位星标联系人与%d个聊天室", len(m.starContacts), len(m.chatroomMap))
		// 保存有效的mc聊天室
		m.conf.McChatrooms = m.mcChatrooms
		m.conf.Save()
		// 清理完成后通知星标联系人
		if len(m.mcChatrooms) == 0 {
			m.broadcastStaredContact("Minecraft聊天室列表已初始化")
		} else {
			// 如果有回复mc聊天室，则换个通知方式
			var nickNames []string
			for _, chatroomUserName := range m.mcChatrooms {
				chatroom := m.chatroomMap[chatroomUserName]
				nickNames = append(nickNames, chatroom.NickName)
			}
			m.broadcastStaredContact(fmt.Sprintf("Minecraft聊天室列表已恢复，当前mc聊天室数量为%d: \\n%s", len(m.mcChatrooms), strings.Join(nickNames, "\\n")))
		}
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
						// 保存设置
						m.conf.McChatrooms = mcChatrooms
						m.conf.Save()
						// 广播此消息
						m.sayToContact(message.FromUserName, "已设置为mc聊天室")
						chatroom := m.chatroomMap[message.FromUserName]
						m.broadcastStaredContact(fmt.Sprintf("已添加新的mc聊天室[%s]，当前mc聊天室数量为%d", chatroom.NickName, len(mcChatrooms)))
					}
				case "unset mc chatroom":
					mcChatrooms := language.ArrayDiff(m.mcChatrooms, []string{message.FromUserName}).([]string)
					if len(mcChatrooms) < len(m.mcChatrooms) {
						m.mcChatrooms = mcChatrooms
						// 保存设置
						m.conf.McChatrooms = mcChatrooms
						m.conf.Save()
						// 广播此消息
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
