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
		myResult.Success(c, gin.H{
			"service":  "example-consumer",
			"desc":     "Nacos 发现 producer，持续互调",
			"registry": config.GetRpcConfig().Registry,
		})
	})

	v1 := engine.Group("/api/v1")
	call := v1.Group("/call")
	{
		call.GET("/producer", func(c *gin.Context) {
			callProducerProxy(c, starter)
		})
		call.GET("/ping", func(c *gin.Context) {
			callProducer(c, starter, "ping", "")
		})
		call.GET("/echo", func(c *gin.Context) {
			message := c.DefaultQuery("message", "hello from consumer")
			callProducer(c, starter, "echo", message)
		})
	}
	v1.GET("/loop/status", func(c *gin.Context) {
		myResult.Success(c, callloop.GetStatus())
	})
}

func callProducerProxy(c *gin.Context, starter *core.Starter) {
	message := c.DefaultQuery("message", "hello via consumer proxy")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	conn, err := starter.GetRpcManager().Client().GetConn("example_producer")
	if err != nil {
		myResult.Error(c, "consumer 调用 producer 失败（Nacos 发现）: "+err.Error())
		return
	}

	reply, err := pb.NewEchoServiceClient(conn).Echo(ctx, &pb.EchoRequest{Message: message})
	if err != nil {
		myResult.Error(c, "consumer 调用 producer gRPC 失败: "+err.Error())
		return
	}

	myResult.Success(c, gin.H{
		"chain":           "http -> consumer -> producer (nacos)",
		"mode":            "nacos",
		"requestMessage":  message,
		"producerReply":   reply.GetMessage(),
		"consumerService": config.GetConfig().AppName,
		"producerService": "example.producer",
	})
}

func callProducer(c *gin.Context, starter *core.Starter, method, message string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	conn, err := starter.GetRpcManager().Client().GetConn("example_producer")
	if err != nil {
		myResult.Error(c, "Nacos 发现/连接 producer 失败: "+err.Error())
		return
	}

	client := pb.NewEchoServiceClient(conn)

	switch method {
	case "ping":
		reply, err := client.Ping(ctx, &pb.PingRequest{})
		if err != nil {
			myResult.Error(c, "gRPC Ping 失败: "+err.Error())
			return
		}
		myResult.Success(c, gin.H{"mode": "nacos", "message": reply.GetMessage()})
	default:
		reply, err := client.Echo(ctx, &pb.EchoRequest{Message: message})
		if err != nil {
			myResult.Error(c, "gRPC Echo 失败: "+err.Error())
			return
		}
		myResult.Success(c, gin.H{"mode": "nacos", "message": reply.GetMessage()})
	}
}
