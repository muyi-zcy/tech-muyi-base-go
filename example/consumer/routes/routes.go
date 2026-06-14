package routes

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/example/callloop"
	"github.com/muyi-zcy/tech-muyi-base-go/example/demo"
	pb "github.com/muyi-zcy/tech-muyi-base-go/example/proto/echo"
	"github.com/muyi-zcy/tech-muyi-base-go/myException"
	"github.com/muyi-zcy/tech-muyi-base-go/myResult"
)

func Register(apiGroup *gin.RouterGroup, starter *core.Starter) {
	apiGroup.GET("/", func(c *gin.Context) {
		myResult.Success(c, gin.H{
			"service":  "example-consumer",
			"desc":     "Nacos 发现 producer，持续互调",
			"registry": config.GetRpcConfig().Registry,
		})
	})

	v1 := apiGroup.Group("/v1")
	test := v1.Group("/test")
	{
		test.GET("/error-demo", errorDemo)
		test.GET("/time", timeDemo)
	}
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

func errorDemo(c *gin.Context) {
	code := c.DefaultQuery("code", "consumer.demo.error")
	args := map[string]string{"field": c.DefaultQuery("field", "username")}
	if reason := c.Query("reason"); reason != "" {
		args["reason"] = reason
	}
	myResult.ErrorWithError(c, myException.NewBizError(code, args))
}

func timeDemo(c *gin.Context) {
	myResult.Success(c, demo.BuildTimeDemo())
}

func callProducerProxy(c *gin.Context, starter *core.Starter) {
	message := c.DefaultQuery("message", "hello via consumer proxy")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	conn, err := starter.GetRpcManager().Client().GetConn("example_producer")
	if err != nil {
		myResult.ErrorWithError(c, myException.NewBizError("consumer.rpc.discovery_failed", map[string]string{
			"peer":   "example_producer",
			"reason": err.Error(),
		}))
		return
	}

	reply, err := pb.NewEchoServiceClient(conn).Echo(ctx, &pb.EchoRequest{Message: message})
	if err != nil {
		myResult.ErrorWithError(c, myException.NewBizError("consumer.rpc.call_failed", map[string]string{
			"method": "Echo",
			"reason": err.Error(),
		}))
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
		myResult.ErrorWithError(c, myException.NewBizError("consumer.rpc.discovery_failed", map[string]string{
			"peer":   "example_producer",
			"reason": err.Error(),
		}))
		return
	}

	client := pb.NewEchoServiceClient(conn)

	switch method {
	case "ping":
		reply, err := client.Ping(ctx, &pb.PingRequest{})
		if err != nil {
			myResult.ErrorWithError(c, myException.NewBizError("consumer.rpc.call_failed", map[string]string{
				"method": "Ping",
				"reason": err.Error(),
			}))
			return
		}
		myResult.Success(c, gin.H{"mode": "nacos", "message": reply.GetMessage()})
	default:
		reply, err := client.Echo(ctx, &pb.EchoRequest{Message: message})
		if err != nil {
			myResult.ErrorWithError(c, myException.NewBizError("consumer.rpc.call_failed", map[string]string{
				"method": "Echo",
				"reason": err.Error(),
			}))
			return
		}
		myResult.Success(c, gin.H{"mode": "nacos", "message": reply.GetMessage()})
	}
}
