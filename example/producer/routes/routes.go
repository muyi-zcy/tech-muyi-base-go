package routes

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/example/callloop"
	pb "github.com/muyi-zcy/tech-muyi-base-go/example/proto/echo"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
)

func Register(engine *gin.Engine, starter *core.Starter) {
	engine.GET("/", func(c *gin.Context) {
		rpcCfg := config.GetRpcConfig()
		myResult.Success(c, gin.H{
			"service":  "example-producer",
			"desc":     "gRPC 生产者，static 直连 consumer，持续互调",
			"grpcPort": rpcCfg.Server.Port,
			"registry": rpcCfg.Registry,
		})
	})

	v1 := engine.Group("/api/v1")
	v1.GET("/test/direct", func(c *gin.Context) {
		callDirect(c, starter)
	})
	v1.GET("/loop/status", func(c *gin.Context) {
		myResult.Success(c, callloop.GetStatus())
	})
}

func callDirect(c *gin.Context, starter *core.Starter) {
	message := c.DefaultQuery("message", "hello")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	conn, err := starter.GetRpcManager().Client().GetConn("example_producer")
	if err != nil {
		myResult.Error(c, "gRPC 直连失败: "+err.Error())
		return
	}

	client := pb.NewEchoServiceClient(conn)
	reply, err := client.Echo(ctx, &pb.EchoRequest{Message: message})
	if err != nil {
		myResult.Error(c, "gRPC 调用失败: "+err.Error())
		return
	}

	myResult.Success(c, gin.H{
		"mode":    "static",
		"message": reply.GetMessage(),
	})
}
