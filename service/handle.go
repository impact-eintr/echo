package service

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/echo/logic"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func RegisterHandle(rg *gin.RouterGroup) {
	// 广播
	go logic.Broadcaster.Start()

	rg.Any("/ws", func(c *gin.Context) {
		conn, err := websocket.Accept(c.Writer, c.Request,
			&websocket.AcceptOptions{})
		if err != nil {
			log.Println("websocket accept error:", err)
			return
		}
		defer conn.Close(websocket.StatusInternalError, "内部错误")

		// 新用户进入
		token := c.Request.FormValue("token")
		nickname := c.Request.FormValue("nickname")

		if l := len(nickname); l < 2 || l > 20 {
			log.Println("illegal nickname: ", nickname)
			wsjson.Write(c.Request.Context(), conn,
				logic.NewErrorMessage("非法昵称 昵称长度："))
			conn.Close(websocket.StatusUnsupportedData, "illegal nickname")
			return
		}
		if !logic.Broadcaster.CanEnterRoom(nickname) {
			log.Println("昵称已经存在", nickname)
			wsjson.Write(c.Request.Context(), conn, logic.NewErrorMessage("该昵称已存在"))
			conn.Close(websocket.StatusUnsupportedData, "nickname exsts!")
			return
		}

		userHasToken := logic.NewUser(conn, token, nickname, c.Request.RemoteAddr)

		// 开启给用户发送消息的goroutine
		go userHasToken.SendMessage(c.Request.Context())

		// 给当前用户发送欢迎消息
		userHasToken.MessageCh <- logic.NewWelcomeMessage(userHasToken)
		// 避免 token 泄露
		tmpUser := *userHasToken
		user := &tmpUser
		user.Token = ""

		// 通知所有用户新用户的到来
		msg := logic.NewUserEnterMessage(user)
		logic.Broadcaster.Broadcast(msg)
		// 将该用户加入广播器的用户列表中
		logic.Broadcaster.UserEntering(user)
		log.Println("user:", nickname, "joins chat")
		// 接收用户消息
		err = user.ReceiveMessage(c.Request.Context())
		// 用户离开
		logic.Broadcaster.UserLeaving(user)
		msg = logic.NewUserLeaveMessage(user)
		logic.Broadcaster.Broadcast(msg)
		log.Println("user:", nickname, "leaves chat")

		// 根据读取时的错误执行不同的 close
		if err == nil {
			conn.Close(websocket.StatusNormalClosure, "")
		} else {
			log.Println("read from client error:", err)
			conn.Close(websocket.StatusInternalError, "Read from client error")
		}
	})
}
