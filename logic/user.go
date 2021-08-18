package logic

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type User struct {
	UID       int           `json:"uid"`
	NickName  string        `json:"nickname"`
	EnterAt   time.Time     `json:"enter_at"`
	Addr      string        `json:"addr"`
	MessageCh chan *Message `json:"-"`
	Token     string        `json:"tocken"`

	conn  *websocket.Conn
	isNew bool
}

var System = &User{}
var globalUID uint64 = 0

func NewUser(conn *websocket.Conn, token, nickname, addr string) *User {
	user := &User{
		NickName:  nickname,
		Addr:      addr,
		EnterAt:   time.Now(),
		MessageCh: make(chan *Message, 32),
		Token:     token,
		conn:      conn,
	}

	if user.Token != "" {
		uid, err := parseTokenAndValidate(token, nickname)
		if err == nil {
			user.UID = uid
		}
	}

	if user.UID == 0 {
		user.UID = int(atomic.AddUint64(&globalUID, 1))
		user.Token = genToken(user.UID, user.NickName)
		user.isNew = true
	}
	return user

}

func parseTokenAndValidate(token, nickname string) (int, error) {
	pos := strings.LastIndex(token, "uid")
	messageMAC, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0, err
	}
	uid := cast.ToInt(token[pos+3:])

	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)

	ok := validateMAC([]byte(message), messageMAC, []byte(secret))
	if ok {
		return uid, nil
	}
	return 0, errors.New("token is illegal")

}

func genToken(uid int, nickname string) string {
	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)
	messageMAC := macSha256([]byte(message), []byte(secret))
	return fmt.Sprintf("%suid%d", base64.StdEncoding.EncodeToString(messageMAC), uid)

}

func macSha256(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return mac.Sum(nil)

}

func validateMAC(message, messageMAC, secret []byte) bool {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)

}

func (u *User) SendMessage(ctx context.Context) {
	for msg := range u.MessageCh {
		wsjson.Write(ctx, u.conn, msg)
	}
}

func (u *User) CloseMessageChannel() {
	close(u.MessageCh)
}

func (u *User) ReceiveMessage(ctx context.Context) error {
	var (
		receiveMsg map[string]interface{}
		err        error
	)

	for {
		err = wsjson.Read(ctx, u.conn, &receiveMsg)
		if err != nil {
			// 判断链接是否关闭了 正常关闭 不认为是错误
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				return nil
			} else if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		// 将内容发送到聊天室
		sendMsg := NewMessage(u, receiveMsg["content"].(string), receiveMsg["bytecontent"].([]byte),
			receiveMsg["send_time"].(string))
		sendMsg.Content = FilterSensitive(sendMsg.Content)

		// 处理 @
		reg := regexp.MustCompile(`@[^\s@]{2,20}`)
		sendMsg.Ats = reg.FindAllString(sendMsg.Content, -1)

		Broadcaster.Broadcast(sendMsg)
	}
}
