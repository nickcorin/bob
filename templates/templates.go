package templates

import _ "embed"

//go:embed docker-compose.tmpl
var Compose string

//go:embed dockerfile.tmpl
var Dockerfile string

//go:embed makefile.tmpl
var Makefile string
