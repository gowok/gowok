module github.com/gowok/gowok

go 1.24.6

require (
	github.com/golang-must/must v1.0.0
	github.com/google/uuid v1.6.0
	github.com/gowok/fp v0.2.1
	github.com/gowok/should v0.0.0-20240831060519-d8ab7c7891fb
	github.com/joho/godotenv v1.5.1
	github.com/ngamux/middleware v0.0.10
	github.com/ngamux/ngamux v1.7.50
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/spf13/cobra v1.10.1
	golang.org/x/crypto v0.40.0
	google.golang.org/grpc v1.76.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

// replace (
// 	github.com/ngamux/middleware => ../../ngamux/middleware
// 	github.com/ngamux/ngamux => ../../ngamux/ngamux
// )
