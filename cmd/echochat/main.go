package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/echo/global"
	"github.com/impact-eintr/echo/service"
)

var (
	banner = `
    ____  ____         ____
   |    ||    ||    | |    |
   |____||     |____| |    |
   |     |     |    | |    |
   |____||____||    | |____|

一个基于websocket的小聊天室: start on %s

`
	addr = ":6430"
)

func init() {
	global.Init()
}

func main() {
	fmt.Printf(banner, addr)

	r := gin.Default()
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	apiv1 := r.Group("/api/v1")

	service.RegisterHandle(apiv1)

	go func() {
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// 优雅关机
	quit := make(chan os.Signal, 1) // 创建一个接受信号的信道
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞在此处

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}

	log.Println("server exit")

}
