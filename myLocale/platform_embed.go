package myLocale

import "embed"

//go:embed platform
var platformFS embed.FS

const platformRoot = "platform"
