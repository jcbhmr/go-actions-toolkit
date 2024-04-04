module github.com/jcbhmr/actions-toolkit.go/tool-cache

go 1.22.1

require (
	github.com/google/uuid v1.6.0
	github.com/jcbhmr/actions-toolkit.go/core v0.0.0-20240404022858-fc4be8a8d3f4
)

replace github.com/jcbhmr/actions-toolkit.go/core => ../core
replace github.com/jcbhmr/actions-toolkit.go/http-client => ../http-client