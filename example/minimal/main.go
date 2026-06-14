package main

import (
	"embed"

	"github.com/muyi-zcy/tech-muyi-base-go/core"
	"github.com/muyi-zcy/tech-muyi-base-go/example/minimal/routes"
)

//go:embed contracts locales
var localeFS embed.FS

func main() {
	starter, err := core.Initialize()
	if err != nil {
		panic(err)
	}

	if err := starter.RegisterLocale(core.LocaleOptionsFromEmbed("example", localeFS)); err != nil {
		panic(err)
	}

	routes.Register(starter.GetAPIGroup())

	if err := starter.Run(); err != nil {
		panic(err)
	}
}
