package mcgate

import (
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"wegate/common"
)

// mc调用传来消息
func (mgt *MCGate) hdSay(session gate.Session, msg map[string]interface{}) (result string, err string) {
	user := common.ForceString(msg["user"])
	content := common.ForceString(msg["content"])
	log.Info("WeCraft %s say: %s", user, content)
	text := content
	if user != "" {
		text = user + ": " + text
	}
	mgt.RpcInvokeNR("WeClient", "McSay", text)
	return
}

// 广播到mc
func (mgt *MCGate) broadcastToMC(text string) (result, err string) {
	for _, session := range mgt.sessionMap {
		session.Send("WeCraft/Say", []byte(text))
	}
	return
}
