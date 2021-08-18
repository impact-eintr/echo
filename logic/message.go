package logic

import (
	"time"

	"github.com/spf13/cast"
)

const (
	MsgTypeNormal = iota
	MsgTypeWelcome
	MsgTypeUserEnter
	MsgTypeUserLeave
	MsgTypeError
)

// 给用户发送的消息
type Message struct {
	// 给哪个用户发送的消息
	User        *User
	Type        int
	Content     string
	ByteContent []byte
	MsgTime     time.Time

	ClientSendTime time.Time

	// 消息@了谁
	Ats []string `json:"ats"`
}

func NewMessage(user *User, content string, bcontent []byte, clientTime string) *Message {
	message := &Message{
		User:        user,
		Type:        MsgTypeWelcome,
		Content:     content,
		ByteContent: bcontent,
		MsgTime:     time.Now(),
	}
	if clientTime != "" {
		message.ClientSendTime = time.Unix(0, cast.ToInt64(clientTime))
	}
	return message

}

func NewWelcomeMessage(user *User) *Message {
	return &Message{
		User:    user,
		Type:    MsgTypeWelcome,
		Content: user.NickName + " 欢迎加入echo!",
		MsgTime: time.Now(),
	}
}

func NewUserEnterMessage(user *User) *Message {
	return &Message{
		User:    user,
		Type:    MsgTypeUserEnter,
		Content: user.NickName + " 加入echo!",
		MsgTime: time.Now(),
	}
}

func NewUserLeaveMessage(user *User) *Message {
	return &Message{
		User:    user,
		Type:    MsgTypeUserEnter,
		Content: user.NickName + " 离开echo!",
		MsgTime: time.Now(),
	}
}

func NewErrorMessage(content string) *Message {
	return &Message{
		User:    System,
		Type:    MsgTypeError,
		Content: content,
		MsgTime: time.Now(),
	}
}
