package mcgate

import (
	"github.com/pkg/errors"
)

// 本文件主要储存对weclient的调用的代理方法

// mcBroadcast 向微信中发送来自mc的消息
// @Param text 消息内容
// @Return err 错误，为空则无错误
func (mgt *MCGate) mcBroadcast(text string) error {
	_, err := mgt.RpcInvoke("WeClient", "McBroadcast", text)
	if err != "" {
		return errors.New("RpcInvoke WeClient.McBroadcast error: " + err)
	}
	return nil
}

// mcSay 向微信中发送来自mc的消息
// @Param text 消息内容
// @Return err 错误，为空则无错误
func (mgt *MCGate) mcSay(text string) error {
	_, err := mgt.RpcInvoke("WeClient", "McSay", text)
	if err != "" {
		return errors.New("RpcInvoke WeClient.McSay error: " + err)
	}
	return nil
}
