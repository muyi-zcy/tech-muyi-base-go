package main

import (
	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/example/minimal/routes"
)

func main() {
	starter, err := core.Initialize()
	if err != nil {
		panic(err)
	}

	routes.Register(starter.GetEngine())

	if err := starter.Run(); err != nil {
		panic(err)
	}
}
