module github.com/xDarkicex/gsx-demo

go 1.25.7

replace github.com/xDarkicex/nanite => ../nanite

replace github.com/xDarkicex/nanite-gsx => ../nanite-gsx

replace github.com/xDarkicex/libravdb => ../libraVDB

replace github.com/xDarkicex/lexer => ../lexer

require (
	github.com/xDarkicex/libravdb v0.0.0-00010101000000-000000000000
	github.com/xDarkicex/nanite v0.5.7
	github.com/xDarkicex/nanite-gsx v0.0.0
	github.com/xDarkicex/nanite-render v0.1.0
	golang.org/x/crypto v0.54.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mmcloughlin/avo v0.6.0 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/tdewolff/minify v2.3.6+incompatible // indirect
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/xDarkicex/lexer v0.1.2 // indirect
	github.com/xDarkicex/memory v1.2.9 // indirect
	github.com/xDarkicex/nanite/sse v0.0.3 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/xDarkicex/nanite-render => ../nanite-render
