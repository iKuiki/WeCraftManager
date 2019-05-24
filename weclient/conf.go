package weclient

import (
	"encoding/json"
	"github.com/ikuiki/storer"
	"github.com/pkg/errors"
)

// config WeClient的配置项
// 会存储当前登陆的user的userName用来防止使用过期的信息
// 储存已经设置好的mc聊天室，当重启时可以载入重用
type config struct {
	// storer 存储器
	storer storer.Storer
	// UserName 当前登陆用户的userName
	UserName string
	// McChatrooms 已设置为mc聊天室的群UserName
	McChatrooms []string
}

var (
	// ErrStorerNotExist 存储器未找到
	ErrStorerNotExist error = errors.New("storer not exist")
)

func (c *config) Load() error {
	if c.storer == nil {
		return ErrStorerNotExist
	}
	b, err := c.storer.Read()
	if err != nil {
		return errors.WithStack(err)
	}
	err = json.Unmarshal(b, c)
	return errors.WithStack(err)
}

func (c *config) Save() error {
	if c.storer == nil {
		return ErrStorerNotExist
	}
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.storer.Truncate()
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.storer.Write(b)
	return errors.WithStack(err)
}

func (c *config) Close() error {
	return c.storer.Close()
}

// 将config重设为初始
func (c *config) Reset() {
	c.UserName = ""
	c.McChatrooms = []string{}
}
