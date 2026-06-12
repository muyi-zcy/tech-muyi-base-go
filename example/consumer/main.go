package main

import (
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/example/callloop"
	"github.com/muyi-zcy/tech-muyi-base-go/example/consumer/routes"
	"github.com/muyi-zcy/tech-muyi-base-go/example/consumer/server"
	"google.golang.org/grpc"
)

func main() {
	starter, err := core.Initialize()
	if err != nil {
		panic(err)
	}

	routes.Register(starter.GetEngine(), starter)

	starter.RegisterGrpcServices(func(s *grpc.Server) {
		server.RegisterEchoService(s)
	})

	// 持续通过 Nacos 调用 producer
	callloop.Start(starter, callloop.Config{
		Self:         "example-consumer",
		PeerKey:      "example_producer",
		Interval:     5 * time.Second,
		InitialDelay: 8 * time.Second,
	})

	if err := starter.Run(); err != nil {
		panic(err)
	}
}
