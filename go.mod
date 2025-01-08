module github.com/gowok/gowok

go 1.22

toolchain go1.22.5

require (
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.23.0
	github.com/google/uuid v1.3.1
	github.com/gowok/should v0.0.0-20240831060519-d8ab7c7891fb
	github.com/ngamux/middleware v0.0.8
	github.com/ngamux/ngamux v1.7.44
	github.com/wagslane/go-rabbitmq v0.13.0
	golang.org/x/crypto v0.31.0
	google.golang.org/grpc v1.59.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/rabbitmq/amqp091-go v1.7.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

// replace (
// 	github.com/ngamux/middleware => ../../ngamux/middleware
// 	github.com/ngamux/ngamux => ../../ngamux/ngamux
// )
