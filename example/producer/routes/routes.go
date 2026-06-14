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
		rpcCfg := config.GetRpcConfig()
		myResult.Success(c, gin.H{
			"service":  "example-producer",
			"desc":     "gRPC 生产者，static 直连 consumer，持续互调",
			"grpcPort": rpcCfg.Server.Port,
			"registry": rpcCfg.Registry,
		})
	})

	v1 := apiGroup.Group("/v1")
	test := v1.Group("/test")
	{
		test.GET("/direct", func(c *gin.Context) {
			callDirect(c, starter)
		})
		test.GET("/error-demo", errorDemo)
		test.GET("/time", timeDemo)
	}
	v1.GET("/loop/status", func(c *gin.Context) {
		myResult.Success(c, callloop.GetStatus())
	})
}

func errorDemo(c *gin.Context) {
	code := c.DefaultQuery("code", "producer.demo.error")
	args := map[string]string{"field": c.DefaultQuery("field", "username")}
	if reason := c.Query("reason"); reason != "" {
		args["reason"] = reason
	}
	myResult.ErrorWithError(c, myException.NewBizError(code, args))
}

func timeDemo(c *gin.Context) {
	myResult.Success(c, demo.BuildTimeDemo())
}

func callDirect(c *gin.Context, starter *core.Starter) {
	message := c.DefaultQuery("message", "hello")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	conn, err := starter.GetRpcManager().Client().GetConn("example_producer")
	if err != nil {
		myResult.ErrorWithError(c, myException.NewBizError("producer.rpc.client_failed", map[string]string{
			"peer":   "example_producer",
			"reason": err.Error(),
		}))
		return
	}

	client := pb.NewEchoServiceClient(conn)
	reply, err := client.Echo(ctx, &pb.EchoRequest{Message: message})
	if err != nil {
		myResult.ErrorWithError(c, myException.NewBizError("producer.rpc.call_failed", map[string]string{
			"method": "Echo",
			"reason": err.Error(),
		}))
		return
	}

	myResult.Success(c, gin.H{
		"mode":    "static",
		"message": reply.GetMessage(),
	})
}
