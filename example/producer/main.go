package main

import (
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/example/callloop"
	"github.com/muyi-zcy/tech-muyi-base-go/example/producer/routes"
	"github.com/muyi-zcy/tech-muyi-base-go/example/producer/server"
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

	// 持续通过 static 调用 consumer
	callloop.Start(starter, callloop.Config{
		Self:         "example-producer",
		PeerKey:      "example_consumer",
		Interval:     5 * time.Second,
		InitialDelay: 10 * time.Second,
	})

	if err := starter.Run(); err != nil {
		panic(err)
	}
}
