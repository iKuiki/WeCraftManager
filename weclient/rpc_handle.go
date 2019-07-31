package weclient

// mcBroadcast minecraft中发的文字
func (m *WeClient) mcBroadcast(text string) (result, err string) {
	m.broadcastMCChatroom(text)
	return
}

// mcSay minecraft中向星标联系人发送消息
func (m *WeClient) mcSay(text string) (result, err string) {
	m.broadcastStaredContact(text)
	return
}
