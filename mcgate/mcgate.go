package mcgate

// 此文件下主要存放供mc客户端调用的方法

import (
	"wegate/common"

	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
)

// hdSay 发送消息
// @Deprecated 这个方法是统一说话的方法，不推荐使用，取而代之的应该是根据事件调取对应事件的方法
// @Param user 发送消息的用户的昵称
// @Param content 发送的消息的内容
// @Return result none
// @Return err 错误消息，为空则无错误
func (mgt *MCGate) hdSay(session gate.Session, msg map[string]interface{}) (result string, err string) {
	user := common.ForceString(msg["user"])
	content := common.ForceString(msg["content"])
	log.Info("WeCraft %s say: %s", user, content)
	text := content
	if user != "" {
		text = user + ": " + text
	}
	if !session.IsGuest() {
		text = "[" + session.GetUserId() + "]" + text
	}
	mgt.mcSay(text)
	return
}

// hdRegister 注册
// @Param clientName 客户端名称，将显示在消息触发的地方
// @Return result none
// @Return err 错误消息，为空则无错误
func (mgt *MCGate) hdRegister(session gate.Session, msg map[string]interface{}) (result string, err string) {
	clientName := common.ForceString(msg["clientName"])
	if clientName == "" {
		err = "client name missing"
		return
	}
	mgt.mcSay("[" + clientName + "] plugin is online")
	session.Bind(clientName)
	return
}

// hdPlayerJoin 玩家加入游戏
// @Param playerName 玩家的名字
// @Return result none
// @Return err 错误消息，为空则无错误
func (mgt *MCGate) hdPlayerJoin(session gate.Session, msg map[string]interface{}) (result, err string) {
	if session.IsGuest() {
		err = "need login"
		return
	}
	playerName := common.ForceString(msg["playerName"])
	if playerName == "" {
		err = "player name missing"
		return
	}
	mgt.mcSay("[" + session.GetUserId() + "]" + playerName + "加入了游戏")
	return
}

// hdPlayerLeave 玩家离开游戏
// @Param playerName 玩家的名字
// @Return result none
// @Return err 错误消息，为空则无错误
func (mgt *MCGate) hdPlayerLeave(session gate.Session, msg map[string]interface{}) (result, err string) {
	if session.IsGuest() {
		err = "need login"
		return
	}
	playerName := common.ForceString(msg["playerName"])
	if playerName == "" {
		err = "player name missing"
		return
	}
	mgt.mcSay("[" + session.GetUserId() + "]" + playerName + "离开了游戏")
	return
}

// hdPlayerDeath 玩家死亡消息
// @Param playerName 玩家的名字
// @Param deathMessage 死亡消息
// @Return result none
// @Return err 错误消息，为空则无错误
func (mgt *MCGate) hdPlayerDeath(session gate.Session, msg map[string]interface{}) (result, err string) {
	if session.IsGuest() {
		err = "need login"
		return
	}
	playerName := common.ForceString(msg["playerName"])
	deathMessage := common.ForceString(msg["deathMessage"])
	if playerName == "" || deathMessage == "" {
		err = "player name or death message missing"
		return
	}
	mgt.mcSay("[" + session.GetUserId() + "]" + deathMessage)
	return
}

// hdPlayerChat 玩家聊天消息
// @Param playerName 玩家的名字
// @Param chatMessage 聊天消息
// @Return result none
// @Return err 错误消息，为空则无错误
func (mgt *MCGate) hdPlayerChat(session gate.Session, msg map[string]interface{}) (result, err string) {
	if session.IsGuest() {
		err = "need login"
		return
	}
	playerName := common.ForceString(msg["playerName"])
	chatMessage := common.ForceString(msg["chatMessage"])
	if playerName == "" || chatMessage == "" {
		err = "player name or chat message missing"
		return
	}
	mgt.mcSay("[" + session.GetUserId() + "]" + playerName + ": " + chatMessage)
	return
}

// hdPlayerAdvancementDone 玩家达成进度
// @Param playerName 玩家的名字
// @Param advancementKey 进度关键字
// @Return result none
// @Return err 错误消息，为空则无错误
func (mgt *MCGate) hdPlayerAdvancementDone(session gate.Session, msg map[string]interface{}) (result, err string) {
	if session.IsGuest() {
		err = "need login"
		return
	}
	playerName := common.ForceString(msg["playerName"])
	advancementKey := common.ForceString(msg["advancementKey"])
	if playerName == "" || advancementKey == "" {
		err = "player name or chat message missing"
		return
	}
	advancement, ok := mgt.advancementMap[advancementKey]
	if ok {
		mgt.mcSay("[" + session.GetUserId() + "]" + playerName + "达成了进度[" + advancement.Advancement + "]\\n" + advancement.InGameDescription)
	} else {
		mgt.mcSay("[" + session.GetUserId() + "]" + playerName + "达成了进度[" + advancementKey + "]")
	}
	return
}
