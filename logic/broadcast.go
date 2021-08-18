package logic

import (
	"expvar"
	"fmt"
	"log"

	"github.com/impact-eintr/echo/global"
)

func init() {
	expvar.Publish("message", expvar.Func(calcMessageQueueLen))
}

func calcMessageQueueLen() interface{} {
	fmt.Println("===len=== : ", len(Broadcaster.messageCh))
	return len(Broadcaster.messageCh)
}

type broadcaster struct {
	users map[string]*User // 聊天室的所有用户

	// 所有 channel 统一管理
	enteringCh chan *User
	leavingCh  chan *User
	messageCh  chan *Message

	// 判断该昵称的用户是否可以进入聊天室
	checkUserCh      chan string
	checkUserCanInCh chan bool

	// 获取用户列表
	requestUsersCh chan struct{}
	usersCh        chan []*User
}

type GroupLevel uint8
type GroupUserLevel uint32

const (
	SMALL_GROUP GroupLevel = (iota + 1)
	MID_GROUP
	BIG_GROUP
	SMALL = 10
	MID   = 100
	BIG   = 1000
)

// 群组广播器
type groupBroadcaster struct {
	groupLevel     GroupUserLevel
	groupUserLevel GroupUserLevel
	users          map[string]*User // 聊天室的所有用户

	// 所有 channel 统一管理
	enteringCh chan *User
	leavingCh  chan *User
	messageCh  chan *Message
}

var Broadcaster = &broadcaster{
	users:            make(map[string]*User, 1<<10),
	enteringCh:       make(chan *User),
	leavingCh:        make(chan *User),
	messageCh:        make(chan *Message, global.MessageQueueLen),
	checkUserCh:      make(chan string),
	checkUserCanInCh: make(chan bool),
	requestUsersCh:   make(chan struct{}),
	usersCh:          make(chan []*User),
}

// 启动广播器
func (b *broadcaster) Start() {
	for {
		select {
		case user := <-b.enteringCh:
			b.users[user.NickName] = user
			OfflineProcessor.Send(user)
		case user := <-b.leavingCh:
			delete(b.users, user.NickName)
			user.CloseMessageChannel()
		case msg := <-b.messageCh:
			for _, user := range b.users {
				if user.UID == msg.User.UID {
					continue
				}
				go func(message *Message) {
					user.MessageCh <- msg
				}(msg)
			}
			OfflineProcessor.Save(msg)
		case nickname := <-b.checkUserCh:
			if _, ok := b.users[nickname]; ok { // 当前用户已经在线
				b.checkUserCanInCh <- false
			} else {
				b.checkUserCanInCh <- true
			}
		case <-b.requestUsersCh:
			userList := make([]*User, 0, len(b.users))
			for _, user := range b.users {
				userList = append(userList, user)
			}
			b.usersCh <- userList
		}
	}
}

func (b *broadcaster) UserEntering(u *User) {
	b.enteringCh <- u
}

func (b *broadcaster) UserLeaving(u *User) {
	b.leavingCh <- u
}

func (b *broadcaster) Broadcast(msg *Message) {
	if len(b.messageCh) >= global.MessageQueueLen {
		log.Println("broadcast queue 已满")
	}
	b.messageCh <- msg
}

func (b *broadcaster) CanEnterRoom(nickname string) bool {
	b.checkUserCh <- nickname
	return <-b.checkUserCanInCh
}
