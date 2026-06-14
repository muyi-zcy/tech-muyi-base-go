package main

import (
	"embed"
	"time"

	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/example/callloop"
	"github.com/muyi-zcy/tech-muyi-base-go/example/producer/routes"
	"github.com/muyi-zcy/tech-muyi-base-go/example/producer/server"
	"google.golang.org/grpc"
)

//go:embed contracts locales
var localeFS embed.FS

func main() {
	starter, err := core.Initialize()
	if err != nil {
		panic(err)
	}

	if err := starter.RegisterLocale(core.LocaleOptionsFromEmbed("producer", localeFS)); err != nil {
		panic(err)
	}

	routes.Register(starter.GetAPIGroup(), starter)

	starter.RegisterGrpcServices(func(s *grpc.Server) {
		server.RegisterEchoService(s)
	})

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
