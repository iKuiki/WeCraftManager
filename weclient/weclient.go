package weclient

import (
	"fmt"
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

func (m *WeClient) broadcastMCChatroom(text string) {
	for _, chatroom := range m.mcChatrooms {
		err := m.sayToContact(chatroom, text)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func (m *WeClient) broadcastStaredContact(text string) {
	for _, contact := range m.starContacts {
		err := m.sayToContact(contact, text)
		if err != nil {
			log.Error(err.Error())
		}
	}
}
