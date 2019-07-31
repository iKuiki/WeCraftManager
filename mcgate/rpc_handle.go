package mcgate

// 此文件下主要存放供别的模块rpc调用mc功能的方法

// 广播到mc
func (mgt *MCGate) broadcastToMC(text string) (result, err string) {
	for _, session := range mgt.sessionMap {
		session.Send("WeCraft/Say", []byte(text))
	}
	return
}
