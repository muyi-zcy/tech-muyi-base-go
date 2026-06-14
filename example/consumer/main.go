package main

import (
	"embed"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/example/callloop"
	"github.com/muyi-zcy/tech-muyi-base-go/example/consumer/routes"
	"github.com/muyi-zcy/tech-muyi-base-go/example/consumer/server"
	"google.golang.org/grpc"
)

//go:embed contracts locales
var localeFS embed.FS

func main() {
	starter, err := core.Initialize()
	if err != nil {
		panic(err)
	}

	if err := starter.RegisterLocale(core.LocaleOptionsFromEmbed("consumer", localeFS)); err != nil {
		panic(err)
	}

	routes.Register(starter.GetAPIGroup(), starter)

	starter.RegisterGrpcServices(func(s *grpc.Server) {
		server.RegisterEchoService(s)
	})

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
