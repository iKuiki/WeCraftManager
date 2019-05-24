package weclient

import (
	"encoding/json"
	"fmt"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/log"
	"github.com/pkg/errors"
	"wegate/common"
)

// mcSay minecraft中发的文字
func (m *WeClient) mcSay(text string) (result, err string) {
	m.broadcastMCChatroom(text)
	return
}

// 向mc中发消息
func (m *WeClient) sayToMC(text string) error {
	m.RpcInvokeNR("MCGate", "BroadcastToMC", text)
	return nil
}

func (m *WeClient) sayToContact(contactUserName, text string) error {
	resp, _ := m.mqttClient.Request("Wechat/HD_Wechat_CallWechat", []byte(fmt.Sprintf(
		`{"fnName":"%s","token":"%s","toUserName":"%s","content":"%s"}`,
		"SendTextMessage", // fnName
		m.wegateToken,     // token
		contactUserName,   // toUserName
		text,              // content
	)))
	if resp.Ret != common.RetCodeOK {
		return errors.Errorf("SendTextMessage to wegate error[%d]: %s", resp.Ret, resp.Msg)
	}
	return nil
}

// 获取当前登陆用户
func (m *WeClient) getUser() (user datastruct.User, err error) {
	resp, _ := m.mqttClient.Request("Wechat/HD_Wechat_CallWechat", []byte(`{"fnName":"GetUser","token":"`+m.wegateToken+`"}`))
	if resp.Ret != common.RetCodeOK {
		err = errors.Errorf("GetUser失败: %s", resp.Msg)
		return
	}
	err = errors.WithStack(json.Unmarshal([]byte(resp.Msg), &user))
	return
}

// 获取联系人
func (m *WeClient) getContacts() (contacts []datastruct.Contact, err error) {
	resp, _ := m.mqttClient.Request("Wechat/HD_Wechat_CallWechat", []byte(`{"fnName":"GetContactList","token":"`+m.wegateToken+`"}`))
	if resp.Ret != common.RetCodeOK {
		err = errors.Errorf("GetContactList失败: %s", resp.Msg)
		return
	}
	err = errors.WithStack(json.Unmarshal([]byte(resp.Msg), &contacts))
	return
}

// 对登记在册的mc聊天室发送广播消息
func (m *WeClient) broadcastMCChatroom(text string) {
	for _, chatroom := range m.mcChatrooms {
		err := m.sayToContact(chatroom, text)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// 对登记在册的星标联系人发送广播消息
func (m *WeClient) broadcastStaredContact(text string) {
	for _, contact := range m.starContacts {
		err := m.sayToContact(contact, text)
		if err != nil {
			log.Error(err.Error())
		}
	}
}
