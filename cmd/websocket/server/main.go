package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func main() {
	router := gin.Default()
	apiv1 := router.Group("/api/v1")
	{
		apiv1.Any("/ws", func(c *gin.Context) {
			conn, err := websocket.Accept(c.Writer, c.Request, nil)
			if err != nil {
				log.Println(err)
				return
			}
			defer conn.Close(websocket.StatusInternalError, "内部错误")

			ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*10)
			defer cancel()

			var v interface{}
			err = wsjson.Read(ctx, conn, &v)
			if err != nil {
				log.Println(err)
				return
			}
			log.Printf("接收到客户端:%v\n", v)

			err = wsjson.Write(ctx, conn, "Hello Websocket Client")
			if err != nil {
				log.Println(err)
				return
			}
			conn.Close(websocket.StatusNormalClosure, "")
		})
	}

	router.Run(":6430")

}
