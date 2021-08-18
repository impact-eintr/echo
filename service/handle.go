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
		//tocken := c.Request.FormValue("token")
		nickname := c.Request.FormValue("nickname")

		if l := len(nickname); l < 2 || l > 20 {
			log.Println("illegal nickname: ", nickname)
			wsjson.Write(c.Request.Context(), conn, nil)
			conn.Close(websocket.StatusUnsupportedData, "illegal nickname")
			return
		}
		// 开启给用户发送消息的goroutine
		// 给当前用户发送欢迎消息
		// 通知所有用户新用户的到来
		// 将该用户加入广播器的用户列表中
		// 接收用户消息
		// 用户离开
	})
}
