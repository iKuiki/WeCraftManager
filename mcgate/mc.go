package mcgate

import (
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"wegate/common"
)

func (mgt *MCGate) say(session gate.Session, msg map[string]interface{}) (result string, err string) {
	user := common.ForceString(msg["user"])
	content := common.ForceString(msg["content"])
	log.Info("WeCraft %s say: %s", user, content)
	session.Send("WeCraft/Say", []byte("got"))
	return
}
